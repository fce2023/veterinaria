package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/auth"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
	"veterinaria/backend/services"
)

// LoginRequest defines login payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Username and password are required"})
		return
	}

	var user models.User
	if err := config.DB.Preload("Roles").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid username or password"})
		return
	}

	if user.Estado != "active" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "User account is suspended"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.CompanyID, user.BranchID, user.RoleType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not generate session token"})
		return
	}

	// Optional: Audit log
	go logAudit(user.ID, "Auth", "Login", "User logged in successfully", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":         user.ID,
				"nombre":     user.Nombre,
				"email":      user.Email,
				"username":   user.Username,
				"company_id": user.CompanyID,
				"branch_id":  user.BranchID,
				"role_type":  user.RoleType,
				"roles":      user.Roles,
			},
		},
	})
}

// GetMe returns current logged in user details
func GetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var user models.User
	if err := config.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	var modules []string
	config.DB.Model(&models.CompanyModule{}).
		Where("company_id = ? AND is_active = true", user.CompanyID).
		Pluck("module_key", &modules)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         user.ID,
			"nombre":     user.Nombre,
			"email":      user.Email,
			"username":   user.Username,
			"company_id": user.CompanyID,
			"branch_id":  user.BranchID,
			"role_type":  user.RoleType,
			"roles":      user.Roles,
			"modules":    modules,
		},
	})
}

// CreateCompany handles onboarding of new businesses (Companies)
func CreateCompany(c *gin.Context) {
	var input struct {
		RUC             string `json:"ruc" binding:"required"`
		RazonSocial     string `json:"razon_social" binding:"required"`
		NombreComercial string `json:"nombre_comercial"`
		Direccion       string `json:"direccion"`
		Telefono        string `json:"telefono"`
		Email           string `json:"email"`
		Sector          string `json:"sector"`
		Ubigeo          string `json:"ubigeo"`
		AdminName       string `json:"admin_name" binding:"required"`
		AdminEmail      string `json:"admin_email" binding:"required"`
		AdminUsername   string `json:"admin_username" binding:"required"`
		AdminPassword   string `json:"admin_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// 1. Create Company with Trial Plan config
	company := models.Company{
		RUC:             input.RUC,
		RazonSocial:     input.RazonSocial,
		NombreComercial: input.NombreComercial,
		Direccion:       input.Direccion,
		Telefono:        input.Telefono,
		Email:           input.Email,
		Estado:          "active",
	}
	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create company: RUC might already exist"})
		return
	}

	// Create Default Trial Subscription
	var basicPlan models.Plan
	if err := tx.Where("nombre = ?", "Básico").First(&basicPlan).Error; err != nil {
		basicPlan = models.Plan{
			Nombre:      "Básico",
			Precio:      49.00,
			MaxBranches: 1,
			MaxUsers:    5,
			Modulos:     "core",
		}
		if err := tx.Create(&basicPlan).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create default plan"})
			return
		}
	}

	sub := models.Subscription{
		CompanyID: company.ID,
		PlanID:    basicPlan.ID,
		Estado:    "TRIAL",
		StartsAt:  time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	}
	if err := tx.Create(&sub).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create subscription"})
		return
	}

	// Activate Core module
	coreMod := models.CompanyModule{
		CompanyID: company.ID,
		ModuleKey: "core",
		IsActive:  true,
	}
	if err := tx.Create(&coreMod).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to activate core module"})
		return
	}

	// Activate Sector module (e.g. "veterinaria", "vidrieria") if provided and valid
	if input.Sector != "" && input.Sector != "core" {
		sectMod := models.CompanyModule{
			CompanyID: company.ID,
			ModuleKey: input.Sector,
			IsActive:  true,
		}
		if err := tx.Create(&sectMod).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to activate sector module"})
			return
		}
	}

	// 2. Create Default Branch (Sucursal Principal)
	branch := models.Branch{
		CompanyID: company.ID,
		Nombre:    "Sucursal Principal",
		Direccion: input.Direccion,
		Telefono:  input.Telefono,
		Estado:    "active",
		IsMain:    true,
		Ubigeo:    input.Ubigeo,
	}
	if err := tx.Create(&branch).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create default branch"})
		return
	}

	// 3. Create Admin User
	pwdHash, err := auth.HashPassword(input.AdminPassword)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to hash password"})
		return
	}

	user := models.User{
		CompanyID:    company.ID,
		BranchID:     branch.ID,
		Nombre:       input.AdminName,
		Email:        input.AdminEmail,
		Username:     input.AdminUsername,
		PasswordHash: pwdHash,
		Estado:       "active",
		RoleType:     "COMPANY_ADMIN", // Owner of the veterinary
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create admin user: username or email already exists"})
		return
	}

	// 4. Associate Admin Role (if Admin role exists)
	var adminRole models.Role
	if err := tx.Where("nombre = ?", "Administrador").First(&adminRole).Error; err == nil {
		tx.Model(&user).Association("Roles").Append(&adminRole)
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"company": company,
			"branch":  branch,
			"user_id": user.ID,
		},
	})
}

// GetCompanies lists companies (SuperAdmin only, or simple list)
func GetCompanies(c *gin.Context) {
	var companies []models.Company
	if err := config.DB.Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": companies})
}

// GetMyCompany retrieves the profile of the current tenant company
func GetMyCompany(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var company models.Company
	if err := config.DB.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": company})
}

// UpdateMyCompany updates the profile of the current tenant company (Razón Social, Dirección, etc)
func UpdateMyCompany(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var input struct {
		RazonSocial     string `json:"razon_social" binding:"required"`
		NombreComercial string `json:"nombre_comercial"`
		Direccion       string `json:"direccion"`
		Telefono        string `json:"telefono"`
		Email           string `json:"email"`
		LogoBase64      string `json:"logo_base64"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var company models.Company
	if err := config.DB.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Company not found"})
		return
	}

	company.RazonSocial = input.RazonSocial
	company.NombreComercial = input.NombreComercial
	company.Direccion = input.Direccion
	company.Telefono = input.Telefono
	company.Email = input.Email
	if input.LogoBase64 != "" {
		// Basic optimization check: 200KB base64 limit (~150KB raw)
		// 400x400 logos should be well under this.
		if len(input.LogoBase64) > 200*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El logo es demasiado grande. Por favor, usa una imagen optimizada (máx 150KB)."})
			return
		}
		company.LogoBase64 = input.LogoBase64
	}

	if err := config.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Try to sync logo with FacturaAPI if billing is active
	if input.LogoBase64 != "" {
		billingService := services.NewBillingService()
		_ = billingService.SyncLogo(company.ID, company.LogoBase64) // Fail silently here as it's secondary
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Datos de negocio actualizados correctamente", "data": company})
}


// --- BRANCH HANDLERS (Tenant isolated) ---

func GetBranches(c *gin.Context) {
	var branches []models.Branch
	if err := config.DB.Scopes(config.TenantFilter(c)).Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": branches})
}

func CreateBranch(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	// Verify Technical Limits (MaxBranches from Active/Trial Subscription Plan)
	var sub models.Subscription
	var maxBranches int = 1
	err := config.DB.Preload("Plan").Where("company_id = ? AND estado IN ('ACTIVE', 'TRIAL')", companyID).First(&sub).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to check active subscription"})
		return
	}
	maxBranches = sub.Plan.MaxBranches

	var currentBranchesCount int64
	config.DB.Model(&models.Branch{}).Where("company_id = ?", companyID).Count(&currentBranchesCount)
	if currentBranchesCount >= int64(maxBranches) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Límite de sucursales alcanzado. Por favor, actualiza tu suscripción para agregar más sedes.",
		})
		return
	}

	var branch models.Branch
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	branch.CompanyID = companyID
	if err := config.DB.Create(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create branch"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": branch})
}

func UpdateBranch(c *gin.Context) {
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid branch ID"})
		return
	}

	var branch models.Branch
	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", branchID).First(&branch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Branch not found"})
		return
	}

	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&branch)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": branch})
}

func DeleteBranch(c *gin.Context) {
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid branch ID"})
		return
	}

	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", branchID).Delete(&models.Branch{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete branch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Branch deleted successfully"})
}

// --- PRODUCT HANDLERS (Tenant isolated) ---

func GetProducts(c *gin.Context) {
	var products []models.Product
	if err := config.DB.Scopes(config.TenantFilter(c)).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": products})
}

func CreateProduct(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	product.CompanyID = companyID
	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": product})
}

func UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", productID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&product)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": product})
}

func DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid product ID"})
		return
	}

	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product deleted successfully"})
}

func GetNextProductCode(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var count int64
	if err := config.DB.Model(&models.Product{}).Where("company_id = ?", companyID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	nextNumber := count + 1
	nextCode := fmt.Sprintf("PROD-%04d", nextNumber)

	c.JSON(http.StatusOK, gin.H{"success": true, "next_code": nextCode})
}

func ValidateProductUniqueness(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	codigo := c.Query("codigo")
	codigoBarras := c.Query("codigo_barras")
	excludeIDStr := c.Query("exclude_id")

	var query = config.DB.Model(&models.Product{}).Where("company_id = ?", companyID)
	if excludeIDStr != "" {
		if exID, err := uuid.Parse(excludeIDStr); err == nil {
			query = query.Where("id != ?", exID)
		}
	}

	var match models.Product
	var err error

	if codigo != "" {
		err = query.Where("codigo = ?", codigo).First(&match).Error
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "valid": false, "field": "codigo", "message": "El código ya se encuentra registrado."})
			return
		}
	}

	if codigoBarras != "" {
		err = query.Where("codigo_barras = ?", codigoBarras).First(&match).Error
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "valid": false, "field": "codigo_barras", "message": "El código de barras ya se encuentra registrado."})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "valid": true})
}

// --- UTILITIES ---

func logAudit(userID uuid.UUID, modulo, accion, descripcion, ip string) {
	log := models.AuditLog{
		UserID:      userID,
		Modulo:      modulo,
		Accion:      accion,
		Descripcion: descripcion,
		IP:          ip,
		Fecha:       time.Now(),
	}
	config.DB.Create(&log)
}

// QueryRUC proxies requests to apiconsulta.sehuacho.com to get RUC details securely
func QueryRUC(c *gin.Context) {
	ruc := c.Param("ruc")
	if len(ruc) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El RUC debe tener exactamente 11 dígitos"})
		return
	}

	req, err := http.NewRequest("GET", "https://apiconsulta.sehuacho.com/api/ruc/"+ruc, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear la petición"})
		return
	}
	req.Header.Add("x-api-key", "tc_a72ee220220a196d5eee19f9f1ba4e00")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "Error al conectar con el servicio de consulta"})
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al leer la respuesta del servicio"})
		return
	}

	if res.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.Unmarshal(body, &errorResponse)
		errMsg := "No se pudo obtener información del RUC"
		if val, ok := errorResponse["error"]; ok {
			errMsg = fmt.Sprintf("%v", val)
		}
		c.JSON(res.StatusCode, gin.H{"success": false, "error": errMsg})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al parsear datos recibidos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// QueryDNI proxies requests to apiconsulta.sehuacho.com to get DNI details securely
func QueryDNI(c *gin.Context) {
	dni := c.Param("dni")
	if len(dni) != 8 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El DNI debe tener exactamente 8 dígitos"})
		return
	}

	req, err := http.NewRequest("GET", "https://apiconsulta.sehuacho.com/api/dni/"+dni, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear la petición"})
		return
	}
	req.Header.Add("x-api-key", "tc_a72ee220220a196d5eee19f9f1ba4e00")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "Error al conectar con el servicio de consulta"})
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al leer la respuesta del servicio"})
		return
	}

	if res.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.Unmarshal(body, &errorResponse)
		errMsg := "No se pudo obtener información del DNI"
		if val, ok := errorResponse["error"]; ok {
			errMsg = fmt.Sprintf("%v", val)
		}
		c.JSON(res.StatusCode, gin.H{"success": false, "error": errMsg})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al parsear datos recibidos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}



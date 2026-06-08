package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/auth"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
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

	token, err := auth.GenerateToken(user.ID, user.CompanyID, user.BranchID)
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
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

	// 1. Create Company
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

	// 2. Create Default Branch (Sucursal Principal)
	branch := models.Branch{
		CompanyID: company.ID,
		Nombre:    "Sucursal Principal",
		Direccion: input.Direccion,
		Telefono:  input.Telefono,
		Estado:    "active",
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

// --- BRANCH HANDLERS (Tenant isolated) ---

func GetBranches(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var branches []models.Branch
	if err := config.DB.Where("company_id = ?", companyID).Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": branches})
}

func CreateBranch(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

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
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid branch ID"})
		return
	}

	var branch models.Branch
	if err := config.DB.Where("id = ? AND company_id = ?", branchID, companyID).First(&branch).Error; err != nil {
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
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid branch ID"})
		return
	}

	if err := config.DB.Where("id = ? AND company_id = ?", branchID, companyID).Delete(&models.Branch{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete branch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Branch deleted successfully"})
}

// --- PRODUCT HANDLERS (Tenant isolated) ---

func GetProducts(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var products []models.Product
	if err := config.DB.Where("company_id = ?", companyID).Find(&products).Error; err != nil {
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
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := config.DB.Where("id = ? AND company_id = ?", productID, companyID).First(&product).Error; err != nil {
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
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid product ID"})
		return
	}

	if err := config.DB.Where("id = ? AND company_id = ?", productID, companyID).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product deleted successfully"})
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

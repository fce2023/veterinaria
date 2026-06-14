package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

// --- HELPERS ---

func logSaasAction(userID, companyID uuid.UUID, action, description, ip string) {
	log := models.SaasAuditLog{
		UserID:      userID,
		CompanyID:   companyID,
		Accion:      action,
		Descripcion: description,
		IP:          ip,
	}
	config.DB.Create(&log)
}

// --- PLANS MANAGEMENT (SuperAdmin) ---

type PlanInput struct {
	Nombre      string  `json:"nombre" binding:"required"`
	Precio      float64 `json:"precio" binding:"required,gt=0"`
	MaxBranches int     `json:"max_branches" binding:"required,gt=0"`
	MaxUsers    int     `json:"max_users" binding:"required,gt=0"`
	Modulos     string  `json:"modulos"`
}

func CreatePlan(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input PlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	plan := models.Plan{
		Nombre:      input.Nombre,
		Precio:      input.Precio,
		MaxBranches: input.MaxBranches,
		MaxUsers:    input.MaxUsers,
		Modulos:     input.Modulos,
	}

	if err := config.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	logSaasAction(userID, uuid.Nil, "CREATE_PLAN", fmt.Sprintf("Creado plan '%s' con precio S/ %.2f", plan.Nombre, plan.Precio), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": plan})
}

func GetPlans(c *gin.Context) {
	var plans []models.Plan
	if err := config.DB.Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": plans})
}

// --- MODULE TOGGLING (SuperAdmin) ---

type ToggleModuleInput struct {
	CompanyID uuid.UUID `json:"company_id" binding:"required"`
	ModuleKey string    `json:"module_key" binding:"required"`
	IsActive  bool      `json:"is_active"`
}

func ToggleCompanyModule(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input ToggleModuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var module models.CompanyModule
	err := config.DB.Where("company_id = ? AND module_key = ?", input.CompanyID, input.ModuleKey).First(&module).Error

	now := time.Now()

	if err != nil {
		module = models.CompanyModule{
			CompanyID: input.CompanyID,
			ModuleKey: input.ModuleKey,
			IsActive:  input.IsActive,
		}
		if input.IsActive {
			module.EnabledAt = &now
			module.EnabledByUserID = &userID
		} else {
			module.DisabledAt = &now
			module.DisabledByUserID = &userID
		}
		config.DB.Create(&module)
	} else {
		if module.IsActive != input.IsActive {
			module.IsActive = input.IsActive
			if input.IsActive {
				module.EnabledAt = &now
				module.EnabledByUserID = &userID
				module.DisabledAt = nil
				module.DisabledByUserID = nil
			} else {
				module.DisabledAt = &now
				module.DisabledByUserID = &userID
			}
		}
		config.DB.Save(&module)
	}

	actionStr := "DISABLE_MODULE"
	if input.IsActive {
		actionStr = "ENABLE_MODULE"
	}
	logSaasAction(userID, input.CompanyID, actionStr, fmt.Sprintf("Módulo '%s' cambiado a activo=%v", input.ModuleKey, input.IsActive), c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"success": true, "data": module})
}

// --- PAYMENTS & SUBSCRIPTIONS (SuperAdmin / Tenants) ---

type CreatePaymentInput struct {
	PlanID         uuid.UUID `json:"plan_id" binding:"required"`
	Monto          float64   `json:"monto" binding:"required,gt=0"`
	MetodoPago     string    `json:"metodo_pago" binding:"required"` // YAPE, PLIN, TRANSFERENCIA
	Referencia     string    `json:"referencia" binding:"required"`
	ComprobanteUrl string    `json:"comprobante_url"`
}

// CreatePaymentRequest logs a renewal/payment request from a client tenant
func CreatePaymentRequest(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var input CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Resolve or create subscription
	var sub models.Subscription
	err := config.DB.Where("company_id = ?", companyID).First(&sub).Error
	if err != nil {
		sub = models.Subscription{
			CompanyID: companyID,
			PlanID:    input.PlanID,
			Estado:    "TRIAL",
			StartsAt:  time.Now(),
			ExpiresAt: time.Now().AddDate(0, 0, 30),
		}
		config.DB.Create(&sub)
	}

	payment := models.Payment{
		CompanyID:      companyID,
		SubscriptionID: sub.ID,
		Monto:          input.Monto,
		MetodoPago:     input.MetodoPago,
		Referencia:     input.Referencia,
		Estado:         "PENDING",
		ComprobanteUrl: input.ComprobanteUrl,
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": payment})
}

// ApprovePayment approves a payment log, activating/renewing the subscription (SuperAdmin)
func ApprovePayment(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	idStr := c.Param("id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid payment ID"})
		return
	}

	var payment models.Payment
	if err := config.DB.First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Payment not found"})
		return
	}

	if payment.Estado != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Payment is not in PENDING state"})
		return
	}

	tx := config.DB.Begin()

	now := time.Now()
	payment.Estado = "APPROVED"
	payment.ApprovedAt = &now
	payment.ApprovedBy = &userID
	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to approve payment"})
		return
	}

	var sub models.Subscription
	if err := tx.First(&sub, payment.SubscriptionID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Subscription not found"})
		return
	}

	// Fetch Plan to check module presets
	var plan models.Plan
	tx.First(&plan, sub.PlanID)

	sub.Estado = "ACTIVE"
	// Renew for 30 days
	if sub.ExpiresAt.After(time.Now()) {
		sub.ExpiresAt = sub.ExpiresAt.AddDate(0, 0, 30)
	} else {
		sub.StartsAt = time.Now()
		sub.ExpiresAt = time.Now().AddDate(0, 0, 30)
	}

	if err := tx.Save(&sub).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update subscription"})
		return
	}

	tx.Commit()

	logSaasAction(userID, payment.CompanyID, "APPROVE_PAYMENT", fmt.Sprintf("Pago aprobado por S/ %.2f. Suscripción extendida.", payment.Monto), c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"success": true, "data": payment})
}

// --- COMPANY SETTINGS ---

type SettingInput struct {
	Clave string `json:"clave" binding:"required"`
	Valor string `json:"valor" binding:"required"`
}

func SaveCompanySetting(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var input SettingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var setting models.CompanySetting
	err := config.DB.Where("company_id = ? AND clave = ?", companyID, input.Clave).First(&setting).Error

	if err != nil {
		setting = models.CompanySetting{
			CompanyID: companyID,
			Clave:     input.Clave,
			Valor:     input.Valor,
		}
		config.DB.Create(&setting)
	} else {
		setting.Valor = input.Valor
		config.DB.Save(&setting)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": setting})
}

func GetCompanySettings(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	var settings []models.CompanySetting
	if err := config.DB.Where("company_id = ?", companyID).Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

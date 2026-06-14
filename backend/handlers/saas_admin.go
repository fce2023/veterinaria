package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

type UpdateSubscriptionInput struct {
	PlanType              string    `json:"plan_type" binding:"required"`
	SubscriptionExpiresAt time.Time `json:"subscription_expires_at" binding:"required"`
	MaxBranches           int       `json:"max_branches" binding:"required,gt=0"`
}

// GetSaaSCompanies lists all tenants (SuperAdmin only)
func GetSaaSCompanies(c *gin.Context) {
	var companies []models.Company
	if err := config.DB.Order("created_at desc").Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var result []gin.H
	for _, comp := range companies {
		var branchCount int64
		config.DB.Model(&models.Branch{}).Where("company_id = ?", comp.ID).Count(&branchCount)

		var userCount int64
		config.DB.Model(&models.User{}).Where("company_id = ?", comp.ID).Count(&userCount)

		// Fetch active subscription and plan details
		var sub models.Subscription
		planName := "Trial/Básico"
		var expiresAt time.Time
		maxBranches := 1

		err := config.DB.Preload("Plan").Where("company_id = ?", comp.ID).First(&sub).Error
		if err == nil {
			planName = sub.Plan.Nombre
			expiresAt = sub.ExpiresAt
			maxBranches = sub.Plan.MaxBranches
		}

		var activeModules []string
		config.DB.Model(&models.CompanyModule{}).Where("company_id = ? AND is_active = ?", comp.ID, true).Pluck("module_key", &activeModules)

		var mainBranch models.Branch
		var ubigeo string
		if err := config.DB.Where("company_id = ? AND is_main = true", comp.ID).First(&mainBranch).Error; err == nil {
			ubigeo = mainBranch.Ubigeo
		}

		result = append(result, gin.H{
			"id":                      comp.ID,
			"ruc":                     comp.RUC,
			"razon_social":            comp.RazonSocial,
			"nombre_comercial":        comp.NombreComercial,
			"direccion":               comp.Direccion,
			"telefono":                comp.Telefono,
			"email":                   comp.Email,
			"estado":                  comp.Estado,
			"plan_type":               planName,
			"subscription_expires_at": expiresAt,
			"max_branches":            maxBranches,
			"branches_count":          branchCount,
			"users_count":             userCount,
			"modules":                 activeModules,
			"ubigeo":                  ubigeo,
			"created_at":              comp.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// UpdateSaaSCompanySubscription updates the limits and subscription of a tenant
func UpdateSaaSCompanySubscription(c *gin.Context) {
	idStr := c.Param("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid company ID"})
		return
	}

	var input UpdateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var company models.Company
	if err := config.DB.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Company not found"})
		return
	}

	// Find or Create Plan matching the PlanType name
	var plan models.Plan
	err = config.DB.Where("nombre = ?", input.PlanType).First(&plan).Error
	if err != nil {
		plan = models.Plan{
			Nombre:      input.PlanType,
			Precio:      99.00,
			MaxBranches: input.MaxBranches,
			MaxUsers:    10,
			Modulos:     "core",
		}
		if err := config.DB.Create(&plan).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create/resolve Plan"})
			return
		}
	} else {
		// Update max branches on plan if modified
		if plan.MaxBranches != input.MaxBranches {
			plan.MaxBranches = input.MaxBranches
			config.DB.Save(&plan)
		}
	}

	// Update or create subscription
	var sub models.Subscription
	err = config.DB.Where("company_id = ?", company.ID).First(&sub).Error
	if err != nil {
		sub = models.Subscription{
			CompanyID: company.ID,
			PlanID:    plan.ID,
			Estado:    "ACTIVE",
			StartsAt:  time.Now(),
			ExpiresAt: input.SubscriptionExpiresAt,
		}
		config.DB.Create(&sub)
	} else {
		sub.PlanID = plan.ID
		sub.Estado = "ACTIVE"
		sub.ExpiresAt = input.SubscriptionExpiresAt
		config.DB.Save(&sub)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": sub})
}

// GetSaaSStats yields global stats across all companies
func GetSaaSStats(c *gin.Context) {
	var totalCompanies int64
	config.DB.Model(&models.Company{}).Count(&totalCompanies)

	var activeCompanies int64
	config.DB.Model(&models.Subscription{}).Where("estado IN ('ACTIVE', 'TRIAL') AND expires_at > ?", time.Now()).Count(&activeCompanies)

	var totalBranches int64
	config.DB.Model(&models.Branch{}).Count(&totalBranches)

	var totalSales float64
	config.DB.Model(&models.Sale{}).Where("estado = 'completed'").Select("COALESCE(SUM(total), 0)").Scan(&totalSales)

	var totalSalesCount int64
	config.DB.Model(&models.Sale{}).Where("estado = 'completed'").Count(&totalSalesCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_companies":   totalCompanies,
			"active_companies":  activeCompanies,
			"total_branches":    totalBranches,
			"total_sales_value": totalSales,
			"total_sales_count": totalSalesCount,
		},
	})
}

// DeleteSaaSCompany deletes a tenant company and its branches/users
func DeleteSaaSCompany(c *gin.Context) {
	idStr := c.Param("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid company ID"})
		return
	}

	tx := config.DB.Begin()

	// Soft delete associated users first
	if err := tx.Where("company_id = ?", companyID).Delete(&models.User{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete company users"})
		return
	}

	// Soft delete associated branches
	if err := tx.Where("company_id = ?", companyID).Delete(&models.Branch{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete company branches"})
		return
	}

	// Soft delete company itself
	if err := tx.Delete(&models.Company{}, companyID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete company"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Empresa y datos asociados eliminados correctamente"})
}

// UpdateSaaSCompany updates basic details of a company
func UpdateSaaSCompany(c *gin.Context) {
	idStr := c.Param("id")
	companyID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid company ID"})
		return
	}

	var input struct {
		RazonSocial     string `json:"razon_social" binding:"required"`
		NombreComercial string `json:"nombre_comercial"`
		Direccion       string `json:"direccion"`
		Telefono        string `json:"telefono"`
		Email           string `json:"email"`
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

	if err := config.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Empresa actualizada correctamente", "data": company})
}

// GetSaaSGlobalSettings retrieves global settings (URL, API Key) for FacturaAPI
func GetSaaSGlobalSettings(c *gin.Context) {
	var settings []models.CompanySetting
	if err := config.DB.Where("company_id = ?", uuid.Nil).Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	res := gin.H{
		"factura_api_url": "",
		"factura_api_key": "",
	}
	for _, s := range settings {
		if s.Clave == "factura_api_url" {
			res["factura_api_url"] = s.Valor
		} else if s.Clave == "factura_api_key" {
			res["factura_api_key"] = s.Valor
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// SaveSaaSGlobalSettings stores global settings (URL, API Key) for FacturaAPI
func SaveSaaSGlobalSettings(c *gin.Context) {
	var input struct {
		FacturaApiURL string `json:"factura_api_url"`
		FacturaApiKey string `json:"factura_api_key"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	keys := map[string]string{
		"factura_api_url": input.FacturaApiURL,
		"factura_api_key": input.FacturaApiKey,
	}

	for k, v := range keys {
		var setting models.CompanySetting
		err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, k).First(&setting).Error
		if err != nil {
			setting = models.CompanySetting{
				CompanyID: uuid.Nil,
				Clave:     k,
				Valor:     v,
			}
			config.DB.Create(&setting)
		} else {
			setting.Valor = v
			config.DB.Save(&setting)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Configuración global guardada correctamente"})
}



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

		result = append(result, gin.H{
			"id":                      comp.ID,
			"ruc":                     comp.RUC,
			"razon_social":            comp.RazonSocial,
			"nombre_comercial":        comp.NombreComercial,
			"email":                   comp.Email,
			"estado":                  comp.Estado,
			"plan_type":               comp.PlanType,
			"subscription_expires_at": comp.SubscriptionExpiresAt,
			"max_branches":            comp.MaxBranches,
			"branches_count":          branchCount,
			"users_count":             userCount,
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

	// Update fields
	company.PlanType = input.PlanType
	company.SubscriptionExpiresAt = input.SubscriptionExpiresAt
	company.MaxBranches = input.MaxBranches

	if err := config.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": company})
}

// GetSaaSStats yields global stats across all companies
func GetSaaSStats(c *gin.Context) {
	var totalCompanies int64
	config.DB.Model(&models.Company{}).Count(&totalCompanies)

	var activeCompanies int64
	config.DB.Model(&models.Company{}).Where("subscription_expires_at > ?", time.Now()).Count(&activeCompanies)

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

package handlers

import (
	"net/http"
	"veterinaria/backend/config"
	"veterinaria/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetBillingConfig retrieves the billing configuration for the company
func GetBillingConfig(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Configuración no encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    billingConfig,
	})
}

// SaveBillingConfig creates or updates the billing configuration
func SaveBillingConfig(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	var input struct {
		ApiURL            string `json:"api_url"`
		ApiKey            string `json:"api_key"`
		TenantUUID        string `json:"tenant_uuid"`
		Modo              string `json:"modo"`
		Estado            string `json:"estado"`
		SolUser           string `json:"sol_user"`
		SolPass           string `json:"sol_pass"`
		CertificadoBase64 string `json:"certificado_base64"`
		ClientID          string `json:"client_id"`
		ClientSecret      string `json:"client_secret"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var billingConfig models.BillingConfig
	err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error

	billingConfig.CompanyID = companyID
	billingConfig.ApiURL = input.ApiURL
	billingConfig.ApiKey = input.ApiKey
	billingConfig.TenantUUID = input.TenantUUID
	billingConfig.Modo = input.Modo
	billingConfig.Estado = input.Estado
	billingConfig.SolUser = input.SolUser
	billingConfig.SolPass = input.SolPass
	billingConfig.ClientID = input.ClientID
	billingConfig.ClientSecret = input.ClientSecret
	if input.CertificadoBase64 != "" {
		billingConfig.CertificadoBase64 = input.CertificadoBase64
	}

	if err != nil {
		// Create new
		if err := config.DB.Create(&billingConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear configuración"})
			return
		}
	} else {
		// Update existing
		if err := config.DB.Save(&billingConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar configuración"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuración guardada exitosamente",
		"data":    billingConfig,
	})
}

package handlers

import (
	"net/http"
	"veterinaria/backend/config"
	"veterinaria/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetPets returns all pets for the company
func GetPets(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	var pets []models.Pet
	
	query := config.DB.Where("company_id = ?", companyID)
	
	customerID := c.Query("customer_id")
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	if err := query.Find(&pets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pets})
}

// CreatePet creates a new pet
func CreatePet(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	var pet models.Pet

	if err := c.ShouldBindJSON(&pet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	pet.CompanyID = companyID.(uuid.UUID)

	if err := config.DB.Create(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pet})
}

// UpdatePet updates an existing pet
func UpdatePet(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	id := c.Param("id")

	var pet models.Pet
	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).First(&pet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Mascota no encontrada"})
		return
	}

	if err := c.ShouldBindJSON(&pet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := config.DB.Save(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pet})
}

// DeletePet deletes a pet
func DeletePet(c *gin.Context) {
	companyID, _ := c.Get("companyID")
	id := c.Param("id")

	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Pet{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mascota eliminada"})
}

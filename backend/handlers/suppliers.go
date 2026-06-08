package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func GetSuppliers(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var suppliers []models.Supplier
	if err := config.DB.Where("company_id = ?", companyID).Find(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": suppliers})
}

func CreateSupplier(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var supplier models.Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	supplier.CompanyID = companyID
	if err := config.DB.Create(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create supplier"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": supplier})
}

func UpdateSupplier(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid supplier ID"})
		return
	}

	var supplier models.Supplier
	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).First(&supplier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Supplier not found"})
		return
	}

	if err := c.ShouldBindJSON(&supplier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&supplier)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": supplier})
}

func DeleteSupplier(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid supplier ID"})
		return
	}

	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Supplier{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete supplier"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Supplier deleted successfully"})
}

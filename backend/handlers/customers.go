package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func GetCustomers(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var customers []models.Customer
	if err := config.DB.Where("company_id = ?", companyID).Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": customers})
}

func CreateCustomer(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	customer.CompanyID = companyID
	if err := config.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create customer"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": customer})
}

func UpdateCustomer(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Customer not found"})
		return
	}

	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&customer)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": customer})
}

func DeleteCustomer(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid customer ID"})
		return
	}

	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.Customer{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete customer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Customer deleted successfully"})
}

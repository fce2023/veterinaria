package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

// Categories
func GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := config.DB.Scopes(config.TenantFilter(c)).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": categories})
}

func CreateCategory(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	category.CompanyID = companyID
	if err := config.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": category})
}

func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid category ID"})
		return
	}

	var category models.Category
	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", id).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Category not found"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&category)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": category})
}

func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid category ID"})
		return
	}

	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", id).Delete(&models.Category{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete category"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category deleted successfully"})
}

// Brands
func GetBrands(c *gin.Context) {
	var brands []models.Brand
	if err := config.DB.Scopes(config.TenantFilter(c)).Find(&brands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": brands})
}

func CreateBrand(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	var brand models.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	brand.CompanyID = companyID
	if err := config.DB.Create(&brand).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create brand"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": brand})
}

func UpdateBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", id).First(&brand).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Brand not found"})
		return
	}

	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.DB.Save(&brand)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": brand})
}

func DeleteBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid brand ID"})
		return
	}

	if err := config.DB.Scopes(config.TenantFilter(c)).Where("id = ?", id).Delete(&models.Brand{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete brand"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand deleted successfully"})
}

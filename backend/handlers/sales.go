package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

type SaleItemInput struct {
	ProductID      uuid.UUID `json:"product_id" binding:"required"`
	Cantidad       float64   `json:"cantidad" binding:"required,gt=0"`
	PrecioUnitario float64   `json:"precio_unitario" binding:"required"`
	Descuento      float64   `json:"descuento"`
}

type CreateSaleInput struct {
	CustomerID uuid.UUID       `json:"customer_id" binding:"required"`
	Items      []SaleItemInput `json:"items" binding:"required,gt=0"`
}

func GetSales(c *gin.Context) {
	var sales []models.Sale
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Customer").Order("created_at desc").Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": sales})
}

func GetSaleDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid sale ID"})
		return
	}

	var sale models.Sale
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Customer").Where("id = ?", id).First(&sale).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Sale not found"})
		return
	}

	var items []models.SaleItem
	if err := config.DB.Where("sale_id = ?", sale.ID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var responseItems []gin.H
	for _, item := range items {
		var product models.Product
		config.DB.Select("nombre, codigo").First(&product, item.ProductID)
		responseItems = append(responseItems, gin.H{
			"id":              item.ID,
			"product_id":      item.ProductID,
			"product_nombre":  product.Nombre,
			"product_codigo":  product.Codigo,
			"cantidad":        item.Cantidad,
			"precio_unitario": item.PrecioUnitario,
			"descuento":       item.Descuento,
			"total":           (item.Cantidad * item.PrecioUnitario) - item.Descuento,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sale":  sale,
			"items": responseItems,
		},
	})
}

func CreateSale(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)
	userID := c.MustGet("userID").(uuid.UUID)

	var input CreateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// 1. Calculate Totals and Validate Stocks
	var total float64
	for _, itemInput := range input.Items {
		// Check Stock
		var stock models.Stock
		err := tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock).Error
		if err != nil || stock.StockActual < itemInput.Cantidad {
			var product models.Product
			tx.Select("nombre").First(&product, itemInput.ProductID)
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Stock insuficiente para el producto: %s. Stock disponible: %.2f", product.Nombre, stock.StockActual),
			})
			return
		}
		total += (itemInput.Cantidad * itemInput.PrecioUnitario) - itemInput.Descuento
	}

	subtotal := total / 1.18
	igv := total - subtotal

	// 2. Create Sale Header
	sale := models.Sale{
		CompanyID:  companyID,
		BranchID:   branchID,
		CustomerID: input.CustomerID,
		UserID:     userID,
		Subtotal:   subtotal,
		IGV:        igv,
		Total:      total,
		Estado:     "completed",
	}

	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale header"})
		return
	}

	// 3. Process Items and update inventory
	for _, itemInput := range input.Items {
		// Save Sale Item
		saleItem := models.SaleItem{
			SaleID:         sale.ID,
			ProductID:      itemInput.ProductID,
			Cantidad:       itemInput.Cantidad,
			PrecioUnitario: itemInput.PrecioUnitario,
			Descuento:      itemInput.Descuento,
		}
		if err := tx.Create(&saleItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale item"})
			return
		}

		// Deduct Stock
		var stock models.Stock
		tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock)
		stockAnterior := stock.StockActual
		stock.StockActual -= itemInput.Cantidad

		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to deduct stock"})
			return
		}

		// Register in Kardex
		kardex := models.Kardex{
			CompanyID:      companyID,
			BranchID:       branchID,
			ProductID:      itemInput.ProductID,
			TipoMovimiento: "VENTA",
			Referencia:     "VENTA-" + sale.ID.String()[:8],
			Cantidad:       itemInput.Cantidad,
			StockAnterior:  stockAnterior,
			StockNuevo:     stock.StockActual,
		}
		if err := tx.Create(&kardex).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create kardex entry"})
			return
		}
	}

	tx.Commit()

	// Log audit action asynchronously
	go logAudit(userID, "Ventas", "Crear Venta", fmt.Sprintf("Venta registrada por S/. %.2f", total), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": sale})
}

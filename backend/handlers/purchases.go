package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

type PurchaseItemInput struct {
	ProductID     uuid.UUID `json:"product_id" binding:"required"`
	Cantidad      float64   `json:"cantidad" binding:"required"`
	CostoUnitario float64   `json:"costo_unitario" binding:"required"`
}

type CreatePurchaseInput struct {
	SupplierID uuid.UUID            `json:"supplier_id" binding:"required"`
	Items      []PurchaseItemInput  `json:"items" binding:"required,gt=0"`
}

func GetPurchases(c *gin.Context) {
	var purchases []models.Purchase
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Supplier").Order("created_at desc").Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": purchases})
}

func GetPurchaseDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid purchase ID"})
		return
	}

	var purchase models.Purchase
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Supplier").Where("id = ?", id).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Purchase not found"})
		return
	}

	var items []models.PurchaseItem
	if err := config.DB.Where("purchase_id = ?", purchase.ID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Fetch product details for items
	var responseItems []gin.H
	for _, item := range items {
		var product models.Product
		config.DB.Select("nombre, codigo").First(&product, item.ProductID)
		responseItems = append(responseItems, gin.H{
			"id":             item.ID,
			"product_id":     item.ProductID,
			"product_nombre": product.Nombre,
			"product_codigo": product.Codigo,
			"cantidad":       item.Cantidad,
			"costo_unitario": item.CostoUnitario,
			"total":          item.Cantidad * item.CostoUnitario,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"purchase": purchase,
			"items":    responseItems,
		},
	})
}

func CreatePurchase(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var input CreatePurchaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// 1. Calculate Totals
	var total float64
	for _, item := range input.Items {
		total += item.Cantidad * item.CostoUnitario
	}
	subtotal := total / 1.18
	igv := total - subtotal

	// 2. Create Purchase Header
	purchase := models.Purchase{
		CompanyID:  companyID,
		BranchID:   branchID,
		SupplierID: input.SupplierID,
		Fecha:      time.Now(),
		Subtotal:   subtotal,
		IGV:        igv,
		Total:      total,
		Estado:     "completed", // Auto-complete purchases
	}

	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create purchase"})
		return
	}

	// 3. Process Items and update inventory
	for _, itemInput := range input.Items {
		// Save Purchase Item
		purchaseItem := models.PurchaseItem{
			PurchaseID:    purchase.ID,
			ProductID:     itemInput.ProductID,
			Cantidad:      itemInput.Cantidad,
			CostoUnitario: itemInput.CostoUnitario,
		}
		if err := tx.Create(&purchaseItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create purchase item"})
			return
		}

		// Update Product latest buy price
		if err := tx.Model(&models.Product{}).Where("id = ? AND company_id = ?", itemInput.ProductID, companyID).Update("precio_compra", itemInput.CostoUnitario).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update product buy price"})
			return
		}

		// Get or Create Stock for this product in branch
		var stock models.Stock
		var stockAnterior float64
		err := tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock).Error
		if err != nil {
			// Create new stock record
			stock = models.Stock{
				CompanyID:   companyID,
				BranchID:    branchID,
				ProductID:   itemInput.ProductID,
				StockActual: itemInput.Cantidad,
			}
			stockAnterior = 0
			if err := tx.Create(&stock).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create stock record"})
				return
			}
		} else {
			stockAnterior = stock.StockActual
			stock.StockActual += itemInput.Cantidad
			if err := tx.Save(&stock).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update stock"})
				return
			}
		}

		// Register in Kardex
		kardex := models.Kardex{
			CompanyID:      companyID,
			BranchID:       branchID,
			ProductID:      itemInput.ProductID,
			TipoMovimiento: "INGRESO",
			Referencia:     "COMPRA-" + purchase.ID.String()[:8],
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
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": purchase})
}

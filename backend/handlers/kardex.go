package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func GetStocks(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var products []models.Product
	if err := config.DB.Where("company_id = ? AND estado = 'active'", companyID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var stocks []models.Stock
	if err := config.DB.Where("company_id = ? AND branch_id = ?", companyID, branchID).Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	stockMap := make(map[uuid.UUID]float64)
	for _, s := range stocks {
		stockMap[s.ProductID] = s.StockActual
	}

	var response []gin.H
	for _, p := range products {
		stockVal := stockMap[p.ID]
		response = append(response, gin.H{
			"product_id":     p.ID,
			"product_nombre": p.Nombre,
			"product_codigo": p.Codigo,
			"precio_compra":  p.PrecioCompra,
			"precio_venta":   p.PrecioVenta,
			"stock_minimo":   p.StockMinimo,
			"stock_actual":   stockVal,
			"estado":         p.Estado,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

func GetKardex(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var movements []models.Kardex
	query := config.DB.Where("company_id = ? AND branch_id = ?", companyID, branchID)

	productIDStr := c.Query("product_id")
	if productIDStr != "" {
		pID, err := uuid.Parse(productIDStr)
		if err == nil {
			query = query.Where("product_id = ?", pID)
		}
	}

	if err := query.Order("created_at desc").Find(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var response []gin.H
	for _, m := range movements {
		var product models.Product
		config.DB.Select("nombre, codigo").First(&product, m.ProductID)
		response = append(response, gin.H{
			"id":              m.ID,
			"product_nombre":  product.Nombre,
			"product_codigo":  product.Codigo,
			"tipo_movimiento": m.TipoMovimiento,
			"referencia":      m.Referencia,
			"cantidad":        m.Cantidad,
			"stock_anterior":  m.StockAnterior,
			"stock_nuevo":     m.StockNuevo,
			"created_at":      m.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

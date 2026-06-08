package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func GetDashboardStats(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Sales today
	var salesToday float64
	config.DB.Model(&models.Sale{}).
		Where("company_id = ? AND branch_id = ? AND created_at >= ? AND estado = 'completed'", companyID, branchID, todayStart).
		Select("COALESCE(SUM(total), 0)").
		Scan(&salesToday)

	// Sales month
	var salesMonth float64
	config.DB.Model(&models.Sale{}).
		Where("company_id = ? AND branch_id = ? AND created_at >= ? AND estado = 'completed'", companyID, branchID, monthStart).
		Select("COALESCE(SUM(total), 0)").
		Scan(&salesMonth)

	// Purchases month
	var purchasesMonth float64
	config.DB.Model(&models.Purchase{}).
		Where("company_id = ? AND branch_id = ? AND created_at >= ? AND estado = 'completed'", companyID, branchID, monthStart).
		Select("COALESCE(SUM(total), 0)").
		Scan(&purchasesMonth)

	// Low stock products count
	var lowStockCount int64
	config.DB.Table("stocks").
		Joins("JOIN products ON products.id = stocks.product_id").
		Where("stocks.company_id = ? AND stocks.branch_id = ? AND stocks.stock_actual <= products.stock_minimo AND products.estado = 'active'", companyID, branchID).
		Count(&lowStockCount)

	// Top Selling Products
	type TopProduct struct {
		Nombre   string  `json:"nombre"`
		Codigo   string  `json:"codigo"`
		Cantidad float64 `json:"cantidad"`
	}
	var topProducts []TopProduct
	config.DB.Table("sale_items").
		Select("products.nombre as nombre, products.codigo as codigo, SUM(sale_items.cantidad) as cantidad").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sales.company_id = ? AND sales.branch_id = ? AND sales.estado = 'completed'", companyID, branchID).
		Group("products.id, products.nombre, products.codigo").
		Order("cantidad desc").
		Limit(5).
		Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sales_today":     salesToday,
			"sales_month":     salesMonth,
			"purchases_month": purchasesMonth,
			"low_stock_count": lowStockCount,
			"top_products":    topProducts,
		},
	})
}

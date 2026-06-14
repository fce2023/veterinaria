package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func TestCreateAndGetPurchase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Connect to the actual test database in Docker
	config.ConnectDB()

	// Seed temporary company, branch, supplier and product
	companyID := uuid.New()
	branchID := uuid.New()
	supplierID := uuid.New()
	productID := uuid.New()

	company := models.Company{
		BaseModel: models.BaseModel{ID: companyID},
		RUC:       "20999999999",
		RazonSocial: "Test Company Purchases",
	}
	config.DB.Create(&company)
	defer config.DB.Unscoped().Delete(&company)

	branch := models.Branch{
		BaseModel: models.BaseModel{ID: branchID},
		CompanyID: companyID,
		Nombre:    "Test Branch Purchases",
	}
	config.DB.Create(&branch)
	defer config.DB.Unscoped().Delete(&branch)

	supplier := models.Supplier{
		BaseModel: models.BaseModel{ID: supplierID},
		CompanyID: companyID,
		RUC:       "20111111111",
		RazonSocial: "Test Supplier",
	}
	config.DB.Create(&supplier)
	defer config.DB.Unscoped().Delete(&supplier)

	product := models.Product{
		BaseModel: models.BaseModel{ID: productID},
		CompanyID: companyID,
		Nombre:    "Test Medicine Product",
		PrecioVenta: 10.0,
		PrecioCompra: 5.0,
		Estado:    "active",
	}
	config.DB.Create(&product)
	defer config.DB.Unscoped().Delete(&product)

	t.Run("Create Purchase", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("companyID", companyID)
			c.Set("branchID", branchID)
			c.Next()
		})
		router.POST("/purchases", CreatePurchase)

		input := CreatePurchaseInput{
			SupplierID: supplierID,
			Items: []PurchaseItemInput{
				{
					ProductID:     productID,
					Cantidad:      5,
					CostoUnitario: 4.0,
				},
			},
		}
		body, _ := json.Marshal(input)
		req, _ := http.NewRequest("POST", "/purchases", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
			return
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		purchaseIDStr := data["id"].(string)

		// Defer deletion of purchase
		purchaseID := uuid.MustParse(purchaseIDStr)
		defer func() {
			config.DB.Unscoped().Where("purchase_id = ?", purchaseID).Delete(&models.PurchaseItem{})
			config.DB.Unscoped().Where("product_id = ? AND branch_id = ?", productID, branchID).Delete(&models.Stock{})
			config.DB.Unscoped().Where("referencia = ?", "COMPRA-"+purchaseIDStr[:8]).Delete(&models.Kardex{})
			config.DB.Unscoped().Delete(&models.Purchase{BaseModel: models.BaseModel{ID: purchaseID}})
		}()

		// Verify Stock is created/updated
		var stock models.Stock
		if err := config.DB.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, productID).First(&stock).Error; err != nil {
			t.Errorf("Expected stock record to exist, got error: %v", err)
		}
		if stock.StockActual != 5 {
			t.Errorf("Expected stock actual to be 5, got %f", stock.StockActual)
		}

		// Verify Kardex entry is created
		var kardex models.Kardex
		if err := config.DB.Where("company_id = ? AND branch_id = ? AND product_id = ? AND tipo_movimiento = 'INGRESO'", companyID, branchID, productID).First(&kardex).Error; err != nil {
			t.Errorf("Expected Kardex record to exist, got error: %v", err)
		}
	})
}

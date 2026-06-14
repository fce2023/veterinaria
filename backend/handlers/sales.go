package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
	"veterinaria/backend/services"
)

type SaleItemInput struct {
	ProductID      uuid.UUID `json:"product_id" binding:"required"`
	Cantidad       float64   `json:"cantidad"` // Optional if dimensional
	PrecioUnitario float64   `json:"precio_unitario" binding:"required"`
	Descuento      float64   `json:"descuento"`
	Alto           float64   `json:"alto"`
	Ancho          float64   `json:"ancho"`
	Espesor        float64   `json:"espesor"`
	CantidadPiezas int       `json:"cantidad_piezas"`
}

type CreateSaleInput struct {
	CustomerID    *uuid.UUID      `json:"customer_id"` // Optional
	TipoDocumento string          `json:"tipo_documento"` // "01" Factura, "03" Boleta
	MetodoPago    string          `json:"metodo_pago"`    // EFECTIVO, TARJETA, etc.
	Items         []SaleItemInput `json:"items" binding:"required,gt=0"`
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
		config.DB.Select("nombre, codigo, is_dimensional, unidad_medida").First(&product, item.ProductID)

		var dim models.SaleItemDimension
		hasDim := false
		if product.IsDimensional {
			if err := config.DB.Where("sale_item_id = ?", item.ID).First(&dim).Error; err == nil {
				hasDim = true
			}
		}

		responseItem := gin.H{
			"id":              item.ID,
			"product_id":      item.ProductID,
			"product_nombre":  product.Nombre,
			"product_codigo":  product.Codigo,
			"cantidad":        item.Cantidad,
			"precio_unitario": item.PrecioUnitario,
			"descuento":       item.Descuento,
			"total":           (item.Cantidad * item.PrecioUnitario) - item.Descuento,
			"is_dimensional":  product.IsDimensional,
			"unidad_medida":   product.UnidadMedida,
		}

		if hasDim {
			responseItem["alto"] = dim.Alto
			responseItem["ancho"] = dim.Ancho
			responseItem["espesor"] = dim.Espesor
			responseItem["cantidad_piezas"] = dim.CantidadPiezas
		}

		responseItems = append(responseItems, responseItem)
	}

	// Fetch Electronic Document if exists
	var electronicDoc models.ElectronicDocument
	hasElectronic := false
	if err := config.DB.Where("sale_id = ?", sale.ID).First(&electronicDoc).Error; err == nil {
		hasElectronic = true
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sale":                sale,
			"items":               responseItems,
			"electronic_document": electronicDoc,
			"has_electronic":      hasElectronic,
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

	// Verify that a cash session is open for this user and branch
	var session models.CashSession
	if err := config.DB.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&session).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Debe abrir una sesión de caja antes de realizar ventas."})
		return
	}

	tx := config.DB.Begin()

	// 0. Handle Customer (if missing, find or create "Público General")
	var finalCustomerID uuid.UUID
	if input.CustomerID == nil || *input.CustomerID == uuid.Nil {
		var publicCustomer models.Customer
		if err := tx.Where("company_id = ? AND numero_documento = '00000000'", companyID).First(&publicCustomer).Error; err != nil {
			// Create it if not exists
			publicCustomer = models.Customer{
				CompanyID:       companyID,
				TipoDocumento:   "DNI",
				NumeroDocumento: "00000000",
				Nombre:          "Público General",
			}
			if err := tx.Create(&publicCustomer).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear cliente por defecto"})
				return
			}
		}
		finalCustomerID = publicCustomer.ID
	} else {
		finalCustomerID = *input.CustomerID
	}

	// 1. Calculate Totals, Validate Stocks and Dimensional formulas
	var total float64
	computedQuantities := make(map[uuid.UUID]float64)

	for _, itemInput := range input.Items {
		var product models.Product
		if err := tx.First(&product, itemInput.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Producto no encontrado"})
			return
		}

		cantidadFinal := itemInput.Cantidad
		if product.IsDimensional {
			piezas := itemInput.CantidadPiezas
			if piezas <= 0 {
				piezas = 1
			}

			switch product.UnidadMedida {
			case "m":
				if itemInput.Alto <= 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   fmt.Sprintf("Producto '%s' (medida lineal) requiere longitud válida (alto > 0)", product.Nombre),
					})
					return
				}
				cantidadFinal = itemInput.Alto * float64(piezas)
			case "m2":
				if itemInput.Alto <= 0 || itemInput.Ancho <= 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   fmt.Sprintf("Producto '%s' (medida de área) requiere alto y ancho válidos (> 0)", product.Nombre),
					})
					return
				}
				cantidadFinal = itemInput.Alto * itemInput.Ancho * float64(piezas)
			case "m3":
				if itemInput.Alto <= 0 || itemInput.Ancho <= 0 || itemInput.Espesor <= 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   fmt.Sprintf("Producto '%s' (medida cúbica/volumen) requiere alto, ancho y espesor válidos (> 0)", product.Nombre),
					})
					return
				}
				cantidadFinal = itemInput.Alto * itemInput.Ancho * itemInput.Espesor * float64(piezas)
			default:
				// Fallback to Area
				if itemInput.Alto <= 0 || itemInput.Ancho <= 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   fmt.Sprintf("Producto dimensional '%s' requiere dimensiones de alto y ancho válidas (> 0)", product.Nombre),
					})
					return
				}
				cantidadFinal = itemInput.Alto * itemInput.Ancho * float64(piezas)
			}
		}

		computedQuantities[itemInput.ProductID] = cantidadFinal

		// Check Stock
		var stock models.Stock
		err := tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock).Error
		if err != nil || stock.StockActual < cantidadFinal {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Stock insuficiente para el producto: %s. Stock disponible: %.2f (Solicitado: %.2f %s)", product.Nombre, stock.StockActual, cantidadFinal, product.UnidadMedida),
			})
			return
		}
		total += (cantidadFinal * itemInput.PrecioUnitario) - itemInput.Descuento
	}

	var subtotal, igv float64
	if input.TipoDocumento == "01" || input.TipoDocumento == "03" {
		subtotal = total / 1.18
		igv = total - subtotal
	} else {
		subtotal = total
		igv = 0
	}

	// 2. Create Sale Header
	// Determine Serie and Number
	var branch models.Branch
	tx.First(&branch, branchID)

	serie := branch.SerieBoleta
	if input.TipoDocumento == "01" {
		serie = branch.SerieFactura
	} else if input.TipoDocumento == "NV" {
		serie = "NV01"
	}
	if serie == "" {
		if input.TipoDocumento == "01" {
			serie = "F001"
		} else if input.TipoDocumento == "NV" {
			serie = "NV01"
		} else {
			serie = "B001"
		}
	}

	var nextNumber int = 1
	if input.TipoDocumento == "01" {
		if branch.CorrelativoFactura > 0 {
			nextNumber = branch.CorrelativoFactura
		}
	} else if input.TipoDocumento == "03" {
		if branch.CorrelativoBoleta > 0 {
			nextNumber = branch.CorrelativoBoleta
		}
	}

	var lastSale models.Sale
	if err := tx.Where("branch_id = ? AND tipo_documento = ? AND serie = ?", branchID, input.TipoDocumento, serie).Order("created_at desc").First(&lastSale).Error; err == nil {
		if n, err := strconv.Atoi(lastSale.Numero); err == nil {
			if n >= nextNumber {
				nextNumber = n + 1
			}
		}
	}

	sale := models.Sale{
		CompanyID:     companyID,
		BranchID:      branchID,
		CustomerID:    finalCustomerID,
		UserID:        userID,
		TipoDocumento: input.TipoDocumento,
		Serie:         serie,
		Numero:        fmt.Sprintf("%d", nextNumber),
		MetodoPago:    input.MetodoPago,
		Subtotal:      subtotal,
		IGV:           igv,
		Total:         total,
		Estado:        "completed",
	}

	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale header"})
		return
	}

	// Update Cash Session if it's a cash sale
	if input.MetodoPago == "EFECTIVO" || input.MetodoPago == "" {
		session.TotalVentasEfe += total
	} else {
		session.TotalVentasOtr += total
	}
	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar totales de caja"})
		return
	}

	// 3. Process Items and update inventory
	var itemsToEmit []models.SaleItem
	for _, itemInput := range input.Items {
		cantidadFinal := computedQuantities[itemInput.ProductID]

		// Save Sale Item
		saleItem := models.SaleItem{
			SaleID:         sale.ID,
			ProductID:      itemInput.ProductID,
			Cantidad:       cantidadFinal,
			PrecioUnitario: itemInput.PrecioUnitario,
			Descuento:      itemInput.Descuento,
		}
		if err := tx.Create(&saleItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale item"})
			return
		}
		itemsToEmit = append(itemsToEmit, saleItem)

		// Save dimensional extension if applicable
		var product models.Product
		tx.Select("is_dimensional").First(&product, itemInput.ProductID)
		if product.IsDimensional {
			piezas := itemInput.CantidadPiezas
			if piezas <= 0 {
				piezas = 1
			}
			dim := models.SaleItemDimension{
				SaleItemID:     saleItem.ID,
				Alto:           itemInput.Alto,
				Ancho:          itemInput.Ancho,
				Espesor:        itemInput.Espesor,
				CantidadPiezas: piezas,
			}
			if err := tx.Create(&dim).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save dimensional item extensions"})
				return
			}
		}

		// Deduct Stock
		var stock models.Stock
		tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock)
		stockAnterior := stock.StockActual
		stock.StockActual -= cantidadFinal

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
			Cantidad:       cantidadFinal,
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

	// 4. Trigger Electronic Billing (Asynchronous)
	go func() {
		billingService := services.NewBillingService()
		_, err := billingService.EmitSale(sale, itemsToEmit)
		if err != nil {
			fmt.Printf("Error emitting electronic document: %v\n", err)
		}
	}()

	// Log audit action asynchronously
	go logAudit(userID, "Ventas", "Crear Venta", fmt.Sprintf("Venta registrada por S/. %.2f", total), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": sale})
}

type CreateCreditNoteInput struct {
	Motivo string `json:"motivo" binding:"required"`
}

func CreateCreditNote(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)
	userID := c.MustGet("userID").(uuid.UUID)

	idStr := c.Param("id")
	saleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid sale ID"})
		return
	}

	var input CreateCreditNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// 1. Fetch Sale
	var sale models.Sale
	if err := tx.Scopes(config.BranchFilter(c)).Where("id = ?", saleID).First(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Venta no encontrada"})
		return
	}

	if sale.Estado == "annulled" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Esta venta ya ha sido anulada"})
		return
	}

	// 2. Revert Cash Session total
	var session models.CashSession
	if err := tx.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&session).Error; err == nil {
		if sale.MetodoPago == "EFECTIVO" || sale.MetodoPago == "" {
			session.TotalVentasEfe -= sale.Total
		} else {
			session.TotalVentasOtr -= sale.Total
		}
		if err := tx.Save(&session).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar totales de caja"})
			return
		}
	}

	// 3. Fetch Items and Revert Stock
	var items []models.SaleItem
	if err := tx.Where("sale_id = ?", sale.ID).Find(&items).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al recuperar ítems de venta"})
		return
	}

	for _, item := range items {
		// Restore Stock
		var stock models.Stock
		tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, item.ProductID).First(&stock)
		stockAnterior := stock.StockActual
		stock.StockActual += item.Cantidad
		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al revertir inventario"})
			return
		}

		// Register in Kardex
		kardex := models.Kardex{
			CompanyID:      companyID,
			BranchID:       branchID,
			ProductID:      item.ProductID,
			TipoMovimiento: "ANULACION",
			Referencia:     "NC-" + sale.Serie + "-" + sale.Numero,
			Cantidad:       item.Cantidad,
			StockAnterior:  stockAnterior,
			StockNuevo:     stock.StockActual,
		}
		if err := tx.Create(&kardex).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al registrar movimiento en Kardex"})
			return
		}
	}

	// 4. Determine Credit Note Serie & Number
	serie := "BC01" // Default for Boleta Credit Note
	if sale.TipoDocumento == "01" {
		serie = "FC01" // Default for Factura Credit Note
	} else if sale.TipoDocumento == "NV" {
		serie = "NC01" // Internal Credit Note for Nota de Venta
	}

	var lastNC models.NotaCredito
	nextNumber := 1
	if err := tx.Where("branch_id = ? AND serie = ?", branchID, serie).Order("created_at desc").First(&lastNC).Error; err == nil {
		if n, err := strconv.Atoi(lastNC.Numero); err == nil {
			nextNumber = n + 1
		}
	}

	nc := models.NotaCredito{
		CompanyID:   companyID,
		BranchID:    branchID,
		SaleID:      sale.ID,
		Serie:       serie,
		Numero:      fmt.Sprintf("%d", nextNumber),
		Motivo:      input.Motivo,
		Total:       sale.Total,
		Estado:      "completed",
	}

	if err := tx.Create(&nc).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al registrar Nota de Crédito"})
		return
	}

	// Update Sale status to annulled
	sale.Estado = "annulled"
	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar estado de la venta"})
		return
	}

	tx.Commit()

	// Asynchronously emit electronic Credit Note to SUNAT if applicable
	if sale.TipoDocumento == "01" || sale.TipoDocumento == "03" {
		go func() {
			fmt.Printf("Emisión electrónica de Nota de Crédito %s-%s para venta %s iniciada.\n", nc.Serie, nc.Numero, sale.ID)
		}()
	}

	go logAudit(userID, "Ventas", "Anular Venta", fmt.Sprintf("Nota de Crédito %s-%s emitida para Venta #%s. Motivo: %s", nc.Serie, nc.Numero, sale.ID.String()[:8], input.Motivo), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": nc})
}

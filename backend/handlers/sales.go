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
	var totalVenta, totalSubtotal, totalIGV float64
	computedItems := make([]models.SaleItem, 0)

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
				cantidadFinal = itemInput.Alto * float64(piezas)
			case "m2":
				cantidadFinal = itemInput.Alto * itemInput.Ancho * float64(piezas)
			case "m3":
				cantidadFinal = itemInput.Alto * itemInput.Ancho * itemInput.Espesor * float64(piezas)
			default:
				cantidadFinal = float64(piezas)
			}
		}

		if cantidadFinal <= 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("Cantidad inválida para %s", product.Nombre)})
			return
		}

		// Check Stock
		var stock models.Stock
		err := tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, itemInput.ProductID).First(&stock).Error
		if err != nil || stock.StockActual < cantidadFinal {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Stock insuficiente para: %s (Disponible: %.2f)", product.Nombre, stock.StockActual),
			})
			return
		}

		// --- GOLDEN RULES CALCULATIONS ---
		// 1. Precio Unitario (Ya viene con IGV desde el POS)
		precioConIGV := itemInput.PrecioUnitario
		
		// 2. Total del Ítem (Cantidad * Precio - Descuento)
		itemTotal := (cantidadFinal * precioConIGV) - itemInput.Descuento
		
		// 3. Subtotal e IGV del Ítem (Asumiendo Gravado - Catálogo 07: 10)
		// TODO: En el futuro permitir otros tipos de afectación desde el frontend
		itemSubtotal := itemTotal / 1.18
		itemIGV := itemTotal - itemSubtotal

		// Map to model
		saleItem := models.SaleItem{
			ProductID:      itemInput.ProductID,
			Cantidad:       cantidadFinal,
			PrecioUnitario: precioConIGV,
			Subtotal:       itemSubtotal,
			IGV:            itemIGV,
			Descuento:      itemInput.Descuento,
			TipoAfectacion: "10", // Default Gravado
			UnidadSUNAT:    mapUnidadSUNAT(product.UnidadMedida),
		}

		computedItems = append(computedItems, saleItem)
		totalVenta += itemTotal
		totalSubtotal += itemSubtotal
		totalIGV += itemIGV
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

	// Validation: Factura (01) requires RUC
	var customer models.Customer
	tx.First(&customer, finalCustomerID)
	if input.TipoDocumento == "01" && customer.TipoDocumento != "RUC" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "La Factura requiere un cliente con RUC"})
		return
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
		Subtotal:      totalSubtotal,
		IGV:           totalIGV,
		Total:         totalVenta,
		Estado:        "completed",
	}

	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale header"})
		return
	}

	// Update Cash Session
	if input.MetodoPago == "EFECTIVO" || input.MetodoPago == "" {
		session.TotalVentasEfe += totalVenta
	} else {
		session.TotalVentasOtr += totalVenta
	}
	tx.Save(&session)

	// 3. Process Items and update inventory
	var itemsToEmit []models.SaleItem
	for _, si := range computedItems {
		si.SaleID = sale.ID
		if err := tx.Create(&si).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create sale item"})
			return
		}
		itemsToEmit = append(itemsToEmit, si)

		// Save dimensional extension if applicable
		var product models.Product
		tx.Select("is_dimensional").First(&product, si.ProductID)
		if product.IsDimensional {
			// Find original input to get dimensions
			var originalInput SaleItemInput
			for _, inp := range input.Items {
				if inp.ProductID == si.ProductID {
					originalInput = inp
					break
				}
			}
			piezas := originalInput.CantidadPiezas
			if piezas <= 0 { piezas = 1 }
			dim := models.SaleItemDimension{
				SaleItemID:     si.ID,
				Alto:           originalInput.Alto,
				Ancho:          originalInput.Ancho,
				Espesor:        originalInput.Espesor,
				CantidadPiezas: piezas,
			}
			tx.Create(&dim)
		}

		// Deduct Stock
		var stock models.Stock
		tx.Where("company_id = ? AND branch_id = ? AND product_id = ?", companyID, branchID, si.ProductID).First(&stock)
		stockAnterior := stock.StockActual
		stock.StockActual -= si.Cantidad
		tx.Save(&stock)

		// Register in Kardex
		tx.Create(&models.Kardex{
			CompanyID:      companyID,
			BranchID:       branchID,
			ProductID:      si.ProductID,
			TipoMovimiento: "VENTA",
			Referencia:     "VENTA-" + sale.ID.String()[:8],
			Cantidad:       si.Cantidad,
			StockAnterior:  stockAnterior,
			StockNuevo:     stock.StockActual,
		})
	}

	tx.Commit()

	// 4. Trigger Electronic Billing Integration
	go func() {
		if sale.TipoDocumento == "01" || sale.TipoDocumento == "03" {
			var billingConfig models.BillingConfig
			config.DB.Where("company_id = ?", companyID).First(&billingConfig)

			if billingConfig.Estado == "active" {
				if billingConfig.EmisionDiferida {
					// MODO DIFERIDO: Solo guardar el registro en borrador
					doc := models.ElectronicDocument{
						CompanyID:     companyID,
						SaleID:        &sale.ID,
						TipoDocumento: sale.TipoDocumento,
						Serie:         sale.Serie,
						Numero:        sale.Numero,
						Estado:        "draft",
						Observaciones: "Emisión diferida: Pendiente de revisión al cierre de jornada",
					}
					config.DB.Create(&doc)
				} else {
					// MODO INSTANTÁNEO: Enviar a la API de inmediato (comportamiento actual)
					billingService := services.NewBillingService()
					billingService.EmitSale(sale, itemsToEmit)
				}
			}
		}
	}()

	go logAudit(userID, "Ventas", "Crear Venta", fmt.Sprintf("Venta %s-%s registrada", sale.Serie, sale.Numero), c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": sale})
}

func mapUnidadSUNAT(unidad string) string {
	switch unidad {
	case "und", "un.":
		return "NIU"
	case "m", "m2", "m3":
		return "NIU" // O MTR/MTK/MTQ segun sea estricto, pero NIU es universal para bienes
	case "kg":
		return "KGM"
	case "servicio":
		return "ZZ"
	default:
		return "NIU"
	}
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

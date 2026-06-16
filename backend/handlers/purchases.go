package handlers

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	SupplierID        uuid.UUID            `json:"supplier_id" binding:"required"`
	Items             []PurchaseItemInput  `json:"items" binding:"required,gt=0"`
	Estado            string               `json:"estado"`
	TipoComprobante   string               `json:"tipo_comprobante"`
	NumeroComprobante string               `json:"numero_comprobante"`
}

func GetPurchases(c *gin.Context) {
	var purchases []models.Purchase
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Supplier").Preload("Attachments").Order("created_at desc").Find(&purchases).Error; err != nil {
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
	if err := config.DB.Scopes(config.BranchFilter(c)).Preload("Supplier").Preload("Attachments").Where("id = ?", id).First(&purchase).Error; err != nil {
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

	estado := "completed"
	if input.Estado != "" {
		estado = input.Estado
	}

	// 2. Create Purchase Header
	purchase := models.Purchase{
		CompanyID:         companyID,
		BranchID:          branchID,
		SupplierID:        input.SupplierID,
		Fecha:             time.Now(),
		Subtotal:          subtotal,
		IGV:               igv,
		Total:             total,
		Estado:            estado,
		TipoComprobante:   input.TipoComprobante,
		NumeroComprobante: input.NumeroComprobante,
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

func UpdatePurchase(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de compra inválido"})
		return
	}

	var purchase models.Purchase
	if err := config.DB.Preload("Supplier").Where("id = ? AND company_id = ?", id, companyID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Compra no encontrada"})
		return
	}

	// Read form values
	supplierIDStr := c.PostForm("supplier_id")
	if supplierIDStr != "" {
		if supplierID, err := uuid.Parse(supplierIDStr); err == nil {
			purchase.SupplierID = supplierID
			// Reload Supplier model if it has changed to keep RUC in sync
			var newSupplier models.Supplier
			if err := config.DB.Where("id = ? AND company_id = ?", supplierID, companyID).First(&newSupplier).Error; err == nil {
				purchase.Supplier = newSupplier
			}
		}
	}

	fechaStr := c.PostForm("fecha")
	if fechaStr != "" {
		if t, err := time.Parse(time.RFC3339, fechaStr); err == nil {
			purchase.Fecha = t
		} else if t, err := time.Parse("2006-01-02", fechaStr); err == nil {
			purchase.Fecha = t
		}
	}

	estado := c.PostForm("estado")
	if estado != "" {
		purchase.Estado = estado
	}

	tipoComprobante := c.PostForm("tipo_comprobante")
	if tipoComprobante != "" {
		purchase.TipoComprobante = tipoComprobante
	}

	numeroComprobante := c.PostForm("numero_comprobante")
	if numeroComprobante != "" {
		purchase.NumeroComprobante = numeroComprobante
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err == nil {
		// Ensure uploads directory exists
		uploadDir := "./uploads/purchases"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo crear la carpeta de subidas"})
			return
		}

		filename := generateProfessionalFilename(purchase, id, file.Filename, "", "")
		dst := filepath.Join(uploadDir, filename)

		// Check if it's an image
		ext := strings.ToLower(filepath.Ext(file.Filename))
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"

		if isImage {
			fileReader, err := file.Open()
			if err == nil {
				defer fileReader.Close()
				if err := compressAndSaveImage(fileReader, dst); err != nil {
					_ = c.SaveUploadedFile(file, dst)
				}
			} else {
				_ = c.SaveUploadedFile(file, dst)
			}
		} else {
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo guardar el archivo"})
				return
			}
		}

		purchase.ComprobanteUrl = "/uploads/purchases/" + filename
	}

	if err := config.DB.Save(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo actualizar la compra"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": purchase})
}

// Helper to compress images (JPEG, PNG, GIF) to JPEG with 75% quality settings
func compressAndSaveImage(src io.Reader, dstPath string) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 75% quality is visually indistinguishable for documents and yields ~80% space saving
	return jpeg.Encode(out, img, &jpeg.Options{Quality: 75})
}

func UploadPurchaseAttachment(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de compra inválido"})
		return
	}

	// Verify purchase belongs to this company and preload supplier to generate professional name
	var purchase models.Purchase
	if err := config.DB.Preload("Supplier").Where("id = ? AND company_id = ?", purchaseID, companyID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Compra no encontrada"})
		return
	}

	// Parse file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No se recibió ningún archivo"})
		return
	}

	// Ensure upload dir exists
	uploadDir := "./uploads/purchases"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo crear la carpeta de subidas"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	attachmentID := uuid.New()

	tipoDoc := c.PostForm("tipo_documento")
	numDoc := c.PostForm("numero_documento")
	descripcion := c.PostForm("descripcion")

	// Generate professional safe filename
	finalFilename := generateProfessionalFilename(purchase, attachmentID, file.Filename, tipoDoc, numDoc)
	dstPath := filepath.Join(uploadDir, finalFilename)

	// Determine if it is an image to compress
	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
	var relativeUrl string

	if isImage {
		fileReader, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo abrir el archivo subido"})
			return
		}
		defer fileReader.Close()

		if err := compressAndSaveImage(fileReader, dstPath); err != nil {
			// Fallback to normal upload if compression fails
			if err := c.SaveUploadedFile(file, dstPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo guardar la imagen"})
				return
			}
		}
		relativeUrl = "/uploads/purchases/" + finalFilename
	} else {
		// Normal upload (e.g. PDF)
		if err := c.SaveUploadedFile(file, dstPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo guardar el archivo"})
			return
		}
		relativeUrl = "/uploads/purchases/" + finalFilename
	}

	// Save attachment to DB
	attachment := models.PurchaseAttachment{
		PurchaseID:  purchaseID,
		Nombre:      finalFilename,
		Url:         relativeUrl,
		Descripcion: descripcion,
	}
	attachment.ID = attachmentID

	if err := config.DB.Create(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo registrar el adjunto en base de datos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": attachment})
}

func DeletePurchaseAttachment(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)

	purchaseIDStr := c.Param("id")
	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de compra inválido"})
		return
	}

	attachmentIDStr := c.Param("attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de adjunto inválido"})
		return
	}

	// Verify purchase belongs to this company
	var purchase models.Purchase
	if err := config.DB.Where("id = ? AND company_id = ?", purchaseID, companyID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Compra no encontrada"})
		return
	}

	// Find the attachment
	var attachment models.PurchaseAttachment
	if err := config.DB.Where("id = ? AND purchase_id = ?", attachmentID, purchaseID).First(&attachment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Adjunto no encontrado"})
		return
	}

	// Delete file from disk if it exists
	localPath := "." + attachment.Url
	if _, err := os.Stat(localPath); err == nil {
		_ = os.Remove(localPath)
	}

	// Delete record from DB
	if err := config.DB.Delete(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo eliminar el registro del adjunto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Adjunto eliminado con éxito"})
}

func generateProfessionalFilename(purchase models.Purchase, id uuid.UUID, originalName string, customTipo string, customNumero string) string {
	ext := strings.ToLower(filepath.Ext(originalName))

	tipo := "COMPROBANTE"
	if customTipo != "" {
		tipo = strings.ToUpper(customTipo)
	} else if purchase.TipoComprobante != "" {
		tipo = strings.ToUpper(purchase.TipoComprobante)
	}
	tipo = sanitizeString(tipo)

	ruc := "NO-RUC"
	if purchase.Supplier.RUC != "" {
		ruc = purchase.Supplier.RUC
	}

	numero := "S-N"
	if customNumero != "" {
		numero = strings.ToUpper(customNumero)
	} else if purchase.NumeroComprobante != "" {
		numero = strings.ToUpper(purchase.NumeroComprobante)
	}
	numero = sanitizeString(numero)

	fecha := "NO-FECHA"
	if !purchase.Fecha.IsZero() {
		fecha = purchase.Fecha.Format("2006-01-02")
	}

	monto := fmt.Sprintf("%.2f", purchase.Total)

	suffix := id.String()[:4]

	// Format: TIPO_RUC-XXX_NUM-XXX_FECHA-XXX_MONTO-XXX_SUF.ext
	filename := fmt.Sprintf("%s_RUC-%s_NUM-%s_FECHA-%s_MONTO-%s_%s%s", tipo, ruc, numero, fecha, monto, suffix, ext)
	return filename
}

func sanitizeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else if r == ' ' || r == '/' || r == '\\' {
			result.WriteRune('-')
		}
	}
	return result.String()
}

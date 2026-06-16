package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
	"veterinaria/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetBillingConfig retrieves the billing configuration for the company
func GetBillingConfig(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Formato de ID de empresa inválido"})
		return
	}
	
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Configuración no encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    billingConfig,
	})
}

// SaveBillingConfig creates or updates the billing configuration
func SaveBillingConfig(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}

	var input struct {
		ApiURL              string `json:"api_url"`
		ApiKey              string `json:"api_key"`
		TenantUUID          string `json:"tenant_uuid"`
		Modo                string `json:"modo"`
		Estado              string `json:"estado"`
		SolUser             string `json:"sol_user"`
		SolPass             string `json:"sol_pass"`
		CertificadoBase64   string `json:"certificado_base64"`
		CertificadoPassword string `json:"certificado_password"`
		ClientID            string `json:"client_id"`
		ClientSecret        string `json:"client_secret"`
		EmisionDiferida     bool   `json:"emision_diferida"`
		CorrelativoPadding  *int   `json:"correlativo_padding"`
		WebhookURL          string `json:"webhook_url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var billingConfig models.BillingConfig
	err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error

	billingConfig.CompanyID = companyID
	billingConfig.ApiURL = input.ApiURL
	billingConfig.ApiKey = input.ApiKey
	billingConfig.TenantUUID = input.TenantUUID
	billingConfig.Modo = input.Modo
	billingConfig.Estado = input.Estado
	billingConfig.SolUser = input.SolUser
	billingConfig.SolPass = input.SolPass
	billingConfig.ClientID = input.ClientID
	billingConfig.ClientSecret = input.ClientSecret
	billingConfig.EmisionDiferida = input.EmisionDiferida
	billingConfig.CorrelativoPadding = input.CorrelativoPadding
	billingConfig.WebhookURL = input.WebhookURL
	if input.CertificadoBase64 != "" {
		billingConfig.CertificadoBase64 = input.CertificadoBase64
	}
	if input.CertificadoPassword != "" {
		billingConfig.CertificadoPassword = input.CertificadoPassword
	}

	// 1. Fetch Company details
	var company models.Company
	if err := config.DB.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se encontró la empresa"})
		return
	}

	// 2. Resolve global API credentials if not overridden
	apiURL := input.ApiURL
	apiKey := input.ApiKey
	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting).Error; err == nil && apiURL == "" {
			apiURL = globalURLSetting.Valor
		}
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting).Error; err == nil && apiKey == "" {
			apiKey = globalKeySetting.Valor
		}
	}
	
	// Sanitize URL
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	var logs []string
	uuidStr := input.TenantUUID

	// Call external FacturaAPI only if API URL & Key are present
	if apiURL != "" && apiKey != "" {
		billingService := services.NewBillingService()
		
		// Ensure company exists in FacturaAPI
		logs = append(logs, "Verificando empresa en FacturaAPI...")
		finalUUID, err := billingService.EnsureCompanyExists(companyID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"logs":    logs,
				"error":   "Error al verificar/registrar empresa en FacturaAPI: " + err.Error(),
			})
			return
		}
		uuidStr = finalUUID
		billingConfig.TenantUUID = uuidStr
		logs = append(logs, "Empresa verificada/registrada correctamente con UUID: "+uuidStr)

		// Save/Update in DB so subsequent steps (which might re-load from DB) see the updated UUID
		config.DB.Save(&billingConfig)

		// Sync full company info (address, ubigeo, etc) to FacturaAPI
		logs = append(logs, "Sincronizando perfil fiscal de la empresa...")
		if syncErr := billingService.SyncCompanyInfo(company); syncErr != nil {
			logs = append(logs, "Aviso: Error al sincronizar perfil fiscal: "+syncErr.Error())
		} else {
			logs = append(logs, "Perfil fiscal sincronizado correctamente.")
		}

		// Step B: PATCH Credentials
		if uuidStr != "" {
			logs = append(logs, "Sincronizando credenciales SUNAT/API GRE...")
			credPayload := map[string]string{
				"usuario_sol":   input.SolUser,
				"clave_sol":     input.SolPass,
				"client_id":     input.ClientID,
				"client_secret": input.ClientSecret,
			}
			jsonCreds, _ := json.Marshal(credPayload)

			// We still use the client directly here for custom patches not in service
			client := &http.Client{Timeout: 20 * time.Second}
			reqCred, err := http.NewRequest("PATCH", apiURL+"/api/v1/config/"+uuidStr+"/sunat-credentials", bytes.NewBuffer(jsonCreds))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al preparar PATCH credenciales: " + err.Error()})
				return
			}
			reqCred.Header.Set("Authorization", "Bearer "+apiKey)
			reqCred.Header.Set("X-API-Key", apiKey)
			reqCred.Header.Set("Content-Type", "application/json")
			reqCred.Header.Set("Accept", "application/json")

			respCred, err := client.Do(reqCred)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error de conexión al sincronizar credenciales: " + err.Error()})
				return
			}
			defer respCred.Body.Close()
			bodyCredBytes, _ := io.ReadAll(respCred.Body)

			if respCred.StatusCode != http.StatusOK && respCred.StatusCode != http.StatusAccepted {
				errMsg := fmt.Sprintf("Error credenciales FacturaAPI (%d): %s", respCred.StatusCode, string(bodyCredBytes))
				c.JSON(http.StatusBadGateway, gin.H{
					"success":   false,
					"logs":      logs,
					"api_error": errMsg,
					"error":     errMsg,
				})
				return
			}
			logs = append(logs, "Credenciales SUNAT/API GRE sincronizadas correctamente.")

			// Step C: POST Certificate
			if input.CertificadoBase64 != "" {
				logs = append(logs, "Subiendo certificado digital .p12...")
				certPassword := input.CertificadoPassword
				if certPassword == "" {
					certPassword = billingConfig.CertificadoPassword
				}
				if certPassword == "" {
					certPassword = input.SolPass // fallback to SOL password
				}
				certPayload := map[string]string{
					"certificate_base64": input.CertificadoBase64,
					"password":           certPassword,
					"extension":          "p12",
				}
				jsonCert, _ := json.Marshal(certPayload)

				reqCert, err := http.NewRequest("POST", apiURL+"/api/v1/config/"+uuidStr+"/certificate", bytes.NewBuffer(jsonCert))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al preparar subida de certificado: " + err.Error()})
					return
				}
				reqCert.Header.Set("Authorization", "Bearer "+apiKey)
				reqCert.Header.Set("X-API-Key", apiKey)
				reqCert.Header.Set("Content-Type", "application/json")
				reqCert.Header.Set("Accept", "application/json")

				respCert, err := client.Do(reqCert)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error de conexión al subir certificado: " + err.Error()})
					return
				}
				defer respCert.Body.Close()
				bodyCertBytes, _ := io.ReadAll(respCert.Body)

				if respCert.StatusCode != http.StatusOK && respCert.StatusCode != http.StatusAccepted {
					errMsg := fmt.Sprintf("Error certificado FacturaAPI (%d): %s", respCert.StatusCode, string(bodyCertBytes))
					c.JSON(http.StatusBadGateway, gin.H{
						"success":   false,
						"logs":      logs,
						"api_error": errMsg,
						"error":     errMsg,
					})
					return
				}
				logs = append(logs, "Certificado digital subido y firmado correctamente.")
			}

			// Step D: POST Logo (Multipart/form-data)
			if company.LogoBase64 != "" {
				logs = append(logs, "Sincronizando logo comercial con FacturaAPI (Multipart)...")

				// Decode base64 to binary
				imgData, err := base64.StdEncoding.DecodeString(company.LogoBase64)
				if err != nil {
					logs = append(logs, "Error al decodificar logo base64: "+err.Error())
				} else {
					body := &bytes.Buffer{}
					writer := multipart.NewWriter(body)
					part, err := writer.CreateFormFile("logo", "logo.png")
					if err != nil {
						logs = append(logs, "Error al crear form file: "+err.Error())
					} else {
						_, _ = io.Copy(part, bytes.NewReader(imgData))
						writer.Close()

						reqLogo, err := http.NewRequest("POST", apiURL+"/api/v1/config/"+uuidStr+"/logo", body)
						if err != nil {
							logs = append(logs, "Error al preparar subida de logo: "+err.Error())
						} else {
							reqLogo.Header.Set("Authorization", "Bearer "+apiKey)
							reqLogo.Header.Set("X-API-Key", apiKey)
							reqLogo.Header.Set("Content-Type", writer.FormDataContentType())
							reqLogo.Header.Set("Accept", "application/json")

							respLogo, err := client.Do(reqLogo)
							if err != nil {
								logs = append(logs, "Error de conexión al subir logo: "+err.Error())
							} else {
								defer respLogo.Body.Close()
								if respLogo.StatusCode != http.StatusOK && respLogo.StatusCode != http.StatusAccepted {
									bodyLogoBytes, _ := io.ReadAll(respLogo.Body)
									logs = append(logs, fmt.Sprintf("Error logo FacturaAPI (%d): %s", respLogo.StatusCode, string(bodyLogoBytes)))
								} else {
									logs = append(logs, "Logo comercial sincronizado correctamente.")
								}
							}
						}
					}
				}
			}
		}
	}

	if err != nil {
		// Create new
		if err := config.DB.Create(&billingConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear configuración"})
			return
		}
	} else {
		// Update existing
		if err := config.DB.Save(&billingConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar configuración"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuración guardada exitosamente",
		"logs":    logs,
		"data":    billingConfig,
	})
}

// GetBillingFiles retrieves the XML, PDF and CDR download links from FacturaAPI
func GetBillingFiles(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}
	docUUID := c.Param("uuid")
	if _, err := uuid.Parse(docUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido (debe ser un UUID)"})
		return
	}

	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Configuración de facturación no encontrada"})
		return
	}

	apiURL := billingConfig.ApiURL
	apiKey := billingConfig.ApiKey
	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting).Error; err == nil && apiURL == "" {
			apiURL = globalURLSetting.Valor
		}
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting).Error; err == nil && apiKey == "" {
			apiKey = globalKeySetting.Valor
		}
	}

	// Sanitize URL
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	if apiURL == "" || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "FacturaAPI global configuration missing"})
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL+"/api/v1/documents/"+docUUID+"/files", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al conectar con FacturaAPI: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"success": false, "error": fmt.Sprintf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetElectronicDocuments retrieves all electronic documents for the company with optional filters
func GetElectronicDocuments(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}

	query := config.DB.Where("electronic_documents.company_id = ?", companyID)

	// Filter by document type
	tipo := c.Query("tipo_documento")
	if tipo != "" {
		query = query.Where("electronic_documents.tipo_documento = ?", tipo)
	}

	// Filter by state
	estado := c.Query("estado")
	if estado != "" {
		query = query.Where("electronic_documents.estado = ?", estado)
	}

	// Search by series or number
	search := c.Query("search")
	if search != "" {
		query = query.Where("electronic_documents.serie ILIKE ? OR electronic_documents.numero ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Filter by date range
	fechaDesde := c.Query("fecha_desde")
	if fechaDesde != "" {
		query = query.Where("electronic_documents.created_at >= ?", fechaDesde)
	}
	fechaHasta := c.Query("fecha_hasta")
	if fechaHasta != "" {
		query = query.Where("electronic_documents.created_at <= ?", fechaHasta+" 23:59:59")
	}

	var docs []models.ElectronicDocument
	if err := query.Preload("Sale").Preload("Sale.Customer").Order("electronic_documents.created_at desc").Find(&docs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    docs,
		"total":   len(docs),
	})
}

// GetDocumentStats returns summary counts grouped by tipo_documento and estado
func GetDocumentStats(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}

	type StatRow struct {
		TipoDocumento string `json:"tipo_documento"`
		Estado        string `json:"estado"`
		Count         int64  `json:"count"`
	}

	var rows []StatRow
	config.DB.Model(&models.ElectronicDocument{}).
		Select("tipo_documento, estado, count(*) as count").
		Where("company_id = ?", companyID).
		Group("tipo_documento, estado").
		Scan(&rows)

	// Also get totals per type
	type TotalRow struct {
		TipoDocumento string `json:"tipo_documento"`
		Total         int64  `json:"total"`
	}
	var totals []TotalRow
	config.DB.Model(&models.ElectronicDocument{}).
		Select("tipo_documento, count(*) as total").
		Where("company_id = ?", companyID).
		Group("tipo_documento").
		Scan(&totals)

	var grandTotal int64
	config.DB.Model(&models.ElectronicDocument{}).
		Where("company_id = ?", companyID).
		Count(&grandTotal)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        rows,
		"totals":      totals,
		"grand_total": grandTotal,
	})
}

// GetBillingSeries returns all branches of the company with their series/correlative configuration
func GetBillingSeries(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}

	var branches []models.Branch
	if err := config.DB.Where("company_id = ?", companyID).Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    branches,
	})
}

// UpdateBillingSeries updates series and correlative settings for a specific branch
func UpdateBillingSeries(c *gin.Context) {
	companyIDRaw, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Sesión inválida"})
		return
	}
	companyID, ok := companyIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Formato de ID de empresa inválido"})
		return
	}

	branchIDStr := c.Param("branchId")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de sucursal inválido"})
		return
	}

	var branch models.Branch
	if err := config.DB.Where("id = ? AND company_id = ?", branchID, companyID).First(&branch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Sucursal no encontrada"})
		return
	}

	var input struct {
		SerieFactura       string `json:"serie_factura"`
		SerieBoleta        string `json:"serie_boleta"`
		CorrelativoFactura int    `json:"correlativo_factura"`
		CorrelativoBoleta  int    `json:"correlativo_boleta"`
		CorrelativoPadding int    `json:"correlativo_padding"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	branch.SerieFactura = input.SerieFactura
	branch.SerieBoleta = input.SerieBoleta
	branch.CorrelativoFactura = input.CorrelativoFactura
	branch.CorrelativoBoleta = input.CorrelativoBoleta
	branch.CorrelativoPadding = input.CorrelativoPadding

	if err := config.DB.Save(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al guardar la configuración"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Series y correlativos actualizados correctamente",
		"data":    branch,
	})
}

// ResetBillingSeries allows resetting correlatives to 1 only if no documents have been emitted
func ResetBillingSeries(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de sucursal inválido"})
		return
	}

	var input struct {
		Type string `json:"type"` // "factura" or "boleta"
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var branch models.Branch
	if err := config.DB.Where("id = ? AND company_id = ?", branchID, companyID).First(&branch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Sucursal no encontrada"})
		return
	}

	// Safety check: check if any electronic documents exist for this branch and type
	var count int64
	docType := "01"
	if input.Type == "boleta" {
		docType = "03"
	}

	config.DB.Model(&models.ElectronicDocument{}).
		Joins("JOIN sales ON electronic_documents.sale_id = sales.id").
		Where("sales.branch_id = ? AND electronic_documents.tipo_documento = ?", branchID, docType).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   fmt.Sprintf("No se puede reiniciar el contador porque ya existen %d comprobantes emitidos de este tipo.", count),
		})
		return
	}

	if input.Type == "factura" {
		branch.CorrelativoFactura = 1
	} else {
		branch.CorrelativoBoleta = 1
	}

	if err := config.DB.Save(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al reiniciar correlativo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contador reiniciado con éxito",
		"data":    branch,
	})
}

// ResendElectronicDocument attempts to re-emit an electronic document
func ResendElectronicDocument(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido"})
		return
	}

	var doc models.ElectronicDocument
	if err := config.DB.Preload("Sale").Where("id = ? AND company_id = ?", docID, companyID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Comprobante no encontrado"})
		return
	}

	if doc.Estado == "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El comprobante ya fue aceptado por SUNAT"})
		return
	}

	if doc.Sale == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Venta asociada no encontrada"})
		return
	}

	// Fetch Sale Items
	var items []models.SaleItem
	if err := config.DB.Where("sale_id = ?", doc.Sale.ID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al recuperar items de la venta"})
		return
	}

	// Emit again (Service now handles updating existing DB records by sale_id)
	billingService := services.NewBillingService()
	_, err = billingService.EmitSale(*doc.Sale, items)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "Error al re-emitir: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comprobante re-enviado exitosamente",
	})
}

// SyncElectronicDocumentStatus queries FacturaAPI for the latest status of a document
func SyncElectronicDocumentStatus(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido"})
		return
	}

	var doc models.ElectronicDocument
	if err := config.DB.Where("id = ? AND company_id = ?", docID, companyID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Comprobante no encontrado"})
		return
	}

	if doc.DocumentUUID == nil || *doc.DocumentUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El comprobante no tiene un identificador de la API"})
		return
	}

	billingService := services.NewBillingService()
	_, err = billingService.SyncDocumentStatus(companyID, &doc)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Estado sincronizado correctamente",
		"data":    doc,
	})
}

// BatchEmitDrafts processes all 'draft' documents for a company
func BatchEmitDrafts(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	// 1. Get all draft documents
	var drafts []models.ElectronicDocument
	if err := config.DB.Where("company_id = ? AND estado = 'draft'", companyID).Find(&drafts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al recuperar borradores"})
		return
	}

	if len(drafts) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "No hay documentos pendientes de emitir"})
		return
	}

	billingService := services.NewBillingService()
	successCount := 0
	errorCount := 0

	for _, doc := range drafts {
		if doc.SaleID == nil {
			continue
		}

		// Load sale and items
		var sale models.Sale
		if err := config.DB.First(&sale, *doc.SaleID).Error; err != nil {
			errorCount++
			continue
		}

		var items []models.SaleItem
		if err := config.DB.Where("sale_id = ?", sale.ID).Find(&items).Error; err != nil {
			errorCount++
			continue
		}

		// Emit
		_, err := billingService.EmitSale(sale, items)
		if err != nil {
			errorCount++
			log.Printf("[BatchEmit] Error emitting %s-%s: %v", doc.Serie, doc.Numero, err)
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Proceso terminado. Éxitos: %d, Errores: %d", successCount, errorCount),
		"data": gin.H{
			"success": successCount,
			"errors":  errorCount,
		},
	})
}

// UpdateElectronicDocumentDraft updates customer data of a draft document
func UpdateElectronicDocumentDraft(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido"})
		return
	}

	var input struct {
		RazonSocial            string `json:"razon_social" binding:"required"`
		Direccion              string `json:"direccion"`
		TipoDocumentoIdentidad string `json:"tipo_documento_identidad"`
		NumeroDocumento        string `json:"numero_documento"`
		Serie                  string `json:"serie"`
		Numero                 string `json:"numero"`
		FechaEmision           string `json:"fecha_emision"`
		Items                  []struct {
			ID             uuid.UUID `json:"id"`
			Cantidad       float64   `json:"cantidad"`
			PrecioUnitario float64   `json:"precio_unitario"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var doc models.ElectronicDocument
	if err := config.DB.Where("id = ? AND company_id = ?", docID, companyID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Comprobante no encontrado"})
		return
	}

	if doc.Estado != "draft" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Solo se pueden editar documentos en estado borrador"})
		return
	}

	if doc.SaleID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El documento no tiene una venta asociada"})
		return
	}

	tx := config.DB.Begin()

	// 1. Update Customer info (Global record update)
	var sale models.Sale
	if err := tx.First(&sale, *doc.SaleID).Error; err == nil {
		var customer models.Customer
		if err := tx.First(&customer, sale.CustomerID).Error; err == nil {
			customer.Nombre = input.RazonSocial
			customer.Direccion = input.Direccion
			if input.TipoDocumentoIdentidad != "" {
				customer.TipoDocumento = input.TipoDocumentoIdentidad
			}
			if input.NumeroDocumento != "" {
				customer.NumeroDocumento = input.NumeroDocumento
			}
			tx.Save(&customer)
		}

		// 2. Update Items and Recalculate Totals
		if len(input.Items) > 0 {
			var newTotalVenta, newTotalSubtotal, newTotalIGV float64

			for _, itemUpdate := range input.Items {
				var si models.SaleItem
				if err := tx.Where("id = ? AND sale_id = ?", itemUpdate.ID, sale.ID).First(&si).Error; err == nil {
					// Apply new values
					si.Cantidad = itemUpdate.Cantidad
					si.PrecioUnitario = itemUpdate.PrecioUnitario
					
					// Recalculate line totals (Assuming 18% IGV - standard for this app)
					lineTotal := si.Cantidad * si.PrecioUnitario
					si.Subtotal = lineTotal / 1.18
					si.IGV = lineTotal - si.Subtotal
					
					tx.Save(&si)

					newTotalVenta += lineTotal
					newTotalSubtotal += si.Subtotal
					newTotalIGV += si.IGV
				}
			}

			// Update Sale header with new totals
			sale.Total = newTotalVenta
			sale.Subtotal = newTotalSubtotal
			sale.IGV = newTotalIGV
		}

		// 3. Update Sale header metadata
		if input.Serie != "" { sale.Serie = input.Serie }
		if input.Numero != "" { 
			// Ensure it has 8 digits
			numInt, err := strconv.Atoi(input.Numero)
			if err == nil {
				sale.Numero = fmt.Sprintf("%08d", numInt)
			} else {
				sale.Numero = input.Numero 
			}
		}
		if input.FechaEmision != "" {
			t, err := time.Parse("2006-01-02", input.FechaEmision)
			if err == nil {
				sale.CreatedAt = t
			}
		}
		tx.Save(&sale)
	}

	// 4. Update Document sync info
	if input.Serie != "" { doc.Serie = input.Serie }
	if input.Numero != "" { 
		numInt, err := strconv.Atoi(input.Numero)
		if err == nil {
			doc.Numero = fmt.Sprintf("%08d", numInt)
		} else {
			doc.Numero = input.Numero
		}
	}
	
	doc.Observaciones = "Editado manualmente (Detalles y Totales). " + doc.Observaciones
	tx.Save(&doc)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comprobante integral actualizado correctamente",
		"data":    doc,
	})
}

// DeleteElectronicDocumentTest removes a document from the system in TEST mode only
func DeleteElectronicDocumentTest(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido"})
		return
	}

	var doc models.ElectronicDocument
	if err := config.DB.Where("id = ? AND company_id = ?", docID, companyID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Comprobante no encontrado"})
		return
	}

	// 1. Check if we are in DEV mode
	var billingConfig models.BillingConfig
	config.DB.Where("company_id = ?", companyID).First(&billingConfig)

	if billingConfig.Modo != "dev" && billingConfig.Modo != "beta" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Esta acción solo está permitida en modo Desarrollo/Beta"})
		return
	}

	// 2. Try to delete from FacturaAPI if UUID exists
	if doc.DocumentUUID != nil && *doc.DocumentUUID != "" {
		billingService := services.NewBillingService()
		_ = billingService.DeleteDocumentBeta(companyID, *doc.DocumentUUID) // Best effort
	}

	// 3. Delete from local DB
	if err := config.DB.Delete(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al eliminar registro: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comprobante eliminado del sistema (Modo Prueba)",
	})
}

// VoidElectronicDocument initiates the legal voiding process (Baja) for Facturas/GRE
func VoidElectronicDocument(c *gin.Context) {
	companyIDRaw, _ := c.Get("companyID")
	companyID := companyIDRaw.(uuid.UUID)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "ID de documento inválido"})
		return
	}

	var input struct {
		Motivo string `json:"motivo" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El motivo de anulación es obligatorio (mín. 10 caracteres)"})
		return
	}

	if len(input.Motivo) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El motivo de anulación debe tener al menos 10 caracteres"})
		return
	}

	var doc models.ElectronicDocument
	if err := config.DB.Where("id = ? AND company_id = ?", docID, companyID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Comprobante no encontrado"})
		return
	}

	if doc.TipoDocumento == "03" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Las Boletas de Venta no se pueden anular por baja. Debe emitir una Nota de Crédito."})
		return
	}

	if doc.Estado == "voided" || doc.Estado == "void_pending" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El comprobante ya está en proceso de anulación o ya fue anulado"})
		return
	}

	if doc.DocumentUUID == nil || *doc.DocumentUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "El comprobante no tiene un identificador de la API para anular"})
		return
	}

	billingService := services.NewBillingService()
	err = billingService.VoidDocument(companyID, *doc.DocumentUUID, input.Motivo)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "Error al anular en FacturaAPI: " + err.Error()})
		return
	}

	// Update local state to void_pending
	doc.Estado = "void_pending"
	doc.Observaciones = "Anulación solicitada: " + input.Motivo
	config.DB.Save(&doc)

	c.JSON(http.StatusOK, gin.H{
	        "success": true,
	        "message": "Solicitud de anulación (Baja) enviada correctamente",
	        "data":    doc,
	})
	}

	// HandleFacturaAPIWebhook receives POST notifications from FacturaAPI when a document changes status
	func HandleFacturaAPIWebhook(c *gin.Context) {
	var payload struct {
	DocumentoID      string `json:"documento_id"`
	Status           string `json:"status"`
	SunatResponse    string `json:"sunat_response"`
	SunatDescription string `json:"sunat_description"`
	SunatNotes       string `json:"sunat_notes"`
	Files            struct {
	XML string `json:"xml"`
	CDR string `json:"cdr"`
	} `json:"files"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
	log.Printf("[Webhook] Error binding JSON: %v", err)
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
	return
	}

	if payload.DocumentoID == "" {
	c.JSON(http.StatusBadRequest, gin.H{"error": "documento_id is required"})
	return
	}

	log.Printf("[Webhook] Received update for document %s: New Status = %s", payload.DocumentoID, payload.Status)

	// 1. Find document by UUID
	var doc models.ElectronicDocument
	if err := config.DB.Where("document_uuid = ?", payload.DocumentoID).First(&doc).Error; err != nil {
	log.Printf("[Webhook] Document %s not found in local DB", payload.DocumentoID)
	c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
	return
	}

	// 2. Update Status
	doc.Estado = payload.Status

	// 3. Update URLs
	if payload.Files.XML != "" {
	doc.XmlURL = payload.Files.XML
	}
	if payload.Files.CDR != "" {
	doc.CdrURL = payload.Files.CDR
	}

	// 4. Process messages
	sunatMsg := payload.SunatResponse
	if sunatMsg == "" {
	sunatMsg = payload.SunatDescription
	}
	if payload.SunatNotes != "" {
	if sunatMsg != "" {
	sunatMsg += "\n" + payload.SunatNotes
	} else {
	sunatMsg = payload.SunatNotes
	}
	}

	billingService := services.NewBillingService()
	doc.SunatResponse = billingService.ParseSunatResponse(sunatMsg)

	// 5. If we have a CDR, try to extract notes
	if doc.CdrURL != "" {
		cdrNotes, err := billingService.ExtractNotesFromCDR(doc.CdrURL)
		if err == nil && cdrNotes != "" {
			if doc.SunatResponse != "" {
				if !strings.Contains(doc.SunatResponse, cdrNotes) {
					doc.SunatResponse += "\n" + cdrNotes
				}
			} else {
				doc.SunatResponse = cdrNotes
			}
		}
	}

	// 6. Smart Status Detection (Fallback for inconsistent API status)
	msgLower := strings.ToLower(doc.SunatResponse)
	isAcceptedMsg := strings.Contains(msgLower, "aceptada") || 
					 strings.Contains(msgLower, "aceptado") || 
					 strings.Contains(msgLower, "exito") || 
					 strings.Contains(msgLower, "éxito") ||
					 doc.CdrURL != ""

	isRejectedMsg := strings.Contains(msgLower, "rechazada") || 
					 strings.Contains(msgLower, "rechazado") || 
					 strings.Contains(msgLower, "error de datos")

	if doc.Estado == "pending" || doc.Estado == "error" || doc.Estado == "" {
		if isAcceptedMsg {
			log.Printf("[Webhook] Smart Detection: Upgrading %s-%s from '%s' to accepted", 
				doc.Serie, doc.Numero, doc.Estado)
			doc.Estado = "accepted"
		} else if isRejectedMsg {
			log.Printf("[Webhook] Smart Detection: Upgrading %s-%s from '%s' to rejected", 
				doc.Serie, doc.Numero, doc.Estado)
			doc.Estado = "rejected"
		}
	}

	// 6.5 Intelligent Observation Detection
	hasObsKeywords := strings.Contains(msgLower, "observaci") || 
					  strings.Contains(msgLower, "advertencia") || 
					  strings.Contains(msgLower, "(")

	if doc.Estado == "accepted" {
		if hasObsKeywords {
			doc.Observaciones = doc.SunatResponse
		} else {
			doc.Observaciones = "" // Clear generic success messages
		}
	}

	// 7. Ensure PDF URL is set
	if doc.PdfURL == "" && doc.DocumentUUID != nil {
		// Resolve API URL for PDF generation
		var billingConfig models.BillingConfig
		if err := config.DB.Where("company_id = ?", doc.CompanyID).First(&billingConfig).Error; err == nil {
			apiURL := billingConfig.ApiURL
			if apiURL == "" {
				var globalURLSetting models.CompanySetting
				if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting).Error; err == nil {
					apiURL = globalURLSetting.Valor
				}
			}
			if apiURL != "" {
				apiURL = strings.TrimSuffix(apiURL, "/")
				apiURL = strings.TrimSuffix(apiURL, "/api/v1")
				doc.PdfURL = apiURL + "/api/v1/download/" + *doc.DocumentUUID + "/pdf"
			}
		}
	}

	// 8. Save with explicit Updates to avoid missing fields
	updateData := map[string]interface{}{
		"estado":          doc.Estado,
		"sunat_response":  doc.SunatResponse,
		"observaciones":   doc.Observaciones,
		"xml_url":         doc.XmlURL,
		"cdr_url":         doc.CdrURL,
		"pdf_url":         doc.PdfURL,
		"facturacion_error": "",
	}
	if err := config.DB.Model(&doc).Updates(updateData).Error; err != nil {
		log.Printf("[Webhook] Error saving document status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
	}

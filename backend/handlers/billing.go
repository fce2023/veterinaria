package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"veterinaria/backend/config"
	"veterinaria/backend/models"

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
	if len(apiURL) > 0 && apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL[:len(apiURL)-1]
	}

	var logs []string
	uuidStr := input.TenantUUID

	// Call external FacturaAPI only if API URL & Key are present
	if apiURL != "" && apiKey != "" {
		client := &http.Client{Timeout: 20 * time.Second}

		// Step A: Register / Search Company
		if uuidStr == "" {
			logs = append(logs, fmt.Sprintf("Buscando empresa con RUC %s en FacturaAPI...", company.RUC))
			
			reqSearch, err := http.NewRequest("GET", apiURL+"/api/v1/companies/search?ruc="+url.QueryEscape(company.RUC), nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al preparar búsqueda de empresa: " + err.Error()})
				return
			}
			reqSearch.Header.Set("Authorization", "Bearer "+apiKey)
			reqSearch.Header.Set("Accept", "application/json")

			respSearch, err := client.Do(reqSearch)
			if err == nil {
				defer respSearch.Body.Close()
				bodyBytes, _ := io.ReadAll(respSearch.Body)
				
				if respSearch.StatusCode == http.StatusOK {
					var result struct {
						UUID string `json:"uuid"`
					}
					if err := json.Unmarshal(bodyBytes, &result); err == nil && result.UUID != "" {
						uuidStr = result.UUID
						logs = append(logs, "Empresa encontrada en FacturaAPI con UUID: "+uuidStr)
					}
				}
			}

			// If still empty, register it
			if uuidStr == "" {
				logs = append(logs, "Empresa no encontrada. Registrando en FacturaAPI...")
				
				regPayload := map[string]string{
					"ruc":          company.RUC,
					"razon_social": company.RazonSocial,
					"direccion":    company.Direccion,
					"modo":         input.Modo,
				}
				jsonPayload, _ := json.Marshal(regPayload)
				
				reqReg, err := http.NewRequest("POST", apiURL+"/api/v1/companies/register", bytes.NewBuffer(jsonPayload))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al preparar registro: " + err.Error()})
					return
				}
				reqReg.Header.Set("Authorization", "Bearer "+apiKey)
				reqReg.Header.Set("Content-Type", "application/json")
				reqReg.Header.Set("Accept", "application/json")

				respReg, err := client.Do(reqReg)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al conectar con FacturaAPI para registro: " + err.Error()})
					return
				}
				defer respReg.Body.Close()
				bodyBytes, _ := io.ReadAll(respReg.Body)

				if respReg.StatusCode != http.StatusCreated && respReg.StatusCode != http.StatusOK {
					c.JSON(http.StatusBadGateway, gin.H{
						"success":   false,
						"logs":      logs,
						"api_error": fmt.Sprintf("Error registro FacturaAPI (%d): %s", respReg.StatusCode, string(bodyBytes)),
					})
					return
				}

				var result struct {
					UUID string `json:"uuid"`
				}
				if err := json.Unmarshal(bodyBytes, &result); err == nil && result.UUID != "" {
					uuidStr = result.UUID
					logs = append(logs, "Empresa registrada exitosamente en FacturaAPI con UUID: "+uuidStr)
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo obtener el UUID tras registro"})
					return
				}
			}
		}
		
		// Update DB BillingConfig struct to hold final TenantUUID
		billingConfig.TenantUUID = uuidStr

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

			reqCred, err := http.NewRequest("PATCH", apiURL+"/api/v1/config/"+uuidStr+"/sunat-credentials", bytes.NewBuffer(jsonCreds))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al preparar PATCH credenciales: " + err.Error()})
				return
			}
			reqCred.Header.Set("Authorization", "Bearer "+apiKey)
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
				c.JSON(http.StatusBadGateway, gin.H{
					"success":   false,
					"logs":      logs,
					"api_error": fmt.Sprintf("Error credenciales FacturaAPI (%d): %s", respCred.StatusCode, string(bodyCredBytes)),
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
					c.JSON(http.StatusBadGateway, gin.H{
						"success":   false,
						"logs":      logs,
						"api_error": fmt.Sprintf("Error certificado FacturaAPI (%d): %s", respCert.StatusCode, string(bodyCertBytes)),
					})
					return
				}
				logs = append(logs, "Certificado digital subido y firmado correctamente.")
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

	if len(apiURL) > 0 && apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL[:len(apiURL)-1]
	}

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

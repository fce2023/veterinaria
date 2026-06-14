# FacturaAPI Credentials Synchronization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to save their SUNAT SOL credentials, API GRE credentials, and digital certificates in the local ERP database and synchronize them automatically with FacturaAPI.

**Architecture:** Extend the `/billing/config` POST endpoint handler in Go to store new fields in the DB, query/onboard the company in FacturaAPI if unregistered, PATCH SUNAT/GRE credentials, and upload certificates to FacturaAPI. Update the Vue frontend to ensure all values are bound and sent.

**Tech Stack:** Go (Gin, GORM), TypeScript (Vue 3, Axios, Pinia).

---

### Task 1: Bind and Save Credentials in Backend DB

**Files:**
- Modify: `backend/handlers/billing.go`

- [ ] **Step 1: Update the input binding struct in `SaveBillingConfig` to capture all the new credential fields.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go):
```go
// Replace the var input struct inside SaveBillingConfig with:
var input struct {
	ApiURL            string `json:"api_url"`
	ApiKey            string `json:"api_key"`
	TenantUUID        string `json:"tenant_uuid"`
	Modo              string `json:"modo"`
	Estado            string `json:"estado"`
	SolUser           string `json:"sol_user"`
	SolPass           string `json:"sol_pass"`
	CertificadoBase64 string `json:"certificado_base64"`
	ClientID          string `json:"client_id"`
	ClientSecret      string `json:"client_secret"`
}
```

- [ ] **Step 2: Assign input values to `billingConfig` fields so they are persisted locally.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go):
```go
// In SaveBillingConfig, map the input fields to the DB config record:
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
```

- [ ] **Step 3: Run the backend compilation to verify that changes compile correctly.**

Run: `go build -o /dev/null ./...` (from `backend` directory)
Expected: Compiles with no errors.

---

### Task 2: Implement Onboarding check (Search or Register) in FacturaAPI

**Files:**
- Modify: `backend/handlers/billing.go`

- [ ] **Step 1: Implement the helper logic to resolve API URL/Key fallbacks and search/register the tenant in FacturaAPI.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go) inside `SaveBillingConfig` before DB transaction or DB write:
```go
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
	var apiError string
	uuidStr := input.TenantUUID

	// Call external FacturaAPI only if API URL & Key are present
	if apiURL != "" && apiKey != "" {
		client := &http.Client{Timeout: 20 * time.Second}

		// Step A: Register / Search Company
		if uuidStr == "" {
			logs = append(logs, fmt.Sprintf("Buscando empresa con RUC %s en FacturaAPI...", company.RUC))
			
			reqSearch, err := http.NewRequest("GET", apiURL+"/api/v1/companies/search?ruc="+company.RUC, nil)
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
						"success":   true,
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
	}
```

- [ ] **Step 2: Add imports (`time`, `bytes`, `encoding/json`, `io`, `fmt`) if missing.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go):
```go
// Verify/Add imports:
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"veterinaria/backend/config"
	"veterinaria/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
```

- [ ] **Step 3: Verify the Go compilation.**

Run: `go build -o /dev/null ./...` (from `backend` directory)
Expected: Compiles with no errors.

---

### Task 3: Implement PATCH Credentials and Certificate Upload in Backend

**Files:**
- Modify: `backend/handlers/billing.go`

- [ ] **Step 1: Implement steps to call PATCH credentials and POST certificate inside `SaveBillingConfig` after onboarding.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go):
```go
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
				c.JSON(http.StatusOK, gin.H{
					"success":   true,
					"logs":      logs,
					"api_error": fmt.Sprintf("Error credenciales FacturaAPI (%d): %s", respCred.StatusCode, string(bodyCredBytes)),
				})
				return
			}
			logs = append(logs, "Credenciales SUNAT/API GRE sincronizadas correctamente.")

			// Step C: POST Certificate
			if input.CertificadoBase64 != "" {
				logs = append(logs, "Subiendo certificado digital .p12...")
				certPayload := map[string]string{
					"certificate_base64": input.CertificadoBase64,
					"password":           input.SolPass,
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
					c.JSON(http.StatusOK, gin.H{
						"success":   true,
						"logs":      logs,
						"api_error": fmt.Sprintf("Error certificado FacturaAPI (%d): %s", respCert.StatusCode, string(bodyCertBytes)),
					})
					return
				}
				logs = append(logs, "Certificado digital subido y firmado correctamente.")
			}
		}
```

- [ ] **Step 2: Update the response block of `SaveBillingConfig` to return logs.**

Modify [backend/handlers/billing.go](file:///home/oyon/sehuacho/veterinaria/backend/handlers/billing.go):
```go
	// Replace the final response inside SaveBillingConfig with:
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuración guardada y sincronizada exitosamente",
		"logs":    logs,
		"data":    billingConfig,
	})
```

- [ ] **Step 3: Run backend compile verification.**

Run: `go build -o /dev/null ./...` (from `backend` directory)
Expected: Compiles successfully.

---

### Task 4: Frontend Data Binding and Verification

**Files:**
- Modify: `frontend/src/components/BillingSection.vue`

- [ ] **Step 1: Verify all key credentials fields are bound and sent correctly in `saveConfig()`.**

Verify/Modify [frontend/src/components/BillingSection.vue](file:///home/oyon/sehuacho/veterinaria/frontend/src/components/BillingSection.vue#L619-L637):
```typescript
async function saveConfig() {
  saving.value = true
  clearLogs()
  try {
    const res = await axios.post('/billing/config', config)
    if (res.data.success) {
      if (res.data.logs) saveLogs.value = res.data.logs
      if (res.data.api_error) {
        apiError.value = res.data.api_error
      }
      await loadConfig()
    }
  } catch (err: any) {
    apiError.value = err.response?.data?.error || 'Error de red o de servidor al intentar guardar.'
    saveLogs.value = ['Error al conectar con el servidor.']
  } finally {
    saving.value = false
  }
}
```

- [ ] **Step 2: Run frontend production check to ensure build compiles.**

Run: `npm run build` (from `frontend` directory)
Expected: Builds Vue assets successfully.

package services

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

var (
	reNote = regexp.MustCompile(`(?s)<(?:[\w-]*:)?Note[^>]*>(.*?)</(?:[\w-]*:)?Note>`)
	reDesc = regexp.MustCompile(`(?s)<(?:[\w-]*:)?Description[^>]*>(.*?)</(?:[\w-]*:)?Description>`)
)

// FacturaAPI Payload Structures
type BillingHeader struct {
	TipoDocumento string `json:"tipo_documento"`
	Serie         string `json:"serie"`
	Numero        int    `json:"numero,omitempty"`
	FechaEmision  string `json:"fecha_emision"`
}

type BillingCustomer struct {
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
	RazonSocial     string `json:"razon_social"`
	Direccion       string `json:"direccion,omitempty"`
}

type BillingIssuer struct {
	Ruc            string `json:"ruc"`
	RazonSocial    string `json:"razon_social"`
	NombreComercial string `json:"nombre_comercial,omitempty"`
	Direccion      string `json:"direccion"`
	Ubigeo         string `json:"ubigeo"`
	Departamento   string `json:"departamento"`
	Provincia      string `json:"provincia"`
	Distrito       string `json:"distrito"`
}

type BillingItem struct {
	Descripcion    string  `json:"descripcion"`
	Cantidad       float64 `json:"cantidad"`
	Unidad         string  `json:"unidad"`          // Catálogo 03
	PrecioUnitario float64 `json:"precio_unitario"` // Con IGV
	Subtotal       float64 `json:"subtotal"`        // Sin IGV
	TipoAfectacion string  `json:"tipo_afectacion"` // Catálogo 07
}

type BillingDetraccion struct {
	CodigoBienServicio string  `json:"codigo_bien_servicio"`
	CodigoMedioPago    string  `json:"codigo_medio_pago"`
	Porcentaje         float64 `json:"porcentaje"`
	Monto              float64 `json:"monto"`
}

type BillingTotals struct {
	TotalVenta     float64 `json:"total_venta"`
	TotalImpuestos float64 `json:"total_impuestos"`
}

type EmitPayload struct {
	EmpresaID  string             `json:"empresa_id"`
	Async      bool               `json:"async"`
	Modo       string             `json:"modo"`
	Header     BillingHeader      `json:"header"`
	Emisor     *BillingIssuer     `json:"emisor,omitempty"`
	Company    *BillingIssuer     `json:"company,omitempty"` // Alternative for compatibility
	Seller     *BillingIssuer     `json:"seller,omitempty"`  // Alternative for compatibility
	Cliente    BillingCustomer    `json:"cliente"`
	Items      []BillingItem      `json:"items"`
	Totales    BillingTotals      `json:"totales"`
	Detraccion *BillingDetraccion `json:"detraccion,omitempty"`
}

type EmitResponse struct {
	Message          string `json:"message"`
	DocumentoID      string `json:"documento_id"`
	Status           string `json:"status"`
	SunatDescription string `json:"sunat_description,omitempty"`
	SunatNotes       string `json:"sunat_notes,omitempty"`
	// Synchronous response fields
	SunatResponse string `json:"sunat_response,omitempty"`
	Files         struct {
		XML string `json:"xml"`
		CDR string `json:"cdr"`
	} `json:"files,omitempty"`
}

// BillingService handles communication with FacturaAPI
type BillingService struct{}

func NewBillingService() *BillingService {
	return &BillingService{}
}

// EnsureCompanyExists verifies that the company exists in FacturaAPI and registers it if not.
// It returns the verified TenantUUID and an error if any.
func (s *BillingService) EnsureCompanyExists(companyID uuid.UUID) (string, error) {
	var company models.Company
	if err := config.DB.First(&company, companyID).Error; err != nil {
		return "", fmt.Errorf("company not found: %w", err)
	}

	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return "", fmt.Errorf("billing configuration not found: %w", err)
	}

	// Resolve API credentials
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

	if apiURL == "" || apiKey == "" {
		return "", fmt.Errorf("FacturaAPI configuration missing")
	}

	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	client := &http.Client{Timeout: 15 * time.Second}
	uuidStr := billingConfig.TenantUUID

	// 1. Verify existing UUID if present
	if uuidStr != "" {
		req, _ := http.NewRequest("GET", apiURL+"/api/v1/config/"+uuidStr, nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				log.Printf("[Billing] UUID %s not found in FacturaAPI for company %s. Will search/register.", uuidStr, company.RUC)
				uuidStr = ""
			}
		} else {
			log.Printf("[Billing] Warning verifying UUID: %v", err)
		}
	}

	// 2. If no UUID (not provided or not found), search by RUC
	if uuidStr == "" {
		log.Printf("[Billing] Searching company by RUC %s in FacturaAPI", company.RUC)
		// Fix URL formatting: use url.Values for safety
		searchURL := fmt.Sprintf("%s/api/v1/companies/search?ruc=%s", apiURL, company.RUC)
		req, _ := http.NewRequest("GET", searchURL, nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var result struct {
					UUID string `json:"uuid"`
				}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &result); err == nil && result.UUID != "" {
					uuidStr = result.UUID
					log.Printf("[Billing] Found existing company in FacturaAPI: %s", uuidStr)
				}
			}
		}
	}

	// 3. Still no UUID? Register it!
	if uuidStr == "" {
		log.Printf("[Billing] Registering company %s in FacturaAPI", company.RUC)
		apiMode := billingConfig.Modo
		if apiMode == "dev" {
			apiMode = "beta"
		} else if apiMode == "prod" {
			apiMode = "produccion"
		}

		regPayload := map[string]string{
			"ruc":              company.RUC,
			"razon_social":     company.RazonSocial,
			"nombre_comercial": company.NombreComercial,
			"direccion":        company.Direccion,
			"ubigeo":           company.Ubigeo,
			"departamento":     company.Departamento,
			"provincia":        company.Provincia,
			"distrito":         company.Distrito,
			"modo":             apiMode,
		}
		jsonPayload, _ := json.Marshal(regPayload)

		req, _ := http.NewRequest("POST", apiURL+"/api/v1/companies/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("error connecting to FacturaAPI for registration: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			var result struct {
				UUID string `json:"uuid"`
			}
			if err := json.Unmarshal(body, &result); err == nil && result.UUID != "" {
				uuidStr = result.UUID
				log.Printf("[Billing] Registered company in FacturaAPI: %s", uuidStr)
			}
		} else {
			return "", fmt.Errorf("FacturaAPI registration failed (%d): %s", resp.StatusCode, string(body))
		}
	}

	// 4. Update local DB if changed
	if uuidStr != billingConfig.TenantUUID {
		log.Printf("[Billing] Updating local TenantUUID from %s to %s", billingConfig.TenantUUID, uuidStr)
		if err := config.DB.Model(&billingConfig).Update("tenant_uuid", uuidStr).Error; err != nil {
			log.Printf("[Billing] Error updating local TenantUUID: %v", err)
		}
	}

	return uuidStr, nil
}

// EmitSale sends a sale to FacturaAPI for electronic billing
func (s *BillingService) EmitSale(sale models.Sale, items []models.SaleItem) (*EmitResponse, error) {
	if sale.TipoDocumento == "NV" || sale.TipoDocumento == "00" || sale.TipoDocumento == "99" {
		return nil, nil // Nota de venta is an internal ticket, no SUNAT emission
	}

	// 1. Ensure company exists in FacturaAPI before emitting
	tenantUUID, err := s.EnsureCompanyExists(sale.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify/register company in FacturaAPI: %w", err)
	}

	// Get updated Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", sale.CompanyID).First(&billingConfig).Error; err != nil {
		return nil, fmt.Errorf("billing configuration not found for company %s", sale.CompanyID)
	}

	if billingConfig.Estado != "active" {
		return nil, fmt.Errorf("billing service is inactive for this company")
	}

	// Fetch global settings fallback if necessary
	apiURL := billingConfig.ApiURL
	apiKey := billingConfig.ApiKey

	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting).Error; err == nil {
			if apiURL == "" {
				apiURL = globalURLSetting.Valor
			}
		}
		if err := config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting).Error; err == nil {
			if apiKey == "" {
				apiKey = globalKeySetting.Valor
			}
		}
	}

	// Sanitize URL
	if apiURL != "" {
		apiURL = strings.TrimSuffix(apiURL, "/")
		apiURL = strings.TrimSuffix(apiURL, "/api/v1")
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("FacturaAPI global configuration missing (URL or Token not configured)")
	}

	// ... (rest remains same but use tenantUUID if needed, although billingConfig.TenantUUID is now updated)

	// 1.5 Fetch Full Company Info for Emisor payload
	var company models.Company
	if err := config.DB.First(&company, sale.CompanyID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve company info: %w", err)
	}

	log.Printf("[Billing] Emitting %s %s-%s for company %s. Fiscal Info: Ubigeo=%s, Dept=%s", 
		sale.TipoDocumento, sale.Serie, sale.Numero, company.RazonSocial, company.Ubigeo, company.Departamento)

	// 1.7 Fetch Branch for padding config
	var branch models.Branch
	padding := 8
	if err := config.DB.First(&branch, sale.BranchID).Error; err == nil {
		if branch.CorrelativoPadding >= 0 {
			padding = branch.CorrelativoPadding
		}
	}

	// 2. Prepare Payload
	var customer models.Customer
	if err := config.DB.First(&customer, sale.CustomerID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve billing customer: %w", err)
	}

	var billingItems []BillingItem
	hasServices := false
	for _, item := range items {
		var product models.Product
		if err := config.DB.Select("nombre, unidad_medida").First(&product, item.ProductID).Error; err != nil {
			return nil, fmt.Errorf("failed to retrieve product %s: %w", item.ProductID, err)
		}

		if product.UnidadMedida == "servicio" {
			hasServices = true
		}

		billingItems = append(billingItems, BillingItem{
			Descripcion:    product.Nombre,
			Cantidad:       item.Cantidad,
			Unidad:         item.UnidadSUNAT,
			PrecioUnitario: item.PrecioUnitario,
			Subtotal:       item.Subtotal,
			TipoAfectacion: item.TipoAfectacion,
		})
	}

	num, _ := strconv.Atoi(sale.Numero)

	// Map dev to beta and prod to produccion for FacturaAPI compatibility
	apiMode := billingConfig.Modo
	if apiMode == "dev" {
		apiMode = "beta"
	} else if apiMode == "prod" {
		apiMode = "produccion"
	}

	// Always true in production to handle queues, false in beta to wait for SUNAT sync response
	isAsync := apiMode == "produccion"

	log.Printf("[Billing] Emitting %s %s-%d for company %s. Env: %s, Async: %v", 
		sale.TipoDocumento, sale.Serie, num, company.RazonSocial, apiMode, isAsync)

	// Ensure we don't send hyphens as per provider instructions
	clean := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "-" {
			return ""
		}
		return s
	}

	payload := EmitPayload{
		EmpresaID: tenantUUID,
		Async:     isAsync,
		Modo:      apiMode,
		Header: BillingHeader{
			TipoDocumento: sale.TipoDocumento,
			Serie:         sale.Serie,
			Numero:        num,
			FechaEmision:  sale.CreatedAt.Format("2006-01-02"),
		},
		// Re-enabling Emisor block to "force" the data into the XML
		// and using the cleaning function to avoid hyphens.
		Emisor: &BillingIssuer{
			Ruc:             clean(company.RUC),
			RazonSocial:     clean(company.RazonSocial),
			NombreComercial: clean(company.NombreComercial),
			Direccion:       clean(company.Direccion),
			Ubigeo:          clean(company.Ubigeo),
			Departamento:    clean(company.Departamento),
			Provincia:       clean(company.Provincia),
			Distrito:        clean(company.Distrito),
		},
		Cliente: prepareBillingCustomer(customer),
		Items:   billingItems,
		Totales: BillingTotals{
			TotalVenta:     sale.Total,
			TotalImpuestos: sale.IGV,
		},
	}

	if sale.TipoDocumento == "01" && hasServices && sale.Total > 700 {
		payload.Detraccion = &BillingDetraccion{
			CodigoBienServicio: "022",
			CodigoMedioPago:    "001",
			Porcentaje:         12.00,
			Monto:              sale.Total * 0.12,
		}
	}

	// 3. Initialize/Get Electronic Document for Payload Tracking
	var electronicDoc models.ElectronicDocument
	dbErr := config.DB.Where("sale_id = ?", sale.ID).First(&electronicDoc).Error

	// IDEMPOTENCY CHECK: If it already has a DocumentUUID and is NOT in 'error' or 'draft', 
	// we should query the API instead of emitting again.
	if dbErr == nil && electronicDoc.DocumentUUID != nil && *electronicDoc.DocumentUUID != "" && 
		electronicDoc.Estado != "error" && electronicDoc.Estado != "draft" && electronicDoc.Estado != "rejected" {
		log.Printf("[Billing] Sale %s already has DocumentUUID %s. Syncing status instead of re-emitting.", sale.ID, *electronicDoc.DocumentUUID)
		return s.SyncDocumentStatus(sale.CompanyID, &electronicDoc)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}
	electronicDoc.CompanyID = sale.CompanyID
	electronicDoc.SaleID = &sale.ID
	electronicDoc.FacturacionPayload = string(jsonData)
	electronicDoc.TipoDocumento = payload.Header.TipoDocumento
	electronicDoc.Serie = payload.Header.Serie
	electronicDoc.Numero = fmt.Sprintf("%0*d", padding, payload.Header.Numero)
	electronicDoc.Estado = "pending"

	if dbErr != nil {
		electronicDoc.ID = uuid.New()
		config.DB.Create(&electronicDoc)
	} else {
		config.DB.Save(&electronicDoc)
	}

	// 4. Send HTTP Request
	req, err := http.NewRequest("POST", apiURL+"/api/v1/documents/emit", bytes.NewBuffer(jsonData))
	if err != nil {
		electronicDoc.Estado = "error"
		electronicDoc.FacturacionError = err.Error()
		config.DB.Save(&electronicDoc)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		electronicDoc.Estado = "error"
		electronicDoc.FacturacionError = err.Error()
		config.DB.Save(&electronicDoc)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 5. Handle Technical Errors (4xx/5xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		electronicDoc.Estado = "error"
		electronicDoc.FacturacionError = string(body)
		config.DB.Save(&electronicDoc)
		return nil, fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	var emitResp EmitResponse
	if err := json.Unmarshal(body, &emitResp); err != nil {
		electronicDoc.Estado = "error"
		electronicDoc.FacturacionError = "JSON Unmarshal Error: " + err.Error()
		config.DB.Save(&electronicDoc)
		return nil, err
	}

	// 6. Update Document with API Response
	electronicDoc.DocumentUUID = &emitResp.DocumentoID
	electronicDoc.Estado = emitResp.Status
	
	// Process SUNAT messages/notes
	sunatMsg := emitResp.SunatResponse
	if sunatMsg == "" {
		sunatMsg = emitResp.SunatDescription
	}
	if emitResp.SunatNotes != "" {
		if sunatMsg != "" {
			sunatMsg += "\n" + emitResp.SunatNotes
		} else {
			sunatMsg = emitResp.SunatNotes
		}
	}
	
	electronicDoc.SunatResponse = s.ParseSunatResponse(sunatMsg)
	electronicDoc.FacturacionError = "" // Clear any previous technical error if success
	
	// Base URLs for downloads
	electronicDoc.PdfURL = apiURL + "/api/v1/download/" + emitResp.DocumentoID + "/pdf"
	electronicDoc.XmlURL = emitResp.Files.XML
	electronicDoc.CdrURL = emitResp.Files.CDR

	// If the CDR URL exists, try to extract notes from it
	if electronicDoc.CdrURL != "" {
		cdrNotes, err := s.ExtractNotesFromCDR(electronicDoc.CdrURL)
		if err == nil && cdrNotes != "" {
			if electronicDoc.SunatResponse != "" {
				if !strings.Contains(electronicDoc.SunatResponse, cdrNotes) {
					electronicDoc.SunatResponse += "\n" + cdrNotes
				}
			} else {
				electronicDoc.SunatResponse = cdrNotes
			}
		}
	}

	// Identify Observations (Accepted but with notes)
	if electronicDoc.Estado == "accepted" && electronicDoc.SunatResponse != "" {
		electronicDoc.Observaciones = electronicDoc.SunatResponse
	}

	config.DB.Save(&electronicDoc)

	return &emitResp, nil
}

// ParseSunatResponse extracts human-readable notes from SUNAT XML/raw response
func (s *BillingService) ParseSunatResponse(raw string) string {
	if raw == "" {
		return ""
	}

	// Check for Note or Description tags (with or without namespaces)
	hasNote := strings.Contains(raw, "Note>")
	hasDesc := strings.Contains(raw, "Description>")

	if !hasNote && !hasDesc {
		return raw
	}

	var notes []string

	// 1. Extract <cbc:Note> or <Note>
	matchesNote := reNote.FindAllStringSubmatch(raw, -1)
	for _, match := range matchesNote {
		if len(match) > 1 {
			note := strings.TrimSpace(match[1])
			if note != "" && !contains(notes, note) {
				notes = append(notes, note)
			}
		}
	}

	// 2. Extract <cbc:Description> or <Description>
	matchesDesc := reDesc.FindAllStringSubmatch(raw, -1)
	for _, match := range matchesDesc {
		if len(match) > 1 {
			desc := strings.TrimSpace(match[1])
			if desc != "" && !contains(notes, desc) {
				notes = append(notes, desc)
			}
		}
	}

	if len(notes) == 0 {
		return raw
	}

	return strings.Join(notes, "\n")
}

// ExtractNotesFromCDR downloads a CDR zip, extracts the XML, and parses the notes
func (s *BillingService) ExtractNotesFromCDR(cdrURL string) (string, error) {
	if cdrURL == "" {
		return "", nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(cdrURL)
	if err != nil {
		return "", fmt.Errorf("error downloading CDR: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading CDR, status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		// Might not be a zip file (e.g. direct XML or error text)
		return s.ParseSunatResponse(string(bodyBytes)), nil
	}

	var notes string
	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlBytes, err := io.ReadAll(rc)
			rc.Close()
			if err == nil {
				// Parse the XML content
				parsedNotes := s.ParseSunatResponse(string(xmlBytes))
				if parsedNotes != "" {
					if notes != "" {
						notes += "\n" + parsedNotes
					} else {
						notes = parsedNotes
					}
				}
			}
		}
	}

	return notes, nil
}

// Helper to check if slice contains string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// SyncCompanyInfo updates the company profile in FacturaAPI
func (s *BillingService) SyncCompanyInfo(company models.Company) error {
	// 1. Ensure company exists in FacturaAPI before syncing info
	tenantUUID, err := s.EnsureCompanyExists(company.ID)
	if err != nil {
		return fmt.Errorf("failed to verify/register company in FacturaAPI: %w", err)
	}

	// 2. Get updated Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", company.ID).First(&billingConfig).Error; err != nil {
		log.Printf("[Billing] SyncCompanyInfo: Configuration not found for company %s", company.ID)
		return nil // Not configured, nothing to sync
	}

	if billingConfig.Estado != "active" {
		log.Printf("[Billing] SyncCompanyInfo: Service inactive for company %s", company.ID)
		return nil
	}

	// 3. Resolve API credentials
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

	if apiURL == "" || apiKey == "" {
		return fmt.Errorf("FacturaAPI configuration missing")
	}

	// Sanitize URL
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	// 4. Prepare Payload
	clean := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "-" {
			return ""
		}
		return s
	}

	apiMode := billingConfig.Modo
	if apiMode == "dev" {
		apiMode = "beta"
	} else if apiMode == "prod" {
		apiMode = "produccion"
	}

	payload := map[string]interface{}{
		"razon_social":     clean(company.RazonSocial),
		"nombre_comercial": clean(company.NombreComercial),
		"direccion":        clean(company.Direccion),
		"ubigeo":           clean(company.Ubigeo),
		"departamento":     clean(company.Departamento),
		"provincia":        clean(company.Provincia),
		"distrito":         clean(company.Distrito),
		"modo":             apiMode,
	}

	if billingConfig.CorrelativoPadding != nil {
		payload["correlativo_padding"] = *billingConfig.CorrelativoPadding
	}
	if billingConfig.WebhookURL != "" {
		payload["webhook_url"] = billingConfig.WebhookURL
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("[Billing] Syncing Profile to FacturaAPI (%s) [Mode: %s] at %s: %s", 
		tenantUUID, apiMode, apiURL+"/api/v1/config/"+tenantUUID, string(jsonPayload))

	// The correct route for updating company profile in FacturaAPI is /api/v1/config/{uuid}
	req, err := http.NewRequest("PATCH", apiURL+"/api/v1/config/"+tenantUUID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Billing] Connection Error: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[Billing] FacturaAPI Response (%d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// VoidDocument initiates the voiding process (Comunicación de Baja) in FacturaAPI
func (s *BillingService) VoidDocument(companyID uuid.UUID, docUUID string, motivo string) error {
	if docUUID == "" {
		return fmt.Errorf("document UUID is required")
	}
	if len(motivo) < 10 {
		return fmt.Errorf("el motivo de anulación debe tener al menos 10 caracteres")
	}

	// 1. Get Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return fmt.Errorf("billing configuration not found")
	}

	apiURL := billingConfig.ApiURL
	apiKey := billingConfig.ApiKey

	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting)
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting)
		if apiURL == "" { apiURL = globalURLSetting.Valor }
		if apiKey == "" { apiKey = globalKeySetting.Valor }
	}

	if apiURL == "" || apiKey == "" {
		return fmt.Errorf("FacturaAPI configuration missing")
	}

	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	// 2. Send Request
	payload := map[string]string{
		"motivo": motivo,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL+"/api/v1/documents/"+docUUID+"/void", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteDocumentBeta removes a document from FacturaAPI (BETA mode only)
func (s *BillingService) DeleteDocumentBeta(companyID uuid.UUID, docUUID string) error {
	if docUUID == "" {
		return nil // Nothing to delete in API
	}

	// 1. Get Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return nil
	}

	apiURL := billingConfig.ApiURL
	apiKey := billingConfig.ApiKey

	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting)
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting)
		if apiURL == "" { apiURL = globalURLSetting.Valor }
		if apiKey == "" { apiKey = globalKeySetting.Valor }
	}

	if apiURL == "" || apiKey == "" {
		return nil
	}

	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	// 2. Send Request
	req, err := http.NewRequest("DELETE", apiURL+"/api/v1/documents/"+docUUID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 403 means it's production, 404 means already deleted, both are "fine" for cleanup attempts
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SyncLogo sends the business logo to FacturaAPI for printed representations
func (s *BillingService) SyncLogo(companyID uuid.UUID, logoBase64 string) error {
	if logoBase64 == "" {
		return nil
	}

	// 1. Ensure company exists in FacturaAPI before syncing logo
	tenantUUID, err := s.EnsureCompanyExists(companyID)
	if err != nil {
		return fmt.Errorf("failed to verify/register company in FacturaAPI: %w", err)
	}

	// 2. Get updated Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return nil // Not configured, nothing to sync
	}

	if billingConfig.Estado != "active" {
		return nil
	}

	// 3. Resolve API credentials
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

	if apiURL == "" || apiKey == "" {
		return fmt.Errorf("FacturaAPI configuration missing")
	}

	// Sanitize URL
	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	// 4. Prepare Multipart Request
	imgData, err := base64.StdEncoding.DecodeString(logoBase64)
	if err != nil {
		return fmt.Errorf("error decoding logo base64: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("logo", "logo.png")
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}
	_, _ = io.Copy(part, bytes.NewReader(imgData))
	writer.Close()

	req, err := http.NewRequest("POST", apiURL+"/api/v1/config/"+tenantUUID+"/logo", body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// mapTipoDoc converts ERP doc types to SUNAT codes (Catálogo 06)
func mapTipoDoc(tipo string) string {
	switch tipo {
	case "DNI":
		return "1"
	case "RUC":
		return "6"
	case "CE":
		return "4"
	case "PASAPORTE":
		return "7"
	case "SIN_DOCUMENTO":
		return "0"
	default:
		return "1"
	}
}

// prepareBillingCustomer prepares the customer structure for FacturaAPI
// applying the generic client rules if applicable.
func prepareBillingCustomer(customer models.Customer) BillingCustomer {
	numDoc := strings.TrimSpace(customer.NumeroDocumento)
	tipoDoc := strings.ToUpper(strings.TrimSpace(customer.TipoDocumento))

	if numDoc == "00000000" || numDoc == "0" || tipoDoc == "SIN_DOCUMENTO" || tipoDoc == "-" {
		return BillingCustomer{
			TipoDocumento:   "-",
			NumeroDocumento: "0",
			RazonSocial:     "CLIENTES VARIOS",
			Direccion:       customer.Direccion,
		}
	}

	return BillingCustomer{
		TipoDocumento:   mapTipoDoc(tipoDoc),
		NumeroDocumento: numDoc,
		RazonSocial:     customer.Nombre,
		Direccion:       customer.Direccion,
	}
}

// SyncDocumentStatus queries FacturaAPI for the latest status of a document and updates the local record
func (s *BillingService) SyncDocumentStatus(companyID uuid.UUID, doc *models.ElectronicDocument) (*EmitResponse, error) {
	if doc.DocumentUUID == nil || *doc.DocumentUUID == "" {
		return nil, fmt.Errorf("document has no UUID")
	}

	// 1. Get Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return nil, fmt.Errorf("billing configuration not found")
	}

	apiURL := billingConfig.ApiURL
	apiKey := billingConfig.ApiKey

	if apiURL == "" || apiKey == "" {
		var globalURLSetting models.CompanySetting
		var globalKeySetting models.CompanySetting
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_url").First(&globalURLSetting)
		config.DB.Where("company_id = ? AND clave = ?", uuid.Nil, "factura_api_key").First(&globalKeySetting)
		if apiURL == "" {
			apiURL = globalURLSetting.Valor
		}
		if apiKey == "" {
			apiKey = globalKeySetting.Valor
		}
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("FacturaAPI configuration missing")
	}

	apiURL = strings.TrimSuffix(apiURL, "/")
	apiURL = strings.TrimSuffix(apiURL, "/api/v1")

	// 2. Query Status
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL+"/api/v1/documents/"+*doc.DocumentUUID+"/status", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	var statusResp struct {
		Status           string `json:"status"`
		SunatResponse    string `json:"sunat_response"`
		SunatDescription string `json:"sunat_description"`
		SunatNotes       string `json:"sunat_notes"`
		Files            struct {
			XML string `json:"xml"`
			CDR string `json:"cdr"`
		} `json:"files"`
	}
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	// 3. Update Local Document
	doc.Estado = statusResp.Status

	// Process messages
	sunatMsg := statusResp.SunatResponse
	if sunatMsg == "" {
		sunatMsg = statusResp.SunatDescription
	}
	if statusResp.SunatNotes != "" {
		if sunatMsg != "" {
			sunatMsg += "\n" + statusResp.SunatNotes
		} else {
			sunatMsg = statusResp.SunatNotes
		}
	}

	doc.SunatResponse = s.ParseSunatResponse(sunatMsg)

	// Update URLs
	if statusResp.Files.XML != "" {
		doc.XmlURL = statusResp.Files.XML
	}
	if statusResp.Files.CDR != "" {
		doc.CdrURL = statusResp.Files.CDR
	}

	// Download/Extract from CDR if possible
	if doc.CdrURL != "" {
		cdrNotes, err := s.ExtractNotesFromCDR(doc.CdrURL)
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

	if doc.Estado == "accepted" && doc.SunatResponse != "" {
		doc.Observaciones = doc.SunatResponse
	}

	config.DB.Save(doc)

	// Return an EmitResponse compatible structure
	return &EmitResponse{
		DocumentoID:      *doc.DocumentUUID,
		Status:           doc.Estado,
		SunatResponse:    doc.SunatResponse,
		SunatDescription: statusResp.SunatDescription,
		SunatNotes:       statusResp.SunatNotes,
	}, nil
}


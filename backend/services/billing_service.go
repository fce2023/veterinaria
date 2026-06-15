package services

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

// EmitSale sends a sale to FacturaAPI for electronic billing
func (s *BillingService) EmitSale(sale models.Sale, items []models.SaleItem) (*EmitResponse, error) {
	if sale.TipoDocumento == "NV" || sale.TipoDocumento == "00" || sale.TipoDocumento == "99" {
		return nil, nil // Nota de venta is an internal ticket, no SUNAT emission
	}

	// 1. Get Billing Config for the company
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
		if apiURL[len(apiURL)-1] == '/' {
			apiURL = apiURL[:len(apiURL)-1]
		}
		suffix := "/api/v1"
		if len(apiURL) >= len(suffix) && apiURL[len(apiURL)-len(suffix):] == suffix {
			apiURL = apiURL[:len(apiURL)-len(suffix)]
		}
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("FacturaAPI global configuration missing (URL or Token not configured)")
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
	payload := EmitPayload{
		EmpresaID: billingConfig.TenantUUID,
		Async:     true,
		Modo:      billingConfig.Modo,
		Header: BillingHeader{
			TipoDocumento: sale.TipoDocumento,
			Serie:         sale.Serie,
			Numero:        num,
			FechaEmision:  sale.CreatedAt.Format("2006-01-02"),
		},
		Cliente: prepareBillingCustomer(customer),
		Items: billingItems,
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

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}
	electronicDoc.CompanyID = sale.CompanyID
	electronicDoc.SaleID = &sale.ID
	electronicDoc.FacturacionPayload = string(jsonData)
	electronicDoc.TipoDocumento = payload.Header.TipoDocumento
	electronicDoc.Serie = payload.Header.Serie
	electronicDoc.Numero = fmt.Sprintf("%08d", payload.Header.Numero)
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
	electronicDoc.DocumentUUID = emitResp.DocumentoID
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

// SyncLogo sends the business logo to FacturaAPI for printed representations
func (s *BillingService) SyncLogo(companyID uuid.UUID, logoBase64 string) error {
	if logoBase64 == "" {
		return nil
	}

	// 1. Get Billing Config
	var billingConfig models.BillingConfig
	if err := config.DB.Where("company_id = ?", companyID).First(&billingConfig).Error; err != nil {
		return nil // Not configured, nothing to sync
	}

	if billingConfig.TenantUUID == "" || billingConfig.Estado != "active" {
		return nil
	}

	// 2. Resolve API credentials
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
	if apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL[:len(apiURL)-1]
	}

	// 3. Prepare Multipart Request
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

	req, err := http.NewRequest("POST", apiURL+"/api/v1/config/"+billingConfig.TenantUUID+"/logo", body)
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

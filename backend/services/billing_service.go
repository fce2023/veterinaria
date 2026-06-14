package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

// FacturaAPI Payload Structures
type BillingHeader struct {
	TipoDocumento string `json:"tipo_documento"`
	Serie         string `json:"serie"`
	Numero        int    `json:"numero"`
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
	PrecioUnitario float64 `json:"precio_unitario"`
	Subtotal       float64 `json:"subtotal"`
}

type BillingTotals struct {
	TotalVenta     float64 `json:"total_venta"`
	TotalImpuestos float64 `json:"total_impuestos"`
}

type EmitPayload struct {
	EmpresaID string          `json:"empresa_id"`
	Async     bool            `json:"async"`
	Modo      string          `json:"modo"`
	Header    BillingHeader   `json:"header"`
	Cliente   BillingCustomer `json:"cliente"`
	Items     []BillingItem   `json:"items"`
	Totales   BillingTotals   `json:"totales"`
}

type EmitResponse struct {
	Message     string `json:"message"`
	DocumentoID string `json:"documento_id"`
	Status      string `json:"status"`
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

	// Sanitize URL: Remove trailing slash or "/api/v1" suffix to prevent duplicate segments
	if apiURL != "" {
		// Remove trailing slash
		if apiURL[len(apiURL)-1] == '/' {
			apiURL = apiURL[:len(apiURL)-1]
		}
		// Remove trailing "/api/v1"
		suffix := "/api/v1"
		if len(apiURL) >= len(suffix) && apiURL[len(apiURL)-len(suffix):] == suffix {
			apiURL = apiURL[:len(apiURL)-len(suffix)]
		}
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("FacturaAPI global configuration missing (URL or Token not configured)")
	}

	// 2. Prepare Payload
	// Map Customer
	var customer models.Customer
	config.DB.First(&customer, sale.CustomerID)

	// Determine Serie and Number (This should ideally come from a counter/branch config)
	// For MVP, we'll try to get it from Branch or use defaults
	var branch models.Branch
	config.DB.First(&branch, sale.BranchID)

	serie := branch.SerieFactura
	if serie == "" {
		serie = "F001" // Default
	}

	// In a real scenario, we should have a document counter. 
	// For now, we'll use a simplified approach or let the user provide it.
	// Since Sale model doesn't have a number yet, we'll use a placeholder or 
	// add a field to Sale later.

	payload := EmitPayload{
		EmpresaID: billingConfig.TenantUUID,
		Async:     true, // Default async for production-ready flow
		Modo:      billingConfig.Modo,
		Header: BillingHeader{
			TipoDocumento: "01", // Default to Factura for now, should be dynamic
			Serie:         serie,
			Numero:        1, // Placeholder: Needs a real counter
			FechaEmision:  sale.CreatedAt.Format("2006-01-02"),
		},
		Cliente: BillingCustomer{
			TipoDocumento:   mapTipoDoc(customer.TipoDocumento),
			NumeroDocumento: customer.NumeroDocumento,
			RazonSocial:     customer.Nombre,
			Direccion:       customer.Direccion,
		},
		Totales: BillingTotals{
			TotalVenta:     sale.Total,
			TotalImpuestos: sale.IGV,
		},
	}

	// Map Items
	for _, item := range items {
		var product models.Product
		config.DB.Select("nombre").First(&product, item.ProductID)

		payload.Items = append(payload.Items, BillingItem{
			Descripcion:    product.Nombre,
			Cantidad:       item.Cantidad,
			PrecioUnitario: item.PrecioUnitario,
			Subtotal:       (item.Cantidad * item.PrecioUnitario) - item.Descuento - (sale.IGV * (item.Cantidad * item.PrecioUnitario / sale.Total)), // Simplified subtotal
		})
	}

	// 3. Send HTTP Request
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL+"/api/v1/documents/emit", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("FacturaAPI error (%d): %s", resp.StatusCode, string(body))
	}

	var emitResp EmitResponse
	if err := json.Unmarshal(body, &emitResp); err != nil {
		return nil, err
	}

	// 4. Save Electronic Document Record
	electronicDoc := models.ElectronicDocument{
		CompanyID:     sale.CompanyID,
		SaleID:        &sale.ID,
		DocumentUUID:  emitResp.DocumentoID,
		TipoDocumento: payload.Header.TipoDocumento,
		Serie:         payload.Header.Serie,
		Numero:        fmt.Sprintf("%08d", payload.Header.Numero),
		Estado:        emitResp.Status,
	}
	config.DB.Create(&electronicDoc)

	return &emitResp, nil
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

// mapTipoDoc converts ERP doc types to SUNAT codes
func mapTipoDoc(tipo string) string {
	switch tipo {
	case "DNI":
		return "1"
	case "RUC":
		return "6"
	case "CE":
		return "4"
	default:
		return "1"
	}
}

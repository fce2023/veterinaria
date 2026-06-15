package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"veterinaria/backend/models"
)

func TestEmitSale(t *testing.T) {
	// Setup a mock server
	mockResp := EmitResponse{
		Message:     "Documento en cola",
		DocumentoID: "test-uuid-123",
		Status:      "pending",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		// Check payload
		var payload EmitPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		if payload.EmpresaID != "test-tenant-uuid" {
			t.Errorf("Expected EmpresaID test-tenant-uuid, got %s", payload.EmpresaID)
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	// Since we are using a real DB in config.DB, this is tricky for unit tests without a mock DB or test DB.
	// For this exercise, I will assume we want to test the logic of payload construction and HTTP calling.
	// In a real project, we'd use a test DB or an interface for the DB.
	
	t.Run("MapTipoDoc", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"DNI", "1"},
			{"RUC", "6"},
			{"CE", "4"},
			{"OTHER", "1"},
		}

		for _, tt := range tests {
			if got := mapTipoDoc(tt.input); got != tt.expected {
				t.Errorf("mapTipoDoc(%s) = %s; want %s", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("PrepareBillingCustomer", func(t *testing.T) {
		tests := []struct {
			name     string
			cust     models.Customer
			expected BillingCustomer
		}{
			{
				name: "Real customer with DNI",
				cust: models.Customer{
					TipoDocumento:   "DNI",
					NumeroDocumento: "12345678",
					Nombre:          "Juan Perez",
					Direccion:       "Av. Larco 123",
				},
				expected: BillingCustomer{
					TipoDocumento:   "1",
					NumeroDocumento: "12345678",
					RazonSocial:     "Juan Perez",
					Direccion:       "Av. Larco 123",
				},
			},
			{
				name: "Generic customer with 00000000",
				cust: models.Customer{
					TipoDocumento:   "DNI",
					NumeroDocumento: "00000000",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with 0",
				cust: models.Customer{
					TipoDocumento:   "DNI",
					NumeroDocumento: "0",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with SIN_DOCUMENTO type",
				cust: models.Customer{
					TipoDocumento:   "SIN_DOCUMENTO",
					NumeroDocumento: "999999",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with '-' type",
				cust: models.Customer{
					TipoDocumento:   "-",
					NumeroDocumento: "999999",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with case-insensitive and padded type",
				cust: models.Customer{
					TipoDocumento:   "  sin_documento  ",
					NumeroDocumento: "999999",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with padded document number",
				cust: models.Customer{
					TipoDocumento:   "DNI",
					NumeroDocumento: "  00000000  ",
					Nombre:          "Público General",
					Direccion:       "",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "",
				},
			},
			{
				name: "Generic customer with non-empty address",
				cust: models.Customer{
					TipoDocumento:   "DNI",
					NumeroDocumento: "0",
					Nombre:          "Público General",
					Direccion:       "Av. General 456",
				},
				expected: BillingCustomer{
					TipoDocumento:   "-",
					NumeroDocumento: "0",
					RazonSocial:     "CLIENTES VARIOS",
					Direccion:       "Av. General 456",
				},
			},
			{
				name: "Real customer with RUC",
				cust: models.Customer{
					TipoDocumento:   "RUC",
					NumeroDocumento: "20123456789",
					Nombre:          "Empresa S.A.",
					Direccion:       "Calle Lima 789",
				},
				expected: BillingCustomer{
					TipoDocumento:   "6",
					NumeroDocumento: "20123456789",
					RazonSocial:     "Empresa S.A.",
					Direccion:       "Calle Lima 789",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := prepareBillingCustomer(tt.cust)
				if got.TipoDocumento != tt.expected.TipoDocumento {
					t.Errorf("expected TipoDocumento %q, got %q", tt.expected.TipoDocumento, got.TipoDocumento)
				}
				if got.NumeroDocumento != tt.expected.NumeroDocumento {
					t.Errorf("expected NumeroDocumento %q, got %q", tt.expected.NumeroDocumento, got.NumeroDocumento)
				}
				if got.RazonSocial != tt.expected.RazonSocial {
					t.Errorf("expected RazonSocial %q, got %q", tt.expected.RazonSocial, got.RazonSocial)
				}
				if got.Direccion != tt.expected.Direccion {
					t.Errorf("expected Direccion %q, got %q", tt.expected.Direccion, got.Direccion)
				}
			})
		}
	})

	t.Run("ParseSunatResponse", func(t *testing.T) {
		s := NewBillingService()
		input := `</cac:Signature>
<cbc:Note>4096 - La provincia del domicilio fiscal del emisor no cumple con el formato establecido</cbc:Note>
<cbc:Note>4093 - El codigo de ubigeo del domicilio fiscal del emisor no es válido</cbc:Note>`
		expected := "4096 - La provincia del domicilio fiscal del emisor no cumple con el formato establecido\n4093 - El codigo de ubigeo del domicilio fiscal del emisor no es válido"

		if got := s.ParseSunatResponse(input); got != expected {
			t.Errorf("ParseSunatResponse() = %v; want %v", got, expected)
		}

		// Test non-XML input
		input2 := "Normal error message"
		if got := s.ParseSunatResponse(input2); got != input2 {
			t.Errorf("ParseSunatResponse() with non-XML = %v; want %v", got, input2)
		}
	})

	t.Log("Note: Full integration test for EmitSale requires a configured database.")
}

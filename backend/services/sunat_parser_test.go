package services

import (
	"strings"
	"testing"
)

func TestParseSunatResponse_Scenarios(t *testing.T) {
	s := NewBillingService()

	tests := []struct {
		name     string
		raw      string
		expected []string
	}{
		{
			name: "Multiple Notes from CDR",
			raw: `</cac:Signature>
<cbc:Note>4096 - La provincia del domicilio fiscal del emisor no cumple con el formato establecido</cbc:Note>
<cbc:Note>4093 - El codigo de ubigeo del domicilio fiscal del emisor no es válido</cbc:Note>`,
			expected: []string{
				"4096 - La provincia del domicilio fiscal del emisor no cumple con el formato establecido",
				"4093 - El codigo de ubigeo del domicilio fiscal del emisor no es válido",
			},
		},
		{
			name: "Description tag with attributes",
			raw: `<cbc:Description languageLocaleID="1000">El comprobante numero F001-00000001 ha sido aceptado</cbc:Description>`,
			expected: []string{
				"El comprobante numero F001-00000001 ha sido aceptado",
			},
		},
		{
			name: "Mixed Note and Description without namespace",
			raw: `<Note>Rechazado por SUNAT</Note><Description>Error en el RUC</Description>`,
			expected: []string{
				"Rechazado por SUNAT",
				"Error en el RUC",
			},
		},
		{
			name: "Duplicate messages",
			raw: `<cbc:Note>Error 123</cbc:Note><cbc:Description>Error 123</cbc:Description>`,
			expected: []string{
				"Error 123",
			},
		},
		{
			name: "Raw non-XML message",
			raw: "Conexion fallida con el servidor de SUNAT",
			expected: []string{
				"Conexion fallida con el servidor de SUNAT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.ParseSunatResponse(tt.raw)
			
			// If expected is empty or single raw string that matches input
			if len(tt.expected) == 1 && tt.expected[0] == tt.raw {
				if got != tt.raw {
					t.Errorf("ParseSunatResponse() = %v, want %v", got, tt.raw)
				}
				return
			}

			// Check if all expected strings are present in the output
			for _, exp := range tt.expected {
				if !strings.Contains(got, exp) {
					t.Errorf("ParseSunatResponse() missing expected note: %v\nGot: %v", exp, got)
				}
			}
			
			// Check number of lines (if applicable)
			lines := strings.Split(got, "\n")
			// Remove empty lines if any from split
			var cleanLines []string
			for _, l := range lines {
				if strings.TrimSpace(l) != "" {
					cleanLines = append(cleanLines, l)
				}
			}
			
			if len(cleanLines) != len(tt.expected) {
				t.Errorf("ParseSunatResponse() expected %d lines, got %d\nFull output: %v", len(tt.expected), len(cleanLines), got)
			}
		})
	}
}

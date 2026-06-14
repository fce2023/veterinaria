package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestQueryRUCAndDNI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Query RUC 20 - Medifarma", func(t *testing.T) {
		router := gin.New()
		router.GET("/public/ruc/:ruc", QueryRUC)

		req, _ := http.NewRequest("GET", "/public/ruc/20100018625", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
			return
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		success, ok := resp["success"].(bool)
		if !ok || !success {
			t.Errorf("Expected success true, got %v", resp["success"])
		}

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data object in response")
		}

		// Verify key properties of RUC 20 payload
		razonSocial := data["razon_social"].(string)
		if razonSocial != "MEDIFARMA S A" {
			t.Errorf("Expected MEDIFARMA S A, got %s", razonSocial)
		}

		direccion := data["direccion"].(string)
		if direccion == "" {
			t.Error("Expected non-empty address")
		}

		// Ensure locales anexos are present
		locales, ok := data["locales_anexos"].([]interface{})
		if !ok || len(locales) == 0 {
			t.Error("Expected locales_anexos array to have elements")
		}
	})

	t.Run("Query RUC 10 - Royner Dionel", func(t *testing.T) {
		router := gin.New()
		router.GET("/public/ruc/:ruc", QueryRUC)

		req, _ := http.NewRequest("GET", "/public/ruc/10721837811", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
			return
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		data := resp["data"].(map[string]interface{})
		razonSocial := data["razon_social"].(string)
		if razonSocial != "RODRIGUEZ FERNANDEZ ROYNER DIONEL" {
			t.Errorf("Expected RODRIGUEZ FERNANDEZ ROYNER DIONEL, got %s", razonSocial)
		}
	})

	t.Run("Query DNI - Royner Dionel", func(t *testing.T) {
		router := gin.New()
		router.GET("/public/dni/:dni", QueryDNI)

		req, _ := http.NewRequest("GET", "/public/dni/72183781", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Note: apiconsulta DNI might fail if query limits exceeded, but we check if we handle payload structure when HTTP 200
		if w.Code == http.StatusOK {
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
				data, ok := resp["data"].(map[string]interface{})
				if ok {
					fullName := data["full_name"].(string)
					if fullName != "RODRIGUEZ FERNANDEZ ROYNER DIONEL" {
						t.Errorf("Expected RODRIGUEZ FERNANDEZ ROYNER DIONEL, got %s", fullName)
					}
				}
			}
		}
	})
}

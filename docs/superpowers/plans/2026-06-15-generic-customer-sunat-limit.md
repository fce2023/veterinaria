# Generic Customer SUNAT Limit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement generic customer mapping for FacturaAPI electronic receipt submission and support the SUNAT limit constraints by ensuring correct JSON payload fields.

**Architecture:** We will create a helper function `prepareBillingCustomer` in `BillingService` that checks if the customer matches any generic customer pattern and structures the `cliente` payload sub-object as required by the API.

**Tech Stack:** Go 1.2x (Gin web framework, GORM)

---

### Task 1: Add unit tests for generic customer mapping

**Files:**
- Modify: `backend/services/billing_service_test.go` ([billing_service_test.go](file:///home/oyon/sehuacho/veterinaria/backend/services/billing_service_test.go))

- [ ] **Step 1: Write the failing tests**

Update `backend/services/billing_service_test.go` to include the unit tests for `prepareBillingCustomer`. Add this block inside `TestEmitSale`:

```go
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
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := prepareBillingCustomer(tt.cust)
				if got.TipoDocumento != tt.expected.TipoDocumento {
					t.Errorf("expected TipoDocumento %s, got %s", tt.expected.TipoDocumento, got.TipoDocumento)
				}
				if got.NumeroDocumento != tt.expected.NumeroDocumento {
					t.Errorf("expected NumeroDocumento %s, got %s", tt.expected.NumeroDocumento, got.NumeroDocumento)
				}
				if got.RazonSocial != tt.expected.RazonSocial {
					t.Errorf("expected RazonSocial %s, got %s", tt.expected.RazonSocial, got.RazonSocial)
				}
				if got.Direccion != tt.expected.Direccion {
					t.Errorf("expected Direccion %s, got %s", tt.expected.Direccion, got.Direccion)
				}
			})
		}
	})
```

- [ ] **Step 2: Run test to verify it fails**

Run the backend tests:
Run: `go test ./...` from the `backend` directory.
Expected: Compilation failure because `prepareBillingCustomer` is not defined yet.

- [ ] **Step 3: Commit**

```bash
git add backend/services/billing_service_test.go
git commit -m "test: add unit tests for generic customer mapping"
```

---

### Task 2: Implement generic customer mapping logic

**Files:**
- Modify: `backend/services/billing_service.go` ([billing_service.go](file:///home/oyon/sehuacho/veterinaria/backend/services/billing_service.go))

- [ ] **Step 1: Write the implementation code**

Add `prepareBillingCustomer` function at the bottom of `backend/services/billing_service.go`:

```go
// prepareBillingCustomer prepares the customer structure for FacturaAPI
// applying the generic client rules if applicable.
func prepareBillingCustomer(customer models.Customer) BillingCustomer {
	numDoc := strings.TrimSpace(customer.NumeroDocumento)
	tipoDoc := strings.TrimSpace(customer.TipoDocumento)

	if numDoc == "00000000" || numDoc == "0" || strings.ToUpper(tipoDoc) == "SIN_DOCUMENTO" || tipoDoc == "-" {
		return BillingCustomer{
			TipoDocumento:   "-",
			NumeroDocumento: "0",
			RazonSocial:     "CLIENTES VARIOS",
			Direccion:       customer.Direccion,
		}
	}

	return BillingCustomer{
		TipoDocumento:   mapTipoDoc(customer.TipoDocumento),
		NumeroDocumento: customer.NumeroDocumento,
		RazonSocial:     customer.Nombre,
		Direccion:       customer.Direccion,
	}
}
```

Then update `EmitSale` inside `backend/services/billing_service.go` to use this new function:

```go
		Cliente: prepareBillingCustomer(customer),
```

Specifically, locate the `Cliente` mapping in the `EmitPayload` struct declaration around line 175-180 and replace:

```go
		Cliente: BillingCustomer{
			TipoDocumento:   mapTipoDoc(customer.TipoDocumento),
			NumeroDocumento: customer.NumeroDocumento,
			RazonSocial:     customer.Nombre,
			Direccion:       customer.Direccion,
		},
```

with:

```go
		Cliente: prepareBillingCustomer(customer),
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test ./...` from the `backend` directory.
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add backend/services/billing_service.go
git commit -m "feat: implement generic customer mapping logic and integrate it"
```

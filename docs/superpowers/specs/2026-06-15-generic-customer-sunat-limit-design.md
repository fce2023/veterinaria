# Generic Customer SUNAT Limit and Payload Mapping - Design Specification

We will implement a mechanism in the backend Go application to map the default/generic customer data into the SUNAT-compliant Generic Client payload structure required by FacturaAPI. This handles boletas emitted to the public general when no identification is provided, conforming to the S/ 700 SUNAT rule.

---

## 1. Backend Design

### 1.1 `BillingService` Changes
File: `backend/services/billing_service.go` ([billing_service.go](file:///home/oyon/sehuacho/veterinaria/backend/services/billing_service.go))

* Implement a new helper function `prepareBillingCustomer(customer models.Customer) BillingCustomer`:
  1. Retrieve `customer.NumeroDocumento` and `customer.TipoDocumento`.
  2. If the document number is `"00000000"`, `"0"`, or the document type is `"SIN_DOCUMENTO"` or `"-"` (case-insensitive and trimmed):
     * Return a `BillingCustomer` with:
       * `TipoDocumento: "-"`
       * `NumeroDocumento: "0"`
       * `RazonSocial: "CLIENTES VARIOS"`
       * `Direccion: customer.Direccion`
  3. Otherwise, return the standard mapped customer details using the existing `mapTipoDoc(customer.TipoDocumento)` function.
* Update `EmitSale`:
  * Replace the inline assignment of the `Cliente` field in the `EmitPayload` struct with a call to `prepareBillingCustomer(customer)`.

### 1.2 Validation Check Context
As specified by the provider, the backend controller will NOT preventively block sending documents above S/ 700 with a generic client; it will let SUNAT validate and return an error, marking the document as `rejected` in the database. 
*(Note: The frontend already preventively blocks creating a sale in `SalesSection.vue` if the Boleta is >= S/ 700 and uses the default generic client `"00000000"`).*

---

## 2. Verification Plan

### 2.1 Unit Tests
File: `backend/services/billing_service_test.go` ([billing_service_test.go](file:///home/oyon/sehuacho/veterinaria/backend/services/billing_service_test.go))

Write unit tests to verify:
1. **Generic Customer Matching**:
   * Verify that a customer with `NumeroDocumento == "00000000"` is mapped to `"-"`, `"0"`, and `"CLIENTES VARIOS"`.
   * Verify that a customer with `NumeroDocumento == "0"` is mapped to `"-"`, `"0"`, and `"CLIENTES VARIOS"`.
   * Verify that a customer with `TipoDocumento == "SIN_DOCUMENTO"` is mapped to `"-"`, `"0"`, and `"CLIENTES VARIOS"`.
2. **Identified Customer Matching**:
   * Verify that a customer with `TipoDocumento == "DNI"` and `NumeroDocumento == "12345678"` maps to `"1"`, `"12345678"`, and their real name.

### 2.2 Security Verification Plan
* **Input Validation & Sanitization**: Ensure that trimmed spaces and casing do not bypass or cause errors in customer mapping.
* **SQL Injection & Data Access**: No database query changes are made; DB access continues to use GORM parameters.
* **Error Handling**: Database errors or serialization issues will fail close and return proper technical errors to logs while returning clean API error messages.

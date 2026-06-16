# Business Location Fields Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add fields and UI inputs to configure company location details (Ubigeo, Departamento, Provincia, Distrito) to prevent SUNAT warnings.

**Architecture:** Extend Company database model and backend API to store and sync these fields, and update frontend settings UI with input fields and a RUC query autocomplete button.

**Tech Stack:** Go (Gin, GORM), Vue 3, Axios.

---

### Task 1: Backend Database & Model Update

**Files:**
- Modify: `backend/models/models.go`

- [ ] **Step 1: Add location fields to `Company` model**
Modify `backend/models/models.go` around line 27:
```go
// Company represents a client tenant
type Company struct {
	BaseModel
	RUC             string `gorm:"type:varchar(11);uniqueIndex;not null" json:"ruc"`
	RazonSocial     string `gorm:"type:varchar(255);not null" json:"razon_social"`
	NombreComercial string `gorm:"type:varchar(255)" json:"nombre_comercial"`
	Direccion       string `gorm:"type:varchar(500)" json:"direccion"`
	Telefono        string `gorm:"type:varchar(50)" json:"telefono"`
	Email           string `gorm:"type:varchar(100)" json:"email"`
	LogoBase64      string `gorm:"type:text" json:"logo_base64"`
	Estado          string `gorm:"type:varchar(50);default:'active'" json:"estado"` // active, inactive
	Ubigeo          string `gorm:"type:varchar(6)" json:"ubigeo"`
	Departamento    string `gorm:"type:varchar(100)" json:"departamento"`
	Provincia       string `gorm:"type:varchar(100)" json:"provincia"`
	Distrito        string `gorm:"type:varchar(100)" json:"distrito"`
}
```

- [ ] **Step 2: Verify it compiles**
Compile the backend to verify that it compiles successfully.
Run command in Go project: `go build ./...`

- [ ] **Step 3: Commit database model changes**
Add and commit the changes:
```bash
git add backend/models/models.go
git commit -m "feat: add ubigeo and location fields to Company model"
```

---

### Task 2: Backend Handler Updates and Sync Logic

**Files:**
- Modify: `backend/handlers/handlers.go`
- Modify: `backend/handlers/handlers_test.go`

- [ ] **Step 1: Update input struct and logic in `UpdateMyCompany` handler**
In `backend/handlers/handlers.go`, modify `UpdateMyCompany` inputs and mapping to assign the new fields (`Ubigeo`, `Departamento`, `Provincia`, `Distrito`) to the company object:
```go
	var input struct {
		RazonSocial     string `json:"razon_social" binding:"required"`
		NombreComercial string `json:"nombre_comercial"`
		Direccion       string `json:"direccion"`
		Telefono        string `json:"telefono"`
		Email           string `json:"email"`
		LogoBase64      string `json:"logo_base64"`
		Ubigeo          string `json:"ubigeo"`
		Departamento    string `json:"departamento"`
		Provincia       string `json:"provincia"`
		Distrito        string `json:"distrito"`
	}
```
And inside `UpdateMyCompany`, map these fields:
```go
	company.RazonSocial = input.RazonSocial
	company.NombreComercial = input.NombreComercial
	company.Direccion = input.Direccion
	company.Telefono = input.Telefono
	company.Email = input.Email
	company.Ubigeo = input.Ubigeo
	company.Departamento = input.Departamento
	company.Provincia = input.Provincia
	company.Distrito = input.Distrito
```
Add the Main Branch Ubigeo sync logic:
```go
	if err := config.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Sync Ubigeo with the main branch if it exists and ubigeo is set
	if company.Ubigeo != "" {
		var mainBranch models.Branch
		if err := config.DB.Where("company_id = ? AND is_main = ?", company.ID, true).First(&mainBranch).Error; err == nil {
			mainBranch.Ubigeo = company.Ubigeo
			_ = config.DB.Save(&mainBranch).Error
		}
	}
```

- [ ] **Step 2: Add integration tests for updating company location**
In `backend/handlers/handlers_test.go`, add the imports if not present (`"bytes"`, `"github.com/google/uuid"`, `"veterinaria/backend/config"`, `"veterinaria/backend/models"`) and implement `TestUpdateMyCompany`:
```go
func TestUpdateMyCompany(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.ConnectDB()

	companyID := uuid.New()
	company := models.Company{
		BaseModel:   models.BaseModel{ID: companyID},
		RUC:         "10721837811",
		RazonSocial: "Test Company Location",
	}
	config.DB.Create(&company)
	defer config.DB.Unscoped().Delete(&company)

	mainBranch := models.Branch{
		BaseModel: models.BaseModel{ID: uuid.New()},
		CompanyID: companyID,
		Nombre:    "Main Branch",
		IsMain:    true,
		Ubigeo:    "000000",
	}
	config.DB.Create(&mainBranch)
	defer config.DB.Unscoped().Delete(&mainBranch)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("companyID", companyID)
		c.Next()
	})
	router.PUT("/companies/me", UpdateMyCompany)

	input := map[string]string{
		"razon_social": "Updated Company Name",
		"ubigeo":       "150801",
		"departamento": "LIMA",
		"provincia":    "HUAURA",
		"distrito":     "HUACHO",
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("PUT", "/companies/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var updatedCompany models.Company
	if err := config.DB.First(&updatedCompany, companyID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated company: %v", err)
	}

	if updatedCompany.Ubigeo != "150801" || updatedCompany.Departamento != "LIMA" || updatedCompany.Provincia != "HUAURA" || updatedCompany.Distrito != "HUACHO" {
		t.Errorf("Fields not updated: %+v", updatedCompany)
	}

	var updatedBranch models.Branch
	if err := config.DB.First(&updatedBranch, mainBranch.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated branch: %v", err)
	}
	if updatedBranch.Ubigeo != "150801" {
		t.Errorf("Main branch ubigeo not synced: expected 150801, got %s", updatedBranch.Ubigeo)
	}
}
```

- [ ] **Step 3: Run the new test and verify it passes**
Run the tests using the docker test container:
`docker run --rm --network host -e DB_HOST=localhost -e DB_PORT=5432 -e DB_USER=veterinaria_user -e DB_PASSWORD=veterinaria_pass123 -e DB_NAME=veterinaria_erp -e DB_SSLMODE=disable -v /home/oyon/sehuacho/veterinaria:/app -w /app/backend golang:1.20-alpine go test ./...`
Expected: PASS

- [ ] **Step 4: Commit handler updates and tests**
Run:
```bash
git add backend/handlers/handlers.go backend/handlers/handlers_test.go
git commit -m "feat: implement location fields in UpdateMyCompany and sync to main branch"
```

---

### Task 3: Frontend UI Fields & Autocomplete Integration

**Files:**
- Modify: `frontend/src/components/BusinessSection.vue`

- [ ] **Step 1: Add location fields to frontend `business` state object**
In `frontend/src/components/BusinessSection.vue`, locate:
```typescript
const business = reactive({
  ruc: '',
  razon_social: '',
  nombre_comercial: '',
  direccion: '',
  telefono: '',
  email: '',
  logo_base64: ''
})
```
and update to:
```typescript
const business = reactive({
  ruc: '',
  razon_social: '',
  nombre_comercial: '',
  direccion: '',
  telefono: '',
  email: '',
  logo_base64: '',
  ubigeo: '',
  departamento: '',
  provincia: '',
  distrito: ''
})
```

- [ ] **Step 2: Update `saveBusinessProfile()` payload**
In `frontend/src/components/BusinessSection.vue`, locate `saveBusinessProfile` and pass the new fields:
```typescript
    const res = await axios.put('/companies/me', {
      razon_social: business.razon_social,
      nombre_comercial: business.nombre_comercial,
      direccion: business.direccion,
      telefono: business.telefono,
      email: business.email,
      logo_base64: business.logo_base64,
      ubigeo: business.ubigeo,
      departamento: business.departamento,
      provincia: business.provincia,
      distrito: business.distrito
    })
```

- [ ] **Step 3: Implement RUC Query Autocomplete function**
In `frontend/src/components/BusinessSection.vue`, add state and query function in the script block:
```typescript
const queryingRuc = ref(false)

async function queryRucFromSunat() {
  if (!business.ruc || business.ruc.length !== 11) {
    businessAlert.msg = 'El RUC debe tener exactamente 11 dígitos.'
    businessAlert.type = 'error'
    return
  }

  queryingRuc.value = true
  businessAlert.msg = ''
  try {
    const res = await axios.get(`/public/ruc/${business.ruc}`)
    if (res.data.success && res.data.data) {
      const data = res.data.data
      if (data.direccion) business.direccion = data.direccion
      if (data.ubigeo) business.ubigeo = String(data.ubigeo)
      if (data.departamento) business.departamento = data.departamento
      if (data.provincia) business.provincia = data.provincia
      if (data.distrito) business.distrito = data.distrito
      businessAlert.msg = '✓ Datos del RUC importados de SUNAT correctamente. Presiona "Guardar" para guardarlos.'
      businessAlert.type = 'success'
    } else {
      businessAlert.msg = 'No se encontraron datos para este RUC en SUNAT.'
      businessAlert.type = 'error'
    }
  } catch (err: any) {
    businessAlert.msg = err.response?.data?.error || 'Error al consultar SUNAT.'
    businessAlert.type = 'error'
  } finally {
    queryingRuc.value = false
  }
}
```

- [ ] **Step 4: Update the form template to include the autocomplete button and location inputs**
Modify the template in `frontend/src/components/BusinessSection.vue`:
Replace the existing RUC input:
```vue
            <div class="field">
              <label class="field-label">RUC (11 dígitos)</label>
              <input
                :value="business.ruc"
                type="text"
                class="field-input field-input--locked"
                disabled
              />
              <p class="field-hint">Fijado en la creación de la cuenta</p>
            </div>
```
with:
```vue
            <div class="field">
              <label class="field-label">RUC (11 dígitos)</label>
              <div class="ruc-input-group">
                <input
                  :value="business.ruc"
                  type="text"
                  class="field-input field-input--locked"
                  disabled
                />
                <button
                  type="button"
                  @click="queryRucFromSunat"
                  class="btn-ruc-query"
                  :disabled="queryingRuc"
                >
                  <svg v-if="queryingRuc" class="spin-btn" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <circle cx="12" cy="12" r="10" stroke="rgba(255,255,255,0.3)"/>
                    <path d="M4 12a8 8 0 0 1 8-8" />
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="btn-icon">
                    <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
                  </svg>
                  <span>{{ queryingRuc ? 'Consultando...' : 'Consultar SUNAT' }}</span>
                </button>
              </div>
              <p class="field-hint">Autocompletar dirección y ubigeo</p>
            </div>
```
Replace the Contacto & Ubicación section:
```vue
          <div class="form-section-label" style="margin-top:24px;">Contacto & Ubicación</div>
          <div class="form-grid form-grid--3">
            <div class="field">
              <label class="field-label">Dirección Fiscal</label>
              <input
                v-model="business.direccion"
                type="text"
                placeholder="Av. Ejemplo 123, Lima..."
                class="field-input"
              />
              <p class="field-hint">Dirección que aparece en los comprobantes</p>
            </div>
            <div class="field">
              <label class="field-label">Teléfono</label>
              <input
                v-model="business.telefono"
                type="text"
                placeholder="01-234-5678..."
                class="field-input"
              />
            </div>
            <div class="field">
              <label class="field-label">Email de Contacto</label>
              <input
                v-model="business.email"
                type="email"
                placeholder="contacto@empresa.com..."
                class="field-input"
              />
            </div>
          </div>
```
with:
```vue
          <div class="form-section-label" style="margin-top:24px;">Ubicación Fiscal</div>
          <div class="form-grid form-grid--3">
            <div class="field" style="grid-column: span 2;">
              <label class="field-label">Dirección Fiscal</label>
              <input
                v-model="business.direccion"
                type="text"
                placeholder="Av. Ejemplo 123, Lima..."
                class="field-input"
              />
              <p class="field-hint">Dirección que aparece en los comprobantes</p>
            </div>
            <div class="field">
              <label class="field-label">Ubigeo (6 dígitos)</label>
              <input
                v-model="business.ubigeo"
                type="text"
                placeholder="150101..."
                class="field-input"
                maxlength="6"
              />
              <p class="field-hint">Ubigeo del domicilio fiscal emisor</p>
            </div>
          </div>

          <div class="form-grid form-grid--3" style="margin-top: 14px;">
            <div class="field">
              <label class="field-label">Departamento</label>
              <input
                v-model="business.departamento"
                type="text"
                placeholder="LIMA..."
                class="field-input"
              />
            </div>
            <div class="field">
              <label class="field-label">Provincia</label>
              <input
                v-model="business.provincia"
                type="text"
                placeholder="HUAURA..."
                class="field-input"
              />
            </div>
            <div class="field">
              <label class="field-label">Distrito</label>
              <input
                v-model="business.distrito"
                type="text"
                placeholder="HUACHO..."
                class="field-input"
              />
            </div>
          </div>

          <div class="form-section-label" style="margin-top:24px;">Contacto</div>
          <div class="form-grid form-grid--2">
            <div class="field">
              <label class="field-label">Teléfono</label>
              <input
                v-model="business.telefono"
                type="text"
                placeholder="01-234-5678..."
                class="field-input"
              />
            </div>
            <div class="field">
              <label class="field-label">Email de Contacto</label>
              <input
                v-model="business.email"
                type="email"
                placeholder="contacto@empresa.com..."
                class="field-input"
              />
            </div>
          </div>
```

- [ ] **Step 5: Add scoped CSS styles for the RUC query button and form grid layouts**
In the style block, add:
```css
.ruc-input-group {
  display: flex;
  gap: 8px;
}
.btn-ruc-query {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 16px;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 9px;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.btn-ruc-query:hover:not(:disabled) {
  background: #e2e8f0;
  color: #1e293b;
  border-color: #94a3b8;
}
.btn-ruc-query:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-icon {
  width: 14px;
  height: 14px;
}
.spin-btn {
  width: 14px;
  height: 14px;
  animation: spin 0.6s linear infinite;
}
.form-grid--2 {
  grid-template-columns: repeat(2, 1fr);
}
@media (max-width: 680px) {
  .form-grid--2 {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 6: Build the frontend to verify compilation**
Verify compilation.

- [ ] **Step 7: Commit frontend changes**
Run:
```bash
git add frontend/src/components/BusinessSection.vue
git commit -m "feat: add business location fields and SUNAT autocomplete to BusinessSection"
```

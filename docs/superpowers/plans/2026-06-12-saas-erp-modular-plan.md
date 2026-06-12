# Core ERP Modular SaaS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the multi-tenant SaaS billing, subscriptions, payment logs, and dynamic modules layer, prioritizing the Veterinary and Glass Shop (Vidriería) modules.

**Architecture:** We will introduce database models for Plans, Subscriptions, Payments, and CompanyModules. A Go/Gin middleware will protect API endpoints. In the POS, products defined as dimensional (M² or M) will dynamically calculate and store their total quantities (m² or meters) using Alto, Ancho, and CantidadPiezas inside a flexible JSONB Metadata field in GORM.

**Tech Stack:** Go (Gin, GORM, PostgreSQL) in the backend, Vue 3 (TypeScript, Tailwind, Vite) in the frontend.

---

### Task 1: Base de Datos de Suscripciones y Módulos (Capa SaaS)

**Files:**
- Modify: `backend/models/models.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Añadir los nuevos structs de datos a models.go**
  Add the GORM models for `Plan`, `Subscription`, `Payment`, `CompanyModule`, and modify `Product` and `SaleItem` with JSONB Metadata.
  
  Edit `backend/models/models.go` by appending or replacing the models:
  ```go
  // Plan defines subscription pricing and limitations
  type Plan struct {
  	BaseModel
  	Nombre      string  `gorm:"type:varchar(100);not null" json:"nombre"`
  	Precio      float64 `gorm:"type:numeric(12,2);not null" json:"precio"`
  	MaxBranches int     `gorm:"type:integer;default:1" json:"max_branches"`
  	MaxUsers    int     `gorm:"type:integer;default:5" json:"max_users"`
  	Modulos     string  `gorm:"type:text" json:"modulos"` // Comma-separated keys
  }
  
  // Subscription represents user licensing status
  type Subscription struct {
  	BaseModel
  	CompanyID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
  	PlanID    uuid.UUID `gorm:"type:uuid;not null" json:"plan_id"`
  	Estado    string    `gorm:"type:varchar(50);default:'TRIAL'" json:"estado"` // TRIAL, ACTIVE, EXPIRED, SUSPENDED
  	StartsAt  time.Time `json:"starts_at"`
  	ExpiresAt time.Time `json:"expires_at"`
  }
  
  // Payment records manual subscription payment approvals
  type Payment struct {
  	BaseModel
  	CompanyID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
  	SubscriptionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"subscription_id"`
  	Monto          float64    `gorm:"type:numeric(12,2);not null" json:"monto"`
  	MetodoPago     string     `gorm:"type:varchar(50);not null" json:"metodo_pago"` // YAPE, PLIN, TRANSFERENCIA
  	Referencia     string     `gorm:"type:varchar(255);uniqueIndex" json:"referencia"`
  	Estado         string     `gorm:"type:varchar(50);default:'PENDING'" json:"estado"` // PENDING, APPROVED, REJECTED
  	ComprobanteUrl string     `gorm:"type:varchar(500)" json:"comprobante_url"`
  	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
  }
  
  // CompanyModule toggles active extensions for tenants
  type CompanyModule struct {
  	BaseModel
  	CompanyID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
  	ModuleKey string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"`
  	IsActive  bool      `gorm:"type:boolean;default:true" json:"is_active"`
  }
  ```
  
  Also, update the `Product` struct to support:
  ```go
  UnidadMedida  string `gorm:"type:varchar(50);default:'und'" json:"unidad_medida"`
  IsDimensional bool   `gorm:"type:boolean;default:false" json:"is_dimensional"`
  ```
  
  And `SaleItem` to support JSONB metadata:
  ```go
  Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"` // For dimensional height/width
  ```

- [ ] **Step 2: Actualizar el listado de AutoMigrate en main.go**
  Add the new models to the `AutoMigrate` chain in `backend/main.go`:
  ```go
  err := config.DB.AutoMigrate(
      &models.Company{},
      &models.Branch{},
      &models.Permission{},
      &models.Role{},
      &models.User{},
      &models.Category{},
      &models.Brand{},
      &models.Product{},
      &models.Stock{},
      &models.Kardex{},
      &models.Supplier{},
      &models.Purchase{},
      &models.PurchaseItem{},
      &models.Customer{},
      &models.Pet{},
      &models.Sale{},
      &models.SaleItem{},
      &models.BillingConfig{},
      &models.ElectronicDocument{},
      &models.AuditLog{},
      &models.Plan{},
      &models.Subscription{},
      &models.Payment{},
      &models.CompanyModule{},
  )
  ```

- [ ] **Step 3: Compilar y ejecutar pruebas locales de migración**
  Run: `cd backend && go build -o /dev/null main.go`
  Expected: Compiled successfully.

- [ ] **Step 4: Realizar commit de cambios de modelos**
  Run:
  ```bash
  git add backend/models/models.go backend/main.go
  git commit -m "feat(db): add saas plans, subscriptions and modular components models"
  ```

---

### Task 2: Middleware Go/Gin para Autorización de Módulos (RequireModule)

**Files:**
- Create/Modify: `backend/auth/auth.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Agregar el middleware RequireModule a auth.go**
  Open `backend/auth/auth.go` and add the middleware function:
  ```go
  func RequireModule(moduleKey string) gin.HandlerFunc {
  	return func(c *gin.Context) {
  		companyIDStr, exists := c.Get("companyID")
  		if !exists {
  			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Acceso no autorizado"})
  			c.Abort()
  			return
  		}
  		companyID, err := uuid.Parse(fmt.Sprintf("%v", companyIDStr))
  		if err != nil {
  			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "ID de empresa inválido"})
  			c.Abort()
  			return
  		}
  		
  		var active bool
  		err = config.DB.Model(&models.CompanyModule{}).
  			Select("is_active").
  			Where("company_id = ? AND module_key = ? AND is_active = true", companyID, moduleKey).
  			Scan(&active).Error
  
  		if err != nil || !active {
  			c.JSON(http.StatusForbidden, gin.H{
  				"success": false, 
  				"error": "Acceso Restringido", 
  				"message": "Tu suscripción no incluye acceso al módulo: " + moduleKey,
  			})
  			c.Abort()
  			return
  		}
  		c.Next()
  	}
  }
  ```

- [ ] **Step 2: Proteger las rutas de Veterinaria mediante RequireModule**
  In `backend/main.go`, filter the `/pets` routes using the new middleware:
  ```go
  petsGroup := protected.Group("/pets")
  petsGroup.Use(auth.RequireModule("veterinaria"))
  {
      petsGroup.GET("", handlers.GetPets)
      petsGroup.POST("", handlers.CreatePet)
      petsGroup.PUT("/:id", handlers.UpdatePet)
      petsGroup.DELETE("/:id", handlers.DeletePet)
  }
  ```

- [ ] **Step 3: Compilar y verificar rutas**
  Run: `cd backend && go run main.go`
  Expected: Runs successfully.

- [ ] **Step 4: Realizar commit**
  Run:
  ```bash
  git add backend/auth/auth.go backend/main.go
  git commit -m "feat(auth): implement require module router middleware"
  ```

---

### Task 3: Aprobación de Pagos y Asignación de Módulos (SuperAdmin APIs)

**Files:**
- Create: `backend/handlers/saas_billing.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Crear las APIs del SuperAdmin para aprobar pagos y activar módulos**
  Create `backend/handlers/saas_billing.go` with functions to approve manual payments and toggle `CompanyModule` items.
  
  ```go
  package handlers
  
  import (
  	"net/http"
  	"time"
  	"github.com/gin-gonic/gin"
  	"github.com/google/uuid"
  	"veterinaria/backend/config"
  	"veterinaria/backend/models"
  )
  
  type ToggleModuleInput struct {
  	CompanyID uuid.UUID `json:"company_id" binding:"required"`
  	ModuleKey string    `json:"module_key" binding:"required"`
  	IsActive  bool      `json:"is_active"`
  }
  
  func ToggleCompanyModule(c *gin.Context) {
  	var input ToggleModuleInput
  	if err := c.ShouldBindJSON(&input); err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
  		return
  	}
  
  	var module models.CompanyModule
  	err := config.DB.Where("company_id = ? AND module_key = ?", input.CompanyID, input.ModuleKey).First(&module).Error
  	
  	if err != nil {
  		module = models.CompanyModule{
  			CompanyID: input.CompanyID,
  			ModuleKey: input.ModuleKey,
  			IsActive:  input.IsActive,
  		}
  		config.DB.Create(&module)
  	} else {
  		module.IsActive = input.IsActive
  		config.DB.Save(&module)
  	}
  
  	c.JSON(http.StatusOK, gin.H{"success": true, "data": module})
  }
  
  func ApprovePayment(c *gin.Context) {
  	idStr := c.Param("id")
  	paymentID, err := uuid.Parse(idStr)
  	if err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid payment ID"})
  		return
  	}
  
  	var payment models.Payment
  	if err := config.DB.First(&payment, paymentID).Error; err != nil {
  		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Payment not found"})
  		return
  	}
  
  	now := time.Now()
  	payment.Estado = "APPROVED"
  	payment.ApprovedAt = &now
  	config.DB.Save(&payment)
  
  	// Activate subscription
  	var sub models.Subscription
  	if err := config.DB.Where("id = ?", payment.SubscriptionID).First(&sub).Error; err == nil {
  		sub.Estado = "ACTIVE"
  		sub.ExpiresAt = time.Now().AddDate(0, 1, 0) // 1 Month extension
  		config.DB.Save(&sub)
  	}
  
  	c.JSON(http.StatusOK, gin.H{"success": true, "data": payment})
  }
  ```

- [ ] **Step 2: Registrar endpoints en main.go**
  Add the routes under the SaaS admin router group in `backend/main.go`:
  ```go
  saasAdmin.POST("/modules/toggle", handlers.ToggleCompanyModule)
  saasAdmin.POST("/payments/:id/approve", handlers.ApprovePayment)
  ```

- [ ] **Step 3: Realizar commit**
  Run:
  ```bash
  git add backend/handlers/saas_billing.go backend/main.go
  git commit -m "feat(saas): add module control and payment approval endpoints"
  ```

---

### Task 4: POS Backend con Cálculo Dimensional para Vidriería

**Files:**
- Modify: `backend/handlers/sales.go`

- [ ] **Step 1: Modificar el struct del Item en sales.go**
  Include dimensions metadata options in `backend/handlers/sales.go`:
  ```go
  type SaleItemInput struct {
  	ProductID      uuid.UUID `json:"product_id" binding:"required"`
  	Cantidad       float64   `json:"cantidad"` // Optional: override if dimensional
  	PrecioUnitario float64   `json:"precio_unitario" binding:"required"`
  	Descuento      float64   `json:"descuento"`
  	Alto           float64   `json:"alto"`
  	Ancho          float64   `json:"ancho"`
  	CantidadPiezas int       `json:"cantidad_piezas"`
  }
  ```

- [ ] **Step 2: Calcular dinámicamente m² en la creación de ventas**
  Add code inside `CreateSale` to verify if a product is dimensional, calculate total m², check stock, and store height/width in metadata JSON string.
  
  Modify the item loop inside `CreateSale` in `backend/handlers/sales.go`:
  ```go
  // Inside the loop for input.Items:
  var product models.Product
  if err := tx.First(&product, itemInput.ProductID).Error; err != nil {
  	tx.Rollback()
  	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Producto no encontrado"})
  	return
  }
  
  cantidadFinal := itemInput.Cantidad
  var metadataStr string
  
  if product.IsDimensional {
  	if itemInput.Alto <= 0 || itemInput.Ancho <= 0 || itemInput.CantidadPiezas <= 0 {
  		tx.Rollback()
  		c.JSON(http.StatusBadRequest, gin.H{
  			"success": false, 
  			"error": fmt.Sprintf("Producto dimensional requiere Alto, Ancho y Piezas válidos: %s", product.Nombre),
  		})
  		return
  	}
  	cantidadFinal = itemInput.Alto * itemInput.Ancho * float64(itemInput.CantidadPiezas)
  	metadataStr = fmt.Sprintf("{\"alto\": %.2f, \"ancho\": %.2f, \"piezas\": %d}", itemInput.Alto, itemInput.Ancho, itemInput.CantidadPiezas)
  }
  ```

- [ ] **Step 3: Almacenar metadatos en la tabla SaleItem**
  Assign the computed `cantidadFinal` and `metadataStr` to `SaleItem`:
  ```go
  saleItem := models.SaleItem{
  	SaleID:         sale.ID,
  	ProductID:      itemInput.ProductID,
  	Cantidad:       cantidadFinal,
  	PrecioUnitario: itemInput.PrecioUnitario,
  	Descuento:      itemInput.Descuento,
  	Metadata:       metadataStr,
  }
  ```

- [ ] **Step 4: Compilar y validar el backend**
  Run: `cd backend && go run main.go`
  Expected: Server starts up with no errors.

- [ ] **Step 5: Realizar commit**
  Run:
  ```bash
  git add backend/handlers/sales.go
  git commit -m "feat(sales): add dimensional item m2 calculations and inventory validation"
  ```

---

### Task 5: Adaptar el Sidebar y Rutas del Frontend (Vue 3)

**Files:**
- Modify: `backend/handlers/handlers.go`
- Modify: `frontend/src/views/Dashboard.vue`

- [ ] **Step 1: Retornar módulos activos en /auth/me**
  In the backend `handlers.go`, load active modules for the user's company and append to `/auth/me` JSON response.
  
  ```go
  var modules []string
  config.DB.Model(&models.CompanyModule{}).
  	Where("company_id = ? AND is_active = true", user.CompanyID).
  	Pluck("module_key", &modules)
  ```

- [ ] **Step 2: Dinamizar la navegación lateral del Dashboard**
  In `frontend/src/views/Dashboard.vue`, load active modules and render tabs conditionally:
  
  ```javascript
  // Example conditional filter in Dashboard.vue component tabs:
  const allowedTabs = computed(() => {
    return allTabs.filter(tab => {
      if (tab.module === 'veterinaria') return activeModules.value.includes('veterinaria');
      if (tab.module === 'vidrieria') return activeModules.value.includes('vidrieria');
      return true; // Core tabs
    });
  });
  ```

- [ ] **Step 3: Realizar commit**
  Run:
  ```bash
  git add backend/handlers/handlers.go frontend/src/views/Dashboard.vue
  git commit -m "feat(ui): implement dynamic sidebar menu tabs based on tenant active modules"
  ```

---

### Task 6: Crear el Calculador Dimensional en el POS Frontend

**Files:**
- Modify: `frontend/src/views/Dashboard.vue`

- [ ] **Step 1: Añadir inputs de Alto, Ancho e Insumo al POS UI**
  If the selected product has `is_dimensional === true`, display inputs: `Alto (m)`, `Ancho (m)`, and `Piezas`.

- [ ] **Step 2: Calcular m² dinámicamente y agregarlo al carrito**
  Compute the square meters total dynamically on keyup. When adding to the checkout list, append the calculated dimensions details into the items payload.

- [ ] **Step 3: Probar el flujo completo**
  Submit a purchase from the POS with a dimensional item and check if the stock is properly deducted by the calculated square meters.

- [ ] **Step 4: Realizar commit final**
  Run:
  ```bash
  git add frontend/src/views/Dashboard.vue
  git commit -m "feat(pos): implement dimensional product calculations in frontend checkout form"
  ```

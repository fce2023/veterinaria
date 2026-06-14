# Modular SaaS ERP and Dimensional POS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement audit enhancements for SaaS module toggles, return company modules in user session, extend product and stock queries with dimensional flags, and integrate dynamic POS dimensional inputs and calculations in Vue 3.

**Architecture:** Update `CompanyModule` database models and saas billing handlers with audit logs. Expand backend `/auth/me` and `/stocks` to return active modules and product dimensional metadata. Extend the Vue 3 Product list and POS component with responsive forms for LINEAR (M), AREA (M²), and VOLUME (M³) calculations, enforcing recalculations securely on the Go backend.

**Tech Stack:** Go (Gin, GORM, PostgreSQL) in the backend, Vue 3 (TypeScript, Tailwind, Vite) in the frontend.

---

### Task 1: Auditoría de Módulos (CompanyModule) y Migraciones

**Files:**
- Modify: `backend/models/models.go`
- Modify: `backend/handlers/saas_billing.go`

- [ ] **Step 1: Añadir campos de auditoría a CompanyModule**
  Open `backend/models/models.go` and update the `CompanyModule` struct:
  ```go
  // CompanyModule defines which domain modules are active per tenant
  type CompanyModule struct {
  	BaseModel
  	CompanyID        uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
  	ModuleKey        string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"` // e.g. "veterinaria", "vidrieria"
  	IsActive         bool       `gorm:"type:boolean;default:true" json:"is_active"`
  	EnabledAt        *time.Time `json:"enabled_at,omitempty"`
  	DisabledAt       *time.Time `json:"disabled_at,omitempty"`
  	EnabledByUserID  *uuid.UUID `gorm:"type:uuid" json:"enabled_by_user_id,omitempty"`
  	DisabledByUserID *uuid.UUID `gorm:"type:uuid" json:"disabled_by_user_id,omitempty"`
  }
  ```

- [ ] **Step 2: Actualizar la lógica de Toggle en saas_billing.go**
  Modify `backend/handlers/saas_billing.go` to populate the audit fields on toggling:
  ```go
  func ToggleCompanyModule(c *gin.Context) {
  	userID := c.MustGet("userID").(uuid.UUID)

  	var input ToggleModuleInput
  	if err := c.ShouldBindJSON(&input); err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
  		return
  	}

  	var module models.CompanyModule
  	err := config.DB.Where("company_id = ? AND module_key = ?", input.CompanyID, input.ModuleKey).First(&module).Error

  	now := time.Now()

  	if err != nil {
  		module = models.CompanyModule{
  			CompanyID: input.CompanyID,
  			ModuleKey: input.ModuleKey,
  			IsActive:  input.IsActive,
  		}
  		if input.IsActive {
  			module.EnabledAt = &now
  			module.EnabledByUserID = &userID
  		} else {
  			module.DisabledAt = &now
  			module.DisabledByUserID = &userID
  		}
  		config.DB.Create(&module)
  	} else {
  		if module.IsActive != input.IsActive {
  			module.IsActive = input.IsActive
  			if input.IsActive {
  				module.EnabledAt = &now
  				module.EnabledByUserID = &userID
  				module.DisabledAt = nil
  				module.DisabledByUserID = nil
  			} else {
  				module.DisabledAt = &now
  				module.DisabledByUserID = &userID
  			}
  		}
  		config.DB.Save(&module)
  	}

  	actionStr := "DISABLE_MODULE"
  	if input.IsActive {
  		actionStr = "ENABLE_MODULE"
  	}
  	logSaasAction(userID, input.CompanyID, actionStr, fmt.Sprintf("Módulo '%s' cambiado a activo=%v", input.ModuleKey, input.IsActive), c.ClientIP())

  	c.JSON(http.StatusOK, gin.H{"success": true, "data": module})
  }
  ```

- [ ] **Step 3: Compilar y validar el backend**
  Run: `cd backend && go build -o /dev/null main.go`
  Expected: Builds with no errors.

- [ ] **Step 4: Commit de cambios**
  ```bash
  git add backend/models/models.go backend/handlers/saas_billing.go
  git commit -m "feat(saas): add audit fields and trace logs for module activation toggles"
  ```

---

### Task 2: Modificar Endpoints /auth/me y /stocks

**Files:**
- Modify: `backend/handlers/handlers.go`
- Modify: `backend/handlers/kardex.go`

- [ ] **Step 1: Incluir módulos en la sesión (/auth/me)**
  Open `backend/handlers/handlers.go` and update `GetMe` (lines 72-85) to return modules:
  ```go
  func GetMe(c *gin.Context) {
  	userID := c.MustGet("userID").(uuid.UUID)

  	var user models.User
  	if err := config.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
  		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
  		return
  	}

  	var modules []string
  	config.DB.Model(&models.CompanyModule{}).
  		Where("company_id = ? AND is_active = true", user.CompanyID).
  		Pluck("module_key", &modules)

  	c.JSON(http.StatusOK, gin.H{
  		"success": true,
  		"data": gin.H{
  			"id":         user.ID,
  			"nombre":     user.Nombre,
  			"email":      user.Email,
  			"username":   user.Username,
  			"company_id": user.CompanyID,
  			"branch_id":  user.BranchID,
  			"role_type":  user.RoleType,
  			"roles":      user.Roles,
  			"modules":    modules,
  		},
  	})
  }
  ```

- [ ] **Step 2: Retornar metadata dimensional en /stocks**
  Open `backend/handlers/kardex.go` and update `GetStocks` (lines 12-46) to return `is_dimensional` and `unidad_medida`:
  ```go
  func GetStocks(c *gin.Context) {
  	var products []models.Product
  	if err := config.DB.Scopes(config.TenantFilter(c)).Where("estado = 'active'").Find(&products).Error; err != nil {
  		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
  		return
  	}

  	var stocks []models.Stock
  	if err := config.DB.Scopes(config.BranchFilter(c)).Find(&stocks).Error; err != nil {
  		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
  		return
  	}

  	stockMap := make(map[uuid.UUID]float64)
  	for _, s := range stocks {
  		stockMap[s.ProductID] = s.StockActual
  	}

  	var response []gin.H
  	for _, p := range products {
  		stockVal := stockMap[p.ID]
  		response = append(response, gin.H{
  			"product_id":     p.ID,
  			"product_nombre": p.Nombre,
  			"product_codigo": p.Codigo,
  			"precio_compra":  p.PrecioCompra,
  			"precio_venta":   p.PrecioVenta,
  			"stock_minimo":   p.StockMinimo,
  			"stock_actual":   stockVal,
  			"estado":         p.Estado,
  			"is_dimensional": p.IsDimensional,
  			"unidad_medida":   p.UnidadMedida,
  		})
  	}

  	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
  }
  ```

- [ ] **Step 3: Compilar y validar el backend**
  Run: `cd backend && go build -o /dev/null main.go`
  Expected: Build success.

- [ ] **Step 4: Commit de cambios**
  ```bash
  git add backend/handlers/handlers.go backend/handlers/kardex.go
  git commit -m "feat(api): append company modules to session and product dimensions to stock query"
  ```

---

### Task 3: Actualizar Formulario de Productos en el Frontend

**Files:**
- Modify: `frontend/src/components/ProductsSection.vue`

- [ ] **Step 1: Añadir campos de Unidad y Dimensiones al Formulario**
  Open `frontend/src/components/ProductsSection.vue`. In the form (between category select and prices, or around line 214), insert inputs for `unidad_medida` and `is_dimensional`:
  ```html
  <div class="grid grid-cols-2 gap-4">
    <div>
      <label class="block text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 px-1">Unidad de Medida</label>
      <select 
        v-model="productForm.unidad_medida"
        class="block w-full px-5 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-indigo-500/10 outline-none transition-all"
      >
        <option value="und">Unidad (UND)</option>
        <option value="m">Metros (M)</option>
        <option value="m2">Metros Cuadrados (M²)</option>
        <option value="m3">Metros Cúbicos (M³)</option>
        <option value="kg">Kilogramos (KG)</option>
        <option value="lt">Litros (LT)</option>
      </select>
    </div>
    <div class="flex items-center pt-8 pl-4">
      <label class="flex items-center gap-3 cursor-pointer">
        <input 
          v-model="productForm.is_dimensional"
          type="checkbox"
          class="w-5 h-5 text-indigo-600 border-slate-300 rounded focus:ring-indigo-500"
        />
        <span class="text-xs font-bold text-slate-700 select-none uppercase tracking-tight">Es Dimensional</span>
      </label>
    </div>
  </div>
  ```

- [ ] **Step 2: Actualizar estado reactivo y reset/edición en script**
  In the `<script setup>` block:
  - Add `unidad_medida` and `is_dimensional` to `productForm` state:
    ```typescript
    const productForm = reactive({
      nombre: '',
      codigo: '',
      codigo_barras: '',
      category_id: '',
      precio_compra: 0,
      precio_venta: 0,
      estado: 'active',
      unidad_medida: 'und',
      is_dimensional: false
    })
    ```
  - Modify `resetForm` (around line 339) to reset these properties:
    ```typescript
    Object.assign(productForm, {
      nombre: '',
      codigo: '',
      codigo_barras: '',
      category_id: '',
      precio_compra: 0,
      precio_venta: 0,
      estado: 'active',
      unidad_medida: 'und',
      is_dimensional: false
    })
    ```
  - Modify `editProduct` (around line 353) to populate them:
    ```typescript
    Object.assign(productForm, {
      nombre: item.nombre,
      codigo: item.codigo,
      codigo_barras: item.codigo_barras,
      category_id: item.category_id || '',
      precio_compra: item.precio_compra,
      precio_venta: item.precio_venta,
      estado: item.estado,
      unidad_medida: item.unidad_medida || 'und',
      is_dimensional: item.is_dimensional || false
    })
    ```

- [ ] **Step 3: Mostrar Unidad de Medida en la Tabla**
  Update the table row display in `frontend/src/components/ProductsSection.vue` to show the unit next to category or in product column, e.g.:
  ```html
  <span class="text-[10px] font-black text-slate-400 mt-0.5">{{ item.codigo || 'SIN-COD' }} • {{ item.unidad_medida?.toUpperCase() }}</span>
  ```

- [ ] **Step 4: Commit de cambios**
  ```bash
  git add frontend/src/components/ProductsSection.vue
  git commit -m "feat(ui): add unit of measure and dimensional flag controls to product catalog"
  ```

---

### Task 4: Adaptar el Sidebar y Rutas basado en Módulos

**Files:**
- Modify: `frontend/src/store/auth.ts`
- Modify: `frontend/src/views/Dashboard.vue`

- [ ] **Step 1: Actualizar interfaz User en Pinia Store**
  Open `frontend/src/store/auth.ts` and update the `User` interface to support modules:
  ```typescript
  export interface User {
    id: string
    nombre: string
    email: string
    username: string
    company_id: string
    branch_id: string
    role_type: string
    roles?: Array<{ nombre: string }>
    modules?: string[]
  }
  ```

- [ ] **Step 2: Ocultar condicionalmente pestañas de Veterinaria**
  Open `frontend/src/views/Dashboard.vue`. In the sidebar/drawer navigation (around line 128), wrap the "Clientes" or add a conditional sub-tab menu or check if we should restrict sections.
  Specifically, since pets management modal is inside `CustomersSection.vue`, we will check dynamic sidebar rendering.
  Wait, let's see: we want to prevent showing Mascotas inside the Customers list. We will pass a prop or update the customer section to read if "veterinaria" module is active.
  Let's add a computed property for modules in `Dashboard.vue`:
  ```typescript
  const activeModules = computed(() => authStore.user?.modules || [])
  const hasVetModule = computed(() => activeModules.value.includes('veterinaria'))
  ```
  Pass `hasVetModule` as a prop to `<CustomersSection />`:
  ```html
  <CustomersSection v-if="currentSection === 'customers'" :has-vet-module="hasVetModule" />
  ```
  Also display a small badge/status on the header next to Plan for the active modules:
  ```html
  <div class="flex flex-col items-end">
    <p class="text-[9px] text-slate-400 font-black uppercase tracking-tighter">Plan de Servicio</p>
    <p class="text-xs font-black text-indigo-600">PROFESSIONAL SAAS</p>
    <span class="text-[8px] font-mono text-slate-500">{{ activeModules.join(', ').toUpperCase() }}</span>
  </div>
  ```

- [ ] **Step 3: Ajustar CustomersSection.vue para recibir el prop**
  Open `frontend/src/components/CustomersSection.vue` and declare the prop:
  ```typescript
  const props = defineProps({
    hasVetModule: {
      type: Boolean,
      default: false
    }
  })
  ```
  And hide the "Mascotas" button (around line 110) in the table/cards list if `props.hasVetModule` is false:
  ```html
  <button 
    v-if="props.hasVetModule"
    @click="openPets(cust)" 
    class="px-2.5 py-1 bg-emerald-50 text-emerald-600 rounded-lg text-[10px] font-black hover:bg-emerald-600 hover:text-white transition-all flex items-center gap-1"
  >
    <i class="pi pi-heart-fill"></i>
    <span>Mascotas</span>
  </button>
  ```

- [ ] **Step 4: Commit de cambios**
  ```bash
  git add frontend/src/store/auth.ts frontend/src/views/Dashboard.vue frontend/src/components/CustomersSection.vue
  git commit -m "feat(ui): restrict pets management visibility based on tenant active modules"
  ```

---

### Task 5: POS Frontend con Calculador Dimensional y Envío al Backend

**Files:**
- Modify: `frontend/src/components/SalesSection.vue`

- [ ] **Step 1: Diseñar Formulario Dimensional en el Carrito**
  Open `frontend/src/components/SalesSection.vue`. In the cart item loop (around line 107), if the item product is dimensional, display the inputs.
  We need to import `stocks` or store product data inside each item inside `saleForm.items`.
  Let's update `quickAddItem` to save dimensional metadata defaults:
  ```typescript
  function quickAddItem(stock: any) {
    const alreadyAdded = saleForm.items.find(item => item.product_id === stock.product_id)
    const currentQty = alreadyAdded ? alreadyAdded.cantidad : 0

    if (stock.is_dimensional) {
      // For dimensional items, we always push as a separate row to let users configure cuts independently
      saleForm.items.push({
        product_id: stock.product_id,
        cantidad: 0, // Calculated dynamically
        precio_unitario: stock.precio_venta,
        descuento: 0,
        alto: 1.0,
        ancho: stock.unidad_medida === 'm' ? 0 : 1.0,
        espesor: stock.unidad_medida === 'm3' ? 1.0 : 0,
        cantidad_piezas: 1,
        is_dimensional: true,
        unidad_medida: stock.unidad_medida
      })
      isCartOpen.value = true
      return
    }

    if (currentQty + 1 > stock.stock_actual) {
      alert('Stock insuficiente')
      return
    }

    if (alreadyAdded) {
      alreadyAdded.cantidad += 1
    } else {
      saleForm.items.push({
        product_id: stock.product_id,
        cantidad: 1,
        precio_unitario: stock.precio_venta,
        descuento: 0,
        is_dimensional: false
      })
    }
  }
  ```

- [ ] **Step 2: Renderizar inputs dimensionales reactivos**
  In the template where items are listed in the cart, if `item.is_dimensional` is true:
  ```html
  <div v-if="item.is_dimensional" class="w-full grid grid-cols-4 gap-2 mt-3 p-3 bg-slate-50 rounded-2xl border border-slate-100">
    <div class="flex flex-col">
      <label class="text-[8px] font-black text-slate-400 uppercase">Largo/Alto</label>
      <input 
        v-model.number="item.alto" 
        type="number" 
        step="0.01" 
        class="w-full bg-white border border-slate-200 rounded-lg p-1 text-[10px] font-bold outline-none"
        @input="recalculateItemQty(item)"
      />
    </div>
    <div v-if="item.unidad_medida !== 'm'" class="flex flex-col">
      <label class="text-[8px] font-black text-slate-400 uppercase">Ancho</label>
      <input 
        v-model.number="item.ancho" 
        type="number" 
        step="0.01" 
        class="w-full bg-white border border-slate-200 rounded-lg p-1 text-[10px] font-bold outline-none"
        @input="recalculateItemQty(item)"
      />
    </div>
    <div v-if="item.unidad_medida === 'm3'" class="flex flex-col">
      <label class="text-[8px] font-black text-slate-400 uppercase">Espesor</label>
      <input 
        v-model.number="item.espesor" 
        type="number" 
        step="0.01" 
        class="w-full bg-white border border-slate-200 rounded-lg p-1 text-[10px] font-bold outline-none"
        @input="recalculateItemQty(item)"
      />
    </div>
    <div class="flex flex-col">
      <label class="text-[8px] font-black text-slate-400 uppercase">Piezas</label>
      <input 
        v-model.number="item.cantidad_piezas" 
        type="number" 
        step="1" 
        class="w-full bg-white border border-slate-200 rounded-lg p-1 text-[10px] font-bold outline-none"
        @input="recalculateItemQty(item)"
      />
    </div>
    <div class="col-span-4 text-[9px] font-black text-slate-500 text-right mt-1">
      Subtotal cantidad: <span class="text-indigo-600">{{ item.cantidad.toFixed(3) }}</span> {{ item.unidad_medida }}
    </div>
  </div>
  ```

- [ ] **Step 3: Agregar lógica de cálculo en script**
  Create `recalculateItemQty` in script block of `SalesSection.vue`:
  ```typescript
  function recalculateItemQty(item: any) {
    const alto = item.alto || 0
    const ancho = item.ancho || 0
    const espesor = item.espesor || 0
    const piezas = item.cantidad_piezas || 1

    if (item.unidad_medida === 'm') {
      item.cantidad = alto * piezas
    } else if (item.unidad_medida === 'm2') {
      item.cantidad = alto * ancho * piezas
    } else if (item.unidad_medida === 'm3') {
      item.cantidad = alto * ancho * espesor * piezas
    } else {
      item.cantidad = alto * (ancho || 1) * piezas
    }
  }
  ```
  Ensure that when adding an item, we call `recalculateItemQty(newItem)` so it doesn't stay at 0.
  Modify the `quickAddItem` to run:
  ```typescript
  if (stock.is_dimensional) {
      const newItem = {
        product_id: stock.product_id,
        cantidad: 0,
        precio_unitario: stock.precio_venta,
        descuento: 0,
        alto: 1.0,
        ancho: stock.unidad_medida === 'm' ? 0 : 1.0,
        espesor: stock.unidad_medida === 'm3' ? 1.0 : 0,
        cantidad_piezas: 1,
        is_dimensional: true,
        unidad_medida: stock.unidad_medida
      }
      recalculateItemQty(newItem)
      saleForm.items.push(newItem)
      isCartOpen.value = true
      return
  }
  ```

- [ ] **Step 4: Mostrar las dimensiones guardadas en el modal de detalles de venta**
  In `SalesSection.vue`, inside the details modal table (around line 227), render dimensions if they are loaded:
  ```html
  <div class="flex-1">
    <h5 class="text-xs font-black text-slate-900 leading-tight">{{ item.product_nombre }}</h5>
    <p v-if="item.is_dimensional" class="text-[9px] text-indigo-500 font-bold mt-1">
      Corte: {{ item.alto }}m 
      <span v-if="item.ancho">x {{ item.ancho }}m</span> 
      <span v-if="item.espesor">x {{ item.espesor }}m</span>
      ({{ item.cantidad_piezas }} piezas)
    </p>
  </div>
  ```

- [ ] **Step 5: Probar flujo completo de compra**
  Test checking out a dimensional product (e.g. glass by m², or cables by m) and verify stocks deduct properly.

- [ ] **Step 6: Commit final**
  ```bash
  git add frontend/src/components/SalesSection.vue
  git commit -m "feat(pos): integrate dynamic dimensional fields and hot quantity recalculations"
  ```

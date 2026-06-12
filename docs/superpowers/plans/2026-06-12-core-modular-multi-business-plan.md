# Plan de Implementación: Core ERP Modular SaaS (Multi-Negocio)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the single-tenant Veterinary ERP into a modular multi-business SaaS core, prioritizing the Veterinaria and Vidriería (glass shop) modules.

**Architecture:** We will introduce a `CompanyModule` table to assign modules to companies. We will implement middleware to protect backend routes and adapt the frontend navigation menu dynamically. The Vidriería module will add support for dimensional items (measured in m²) in the catalog and calculated subtotals (height × width × quantity) in the POS.

**Tech Stack:** Go (Gin, GORM, PostgreSQL) in the backend, Vue 3 (Vite, TypeScript, Tailwind) in the frontend.

---

### Task 1: Actualizar el Modelo de Base de Datos y Migraciones

**Files:**
- Modify: `backend/models/models.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Modificar el modelo de datos en models.go**
  Añadir el modelo `CompanyModule`, y los campos multi-negocio en `Company`, `Product` y `SaleItem`.
  
  Reemplazar el contenido correspondiente en `backend/models/models.go` con las nuevas definiciones:
  ```go
  // CompanyModule representa un módulo activado para una empresa específica
  type CompanyModule struct {
  	BaseModel
  	CompanyID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
  	ModuleKey  string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"` // "veterinaria", "vidrieria", "restaurante", etc.
  	IsActive   bool      `gorm:"type:boolean;default:true" json:"is_active"`
  }
  
  // Agregar campos nuevos a Product:
  // UnidadMedida (string), IsDimensional (bool), RequiresLot (bool)
  
  // Agregar campos nuevos a SaleItem:
  // Alto (float64), Ancho (float64), CantidadPiezas (int)
  ```

- [ ] **Step 2: Registrar el nuevo modelo para AutoMigrate en main.go**
  Añadir `&models.CompanyModule{}` a la lista de migración en `backend/main.go`.
  
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
      &models.CompanyModule{}, // <--- Nuevo modelo
  )
  ```

- [ ] **Step 3: Ejecutar pruebas de compilación del backend**
  Run: `go build -o /dev/null backend/main.go` from root.
  Expected: Compila exitosamente sin errores.

---

### Task 2: Implementar el Middleware de Validación de Módulos

**Files:**
- Create/Modify: `backend/auth/auth.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Crear la función de middleware RequireModule en auth.go**
  Implementar el middleware para restringir peticiones HTTP si la empresa del usuario no tiene contratado el módulo solicitado.
  
  Agregar a `backend/auth/auth.go`:
  ```go
  func RequireModule(moduleKey string) gin.HandlerFunc {
  	return func(c *gin.Context) {
  		companyIDStr, exists := c.Get("companyID")
  		if !exists {
  			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Acceso no autorizado"})
  			c.Abort()
  			return
  		}
  		companyID := companyIDStr.(uuid.UUID)
  		var active bool
  		
  		err := config.DB.Model(&models.CompanyModule{}).
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

- [ ] **Step 2: Proteger las rutas específicas en main.go**
  Actualizar las rutas específicas en `backend/main.go` para que utilicen el middleware.
  
  Modificar el grupo de rutas en `backend/main.go`:
  ```go
  // Pets (Veterinaria) protegido por RequireModule("veterinaria")
  petsGroup := protected.Group("/pets")
  petsGroup.Use(auth.RequireModule("veterinaria"))
  {
      petsGroup.GET("", handlers.GetPets)
      petsGroup.POST("", handlers.CreatePet)
      petsGroup.PUT("/:id", handlers.UpdatePet)
      petsGroup.DELETE("/:id", handlers.DeletePet)
  }
  ```

---

### Task 3: Modificar la Base de Semillas (Seeds) de Base de Datos

**Files:**
- Modify: `backend/main.go`

- [ ] **Step 1: Habilitar módulos por defecto al crear la empresa demo**
  Modificar el método `seedDatabase()` en `backend/main.go` para insertar las tuplas de activación de módulos para la veterinaria por defecto.
  
  Agregar el siguiente bloque después de crear la empresa en `backend/main.go`:
  ```go
  // Activar módulos de veterinaria y vidrieria para la empresa demo
  config.DB.Create(&models.CompanyModule{CompanyID: company.ID, ModuleKey: "veterinaria", IsActive: true})
  config.DB.Create(&models.CompanyModule{CompanyID: company.ID, ModuleKey: "vidrieria", IsActive: true})
  ```

- [ ] **Step 2: Levantar el servidor para aplicar migraciones y cargar semillas**
  Run: `cd backend && go run main.go`
  Expected: Ver la consola imprimiendo que las migraciones y semillas finalizaron con éxito.

---

### Task 4: Modificar el POS Backend para Cálculos Dimensionales (Vidriería)

**Files:**
- Modify: `backend/handlers/sales.go`

- [ ] **Step 1: Actualizar estructuras de input en sales.go**
  Añadir los campos dimensionales opcionales al struct de item en la venta en `backend/handlers/sales.go`.
  
  ```go
  type SaleItemInput struct {
  	ProductID      uuid.UUID `json:"product_id" binding:"required"`
  	Cantidad       float64   `json:"cantidad"` // Si es dimensional, se calculará en el backend
  	PrecioUnitario float64   `json:"precio_unitario" binding:"required"`
  	Descuento      float64   `json:"descuento"`
  	Alto           float64   `json:"alto"`
  	Ancho          float64   `json:"ancho"`
  	CantidadPiezas int       `json:"cantidad_piezas"`
  }
  ```

- [ ] **Step 2: Calcular dinámicamente la cantidad si el producto es dimensional**
  Dentro de `CreateSale` en `backend/handlers/sales.go`, obtener el producto y validar/calcular su área total.
  
  Insertar dentro del bucle de items:
  ```go
  var product models.Product
  if err := tx.First(&product, itemInput.ProductID).Error; err != nil {
      tx.Rollback()
      c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Producto no encontrado"})
      return
  }
  
  cantidadFinal := itemInput.Cantidad
  if product.IsDimensional {
      if itemInput.Alto <= 0 || itemInput.Ancho <= 0 || itemInput.CantidadPiezas <= 0 {
          tx.Rollback()
          c.JSON(http.StatusBadRequest, gin.H{
              "success": false,
              "error": fmt.Sprintf("Los productos dimensionales requieren Alto, Ancho y Piezas válidos (%s)", product.Nombre),
          })
          return
      }
      cantidadFinal = itemInput.Alto * itemInput.Ancho * float64(itemInput.CantidadPiezas)
  }
  
  // Usar cantidadFinal para las validaciones de stock, cálculo de Kardex y deducción.
  ```

- [ ] **Step 3: Guardar detalles dimensionales en la tabla SaleItem**
  Guardar los campos `Alto`, `Ancho` y `CantidadPiezas` en la instancia final de `models.SaleItem` creada.
  
  ```go
  saleItem := models.SaleItem{
      SaleID:         sale.ID,
      ProductID:      itemInput.ProductID,
      Cantidad:       cantidadFinal,
      PrecioUnitario: itemInput.PrecioUnitario,
      Descuento:      itemInput.Descuento,
      Alto:           itemInput.Alto,
      Ancho:          itemInput.Ancho,
      CantidadPiezas: itemInput.CantidadPiezas,
  }
  ```

---

### Task 5: Ajustar el Auth Payload y Menú Lateral Dinámico (Frontend)

**Files:**
- Modify: `backend/handlers/handlers.go` (Endpoint GetMe)
- Modify: `frontend/src/views/Dashboard.vue` (o archivo de navegación lateral)

- [ ] **Step 1: Retornar los módulos activos de la empresa en el endpoint GetMe**
  Modificar el handler `GetMe` en el backend para incluir los códigos de los módulos activos.
  
  ```go
  // En handlers.go, modificar el struct de respuesta de GetMe
  var modules []string
  config.DB.Model(&models.CompanyModule{}).
      Where("company_id = ? AND is_active = true", user.CompanyID).
      Pluck("module_key", &modules)
  
  c.JSON(http.StatusOK, gin.H{
      "success": true,
      "data": gin.H{
          "user": user,
          "modules": modules, // <--- Lista de strings: ["veterinaria", "vidrieria"]
      },
  })
  ```

- [ ] **Step 2: Modificar el componente de navegación del Dashboard en Vue**
  Actualizar la lógica de carga del menú lateral para ocultar o mostrar dinámicamente "Mascotas" e "Historias Clínicas" si el módulo `veterinaria` está presente en la respuesta del usuario.
  
  (Si `Dashboard.vue` maneja el menú, filtrar los items del menú con base a `modules` obtenidos del Login/Session).

---

### Task 6: Crear el Formulario Dimensional en el POS (Frontend)

**Files:**
- Modify: `frontend/src/views/Dashboard.vue` (POS tab/component)

- [ ] **Step 1: Renderizar campos de Alto/Ancho en el POS si el producto es dimensional**
  En el formulario de selección del item antes de añadirlo al carrito de ventas, verificar si `product.is_dimensional` es true.
  Mostrar los campos: `Alto (m)`, `Ancho (m)`, `Piezas`.

- [ ] **Step 2: Calcular y mostrar el subtotal dinámico en la interfaz**
  Calcular reactivamente:
  $$\text{Cantidad Equivalente (m2)} = \text{Alto} \times \text{Ancho} \times \text{Piezas}$$
  Multiplicar por `Precio de Venta` y descontar el descuento ingresado.

- [ ] **Step 3: Enviar los datos dimensionales en la llamada POST al endpoint de venta**
  Actualizar el payload enviado a `/api/v1/sales` para pasar las propiedades `alto`, `ancho` y `cantidad_piezas`.

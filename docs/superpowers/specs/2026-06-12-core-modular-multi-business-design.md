# Especificación de Diseño: Core ERP Modular SaaS (Multi-Negocio)

*   **Fecha:** 2026-06-12
*   **Estado:** En revisión
*   **Autor:** Antigravity (AI Architect)

---

## 1. Resumen Ejecutivo
Esta especificación define la transformación del ERP, inicialmente diseñado para veterinarias, en una plataforma SaaS modular y multiempresa de núcleo común. La arquitectura permitirá a cada empresa cliente (tenant) operar de acuerdo a su rubro específico (Restaurante, Veterinaria, Ferretería, Vidriería, Farmacia, etc.) sin ver módulos irrelevantes para su negocio.

La activación de características estará gobernada por el **SuperAdmin** mediante un panel de control de módulos por empresa, y la interfaz (menú y rutas) se adaptará dinámicamente.

---

## 2. Elementos Estrictamente Comunes (El Core)
Todas las empresas compartirán la misma base tecnológica para los procesos universales de negocio:
*   **Seguridad y Tenancy**: Autenticación, registro, gestión de sucursales, roles y permisos.
*   **Inventario Core**: Registro de productos, categorías, marcas, stock por sucursal y Kardex.
*   **Compras**: Proveedores y órdenes de compra con ingreso a stock.
*   **Ventas y Caja**: Apertura y cierre de caja, base de clientes, flujo POS y comprobantes de venta.
*   **Facturación Electrónica**: Integración con **FacturaAPI** para la emisión y envío de documentos a la SUNAT.
*   **Auditoría**: Logs inmutables de operaciones críticas.

---

## 3. Modelo de Datos (Base de Datos)

### 3.1 Gestión de Módulos (Tabla Nueva)
Para vincular qué características están activas para qué empresas:

```go
// CompanyModule representa un módulo activado para una empresa específica
type CompanyModule struct {
	BaseModel
	CompanyID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
	ModuleKey  string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"` // ej. "veterinaria", "restaurante", "farmacia", "vidrieria"
	IsActive   bool      `gorm:"type:boolean;default:true" json:"is_active"`
}
```

### 3.2 Adaptación del Catálogo de Productos (`Product`)
Se agregan campos para dar soporte al módulo de **Vidriería** y **Farmacia** (lotes/medidas dimensionales):

```go
type Product struct {
	BaseModel
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	CategoryID    uuid.UUID `gorm:"type:uuid;index" json:"category_id"`
	BrandID       uuid.UUID `gorm:"type:uuid;index" json:"brand_id"`
	Codigo        string    `gorm:"type:varchar(100)" json:"codigo"`
	CodigoBarras  string    `gorm:"type:varchar(100)" json:"codigo_barras"`
	Nombre        string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion   string    `gorm:"type:text" json:"descripcion"`
	PrecioCompra  float64   `gorm:"type:numeric(12,4);default:0" json:"precio_compra"`
	PrecioVenta   float64   `gorm:"type:numeric(12,4);default:0" json:"precio_venta"`
	StockMinimo   float64   `gorm:"type:numeric(12,4);default:0" json:"stock_minimo"`
	Estado        string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
	
	// NUEVOS CAMPOS PARA SOPORTE MULTI-NEGOCIO
	UnidadMedida  string    `gorm:"type:varchar(50);default:'und'" json:"unidad_medida"` // und, m2, m, kg, lt
	IsDimensional bool      `gorm:"type:boolean;default:false" json:"is_dimensional"`    // true para Vidriería/Corte
	RequiresLot   bool      `gorm:"type:boolean;default:false" json:"requires_lot"`      // true para Farmacia (lotes/vencimiento)
}
```

### 3.3 Soporte de Ventas por Dimensiones (`SaleItem` y `Kardex`)
Para el POS de **Vidrierías** (alto x ancho x cantidad de piezas) y descuento correcto de stock:

```go
type SaleItem struct {
	BaseModel
	SaleID         uuid.UUID `gorm:"type:uuid;not null;index" json:"sale_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Cantidad       float64   `gorm:"type:numeric(12,4);not null" json:"cantidad"` // Representa la cantidad total (ej. m2 o unidades)
	PrecioUnitario float64   `gorm:"type:numeric(12,4);not null" json:"precio_unitario"` // Precio por m2 o por Unidad
	Descuento      float64   `gorm:"type:numeric(12,4);default:0" json:"descuento"`
	
	// DETALLE DIMENSIONAL (VIDRIERÍA/CORTE)
	Alto           float64   `gorm:"type:numeric(12,4);default:0" json:"alto,omitempty"`  // metros
	Ancho          float64   `gorm:"type:numeric(12,4);default:0" json:"ancho,omitempty"` // metros
	CantidadPiezas int       `gorm:"type:integer;default:0" json:"cantidad_piezas,omitempty"`
}
```

---

## 4. Backend (Go/Gin) - Protección de APIs

Se implementará un Middleware en Go para restringir el acceso a endpoints de módulos no contratados:

```go
package auth

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

func RequireModule(moduleKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// El CompanyID es extraído previamente por el AuthMiddleware
		companyIDStr, exists := c.Get("company_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
			c.Abort()
			return
		}
		
		companyID := companyIDStr.(uuid.UUID)
		var active bool
		
		// Verificar en base de datos si el módulo está habilitado para el cliente
		err := config.DB.Model(&models.CompanyModule{}).
			Select("is_active").
			Where("company_id = ? AND module_key = ?", companyID, moduleKey).
			Scan(&active).Error

		if err != nil || !active {
			c.JSON(http.StatusForbidden, gin.H{
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

---

## 5. Frontend (Vue 3) - Rutas y Sidebar Dinámico

1.  **Carga de Configuración**: Al cargar la aplicación (después del Login), se almacena en Pinia/Vuex la lista de módulos activos obtenidos de `/api/v1/auth/me`.
2.  **Menú Sidebar**:
    ```typescript
    const menuItems = computed(() => {
      const baseMenu = [
        { name: 'Dashboard', route: '/dashboard', icon: 'ChartPieIcon' },
        { name: 'Inventario', route: '/inventario', icon: 'ArchiveBoxIcon' },
        { name: 'Ventas (POS)', route: '/pos', icon: 'ShoppingCartIcon' },
        { name: 'Facturación', route: '/facturacion', icon: 'DocumentTextIcon' },
      ];

      // Inyección dinámica de módulos específicos
      if (store.activeModules.includes('veterinaria')) {
        baseMenu.push(
          { name: 'Mascotas', route: '/mascotas', icon: 'HeartIcon' },
          { name: 'Historias Clínicas', route: '/historias', icon: 'FolderOpenIcon' }
        );
      }
      if (store.activeModules.includes('restaurante')) {
        baseMenu.push(
          { name: 'Control de Mesas', route: '/mesas', icon: 'TableIcon' },
          { name: 'Cocina', route: '/cocina', icon: 'FireIcon' }
        );
      }
      return baseMenu;
    });
    ```
3.  **POS Dinámico**:
    *   Si el producto a vender tiene `is_dimensional == true`, el formulario del POS requiere ingresar `Alto`, `Ancho` y `Piezas`, y calcula el subtotal de manera automática antes de agregar el item.

---

## 6. Onboarding y Plantillas de Negocio
El SuperAdmin podrá clasificar el negocio mediante perfiles que activan automáticamente los siguientes módulos por defecto:

| Perfil / Giro | Módulos Activos por Defecto |
| :--- | :--- |
| **Veterinaria** | `Core`, `veterinaria` |
| **Restaurante** | `Core`, `restaurante` |
| **Vidriería** | `Core`, `vidrieria` (Habilita el cálculo dimensional Alto x Ancho) |
| **Farmacia** | `Core`, `farmacia` (Lotes y Vencimientos en Inventario) |
| **Comercio General** (Minimarket, Ferretería, Ropa) | `Core` |

---

## 7. Pruebas y Criterios de Aceptación
1.  **Aislamiento**: Una veterinaria no puede consultar `/api/v1/restaurant/tables` ni `/api/v1/pets` si no tiene el módulo contratado (debe retornar `403 Forbidden`).
2.  **Cálculo Dimensional**: Una vidriería puede vender un vidrio de 1.50m x 0.80m (1.20 m²), y el stock del producto debe reducirse en `1.20` m² exactos en el Kardex.
3.  **UI Adaptable**: El menú debe ocultar por completo las secciones no contratadas inmediatamente después del inicio de sesión.

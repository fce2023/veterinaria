# Especificación de Arquitectura y Diseño: ERP SaaS Modular Multiempresa

*   **Fecha:** 2026-06-12
*   **Autor:** Antigravity (AI SaaS Architect)
*   **Estado:** Aprobado

---

## 1. Filosofía de Arquitectura y Aislamiento

El sistema se estructura como un **Monolito Modular** multiempresa (multi-tenant) con base de datos compartida y aislamiento lógico. El objetivo clave es permitir activar o desactivar módulos dinámicamente según el rubro de cada empresa (`Company`) y su plan de suscripción (`Subscription`), reutilizando una única base de código.

```mermaid
graph TD
    subgraph Capa 1: Plataforma SaaS (Billing & Admin)
        SaaSAdmin[SuperAdmin Panel]
        TenantMgr[Gestión de Clientes / Empresas]
        SubscrMgr[Motor de Suscripciones & Pagos]
        ModuleRegistry[Registro & Activación de Módulos]
        SaasAudit[SaasAuditLog]
    end

    subgraph Capa 2: Core ERP (Shared Core)
        Auth[Autenticación & JWT]
        Inventory[Catálogo, Stock & Kardex]
        Sales[Ventas, POS & Caja]
        CompanySetting[CompanySettings]
        SaleItem[SaleItems]
    end

    subgraph Capa 3: Módulos Verticales (Extensions)
        VetMod[Veterinaria: Mascotas, Historias]
        GlassItemDim[SaleItemDimension: M, M², M³]
        MelamineMod[Melamina / Vidriería / Cables]
    end

    SaaSAdmin --> TenantMgr
    TenantMgr --> ModuleRegistry
    ModuleRegistry -->|Filtra Acceso| Capa 2
    ModuleRegistry -->|Filtra Acceso| Capa 3
    Capa 2 --> Capa 3
    SaleItem -->|Tiene Opcional| GlassItemDim
```

---

## 2. Esquema de Base de Datos y Entidades SaaS/Core

### 2.1 Modelos de Negocio SaaS

#### Plan
Representa los paquetes comerciales ofrecidos.
*   `Nombre`: Básico, Profesional, Empresarial.
*   `Precio`: Monto mensual base.
*   `MaxBranches`: Sucursales máximas permitidas.
*   `MaxUsers`: Usuarios máximos permitidos.
*   `Modulos`: Llaves de módulos incluidos por defecto (ej: `"core,veterinaria"`).

#### Subscription
Representa el estado actual de licenciamiento por empresa.
*   `CompanyID`: Tenant asociado.
*   `PlanID`: Plan contratado.
*   `Estado`: `TRIAL`, `ACTIVE`, `EXPIRED`, `SUSPENDED`.
*   `StartsAt` y `ExpiresAt`: Periodo de validez.

#### Payment
Historial de transacciones para control y aprobación manual.
*   `CompanyID` y `SubscriptionID`: Relación con el inquilino y su suscripción.
*   `Monto`: Total pagado.
*   `MetodoPago`: `YAPE`, `PLIN`, `TRANSFERENCIA`, `DEPOSITO`.
*   `Referencia`: Código de operación único del banco.
*   `Estado`: `PENDING`, `APPROVED`, `REJECTED`.
*   `ComprobanteUrl`: Comprobante de pago subido.
*   `ApprovedAt` / `ApprovedBy`: Auditoría de aprobación manual por SuperAdmin.

#### CompanyModule
Controla qué verticales están habilitados por inquilino, integrando auditoría de activación.
```go
type CompanyModule struct {
	BaseModel
	CompanyID        uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
	ModuleKey        string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"`
	IsActive         bool       `gorm:"type:boolean;default:true" json:"is_active"`
	EnabledAt        *time.Time `json:"enabled_at,omitempty"`
	DisabledAt       *time.Time `json:"disabled_at,omitempty"`
	EnabledByUserID  *uuid.UUID `gorm:"type:uuid" json:"enabled_by_user_id,omitempty"`
	DisabledByUserID *uuid.UUID `gorm:"type:uuid" json:"disabled_by_user_id,omitempty"`
}
```

### 2.2 Modelos de Configuración y Auditoría SaaS

#### CompanySetting
Repositorio central de configuraciones llave-valor por inquilino.
```go
type CompanySetting struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_setting" json:"company_id"`
	Clave     string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_setting" json:"clave"` // e.g. "receipt_footer", "timezone"
	Valor     string    `gorm:"type:text" json:"valor"`
}
```

#### SaasAuditLog
Bitácora de auditoría inmutable para acciones de nivel SuperAdmin en el plano de control.
```go
type SaasAuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`       // SuperAdmin
	CompanyID   uuid.UUID `gorm:"type:uuid;index" json:"company_id"`    // Tenant afectado
	Accion      string    `gorm:"type:varchar(100);not null" json:"accion"` // e.g. "APPROVE_PAYMENT", "TOGGLE_MODULE"
	Descripcion string    `gorm:"type:text" json:"descripcion"`
	IP          string    `gorm:"type:varchar(45)" json:"ip"`
	Fecha       time.Time `json:"fecha"`
}
```

### 2.3 Modelos Dimensionales Genéricos (Core/Verticals)

#### Product
*   `UnidadMedida`: `und`, `m`, `m2`, `m3`, `kg`, `lt`.
*   `IsDimensional`: Flag booleano que determina si el producto requiere cálculo de dimensiones físicas al venderse.

#### SaleItemDimension
Tabla de extensión relacional que desglosa el cálculo dimensional del item de venta.
```go
type SaleItemDimension struct {
	BaseModel
	SaleItemID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;index" json:"sale_item_id"`
	Alto           float64   `gorm:"type:numeric(12,4);default:0" json:"alto"`            // Altura / Longitud (m, m2, m3)
	Ancho          float64   `gorm:"type:numeric(12,4);default:0" json:"ancho"`           // Ancho (m2, m3)
	Espesor        float64   `gorm:"type:numeric(12,4);default:0" json:"espesor"`         // Espesor (m3)
	CantidadPiezas int       `gorm:"type:integer;default:1" json:"cantidad_piezas"`
}
```

---

## 4. Reglas del Calculador Dimensional del Backend

Para evitar fraudes o discrepancias, **el backend nunca confía en el valor de la cantidad calculado por el frontend**. Al procesar una venta (`CreateSale`), se ejecuta el siguiente algoritmo estricto:

1.  Se recibe la lista de items. Para cada item, se consulta el producto en la base de datos.
2.  Si `Product.IsDimensional == false`:
    *   La cantidad final cobrada y descontada del stock es exactamente `ItemInput.Cantidad`.
3.  Si `Product.IsDimensional == true`:
    *   Se lee `Product.UnidadMedida`.
    *   **Cálculo Lineal (`m`)**: Requiere `Alto > 0`. Cantidad final = $\text{Alto} \times \text{Piezas}$.
    *   **Cálculo de Área (`m2`)**: Requiere `Alto > 0` y `Ancho > 0`. Cantidad final = $\text{Alto} \times \text{Ancho} \times \text{Piezas}$.
    *   **Cálculo de Volumen (`m3`)**: Requiere `Alto > 0`, `Ancho > 0` y `Espesor > 0`. Cantidad final = $\text{Alto} \times \text{Ancho} \times \text{Espesor} \times \text{Piezas}$.
    *   Cualquier valor menor o igual a cero en los campos requeridos provocará un rechazo inmediato de la venta (`400 Bad Request`).
4.  Se valida el stock físico actual contra la cantidad calculada en el backend.
5.  Se genera la venta, se crea la fila correspondiente en `SaleItemDimension` y se descuenta el stock en m, m² o m³ en la tabla de Kardex/Inventario.

---

## 5. Implementación del Frontend Vue 3

### 5.1 Menú Lateral Dinámico
*   El componente `Dashboard.vue` consulta `/auth/me` al cargar.
*   La respuesta incluye el listado `modules` activos para el inquilino.
*   Si `"veterinaria"` está ausente, los accesos a "Mascotas" e "Historias Clínicas" en la vista de Clientes quedan bloqueados o invisibles.

### 5.2 POS Adaptativo (`SalesSection.vue`)
*   Al seleccionar un producto en el formulario de ventas, se inspecciona reactivamente su flag `is_dimensional`.
*   Si es verdadero, se despliega un sub-formulario en el modal/fila de venta solicitando `Alto (m)`, `Ancho (m)`, `Espesor (m)` y `Piezas` según su unidad de medida.
*   Se calcula en tiempo real el área/volumen y se muestra el total proyectado antes de enviar la orden de venta al backend.

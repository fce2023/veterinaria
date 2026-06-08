# 📑 Plan de Implementación: Estructura SaaS Profesional - ERP Veterinaria

Este documento detalla la hoja de ruta para transformar el prototipo actual en un sistema ERP SaaS (Software as a Service) profesional, escalable, seguro y listo para comercialización.

---

## 1. Refactorización del Modelo de Datos (Cimientos)
Para soportar una operación profesional, el esquema de base de datos debe evolucionar para incluir control de licencias y detalles operativos por sede necesarios para la facturación electrónica.

*   **Entidad `Company` (Empresa / Inquilino):**
    *   `PlanType`: Tipo de plan contratado (`Basic`, `Premium`, `Enterprise`).
    *   `SubscriptionExpiresAt`: Fecha y hora de expiración de la licencia.
    *   `MaxBranches`: Límite de sucursales permitidas por el plan (ej: 1 para Basic, 3 para Premium, ilimitado para Enterprise).
*   **Entidad `Branch` (Sucursal / Sede):**
    *   `Email`: Correo exclusivo de contacto de la sede para el envío de comprobantes electrónicos.
    *   `IsMain`: Booleano para identificar la sede administrativa central.
    *   `Phone` y `Address`: Datos de contacto y dirección física.
    *   `Ubigeo`: Código geográfico de 6 dígitos (indispensable para facturación electrónica en Perú).
    *   `SerieFactura`: Prefijo de 4 dígitos para facturas emitidas desde esta sucursal (ej: `F001`).
    *   `SerieBoleta`: Prefijo de 4 dígitos para boletas emitidas desde esta sucursal (ej: `B001`).
*   **Entidad `User`:**
    *   `RoleType`: Enum de acceso directo para control de acceso:
        *   `SUPER_ADMIN`: Administrador de la plataforma SaaS (dueño del software, bypass total de multi-tenant para soporte y facturación).
        *   `COMPANY_ADMIN`: Administrador de la veterinaria (dueño del negocio, acceso total a todas las sedes de su empresa).
        *   `BRANCH_USER`: Empleado de veterinaria anclado a una o más sedes operativas específicas.

---

## 2. Capa de Seguridad y Aislamiento (Tenant Isolation)
Garantizar que los datos nunca se mezclen entre empresas o sucursales no autorizadas en ninguna consulta de base de datos.

### A. Implementación de GORM Scopes en el Backend
En lugar de inyectar filtros manuales en cada controlador SQL, se utilizarán Scopes de GORM para aplicar filtros automáticos en base al contexto de la petición (`gin.Context`).

```go
package config

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantFilter aplica automáticamente el aislamiento de datos por Empresa y Sucursal
func TenantFilter(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 1. Filtro obligatorio de Empresa
		companyID, exists := c.Get("companyID")
		if exists {
			db = db.Where("company_id = ?", companyID.(uuid.UUID))
		}

		// 2. Filtro opcional de Sucursal si el usuario no es Company Admin
		roleType, hasRole := c.Get("roleType")
		if hasRole && roleType.(string) == "BRANCH_USER" {
			branchID, hasBranch := c.Get("branchID")
			if hasBranch {
				db = db.Where("branch_id = ?", branchID.(uuid.UUID))
			}
		}

		return db
	}
}
```

* **Ejemplo de aplicación en handlers:**
```go
// Filtrado automático y transparente
config.DB.Scopes(config.TenantFilter(c)).Find(&products)
```

### B. RBAC (Role Based Access Control) y Middleware
* **Nivel 0 (SaaS System):** Middleware valida `SUPER_ADMIN` para rutas `/api/v1/saas-admin/*`.
* **Nivel 1 (Company):** Middleware valida `COMPANY_ADMIN` para rutas de reportes consolidados y creación de personal/sedes.
* **Nivel 2 (Branch):** Middleware valida accesos transaccionales (POS/Compras) limitados a la sede activa en el contexto.

---

## 3. Desarrollo del Panel de Administración SaaS (SuperAdmin)
Centro de mando global para gestionar el negocio de suscripciones.

*   **Ruta Frontend:** `/saas-admin` (Protegida en Vue Router y oculta en la barra de navegación para usuarios comunes).
*   **Funcionalidades:**
    *   **Dashboard Global:** Métricas de MRR (Ingresos Recurrentes Mensuales), cantidad de inquilinos activos, porcentaje de retención y volumen de comprobantes electrónicos procesados.
    *   **Gestión de Inquilinos:** Listado completo de empresas registradas con opciones para extender suscripciones, suspender cuentas (por falta de pago) y migrar planes.
    *   **Configuración de Planes:** Formulario para definir precios, límites técnicos (sucursales, usuarios, almacenamiento de imágenes) y periodos de prueba gratis.
    *   **Auditoría SaaS:** Registro centralizado de logs de auditoría de todas las empresas.

---

## 4. Control de Restricciones y Periodos de Prueba (Límites Técnicos)
*   **Onboarding Trial:** Al registrarse públicamente una empresa mediante `/companies/register`, el backend le asignará automáticamente:
    *   `PlanType = "Basic"`
    *   `MaxBranches = 1`
    *   `SubscriptionExpiresAt = time.Now().AddDate(0, 0, 30)` (30 días de prueba gratis).
*   **Validación de Sucursales:** En el handler `CreateBranch`, el backend consultará el número de sedes activas de la empresa. Si el conteo iguala o supera `MaxBranches`, denegará la petición devolviendo un código `403 Forbidden` informando que debe actualizar su plan de suscripción.

---

## 5. Experiencia de Usuario Multi-Sede (Frontend)
*   **Selector de Contexto (Login / Switcher):** 
    *   Si el usuario logueado posee rol `COMPANY_ADMIN` (o tiene acceso a múltiples sedes), al iniciar sesión aparecerá un modal modal solicitándole elegir la sede en la que operará en esa sesión.
    *   Se agregará un menú de cambio rápido de sucursal en el Navbar superior para facilitar el flujo sin tener que desloguearse.
*   **Reportes Consolidados (Dashboard BI):** 
    *   Comparativas gráficas en tiempo real del total de ventas entre locales para identificar las sucursales más rentables.

---

## Cronograma Sugerido de Implementación
1.  **Fase 1: Backend & Seguridad**
    *   Migración de base de datos con los nuevos campos (`PlanType`, `SubscriptionExpiresAt`, `MaxBranches`, `RoleType`, etc.).
    *   Implementación del Scope `TenantFilter` y middleware de control de accesos RBAC.
    *   Validaciones de límites comerciales (ej. bloqueo de creación de sedes al superar límites).
2.  **Fase 2: Panel SuperAdmin**
    *   Desarrollo de las API para el panel SaaS-Admin.
    *   Construcción de la interfaz de `/saas-admin` en el frontend para el control de empresas y licencias.
3.  **Fase 3: Refactorización del Panel de Usuario**
    *   Creación del componente Selector de Sucursal en el Login.
    *   Formularios para configurar datos geográficos/fiscales por sede y asignación de personal a turnos y locales en el frontend.

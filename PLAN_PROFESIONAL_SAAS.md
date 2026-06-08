# 📑 Plan de Implementación: Estructura SaaS Profesional - ERP Veterinaria

Este documento detalla la hoja de ruta para transformar el prototipo actual en un sistema ERP SaaS (Software as a Service) profesional, escalable y seguro.

---

## 1. Refactorización del Modelo de Datos (Cimientos)
Para soportar una operación profesional, el esquema de base de datos debe evolucionar para incluir control de licencias y detalles operativos por sede.

*   **Entidad `Company` (Empresa):**
    *   `PlanType`: (Basic, Premium, Enterprise).
    *   `SubscriptionExpiresAt`: Fecha de caducidad de la licencia.
    *   `MaxBranches`: Límite de sucursales según el plan contratado.
*   **Entidad `Branch` (Sucursal):**
    *   `Email`: Correo específico para comprobantes y contacto de la sede.
    *   `IsMain`: Indicador de si es la sede administrativa central.
    *   `Phone` y `Ubigeo/Address`: Datos geográficos y fiscales para facturación electrónica por sede.
*   **Entidad `User`:**
    *   `RoleType`: Clasificación clara entre `SUPER_ADMIN` (Dueño del software), `COMPANY_ADMIN` (Dueño de la veterinaria) y `BRANCH_USER` (Empleado de sede).

## 2. Capa de Seguridad y Aislamiento (Tenant Isolation)
Garantizar que los datos nunca se mezclen entre empresas o sucursales no autorizadas.

*   **Global Filter Middleware (Backend):** Interceptar todas las consultas SQL para inyectar automáticamente el `company_id` y, si el usuario no es admin, el `branch_id`.
*   **RBAC (Role Based Access Control):**
    *   **Nivel 0 (System):** Acceso total (Soporte técnico / SuperAdmin).
    *   **Nivel 1 (Company):** Acceso a todas las sedes de su propia empresa.
    *   **Nivel 2 (Branch):** Acceso restringido únicamente a la sede asignada.

## 3. Desarrollo del Panel de Administración SaaS (SuperAdmin)
Centro de mando para la gestión del negocio SaaS.

*   **Ruta:** `/saas-admin` (Protegida y oculta para clientes).
*   **Funcionalidades:**
    *   Dashboard de métricas globales e ingresos.
    *   Listado de empresas con estados de suscripción (Activo, Vencido, Suspendido).
    *   Gestión de planes y límites técnicos.
    *   Módulo de soporte y auditoría global.

## 4. Módulo de Gestión de Sucursales y Usuarios (Admin Empresa)
Herramientas para que el cliente gestione su organización.

*   **Gestión de Sedes:** Configuración de correos, series de facturación y datos de contacto por sucursal.
*   **Gestión de Personal:**
    *   Creación y edición de perfiles de empleados.
    *   Asignación de "Sede de Trabajo" (Anclada o Multisede).
    *   Control de acceso basado en turnos o roles.

## 5. Experiencia de Usuario Multi-Sede (Frontend)
*   **Selector de Contexto:** Al iniciar sesión, si el usuario tiene acceso a varias sedes, el sistema solicitará elegir la sede de operación actual.
*   **Reportes Consolidados:** Herramientas de Business Intelligence para que el dueño compare el rendimiento entre sus diferentes locales.

---

## Cronograma Sugerido
1.  **Fase 1:** Actualización de base de datos y lógica de aislamiento en el Backend.
2.  **Fase 2:** Interfaz de SuperAdmin para gestión de empresas y licencias.
3.  **Fase 3:** Refactorización del panel de usuario para gestión de sedes y personal.

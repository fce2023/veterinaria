Plan de Integración con FacturaAPI
Objetivo

Permitir que cualquier módulo del ERP pueda emitir:

Facturas
Boletas
Notas de crédito
Notas de débito
Guías de remisión

sin conocer los detalles técnicos de SUNAT.

Arquitectura
Frontend Vue

       ↓

ERP Core (Go)

       ↓

FacturaAPI Client

       ↓

FacturaAPI

       ↓

SUNAT
Módulo: FacturaAPI Client
Responsabilidades
Empresas

Sincronización automática:

Registrar Empresa
Buscar Empresa
Actualizar Configuración
Consultar Estado
Certificados
Validar Certificado
Subir Certificado
Actualizar Certificado
Consultar Estado
Comprobantes
Emitir Factura
Emitir Boleta
Emitir Nota Crédito
Emitir Nota Débito
Emitir GRE
Consultas
Consultar Estado
Descargar XML
Descargar PDF
Descargar CDR
Configuración ERP
Tabla: facturacion_config
id
empresa_id
api_url
api_key
tenant_uuid
modo
estado

Ejemplo:

empresa_id = 1

api_url = https://api.facturaapi.com

api_key = ********

tenant_uuid =
019e6cd4-1858-7290-8dc0-c65a822e6160

modo = produccion
Flujo de Onboarding
Paso 1

Registrar empresa en ERP.

Empresa
↓
Sucursal
↓
Usuarios
Paso 2

ERP llama:

POST /companies/register

Respuesta:

{
  "uuid": "019e..."
}

Guardar:

tenant_uuid
Paso 3

Subir certificado.

ERP:

POST /certificate
Paso 4

Validar.

POST /certificate/validate
Paso 5

Empresa lista para emitir.

Flujo de Venta
Usuario
POS
↓
Venta
↓
Confirmar
ERP

Guardar:

venta
detalle
stock
kardex
ERP → FacturaAPI
POST /documents/emit

Payload generado automáticamente.

Respuesta
{
  "uuid": "019e...",
  "status": "pending"
}

Guardar:

document_uuid
estado
Tabla documentos_electronicos
id
venta_id
empresa_id

document_uuid

tipo_documento

serie
numero

estado

xml_url
pdf_url
cdr_url

fecha_emision
Estados soportados
draft

pending

accepted

rejected

voided
Servicio de Sincronización

Cada cierto tiempo:

Consultar documentos pendientes
↓
Consultar estado FacturaAPI
↓
Actualizar ERP

Endpoints utilizados:

GET /documents/{uuid}/status
Descarga de Archivos

ERP:

GET /documents/{uuid}/files

Guardar:

pdf_url
xml_url
cdr_url

Mostrar:

Ver PDF
Descargar XML
Descargar CDR
Módulo de Empresas

Dentro del ERP:

Facturación Electrónica
Información
RUC
Razón Social
Dirección
Modo
Certificado
Estado

Válido
Vencido
No cargado
Logo
Subir Logo
Branding
Colores
Pie de página
Mensaje comercial
Módulo GRE

Cuando el usuario genera una guía:

Transferencia
↓
Generar GRE
↓
FacturaAPI
↓
SUNAT
Módulo API Interna

El frontend nunca debe consumir FacturaAPI directamente.

Siempre:

Vue
↓
ERP API
↓
FacturaAPI
Interfaces Go

Ejemplo conceptual:

type BillingProvider interface {
    RegisterCompany()
    UploadCertificate()
    EmitDocument()
    GetStatus()
    GetFiles()
    VoidDocument()
}

Implementación:

type FacturaAPIProvider struct {}

Esto te permitirá en el futuro soportar:

FacturaAPI
Nubefact
Efact
Otro proveedor

sin modificar el ERP.

Fase 1 + FacturaAPI
Módulos incluidos
Core
Usuarios
Roles
Permisos
Auditoría
Empresas
Multiempresa
Multisucursal
Inventario
Productos
Categorías
Stock
Kardex
Compras
Proveedores
Órdenes
Recepciones
Ventas
POS
Clientes
Caja
Facturación
Integración FacturaAPI
Facturas
Boletas
Notas
GRE
Descargas PDF/XML/CDR
Reportes
Ventas
Compras
Inventario
Dashboard
Indicadores principales
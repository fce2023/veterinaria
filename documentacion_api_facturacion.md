# FacturaAPI - Guía de Integración Enterprise

Bienvenido a la documentación oficial de **FacturaAPI**, una solución completa y robusta para la emisión de comprobantes electrónicos en Perú (SUNAT). 

Esta guía está diseñada para desarrolladores. Aquí encontrarás toda la información necesaria para integrar nuestra API en tus sistemas, desde la autenticación inicial hasta el paso a producción.

## Características Principales

*   **API RESTful:** Estándar de la industria, fácil de consumir desde cualquier lenguaje (PHP, Python, JS, Go, Java, etc.).
*   **Emisión Síncrona y Asíncrona:** Soporte para ambos modos mediante el parámetro `async`. Por defecto, la emisión es **asíncrona** (HTTP 202) en producción y **síncrona** (HTTP 201) en modo testing.
*   **Soporte Multiformato:** Facturas (01), Boletas (03), Notas de Crédito (07), Notas de Débito (08), y Guías de Remisión (09).
*   **Soporte de Detracciones:** Emisión de facturas sujetas a detracción con tipo de operación `1001` y leyenda obligatoria automática.
*   **Webhooks (Próximamente):** Notificaciones en tiempo real del estado de los documentos.
*   **Generación de PDF y XML:** Archivos con calidad de impresión (A4, Ticket) e interoperabilidad legal (XML UBL 2.1, CDR).
*   **Gestión de Empresas vía API:** Registro automático de nuevos emisores (tenants).

## Flujo Completo de Integración

Para enviar tu primer comprobante, sigue este flujo:

1.  **[Autenticación](#autenticacion):** Obtén tu API Key desde el Panel Admin.
2.  **[Configuración de Empresa](#gestion-y-configuracion-de-empresas):** Registra tu empresa (Emisor) vía API o en el Panel.
3.  **[Certificados](#gestion-de-certificados-digitales):** Sube y valida tu certificado digital (`.p12`, `.pfx` o `.pem`).
4.  **[Emisión de Documentos](#emision-y-gestion-de-documentos):** Envía el payload JSON con los datos de tu factura o boleta.
5.  **Consulta de Estado:** Monitorea el documento hasta obtener el estado `accepted`.
6.  **Descargas:** Obtén el XML, el PDF generado y el CDR.

## Entornos (Modos)

Nuestra API soporta dos entornos por empresa, configurables a nivel global o por documento:

*   **Beta (`modo: "beta"`):** Entorno de homologación y pruebas.
*   **Producción (`modo: "produccion"`):** Entorno en vivo. **Documentos con validez legal.**

# Autenticación

FacturaAPI utiliza validación mediante API Keys para autenticar las peticiones.

## Uso del API Key

Todas las peticiones a la API (rutas bajo `/api/v1/*`) deben incluir tu API Key en los headers HTTP.

### Headers Requeridos

```http
Authorization: Bearer {TU_API_KEY}
Accept: application/json
Content-Type: application/json
```

# Gestión y Configuración de Empresas

Puedes interactuar con las empresas emisoras (tenants) directamente a través de la API.

## 1. Registrar Empresa

Registra un nuevo emisor en el sistema.

**Endpoint:** `POST /api/v1/companies/register`

**Body (JSON):**
```json
{
    "ruc": "20123456789",
    "razon_social": "MI EMPRESA S.A.C.",
    "direccion": "AV. SIEMPRE VIVA 123",
    "modo": "beta"
}
```

**Respuesta (201 Created):**
```json
{
    "message": "Empresa registrada correctamente",
    "uuid": "019e6cd4-1858-7290-8dc0-c65a822e6160"
}
```

## 2. Buscar Empresa por RUC

Recupera el UUID de una empresa ya existente usando su número de RUC.

**Endpoint:** `GET /api/v1/companies/search?ruc=20123456789`

**Respuesta (200 OK):**
```json
{
    "uuid": "019e6cd4-1858-7290-8dc0-c65a822e6160",
    "ruc": "20123456789",
    "razon_social": "MI EMPRESA S.A.C."
}
```

## 3. Consultar Configuración

Retorna los datos de la empresa, incluyendo su estado y modo actual.

**Endpoint:** `GET /api/v1/config/{id}`

**Ejemplo de Respuesta (200 OK):**
```json
{
    "id": "019e6cd4-1858-7290-8dc0-c65a822e6160",
    "razon_social": "MI EMPRESA S.A.C.",
    "ruc": "20123456789",
    "modo": "beta"
}
```

## 4. Actualizar Logo

Sube o actualiza el logo comercial para las representaciones impresas.

**Endpoint:** `POST /api/v1/config/{id}/logo`

## 5. Actualizar Branding

Permite inyectar configuraciones de diseño (colores, fuentes) para el PDF.

**Endpoint:** `PATCH /api/v1/config/{id}/branding`

# Gestión de Certificados Digitales

FacturaAPI se encarga de la firma digital de los XML. Solo necesitas subir el certificado de la empresa.

## 1. Validar Certificado (Sin Guardar)

Valida la integridad y la contraseña de un certificado antes de realizar la subida definitiva.

**Endpoint:** `POST /api/v1/config/{id}/certificate/validate`

**Body (JSON):**
```json
{
    "certificate_base64": "MIIKrgIBAzCCCm8GCSqGSIb3DQEHAa...",
    "extension": "pfx",
    "password": "mi_clave_secreta"
}
```

## 2. Subir / Actualizar Certificado

Sube el certificado digital. Acepta tanto un archivo multipart como un string en Base64.

**Endpoint:** `POST /api/v1/config/{id}/certificate`

**Opción A — Multipart (archivo):**
```
POST /api/v1/config/{id}/certificate
Content-Type: multipart/form-data

certificate: [archivo .p12 / .pfx / .pem]
password: mi_clave_secreta
```
> La API detecta automáticamente la extensión del archivo subido (`p12`, `pfx`, `pem`, etc.) y la usa para nombrar el certificado guardado.

**Opción B — JSON (Base64):**
```json
{
    "certificate_base64": "MIIKrgIBAzCCCm8GCSqGSIb3DQEHAa...",
    "extension": "p12",
    "password": "mi_clave_secreta"
}
```
> El campo `extension` es opcional. Si se omite, se asume `pem` por defecto.

**Respuesta Exitosa (200 OK):**
```json
{
    "message": "Certificado subido correctamente"
}
```

# Emisión y Gestión de Documentos

El endpoint principal es `/api/v1/documents/emit`. Procesa Facturas, Boletas, Notas y Guías.

## 1. Emitir Documento

**Endpoint:** `POST /api/v1/documents/emit`

### Modos de Procesamiento (parámetro `async`)

La emisión soporta dos modos de procesamiento controlables con el parámetro `async`:

| Valor `async` | Comportamiento | Código HTTP |
|---|---|---|
| `true` (defecto en producción) | La emisión se delega a una cola de trabajo. La respuesta es inmediata. | **202 Accepted** |
| `false` | La emisión es síncrona. La respuesta contiene el resultado completo de SUNAT. | **201 Created** |

Si no se envía el parámetro `async`, el sistema detecta el entorno automáticamente: en producción usa colas (`202`); en modo de integración o testing, espera el resultado de SUNAT (`201`).

**Respuesta Asíncrona (202 — modo por defecto en producción):**
```json
{
    "message": "Documento en cola",
    "documento_id": "019e...",
    "status": "pending"
}
```

**Respuesta Síncrona (201 — cuando `async: false` o en entorno testing):**
```json
{
    "message": "Documento emitido",
    "documento_id": "019e...",
    "sunat_response": "La Factura numero F001-100, ha sido aceptada",
    "files": {
        "xml": "sunat/beta/20123456789/20123456789-01-F001-100.xml",
        "cdr": "sunat/beta/20123456789/R-20123456789-01-F001-100.zip"
    }
}
```

### Payload Base

```json
{
    "empresa_id": "UUID_DE_LA_EMPRESA",
    "async": false,
    "modo": "produccion", 
    "header": {
        "tipo_documento": "01", 
        "serie": "F001",
        "numero": 105,
        "fecha_emision": "2026-05-19"
    },
    "cliente": {
        "tipo_documento": "6",
        "numero_documento": "20123456789",
        "razon_social": "CLIENTE EJEMPLO S.A.C."
    },
    "items": [
        {
            "descripcion": "Servicio de Consultoría",
            "cantidad": 1,
            "precio_unitario": 1180.00,
            "subtotal": 1000.00
        }
    ],
    "totales": {
        "total_venta": 1000.00,
        "total_impuestos": 180.00
    }
}
```

> **Importante:** El campo `precio_unitario` debe contener el **precio de venta al público con IGV incluido**. El campo `subtotal` debe ser el valor sin IGV (base imponible). La API calcula los campos internos de Greenter de forma automática.

## 2. Consultar Estado

Obtiene el estado actual del documento en SUNAT.

**Endpoint:** `GET /api/v1/documents/{uuid}/status`

## 3. Obtener Enlaces de Descarga

Obtiene las URLs seguras para descargar el PDF, XML y CDR del documento. Las URLs apuntan al sistema de consulta pública, que requiere autenticación.

**Endpoint:** `GET /api/v1/documents/{uuid}/files`

**Respuesta (200 OK):**
```json
{
    "files": {
        "pdf": "https://apifacturacion.sehuacho.com/consulta/doc/{uuid}/pdf",
        "xml": "https://apifacturacion.sehuacho.com/consulta/doc/{uuid}/xml",
        "cdr": "https://apifacturacion.sehuacho.com/consulta/doc/{uuid}/cdr"
    }
}
```

## 4. Anular Documento (Baja)

Inicia el proceso para anular un comprobante ya aceptado.

**Endpoint:** `POST /api/v1/documents/{uuid}/void`

## 5. Eliminar Documentos de Prueba

Elimina definitivamente de la base de datos y del disco duro (archivos PDF y XML) un comprobante emitido.
**Importante:** Por seguridad, este endpoint *solo funciona* si el documento fue generado en `modo = "beta"`. Si intentas eliminar un documento de `produccion`, la API devolverá un error HTTP 403 (Forbidden).

**Endpoint:** `DELETE /api/v1/documents/{uuid}`

**Respuesta Exitosa (200 OK):**
```json
{
    "message": "Documento de prueba eliminado correctamente"
}
```

# Auto-cálculo Inteligente de Precios

Para facilitar la integración desde cualquier sistema o lenguaje, **FacturaAPI calcula automáticamente los valores unitarios internos** si solo envías `cantidad`, `precio_unitario` (con IGV) y `subtotal` (sin IGV). 

Los campos opcionales `valor_unitario` y `igv` pueden omitirse y la API los calculará con hasta 6 decimales de precisión para evitar el error SUNAT 3270 ("precio unitario difiere de los cálculos").

# Notas de Crédito (Tipo 07) y Débito (Tipo 08)

Para emitir una Nota de Crédito o Débito, debes especificar la referencia al documento original. La API admite dos formatos:

**Formato A — Bloque `referencia` a nivel raíz (recomendado):**
```json
{
    "header": { "tipo_documento": "07", "serie": "FC01", "numero": 124 },
    "referencia": {
        "tipo_documento": "01",
        "numero_documento": "F001-123",
        "codigo_motivo": "01",
        "descripcion_motivo": "Anulación de la operación"
    },
    "cliente": { "...": "..." },
    "items": [ "..." ]
}
```

**Formato B — Bloque `documento_referencia` y `motivo` dentro de `header` (alternativo):**
```json
{
    "header": {
        "tipo_documento": "07",
        "serie": "FC01",
        "numero": 124,
        "documento_referencia": {
            "tipo_documento": "01",
            "serie": "F001",
            "numero": 123
        },
        "motivo": {
            "codigo": "01",
            "descripcion": "Anulación de la operación"
        }
    },
    "cliente": { "...": "..." },
    "items": [ "..." ]
}
```

# Guías de Remisión Electrónica (GRE)

Las Guías de Remisión (código `09`) se procesan a través de la nueva API REST de SUNAT (no por SOAP como las facturas). FacturaAPI abstrae esta complejidad por ti.

**Requisito previo:** Debes tener configurados un *Client ID* y *Client Secret* para la API GRE desde tu portal SOL, y haberlos registrado en la empresa vía Panel o API.

La API admite dos formatos para los datos del envío:

**Formato A — Bloque `envio` a nivel raíz (recomendado):**
```json
{
    "header": { "tipo_documento": "09", "serie": "T001", "numero": 100 },
    "cliente": { "...": "..." },
    "envio": {
        "modalidad_traslado": "01",
        "motivo_traslado": "01",
        "descripcion_motivo": "Venta",
        "peso_total": 50.5,
        "unidad_peso": "KGM",
        "numero_bultos": 2,
        "partida": { "ubigeo": "150101", "direccion": "AV. PRINCIPAL 123" },
        "llegada": { "ubigeo": "150131", "direccion": "AV. LLEGADA 456" }
    },
    "items": [ { "descripcion": "Televisor", "cantidad": 2, "unidad": "NIU" } ]
}
```

**Formato B — Bloque `guia_detalles` dentro de `header` (alternativo):**
```json
{
    "header": {
        "tipo_documento": "09",
        "serie": "T001",
        "numero": 100,
        "guia_detalles": {
            "modalidad_traslado": "02",
            "motivo_traslado": "01",
            "peso_total": 15.5,
            "unidad_peso": "KGM",
            "partida": { "ubigeo": "150101", "direccion": "AV. ORIGEN 123" },
            "llegada": { "ubigeo": "150101", "direccion": "AV. DESTINO 456" }
        }
    },
    "cliente": { "...": "..." },
    "items": [ { "descripcion": "MUEBLES", "cantidad": 5 } ]
}
```

# Facturas con Detracciones (Sistema de Pago Adelantado)

Para emitir facturas sujetas al sistema de detracciones de SUNAT, agrega el bloque `detraccion` al payload. La API configurará automáticamente el tipo de operación UBL a `1001` y añadirá la leyenda obligatoria "Operación sujeta a detracción".

```json
{
    "header": { "tipo_documento": "01", "serie": "F001", "numero": 500 },
    "cliente": { "...": "..." },
    "items": [
        {
            "descripcion": "Servicio de Consultoría",
            "cantidad": 1,
            "precio_unitario": 1180.00,
            "subtotal": 1000.00
        }
    ],
    "totales": { "total_venta": 1000.00, "total_impuestos": 180.00 },
    "detraccion": {
        "codigo_bien_servicio": "022",
        "codigo_medio_pago": "001",
        "porcentaje": 12.00,
        "monto": 141.60
    }
}
```

> **Nota:** La empresa emisora debe tener configurada su cuenta bancaria de detracciones en el campo `cuenta_detracciones` para que se incluya en el XML correctamente.



# NOTA IMPORTANTE

En un entorno SaaS real, los roles se dividen precisamente
  así:

  1. El Administrador del SaaS configura la API Key global en
  el backend de su servidor. Esta llave asegura que nadie más pueda
  consumir el microservicio, solo tu SaaS.
  2. Los Inquilinos (Tenants) entran a su panel de
  configuración dentro de tu SaaS y allí ingresan sus datos: su
  RUC, sus credenciales SOL y suben su certificado digital  .  
  p12 .
  3. El SaaS toma esos archivos y datos del inquilino y los
  envía automáticamente al microservicio mediante la API para
  guardarlos en el registro de su empresa, quedando
  listos para firmar y emitir comprobantes a la SUNAT.
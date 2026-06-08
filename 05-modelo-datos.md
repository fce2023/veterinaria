# Modelo de Datos

## Convenciones

Todas las tablas deben incluir:

* id (UUID)
* created_at
* updated_at
* deleted_at (soft delete)
* empresa_id (cuando aplique)

---

# Seguridad

## users

* id
* empresa_id
* sucursal_id
* nombre
* email
* username
* password_hash
* estado

## roles

* id
* nombre
* descripcion

## permissions

* id
* codigo
* descripcion

## role_permissions

* role_id
* permission_id

## user_roles

* user_id
* role_id

---

# Empresas

## companies

* id
* ruc
* razon_social
* nombre_comercial
* direccion
* telefono
* email
* estado

## branches

* id
* empresa_id
* nombre
* direccion
* telefono
* estado

---

# Inventario

## categories

* id
* empresa_id
* nombre

## brands

* id
* empresa_id
* nombre

## products

* id
* empresa_id
* category_id
* brand_id
* codigo
* codigo_barras
* nombre
* descripcion
* precio_compra
* precio_venta
* stock_minimo
* estado

## stock

* id
* empresa_id
* sucursal_id
* producto_id
* stock_actual

## kardex

* id
* empresa_id
* sucursal_id
* producto_id
* tipo_movimiento
* referencia
* cantidad
* stock_anterior
* stock_nuevo

---

# Compras

## suppliers

* id
* empresa_id
* ruc
* razon_social
* direccion
* telefono

## purchases

* id
* empresa_id
* sucursal_id
* proveedor_id
* fecha
* subtotal
* igv
* total
* estado

## purchase_items

* id
* compra_id
* producto_id
* cantidad
* costo_unitario

---

# Ventas

## customers

* id
* empresa_id
* tipo_documento
* numero_documento
* nombre
* direccion
* telefono

## sales

* id
* empresa_id
* sucursal_id
* cliente_id
* usuario_id
* subtotal
* igv
* total
* estado

## sale_items

* id
* venta_id
* producto_id
* cantidad
* precio_unitario
* descuento

---

## billing_configs

* id
* empresa_id
* api_url
* api_key
* tenant_uuid
* modo
* estado

## electronic_documents

* id
* empresa_id
* venta_id
* document_uuid
* tipo_documento
* serie
* numero
* estado
* pdf_url
* xml_url
* cdr_url

---

# Auditoría

## audit_logs

* id
* usuario_id
* modulo
* accion
* descripcion
* ip
* fecha

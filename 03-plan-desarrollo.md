FASE 1 - ERP CORE (MVP COMERCIAL)
Objetivo

Tener un sistema que permita:

Gestionar múltiples empresas.
Gestionar múltiples sucursales.
Controlar inventario.
Realizar compras.
Registrar ventas.
Conectarse a tu API de facturación.
Mostrar reportes básicos.

Con esto ya puedes venderlo a:

Veterinarias
Farmacias
Ferreterías
Minimarkets
Restaurantes
Distribuidoras
Módulo 1: Seguridad y Accesos
Usuarios

Campos:

id
nombre
email
usuario
password_hash
estado
fecha_creacion

Funciones:

Login
Logout
Cambio de contraseña
Recuperación de contraseña
Roles

Ejemplos:

SuperAdmin
Administrador
Gerente
Cajero
Almacenero
Comprador
Permisos

Ejemplos:

productos.ver
productos.crear
productos.editar
productos.eliminar

ventas.ver
ventas.crear

compras.ver
compras.crear

reportes.ver
Auditoría

Registrar:

usuario
acción
fecha
IP
módulo

Ejemplo:

Juan Perez
Creó Producto
2026-06-08 15:10
Módulo 2: Multiempresa
Empresas

Tabla:

id
ruc
razon_social
nombre_comercial
direccion
telefono
email
estado
Sucursales

Tabla:

id
empresa_id
nombre
direccion
telefono
estado

Ejemplo:

Veterinaria San Martín

Sucursal Centro
Sucursal Plaza Sol
Sucursal Huaura
Módulo 3: Inventario
Categorías
Medicamentos
Vacunas
Accesorios
Alimentos
Marcas
Royal Canin
Purina
Bayer
Zoetis
Productos

Campos:

id
codigo
codigo_barras
nombre
categoria_id
marca_id
descripcion
precio_compra
precio_venta
stock_minimo
estado
Stock por sucursal

Tabla fundamental:

producto_id
sucursal_id
stock_actual
Kardex

Todo movimiento debe registrarse.

Campos:

fecha
producto
tipo_movimiento

INGRESO
VENTA
AJUSTE
TRANSFERENCIA

cantidad
stock_anterior
stock_nuevo
Módulo 4: Compras
Proveedores
ruc
razon_social
telefono
email
direccion
Ordenes de Compra

Cabecera:

numero
proveedor
fecha
estado

Detalle:

producto
cantidad
costo
subtotal
Recepción de Mercadería

Al confirmar:

actualizar stock
registrar kardex
Módulo 5: Clientes

Campos:

tipo_documento
numero_documento
nombre
direccion
telefono
email
Módulo 6: Ventas
Cabecera
serie
numero
cliente
fecha
subtotal
igv
total
estado
Detalle
producto
cantidad
precio
descuento
subtotal
Al confirmar

Debe:

descontar stock
registrar kardex
llamar API Facturación
guardar respuesta
Integración con tu Microservicio

Tu ERP nunca genera XML.

Solo envía:

{
  "empresa_id": 1,
  "tipo_comprobante": "01",
  "cliente": {},
  "items": []
}

Tu microservicio responde:

{
  "estado": "aceptado",
  "cdr": "...",
  "hash": "...",
  "pdf": "...",
  "xml": "..."
}

El ERP solo almacena la respuesta.

Módulo 7: Dashboard
Ventas
Hoy
Semana
Mes
Compras
Mes actual
Inventario
Productos bajo stock
Top productos
Más vendidos
Módulo 8: Reportes
Ventas

Filtros:

Fecha
Sucursal
Compras

Filtros:

Fecha
Proveedor
Inventario
Stock actual
Stock crítico
Base de Datos (mínima)
usuarios
roles
permisos

empresas
sucursales

categorias
marcas
productos

stock_sucursal
kardex

proveedores
compras
compra_detalle

clientes

ventas
venta_detalle

auditoria
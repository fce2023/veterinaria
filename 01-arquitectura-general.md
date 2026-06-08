# Arquitectura General del ERP SaaS Veterinario

## Resumen

Este proyecto consiste en el desarrollo de una plataforma ERP SaaS orientada inicialmente a veterinarias, pero diseñada desde su núcleo para ser reutilizable en otros sectores como farmacias, restaurantes, minimarkets, ferreterías y comercios en general.

La plataforma será multiempresa, multisucursal y modular.

La facturación electrónica será gestionada mediante un microservicio independiente denominado FacturaAPI.

---

# Documentos del Proyecto

La documentación se divide en los siguientes archivos:

## 1. 03-plan-desarrollo.md

Describe:

* Fases de desarrollo.
* Módulos del ERP.
* Roadmap.
* Prioridades de implementación.
* MVP comercial.

Debe utilizarse como referencia funcional.

---

## 2. 02-tecnologias-usar.md

Describe:

* Stack tecnológico.
* Arquitectura frontend.
* Arquitectura backend.
* Base de datos.
* Infraestructura.
* Seguridad.
* Despliegue.

Debe utilizarse como referencia técnica.

---

## 3. 04-integracion-facturaapi.md

Describe:

* Integración con FacturaAPI.
* Flujo de emisión.
* Gestión de empresas.
* Certificados.
* Comprobantes electrónicos.
* GRE.

Debe utilizarse como referencia de integración tributaria.

---

# Objetivo de la Plataforma

La plataforma deberá permitir:

* Multiempresa.
* Multisucursal.
* Control de usuarios.
* Control de roles.
* Inventario.
* Compras.
* Ventas.
* Caja.
* Reportes.
* Dashboard.
* Integración con FacturaAPI.

Posteriormente:

* Historia clínica.
* Agenda veterinaria.
* Vacunación.
* Automatizaciones.
* WhatsApp.
* CRM.

---

# Arquitectura General

Frontend Vue

↓

API ERP Core (Go)

↓

PostgreSQL

↓

Redis

↓

FacturaAPI

↓

SUNAT

---

# Principios de Desarrollo

## Modularidad

Cada módulo debe ser independiente.

Ejemplos:

* Auth
* Users
* Companies
* Inventory
* Purchases
* Sales
* Reports

No debe existir lógica de negocio duplicada.

---

## Multiempresa

Todas las entidades operativas deben pertenecer a una empresa.

Ejemplo:

empresa_id

Esto permitirá operar múltiples clientes en una misma instalación SaaS.

---

## Multisucursal

Todos los movimientos operativos deben identificar la sucursal.

Ejemplo:

sucursal_id

Aplicable a:

* Inventario
* Compras
* Ventas
* Caja
* Reportes

---

## Auditoría

Toda acción importante debe registrarse.

Ejemplos:

* Crear producto.
* Editar producto.
* Registrar compra.
* Registrar venta.
* Anular documento.

---

## API First

Toda funcionalidad deberá exponerse mediante API REST.

El frontend nunca accederá directamente a la base de datos.

---

# Estructura de Módulos

ERP Core

├── Auth

├── Users

├── Roles

├── Permissions

├── Audit

├── Companies

├── Branches

├── Inventory

├── Purchases

├── Suppliers

├── Customers

├── Sales

├── POS

├── Dashboard

├── Reports

├── BillingClient

└── Settings

---

# Flujo de Venta

Venta

↓

Validar Stock

↓

Registrar Kardex

↓

Guardar Venta

↓

Emitir Documento

↓

FacturaAPI

↓

SUNAT

↓

Guardar Estado

---

# Flujo de Compra

Orden Compra

↓

Recepción

↓

Actualizar Stock

↓

Registrar Kardex

↓

Actualizar Costos

---

# Flujo de Inventario

Movimiento

↓

Kardex

↓

Stock Actual

↓

Reportes

---

# Objetivo MVP

La primera versión comercial debe incluir:

* Usuarios
* Roles
* Empresas
* Sucursales
* Productos
* Categorías
* Compras
* Proveedores
* Ventas
* Clientes
* Kardex
* Caja
* Facturación Electrónica
* Dashboard
* Reportes

No incluir módulos veterinarios avanzados en la primera liberación.

---

# Objetivo Fase 2

Agregar:

* Transferencias
* Lotes
* Vencimientos
* Inventario físico
* Dashboard avanzado

---

# Objetivo Fase 3

Agregar:

* Mascotas
* Historia clínica
* Tratamientos
* Vacunas
* Agenda

---

# Objetivo Final

Construir una plataforma ERP SaaS modular y escalable que permita administrar múltiples negocios desde una sola base tecnológica, utilizando FacturaAPI como servicio centralizado para la facturación electrónica.

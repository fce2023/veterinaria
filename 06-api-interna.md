# API Interna ERP

Base URL

/api/v1

---

# Auth

POST /auth/login

POST /auth/logout

POST /auth/refresh

GET /auth/me

---

# Usuarios

GET /users

POST /users

GET /users/{id}

PUT /users/{id}

DELETE /users/{id}

---

# Roles

GET /roles

POST /roles

PUT /roles/{id}

DELETE /roles/{id}

---

# Empresas

GET /companies

POST /companies

GET /companies/{id}

PUT /companies/{id}

---

# Sucursales

GET /branches

POST /branches

PUT /branches/{id}

DELETE /branches/{id}

---

# Productos

GET /products

POST /products

GET /products/{id}

PUT /products/{id}

DELETE /products/{id}

---

# Categorías

GET /categories

POST /categories

PUT /categories/{id}

DELETE /categories/{id}

---

# Inventario

GET /inventory/stock

GET /inventory/kardex

POST /inventory/adjustment

---

# Compras

GET /purchases

POST /purchases

GET /purchases/{id}

POST /purchases/{id}/receive

---

# Proveedores

GET /suppliers

POST /suppliers

PUT /suppliers/{id}

DELETE /suppliers/{id}

---

# Clientes

GET /customers

POST /customers

PUT /customers/{id}

DELETE /customers/{id}

---

# Ventas

GET /sales

POST /sales

GET /sales/{id}

POST /sales/{id}/cancel

---

# Caja

POST /cash/open

POST /cash/close

GET /cash/history

---

# Dashboard

GET /dashboard/general

GET /dashboard/sales

GET /dashboard/inventory

---

# Reportes

GET /reports/sales

GET /reports/purchases

GET /reports/inventory

---

# Facturación

POST /billing/emit

GET /billing/status/{uuid}

GET /billing/files/{uuid}

POST /billing/void/{uuid}

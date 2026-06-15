# Design Spec: SUNAT Business Address Configuration

## Goals
Allow configuring missing business address metadata (**Ubigeo**, **Departamento**, **Provincia**, **Distrito**) for the issuer/company (RUC `10721837811` and others) to prevent SUNAT warnings (4096, 4093, 4097, 4098) in generated XMLs.

## Proposed Solution (Enfoque 1)
1. **Database / Backend Model Update**:
   - Extend the `Company` model in `backend/models/models.go` to store `Ubigeo`, `Departamento`, `Provincia`, and `Distrito`. GORM `AutoMigrate` on startup will add these fields to the `companies` table.
2. **Backend Handler Update**:
   - Modify the `UpdateMyCompany` handler in `backend/handlers/handlers.go` to accept, validate, and persist the new location fields.
   - Sync the company's `Ubigeo` field to the main branch (`IsMain == true`) in the `branches` table to maintain consistency.
3. **Frontend UI Update**:
   - In `BusinessSection.vue`:
     - Add input fields for `Ubigeo` (max 6 digits), `Departamento`, `Provincia`, and `Distrito`.
     - Add a "Consultar SUNAT" query button next to the locked/disabled RUC input.
     - When clicked, this button triggers `axios.get('/public/ruc/' + business.ruc)`, populating the fields automatically with the results from the RUC lookup.
     - Maintain consistent layout/design matching the existing typography and form grids.

## Security Controls (SecureCoder Guidelines)
- **SQL Injection Prevention**: All database operations use the GORM ORM (`DB.Save`, `DB.First`), which uses prepared statements and parameterized queries.
- **XSS Prevention**: Frontend rendering uses Vue 3 framework-native interpolation (`v-model` binding on text inputs) which automatically escapes values.
- **Input Validation**: Check that `ubigeo` is sanitized (trimmed) and optionally validates against a 6-digit alphanumeric format.
- **Least Privilege**: The handler operates within the scope of the authenticated client tenant (`companyID` from the JWT session claims).

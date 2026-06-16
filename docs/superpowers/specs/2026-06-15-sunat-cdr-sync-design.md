# SUNAT CDR Message Synchronization - Design Specification

We will implement a mechanism to fetch, unzip, and parse the SUNAT Constancia de Recepción (CDR) ZIP file on the backend when the FacturaAPI status response doesn't provide a detailed `sunat_response`. In the frontend, we will automatically trigger synchronization when opening the messages modal if it's empty, and provide manual sync buttons.

---

## 1. Backend Design

### 1.1 `BillingService` Changes
File: `backend/services/billing_service.go`
* Import `"archive/zip"`
* Implement `GetSunatMessageFromCDR(apiURL, apiKey, docUUID string) (string, error)`:
  1. Perform a HTTP GET call to `/api/v1/documents/{uuid}/files` with the authorization header.
  2. Parse the CDR file download URL (`files.cdr` or `data.files.cdr`).
  3. Perform a HTTP GET call to download the ZIP file.
  4. Unzip in memory, locate the XML file, read its content, and pass it to the existing `ParseSunatResponse` method to extract notes/descriptions.

### 1.2 Handler Changes
File: `backend/handlers/billing.go`
* Modify `SyncElectronicDocumentStatus`:
  1. Capture additional JSON response fields (`message`, `error`) from the status API.
  2. If the parsed `SunatResponse` is empty and the status is `accepted` or `rejected`, call `GetSunatMessageFromCDR`.
  3. Update `doc.SunatResponse` with the parsed message/error and save to database.

---

## 2. Frontend Design

File: `frontend/src/components/BillingSection.vue`
* **Table List**: Show the inline sync button for all statuses (not just `pending`/`error`), so users can refresh any document.
* **Message Modal**:
  - If `selectedDoc.sunat_response` is empty:
    - Display a clear placeholder and a sync button: `"Consultar/Actualizar desde SUNAT"`.
    - Automatically trigger the sync action when the modal is opened if the message is empty.
  - Implement `syncModalDoc(id)` to call `syncDocumentStatus(id)` and update `selectedDoc` reactively.

---

## 3. Verification Plan

1. **Unit Tests**:
   - Write a unit test in `backend/services/billing_service_test.go` or a new file to verify unzipping and parsing a simulated CDR ZIP.
2. **Integration Verification**:
   - Trigger the sync API manually or via the UI for an accepted document with empty `sunat_response` to verify it fetches the CDR, unzips/parses the XML, and populates the database.

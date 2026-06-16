package services

import (
	"log"
	"time"

	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

// StartBillingSyncWorker starts a background worker that periodically
// checks for pending documents and syncs their status with FacturaAPI.
func StartBillingSyncWorker() {
	go func() {
		// Run every 10 seconds
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			syncPendingDocuments()
		}
	}()
}

func syncPendingDocuments() {
	var docs []models.ElectronicDocument
	
	// Find documents that need syncing: pending, void_pending or empty/unknown status
	err := config.DB.Where("estado IN (?) OR estado IS NULL OR estado = ''", []string{"pending", "void_pending"}).Find(&docs).Error
	if err != nil {
		log.Printf("[Billing Worker] Error fetching pending documents: %v", err)
		return
	}

	if len(docs) == 0 {
		return
	}

	billingService := NewBillingService()

	for _, doc := range docs {
		// Skip documents without UUID
		if doc.DocumentUUID == nil || *doc.DocumentUUID == "" {
			continue
		}

		// Re-fetch document to ensure we have the latest state before syncing
		var currentDoc models.ElectronicDocument
		if err := config.DB.First(&currentDoc, doc.ID).Error; err != nil {
			continue
		}
		
		// Ensure it's still in a syncable state (pending, void_pending or empty)
		if currentDoc.Estado != "pending" && currentDoc.Estado != "void_pending" && currentDoc.Estado != "" {
			continue
		}

		_, err := billingService.SyncDocumentStatus(currentDoc.CompanyID, &currentDoc)
		if err != nil {
			log.Printf("[Billing Worker] Error syncing document %s: %v", *currentDoc.DocumentUUID, err)
			continue
		}
	}
}

package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

// GetActiveCashSession returns the currently open session for the user
func GetActiveCashSession(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var session models.CashSession
	if err := config.DB.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&session).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": session})
}

// OpenCashSession starts a new cash session
func OpenCashSession(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)
	userID := c.MustGet("userID").(uuid.UUID)

	var input struct {
		SaldoInicial float64 `json:"saldo_inicial"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if already has an open session
	var existing models.CashSession
	if err := config.DB.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Ya tienes una caja abierta"})
		return
	}

	session := models.CashSession{
		CompanyID:    companyID,
		BranchID:     branchID,
		UserID:       userID,
		OpenedAt:     time.Now(),
		SaldoInicial: input.SaldoInicial,
		Estado:       "OPEN",
	}

	if err := config.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo abrir la caja"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": session})
}

// CloseCashSession closes the active session
func CloseCashSession(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var input struct {
		SaldoFinalReal float64 `json:"saldo_final_real"`
		Notas          string  `json:"notas"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var session models.CashSession
	if err := config.DB.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No hay una caja abierta para cerrar"})
		return
	}

	now := time.Now()
	session.ClosedAt = &now
	session.SaldoFinalReal = input.SaldoFinalReal
	session.Notas = input.Notas
	session.Estado = "CLOSED"

	if err := config.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al cerrar la caja"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Caja cerrada correctamente", "data": session})
}

// CreateCashMovement registers a manual income or expense
func CreateCashMovement(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	branchID := c.MustGet("branchID").(uuid.UUID)

	var input struct {
		Tipo   string  `json:"tipo" binding:"required"` // INGRESO, EGRESO
		Monto  float64 `json:"monto" binding:"required"`
		Motivo string  `json:"motivo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var session models.CashSession
	if err := config.DB.Where("user_id = ? AND branch_id = ? AND estado = 'OPEN'", userID, branchID).First(&session).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Debe abrir caja antes de registrar movimientos"})
		return
	}

	movement := models.CashMovement{
		CashSessionID: session.ID,
		Tipo:          input.Tipo,
		Monto:         input.Monto,
		Motivo:        input.Motivo,
	}

	tx := config.DB.Begin()
	if err := tx.Create(&movement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al crear movimiento"})
		return
	}

	// Update session totals
	if input.Tipo == "INGRESO" {
		session.TotalIngresos += input.Monto
	} else {
		session.TotalEgresos += input.Monto
	}

	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al actualizar totales de caja"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": movement})
}

// GetCashSessions returns a list of previous sessions
func GetCashSessions(c *gin.Context) {
	branchID := c.MustGet("branchID").(uuid.UUID)

	var sessions []models.CashSession
	if err := config.DB.Where("branch_id = ?", branchID).Order("opened_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": sessions})
}

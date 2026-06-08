package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/auth"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

type SwitchBranchInput struct {
	BranchID uuid.UUID `json:"branch_id" binding:"required"`
}

func SwitchBranch(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	userID := c.MustGet("userID").(uuid.UUID)
	roleType := c.GetString("roleType")

	var input SwitchBranchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 1. Verify that the branch belongs to the user's company
	var branch models.Branch
	if err := config.DB.Where("id = ? AND company_id = ?", input.BranchID, companyID).First(&branch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Branch not found or unauthorized"})
		return
	}

	// 2. Generate a new JWT with updated branch context
	newToken, err := auth.GenerateToken(userID, companyID, input.BranchID, roleType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not switch branch session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":     newToken,
			"branch_id": input.BranchID,
		},
	})
}

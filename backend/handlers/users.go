package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"veterinaria/backend/auth"
	"veterinaria/backend/config"
	"veterinaria/backend/models"
)

type CreateUserInput struct {
	Nombre   string    `json:"nombre" binding:"required"`
	Email    string    `json:"email" binding:"required,email"`
	Username string    `json:"username" binding:"required"`
	Password string    `json:"password" binding:"required,min=8"`
	RoleType string    `json:"role_type" binding:"required"`
	BranchID uuid.UUID `json:"branch_id" binding:"required"`
}

type UpdateUserInput struct {
	Nombre   string `json:"nombre" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	RoleType string `json:"role_type" binding:"required"`
	Estado   string `json:"estado" binding:"required"`
}

func GetUsers(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	roleType := c.GetString("roleType")

	var users []models.User
	query := config.DB.Where("company_id = ?", companyID)

	if roleType == "BRANCH_USER" || roleType == "BRANCH_ADMIN" {
		branchID, exists := c.Get("branchID")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "No branch assigned"})
			return
		}
		query = query.Where("branch_id = ?", branchID.(uuid.UUID))
	}

	if err := query.Order("created_at desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Could not fetch users"})
		return
	}

	// Remove password hashes from response
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": users})
}

func CreateUser(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	callerRole := c.GetString("roleType")

	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos: " + err.Error()})
		return
	}

	// Security: Authorization Checks
	if callerRole != "COMPANY_ADMIN" && callerRole != "BRANCH_ADMIN" && callerRole != "SUPER_ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Insuficientes permisos para crear usuarios"})
		return
	}

	if callerRole == "BRANCH_ADMIN" {
		callerBranchID, exists := c.Get("branchID")
		if !exists || callerBranchID.(uuid.UUID) != input.BranchID {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "No puede crear usuarios en otra sucursal"})
			return
		}
		if input.RoleType != "CASHIER" && input.RoleType != "BRANCH_USER" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Sólo puede crear usuarios de tipo CAJERO (CASHIER)"})
			return
		}
	}

	// Sanitize and Validate
	input.RoleType = strings.ToUpper(strings.TrimSpace(input.RoleType))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate Branch belongs to Company
	var branch models.Branch
	if err := config.DB.Where("id = ? AND company_id = ?", input.BranchID, companyID).First(&branch).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "La sucursal indicada no existe o no pertenece a la empresa"})
		return
	}

	// Check if user already exists
	var existing models.User
	if err := config.DB.Where("email = ? OR username = ?", input.Email, input.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "El correo o nombre de usuario ya está registrado"})
		return
	}

	pwdHash, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error al procesar la contraseña"})
		return
	}

	user := models.User{
		CompanyID:    companyID,
		BranchID:     input.BranchID,
		Nombre:       input.Nombre,
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: pwdHash,
		Estado:       "active",
		RoleType:     input.RoleType,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo crear el usuario"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": user})
}

func UpdateUser(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	callerRole := c.GetString("roleType")
	userID := c.Param("id")

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Datos inválidos: " + err.Error()})
		return
	}

	if callerRole != "COMPANY_ADMIN" && callerRole != "SUPER_ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Solo el administrador de empresa puede editar usuarios"})
		return
	}

	var user models.User
	if err := config.DB.Where("id = ? AND company_id = ?", userID, companyID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Usuario no encontrado"})
		return
	}

	// Check duplicate email or username for other users
	var duplicate models.User
	if err := config.DB.Where("(email = ? OR username = ?) AND id != ?", input.Email, input.Username, userID).First(&duplicate).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "El correo o nombre de usuario ya está en uso por otra cuenta"})
		return
	}

	user.Nombre = input.Nombre
	user.Email = strings.ToLower(strings.TrimSpace(input.Email))
	user.Username = input.Username
	user.RoleType = strings.ToUpper(strings.TrimSpace(input.RoleType))
	user.Estado = input.Estado

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo actualizar el usuario"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

func DeleteUser(c *gin.Context) {
	companyID := c.MustGet("companyID").(uuid.UUID)
	callerRole := c.GetString("roleType")
	userID := c.Param("id")

	if callerRole != "COMPANY_ADMIN" && callerRole != "SUPER_ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Solo el administrador de empresa puede eliminar usuarios"})
		return
	}

	// Soft delete
	if err := config.DB.Where("id = ? AND company_id = ?", userID, companyID).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "No se pudo eliminar el usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Usuario eliminado correctamente"})
}

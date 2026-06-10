package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func getJWTSecret() []byte {
	if len(jwtSecret) == 0 {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "veterinaria_default_secret_key_2026"
		}
		jwtSecret = []byte(secret)
	}
	return jwtSecret
}

// Claims defines the JWT payload structure
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	CompanyID uuid.UUID `json:"company_id"`
	BranchID  uuid.UUID `json:"branch_id"`
	RoleType  string    `json:"role_type"`
	jwt.RegisteredClaims
}

// HashPassword hashes a raw password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPasswordHash compares password hash with raw input
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a JWT for a user session
func GenerateToken(userID, companyID, branchID uuid.UUID, roleType string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:    userID,
		CompanyID: companyID,
		BranchID:  branchID,
		RoleType:  roleType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ParseToken parses and validates a token string
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AuthMiddleware intercepts requests and validates JWT bearer tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Authorization format must be Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Inject details into Gin context
		c.Set("userID", claims.UserID)
		c.Set("companyID", claims.CompanyID)
		c.Set("branchID", claims.BranchID)
		c.Set("roleType", claims.RoleType)

		c.Next()
	}
}

// SuperAdminMiddleware restricts routes to SUPER_ADMIN users
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleType, exists := c.Get("roleType")
		if !exists || roleType.(string) != "SUPER_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso denegado: Se requieren privilegios de Super Administrador"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CompanyAdminMiddleware restricts routes to COMPANY_ADMIN and SUPER_ADMIN users
func CompanyAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleType, exists := c.Get("roleType")
		if !exists || (roleType.(string) != "SUPER_ADMIN" && roleType.(string) != "COMPANY_ADMIN") {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso denegado: Se requieren privilegios de Administrador de Empresa"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// BranchAdminMiddleware restricts routes to BRANCH_ADMIN, COMPANY_ADMIN and SUPER_ADMIN users
func BranchAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleType, exists := c.Get("roleType")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso denegado"})
			c.Abort()
			return
		}
		
		role := roleType.(string)
		if role != "SUPER_ADMIN" && role != "COMPANY_ADMIN" && role != "BRANCH_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Acceso denegado: Se requieren privilegios de Encargado de Sucursal o superior"})
			c.Abort()
			return
		}
		c.Next()
	}
}


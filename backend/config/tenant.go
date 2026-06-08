package config

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantFilter automatically applies data isolation by Company and Branch
func TenantFilter(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 1. Check if user is SUPER_ADMIN - if so, bypass tenant isolation
		roleType, hasRole := c.Get("roleType")
		if hasRole && roleType.(string) == "SUPER_ADMIN" {
			return db
		}

		// 2. Mandatory Company isolation
		companyID, exists := c.Get("companyID")
		if exists {
			db = db.Where("company_id = ?", companyID.(uuid.UUID))
		}

		// 3. Branch isolation for regular users
		if hasRole && roleType.(string) == "BRANCH_USER" {
			branchID, hasBranch := c.Get("branchID")
			if hasBranch {
				db = db.Where("branch_id = ?", branchID.(uuid.UUID))
			}
		}

		return db
	}
}

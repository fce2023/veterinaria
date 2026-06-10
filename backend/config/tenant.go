package config

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantFilter automatically applies data isolation by Company only
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

		return db
	}
}

// BranchFilter applies both Company isolation and Branch isolation
func BranchFilter(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 1. Check if user is SUPER_ADMIN
		roleType, hasRole := c.Get("roleType")
		if hasRole && roleType.(string) == "SUPER_ADMIN" {
			return db
		}

		// 2. Mandatory Company isolation
		companyID, exists := c.Get("companyID")
		if exists {
			db = db.Where("company_id = ?", companyID.(uuid.UUID))
		}

		// 3. Branch isolation for regular users and branch admins
		if hasRole {
			role := roleType.(string)
			if role == "BRANCH_USER" || role == "BRANCH_ADMIN" || role == "CASHIER" {
				branchID, hasBranch := c.Get("branchID")
				if hasBranch {
					db = db.Where("branch_id = ?", branchID.(uuid.UUID))
				}
			}
		}

		return db
	}
}

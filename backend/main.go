package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"veterinaria/backend/auth"
	"veterinaria/backend/config"
	"veterinaria/backend/handlers"
	"veterinaria/backend/models"
)

func main() {
	log.Println("Starting ERP Core Backend...")

	// 1. Connect to Database & Redis
	config.ConnectDB()
	config.ConnectRedis()

	// 2. Auto-migrate DB Schema
	log.Println("Running database migrations...")
	err := config.DB.AutoMigrate(
		&models.Company{},
		&models.Branch{},
		&models.Permission{},
		&models.Role{},
		&models.User{},
		&models.Category{},
		&models.Brand{},
		&models.Product{},
		&models.Stock{},
		&models.Kardex{},
		&models.Supplier{},
		&models.Purchase{},
		&models.PurchaseItem{},
		&models.Customer{},
		&models.Sale{},
		&models.SaleItem{},
		&models.BillingConfig{},
		&models.ElectronicDocument{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrations completed successfully.")

	// 3. Seed Database with Initial Data
	seedDatabase()

	// 4. Setup Gin Router
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()

	// CORS Setup
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API Routing Group
	v1 := router.Group("/api/v1")
	{
		// Public Routes
		v1.POST("/auth/login", handlers.Login)
		v1.POST("/companies/register", handlers.CreateCompany) // Onboarding endpoint

		// Protected Routes
		protected := v1.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			// Auth & Switch Branch
			protected.GET("/auth/me", handlers.GetMe)
			protected.POST("/auth/switch-branch", handlers.SwitchBranch)

			// SaaS Admin Routes (SuperAdmin restricted)
			saasAdmin := protected.Group("/saas-admin")
			saasAdmin.Use(auth.SuperAdminMiddleware())
			{
				saasAdmin.GET("/companies", handlers.GetSaaSCompanies)
				saasAdmin.PUT("/companies/:id/subscription", handlers.UpdateSaaSCompanySubscription)
				saasAdmin.GET("/stats", handlers.GetSaaSStats)
			}

			// Companies
			protected.GET("/companies", handlers.GetCompanies)

			// Branches
			protected.GET("/branches", handlers.GetBranches)
			protected.POST("/branches", handlers.CreateBranch)
			protected.PUT("/branches/:id", handlers.UpdateBranch)
			protected.DELETE("/branches/:id", handlers.DeleteBranch)

			// Products
			protected.GET("/products", handlers.GetProducts)
			protected.POST("/products", handlers.CreateProduct)
			protected.PUT("/products/:id", handlers.UpdateProduct)
			protected.DELETE("/products/:id", handlers.DeleteProduct)

			// Categories
			protected.GET("/categories", handlers.GetCategories)
			protected.POST("/categories", handlers.CreateCategory)
			protected.PUT("/categories/:id", handlers.UpdateCategory)
			protected.DELETE("/categories/:id", handlers.DeleteCategory)

			// Brands
			protected.GET("/brands", handlers.GetBrands)
			protected.POST("/brands", handlers.CreateBrand)
			protected.PUT("/brands/:id", handlers.UpdateBrand)
			protected.DELETE("/brands/:id", handlers.DeleteBrand)

			// Suppliers
			protected.GET("/suppliers", handlers.GetSuppliers)
			protected.POST("/suppliers", handlers.CreateSupplier)
			protected.PUT("/suppliers/:id", handlers.UpdateSupplier)
			protected.DELETE("/suppliers/:id", handlers.DeleteSupplier)

			// Customers
			protected.GET("/customers", handlers.GetCustomers)
			protected.POST("/customers", handlers.CreateCustomer)
			protected.PUT("/customers/:id", handlers.UpdateCustomer)
			protected.DELETE("/customers/:id", handlers.DeleteCustomer)

			// Purchases
			protected.GET("/purchases", handlers.GetPurchases)
			protected.GET("/purchases/:id", handlers.GetPurchaseDetails)
			protected.POST("/purchases", handlers.CreatePurchase)

			// Sales
			protected.GET("/sales", handlers.GetSales)
			protected.GET("/sales/:id", handlers.GetSaleDetails)
			protected.POST("/sales", handlers.CreateSale)

			// Stocks & Kardex
			protected.GET("/stocks", handlers.GetStocks)
			protected.GET("/kardex", handlers.GetKardex)

			// Dashboard Stats
			protected.GET("/dashboard/stats", handlers.GetDashboardStats)
		}
	}

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().Format(time.RFC3339)})
	})

	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func seedDatabase() {
	log.Println("Seeding database...")

	// 1. Roles
	var rolesCount int64
	config.DB.Model(&models.Role{}).Count(&rolesCount)
	if rolesCount == 0 {
		adminRole := models.Role{
			Nombre:      "Administrador",
			Descripcion: "Administrador con acceso total al sistema de su empresa",
		}
		vendedorRole := models.Role{
			Nombre:      "Vendedor",
			Descripcion: "Rol para atención al público y ventas",
		}
		config.DB.Create(&adminRole)
		config.DB.Create(&vendedorRole)
		log.Println("Seeded roles: Administrador, Vendedor.")
	}

	// 2. SaaS SuperAdmin User
	var superAdminCount int64
	config.DB.Model(&models.User{}).Where("username = ?", "saas_admin").Count(&superAdminCount)
	if superAdminCount == 0 {
		pwdHash, _ := auth.HashPassword("saas123")
		sa := models.User{
			Nombre:       "Super Administrador SaaS",
			Email:        "saas@veterinaria.com",
			Username:     "saas_admin",
			PasswordHash: pwdHash,
			Estado:       "active",
			RoleType:     "SUPER_ADMIN",
		}
		config.DB.Create(&sa)
		log.Println("Seeded SaaS SuperAdmin: saas_admin / saas123")
	}

	// 3. Default Company, Branch and Admin User
	var companiesCount int64
	config.DB.Model(&models.Company{}).Count(&companiesCount)
	if companiesCount == 0 {
		company := models.Company{
			RUC:                   "20123456789",
			RazonSocial:           "Clínica Veterinaria San Martín S.A.C.",
			NombreComercial:       "Veterinaria San Martín",
			Direccion:             "Av. Universitaria 1234, Lima",
			Telefono:              "014567890",
			Email:                 "contacto@vetsanmartin.com",
			Estado:                "active",
			PlanType:              "Premium",
			SubscriptionExpiresAt: time.Now().AddDate(1, 0, 0), // 1 year
			MaxBranches:           5,
		}
		config.DB.Create(&company)

		branch := models.Branch{
			CompanyID: company.ID,
			Nombre:    "Sede Principal",
			Direccion: "Av. Universitaria 1234, Lima",
			Telefono:  "014567890",
			Estado:    "active",
			IsMain:    true,
		}
		config.DB.Create(&branch)

		// Admin user credentials: admin / admin123
		pwdHash, _ := auth.HashPassword("admin123")
		user := models.User{
			CompanyID:    company.ID,
			BranchID:     branch.ID,
			Nombre:       "Administrador ERP",
			Email:        "admin@veterinaria.com",
			Username:     "admin",
			PasswordHash: pwdHash,
			Estado:       "active",
			RoleType:     "COMPANY_ADMIN",
		}
		config.DB.Create(&user)

		// Link role
		var adminRole models.Role
		if err := config.DB.Where("nombre = ?", "Administrador").First(&adminRole).Error; err == nil {
			config.DB.Model(&user).Association("Roles").Append(&adminRole)
		}

		log.Println("Seeded default tenant: Company='Veterinaria San Martín', Branch='Sede Principal', User='admin/admin123'.")
	}
}

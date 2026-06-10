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
		&models.Pet{},
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
			protected.POST("/branches", auth.CompanyAdminMiddleware(), handlers.CreateBranch)
			protected.PUT("/branches/:id", auth.CompanyAdminMiddleware(), handlers.UpdateBranch)
			protected.DELETE("/branches/:id", auth.CompanyAdminMiddleware(), handlers.DeleteBranch)

			// Users (Personal)
			protected.GET("/users", auth.BranchAdminMiddleware(), handlers.GetUsers)
			protected.POST("/users", auth.BranchAdminMiddleware(), handlers.CreateUser)
			protected.PUT("/users/:id", auth.BranchAdminMiddleware(), handlers.UpdateUser)
			protected.DELETE("/users/:id", auth.BranchAdminMiddleware(), handlers.DeleteUser)

			// Products
			protected.GET("/products", handlers.GetProducts)
			protected.POST("/products", auth.CompanyAdminMiddleware(), handlers.CreateProduct)
			protected.PUT("/products/:id", auth.CompanyAdminMiddleware(), handlers.UpdateProduct)
			protected.DELETE("/products/:id", auth.CompanyAdminMiddleware(), handlers.DeleteProduct)

			// Categories
			protected.GET("/categories", handlers.GetCategories)
			protected.POST("/categories", auth.CompanyAdminMiddleware(), handlers.CreateCategory)
			protected.PUT("/categories/:id", auth.CompanyAdminMiddleware(), handlers.UpdateCategory)
			protected.DELETE("/categories/:id", auth.CompanyAdminMiddleware(), handlers.DeleteCategory)

			// Brands
			protected.GET("/brands", handlers.GetBrands)
			protected.POST("/brands", auth.CompanyAdminMiddleware(), handlers.CreateBrand)
			protected.PUT("/brands/:id", auth.CompanyAdminMiddleware(), handlers.UpdateBrand)
			protected.DELETE("/brands/:id", auth.CompanyAdminMiddleware(), handlers.DeleteBrand)

			// Suppliers
			protected.GET("/suppliers", handlers.GetSuppliers)
			protected.POST("/suppliers", auth.CompanyAdminMiddleware(), handlers.CreateSupplier)
			protected.PUT("/suppliers/:id", auth.CompanyAdminMiddleware(), handlers.UpdateSupplier)
			protected.DELETE("/suppliers/:id", auth.CompanyAdminMiddleware(), handlers.DeleteSupplier)

			// Customers
			protected.GET("/customers", handlers.GetCustomers)
			protected.POST("/customers", handlers.CreateCustomer)
			protected.PUT("/customers/:id", handlers.UpdateCustomer)
			protected.DELETE("/customers/:id", handlers.DeleteCustomer)

			// Pets
			protected.GET("/pets", handlers.GetPets)
			protected.POST("/pets", handlers.CreatePet)
			protected.PUT("/pets/:id", handlers.UpdatePet)
			protected.DELETE("/pets/:id", handlers.DeletePet)

			// Purchases
			protected.GET("/purchases", auth.BranchAdminMiddleware(), handlers.GetPurchases)
			protected.GET("/purchases/:id", auth.BranchAdminMiddleware(), handlers.GetPurchaseDetails)
			protected.POST("/purchases", auth.BranchAdminMiddleware(), handlers.CreatePurchase)

			// Sales
			protected.GET("/sales", handlers.GetSales)
			protected.GET("/sales/:id", handlers.GetSaleDetails)
			protected.POST("/sales", handlers.CreateSale)

			// Stocks & Kardex
			protected.GET("/stocks", handlers.GetStocks)
			protected.GET("/kardex", handlers.GetKardex)

			// Dashboard Stats
			protected.GET("/dashboard/stats", auth.BranchAdminMiddleware(), handlers.GetDashboardStats)

			// Billing Configuration
			protected.GET("/billing/config", auth.CompanyAdminMiddleware(), handlers.GetBillingConfig)
			protected.POST("/billing/config", auth.CompanyAdminMiddleware(), handlers.SaveBillingConfig)
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
		
		// Seed additional test data for the default tenant
		seedTestData(company, branch)
	} else {
		// Company exists, let's check if we need to seed products
		var company models.Company
		var branch models.Branch
		if err := config.DB.First(&company).Error; err == nil {
			if err := config.DB.First(&branch, "company_id = ?", company.ID).Error; err == nil {
				var productsCount int64
				config.DB.Model(&models.Product{}).Count(&productsCount)
				if productsCount == 0 {
					seedTestData(company, branch)
				}
			}
		}
	}
}

func seedTestData(company models.Company, branch models.Branch) {
	log.Println("Seeding test data (Categories, Brands, Products, Customers, Suppliers)...")

	// Categories
	cat1 := models.Category{CompanyID: company.ID, Nombre: "Alimentos"}
	cat2 := models.Category{CompanyID: company.ID, Nombre: "Medicinas"}
	cat3 := models.Category{CompanyID: company.ID, Nombre: "Accesorios"}
	config.DB.Create(&cat1)
	config.DB.Create(&cat2)
	config.DB.Create(&cat3)

	// Brands
	b1 := models.Brand{CompanyID: company.ID, Nombre: "Ricocan"}
	b2 := models.Brand{CompanyID: company.ID, Nombre: "Bravecto"}
	b3 := models.Brand{CompanyID: company.ID, Nombre: "KONG"}
	config.DB.Create(&b1)
	config.DB.Create(&b2)
	config.DB.Create(&b3)

	// Products
	p1 := models.Product{
		CompanyID:    company.ID,
		CategoryID:   cat1.ID,
		BrandID:      b1.ID,
		Codigo:       "PROD-001",
		Nombre:       "Ricocan Adulto 15kg",
		PrecioCompra: 120.00,
		PrecioVenta:  150.00,
		StockMinimo:  5,
		Estado:       "active",
	}
	p2 := models.Product{
		CompanyID:    company.ID,
		CategoryID:   cat2.ID,
		BrandID:      b2.ID,
		Codigo:       "PROD-002",
		Nombre:       "Pastilla Bravecto 10-20kg",
		PrecioCompra: 60.00,
		PrecioVenta:  95.00,
		StockMinimo:  2,
		Estado:       "active",
	}
	config.DB.Create(&p1)
	config.DB.Create(&p2)

	// Stock
	config.DB.Create(&models.Stock{CompanyID: company.ID, BranchID: branch.ID, ProductID: p1.ID, StockActual: 10})
	config.DB.Create(&models.Stock{CompanyID: company.ID, BranchID: branch.ID, ProductID: p2.ID, StockActual: 15})

	// Customers
	c1 := models.Customer{
		CompanyID:       company.ID,
		TipoDocumento:   "DNI",
		NumeroDocumento: "12345678",
		Nombre:          "Juan Perez",
		Email:           "juan.perez@example.com",
		Telefono:        "987654321",
	}
	config.DB.Create(&c1)

	// Pets
	config.DB.Create(&models.Pet{
		CompanyID:  company.ID,
		CustomerID: c1.ID,
		Nombre:     "Fido",
		Especie:    "Perro",
		Raza:       "Mestizo",
		Sexo:       "Macho",
		Peso:       15.5,
	})

	// Suppliers
	s1 := models.Supplier{
		CompanyID:   company.ID,
		RUC:         "20987654321",
		RazonSocial: "Distribuidora Mascotas SAC",
		Direccion:   "Av. Los Pinos 456",
		Telefono:    "012345678",
	}
	config.DB.Create(&s1)
	log.Println("Test data successfully seeded.")
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains standard fields with UUID primary key
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook generates a UUID if not set
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}

// Company represents a client tenant
type Company struct {
	BaseModel
	RUC                   string    `gorm:"type:varchar(11);uniqueIndex;not null" json:"ruc"`
	RazonSocial           string    `gorm:"type:varchar(255);not null" json:"razon_social"`
	NombreComercial       string    `gorm:"type:varchar(255)" json:"nombre_comercial"`
	Direccion             string    `gorm:"type:varchar(500)" json:"direccion"`
	Telefono              string    `gorm:"type:varchar(50)" json:"telefono"`
	Email                 string    `gorm:"type:varchar(100)" json:"email"`
	Estado                string    `gorm:"type:varchar(50);default:'active'" json:"estado"` // active, inactive
	PlanType              string    `gorm:"type:varchar(50);default:'Basic'" json:"plan_type"`
	SubscriptionExpiresAt time.Time `json:"subscription_expires_at"`
	MaxBranches           int       `gorm:"type:integer;default:1" json:"max_branches"`
}

// Branch belongs to a Company
type Branch struct {
	BaseModel
	CompanyID    uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Nombre       string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Direccion    string    `gorm:"type:varchar(500)" json:"direccion"`
	Telefono     string    `gorm:"type:varchar(50)" json:"telefono"`
	Email        string    `gorm:"type:varchar(100)" json:"email"`
	IsMain       bool      `gorm:"type:boolean;default:false" json:"is_main"`
	Ubigeo       string    `gorm:"type:varchar(6)" json:"ubigeo"`
	SerieFactura string    `gorm:"type:varchar(4)" json:"serie_factura"`
	SerieBoleta  string    `gorm:"type:varchar(4)" json:"serie_boleta"`
	Estado       string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
}

// Permission defines a policy access code
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Codigo      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"codigo"` // e.g. "products.create"
	Descripcion string    `gorm:"type:varchar(255)" json:"descripcion"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// Role defines a group of permissions
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;" json:"id"`
	Nombre      string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"nombre"` // e.g. "SuperAdmin", "Cajero"
	Descripcion string       `gorm:"type:varchar(255)" json:"descripcion"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

// User belongs to a Company and Branch, and has Roles
type User struct {
	BaseModel
	CompanyID    uuid.UUID `gorm:"type:uuid;index" json:"company_id,omitempty"`
	BranchID     uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Nombre       string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Estado       string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
	RoleType     string    `gorm:"type:varchar(50);default:'BRANCH_USER'" json:"role_type"`
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Category for Products
type Category struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Nombre    string    `gorm:"type:varchar(100);not null" json:"nombre"`
}

// Brand for Products
type Brand struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Nombre    string    `gorm:"type:varchar(100);not null" json:"nombre"`
}

// Product in Inventory
type Product struct {
	BaseModel
	CompanyID    uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	CategoryID   uuid.UUID `gorm:"type:uuid;index" json:"category_id"`
	BrandID      uuid.UUID `gorm:"type:uuid;index" json:"brand_id"`
	Codigo       string    `gorm:"type:varchar(100)" json:"codigo"`
	CodigoBarras string    `gorm:"type:varchar(100)" json:"codigo_barras"`
	Nombre       string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion  string    `gorm:"type:text" json:"descripcion"`
	PrecioCompra float64   `gorm:"type:numeric(12,4);default:0" json:"precio_compra"`
	PrecioVenta  float64   `gorm:"type:numeric(12,4);default:0" json:"precio_venta"`
	StockMinimo  float64   `gorm:"type:numeric(12,4);default:0" json:"stock_minimo"`
	Estado       string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
}

// Stock by Branch
type Stock struct {
	BaseModel
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	StockActual float64   `gorm:"type:numeric(12,4);default:0" json:"stock_actual"`
}

// Kardex transaction history
type Kardex struct {
	BaseModel
	CompanyID      uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID       uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	TipoMovimiento string    `gorm:"type:varchar(50);not null" json:"tipo_movimiento"` // INGRESO, VENTA, AJUSTE, TRANSFERENCIA
	Referencia     string    `gorm:"type:varchar(255)" json:"referencia"`              // Doc number or ID
	Cantidad       float64   `gorm:"type:numeric(12,4);not null" json:"cantidad"`
	StockAnterior  float64   `gorm:"type:numeric(12,4);not null" json:"stock_anterior"`
	StockNuevo     float64   `gorm:"type:numeric(12,4);not null" json:"stock_nuevo"`
}

// Supplier for purchases
type Supplier struct {
	BaseModel
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	RUC         string    `gorm:"type:varchar(11);not null" json:"ruc"`
	RazonSocial string    `gorm:"type:varchar(255);not null" json:"razon_social"`
	Direccion   string    `gorm:"type:varchar(500)" json:"direccion"`
	Telefono    string    `gorm:"type:varchar(50)" json:"telefono"`
}

// Purchase header
type Purchase struct {
	BaseModel
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	SupplierID  uuid.UUID `gorm:"type:uuid;not null;index" json:"supplier_id"`
	Fecha       time.Time `json:"fecha"`
	Subtotal    float64   `gorm:"type:numeric(12,4);default:0" json:"subtotal"`
	IGV         float64   `gorm:"type:numeric(12,4);default:0" json:"igv"`
	Total       float64   `gorm:"type:numeric(12,4);default:0" json:"total"`
	Estado      string    `gorm:"type:varchar(50);default:'draft'" json:"estado"` // draft, completed, cancelled
}

// PurchaseItem detail
type PurchaseItem struct {
	BaseModel
	PurchaseID    uuid.UUID `gorm:"type:uuid;not null;index" json:"purchase_id"`
	ProductID     uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Cantidad      float64   `gorm:"type:numeric(12,4);not null" json:"cantidad"`
	CostoUnitario float64   `gorm:"type:numeric(12,4);not null" json:"costo_unitario"`
}

// Customer for sales and owners of pets
type Customer struct {
	BaseModel
	CompanyID       uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	TipoDocumento   string    `gorm:"type:varchar(50);not null" json:"tipo_documento"` // DNI, RUC, CE
	NumeroDocumento string    `gorm:"type:varchar(50);not null" json:"numero_documento"`
	Nombre          string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Direccion       string    `gorm:"type:varchar(500)" json:"direccion"`
	Telefono        string    `gorm:"type:varchar(50)" json:"telefono"`
	Email           string    `gorm:"type:varchar(100)" json:"email"`
	Pets            []Pet     `gorm:"foreignKey:CustomerID" json:"pets,omitempty"`
}

// Pet belongs to a Customer
type Pet struct {
	BaseModel
	CompanyID  uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Nombre     string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Especie    string    `gorm:"type:varchar(100)" json:"especie"` // Perro, Gato, etc.
	Raza       string    `gorm:"type:varchar(100)" json:"raza"`
	Sexo       string    `gorm:"type:varchar(20)" json:"sexo"` // Macho, Hembra
	FechaNac   *time.Time `json:"fecha_nac"`
	Peso       float64   `gorm:"type:numeric(12,2)" json:"peso"`
	Notas      string    `gorm:"type:text" json:"notas"`
}

// Sale header
type Sale struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID  uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Subtotal  float64   `gorm:"type:numeric(12,4);default:0" json:"subtotal"`
	IGV       float64   `gorm:"type:numeric(12,4);default:0" json:"igv"`
	Total     float64   `gorm:"type:numeric(12,4);default:0" json:"total"`
	Estado    string    `gorm:"type:varchar(50);default:'completed'" json:"estado"` // completed, cancelled
}

// SaleItem detail
type SaleItem struct {
	BaseModel
	SaleID         uuid.UUID `gorm:"type:uuid;not null;index" json:"sale_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Cantidad       float64   `gorm:"type:numeric(12,4);not null" json:"cantidad"`
	PrecioUnitario float64   `gorm:"type:numeric(12,4);not null" json:"precio_unitario"`
	Descuento      float64   `gorm:"type:numeric(12,4);default:0" json:"descuento"`
}

// BillingConfig configurations for FacturaAPI
type BillingConfig struct {
	BaseModel
	CompanyID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	ApiURL     string    `gorm:"type:varchar(255)" json:"api_url"`
	ApiKey     string    `gorm:"type:varchar(255)" json:"api_key"`
	TenantUUID string    `gorm:"type:varchar(255)" json:"tenant_uuid"`
	Modo       string    `gorm:"type:varchar(50);default:'dev'" json:"modo"` // dev, prod
	Estado     string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
}

// ElectronicDocument billing record
type ElectronicDocument struct {
	BaseModel
	CompanyID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	SaleID         *uuid.UUID `gorm:"type:uuid;index" json:"sale_id,omitempty"`
	DocumentUUID   string     `gorm:"type:varchar(255);uniqueIndex" json:"document_uuid"`
	TipoDocumento  string     `gorm:"type:varchar(2);not null" json:"tipo_documento"` // 01=Factura, 03=Boleta, etc.
	Serie          string     `gorm:"type:varchar(4);not null" json:"serie"`
	Numero         string     `gorm:"type:varchar(8);not null" json:"numero"`
	Estado         string     `gorm:"type:varchar(50);default:'pending'" json:"estado"` // pending, accepted, rejected, voided
	PdfURL         string     `gorm:"type:varchar(500)" json:"pdf_url"`
	XmlURL         string     `gorm:"type:varchar(500)" json:"xml_url"`
	CdrURL         string     `gorm:"type:varchar(500)" json:"cdr_url"`
}

// AuditLog audit action registry
type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Modulo      string    `gorm:"type:varchar(100)" json:"modulo"`
	Accion      string    `gorm:"type:varchar(100)" json:"accion"`
	Descripcion string    `gorm:"type:text" json:"descripcion"`
	IP          string    `gorm:"type:varchar(45)" json:"ip"`
	Fecha       time.Time `json:"fecha"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Fecha.IsZero() {
		a.Fecha = time.Now()
	}
	return
}

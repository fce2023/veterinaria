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
	RUC             string `gorm:"type:varchar(11);uniqueIndex;not null" json:"ruc"`
	RazonSocial     string `gorm:"type:varchar(255);not null" json:"razon_social"`
	NombreComercial string `gorm:"type:varchar(255)" json:"nombre_comercial"`
	Direccion       string `gorm:"type:varchar(500)" json:"direccion"`
	Telefono        string `gorm:"type:varchar(50)" json:"telefono"`
	Email           string `gorm:"type:varchar(100)" json:"email"`
	LogoBase64      string `gorm:"type:text" json:"logo_base64"`
	Estado          string `gorm:"type:varchar(50);default:'active'" json:"estado"` // active, inactive
	Ubigeo          string `gorm:"type:varchar(6)" json:"ubigeo"`
	Departamento    string `gorm:"type:varchar(100)" json:"departamento"`
	Provincia       string `gorm:"type:varchar(100)" json:"provincia"`
	Distrito        string `gorm:"type:varchar(100)" json:"distrito"`
}

// Branch belongs to a Company
type Branch struct {
	BaseModel
	CompanyID          uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Nombre             string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Direccion          string    `gorm:"type:varchar(500)" json:"direccion"`
	Telefono           string    `gorm:"type:varchar(50)" json:"telefono"`
	Email              string    `gorm:"type:varchar(100)" json:"email"`
	IsMain             bool      `gorm:"type:boolean;default:false" json:"is_main"`
	Ubigeo             string    `gorm:"type:varchar(6)" json:"ubigeo"`
	SerieFactura       string    `gorm:"type:varchar(4)" json:"serie_factura"`
	SerieBoleta        string    `gorm:"type:varchar(4)" json:"serie_boleta"`
	CorrelativoFactura int       `gorm:"type:integer;default:0" json:"correlativo_factura"`
	CorrelativoBoleta  int       `gorm:"type:integer;default:0" json:"correlativo_boleta"`
	Estado             string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
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
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_product_code;uniqueIndex:idx_company_product_barcode" json:"company_id"`
	CategoryID    uuid.UUID `gorm:"type:uuid;index" json:"category_id"`
	BrandID       uuid.UUID `gorm:"type:uuid;index" json:"brand_id"`
	Codigo        string    `gorm:"type:varchar(100);uniqueIndex:idx_company_product_code" json:"codigo"`
	CodigoBarras  string    `gorm:"type:varchar(100);uniqueIndex:idx_company_product_barcode" json:"codigo_barras"`
	Nombre        string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion   string    `gorm:"type:text" json:"descripcion"`
	PrecioCompra  float64   `gorm:"type:numeric(12,4);default:0" json:"precio_compra"`
	PrecioVenta   float64   `gorm:"type:numeric(12,4);default:0" json:"precio_venta"`
	StockMinimo   float64   `gorm:"type:numeric(12,4);default:0" json:"stock_minimo"`
	Estado        string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
	UnidadMedida  string    `gorm:"type:varchar(50);default:'und'" json:"unidad_medida"`
	IsDimensional bool      `gorm:"type:boolean;default:false" json:"is_dimensional"`
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
	CompanyID         uuid.UUID            `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID          uuid.UUID            `gorm:"type:uuid;not null;index" json:"branch_id"`
	SupplierID        uuid.UUID            `gorm:"type:uuid;not null;index" json:"supplier_id"`
	Supplier          Supplier             `gorm:"foreignKey:SupplierID" json:"supplier"`
	Fecha             time.Time            `json:"fecha"`
	Subtotal          float64              `gorm:"type:numeric(12,4);default:0" json:"subtotal"`
	IGV               float64              `gorm:"type:numeric(12,4);default:0" json:"igv"`
	Total             float64              `gorm:"type:numeric(12,4);default:0" json:"total"`
	Estado            string               `gorm:"type:varchar(50);default:'draft'" json:"estado"` // draft, completed, cancelled
	TipoComprobante   string               `gorm:"type:varchar(100)" json:"tipo_comprobante"`
	NumeroComprobante string               `gorm:"type:varchar(100)" json:"numero_comprobante"`
	ComprobanteUrl    string               `gorm:"type:varchar(500)" json:"comprobante_url"`
	Attachments       []PurchaseAttachment `gorm:"foreignKey:PurchaseID" json:"attachments,omitempty"`
}

// PurchaseAttachment stores files associated with a purchase
type PurchaseAttachment struct {
	BaseModel
	PurchaseID  uuid.UUID `gorm:"type:uuid;not null;index" json:"purchase_id"`
	Nombre      string    `gorm:"type:varchar(255);not null" json:"nombre"`
	Url         string    `gorm:"type:varchar(500);not null" json:"url"`
	Descripcion string    `gorm:"type:varchar(500)" json:"descripcion"`
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
	CompanyID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	CustomerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	Nombre     string     `gorm:"type:varchar(255);not null" json:"nombre"`
	Especie    string     `gorm:"type:varchar(100)" json:"especie"` // Perro, Gato, etc.
	Raza       string     `gorm:"type:varchar(100)" json:"raza"`
	Sexo       string     `gorm:"type:varchar(20)" json:"sexo"` // Macho, Hembra
	FechaNac   *time.Time `json:"fecha_nac"`
	Peso       float64    `gorm:"type:numeric(12,2)" json:"peso"`
	Notas      string     `gorm:"type:text" json:"notes"`
}

// Sale header
type Sale struct {
	BaseModel
	CompanyID     uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID      uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	CustomerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer      Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TipoDocumento string    `gorm:"type:varchar(2);default:'03'" json:"tipo_documento"` // 01=Factura, 03=Boleta
	Serie         string    `gorm:"type:varchar(4)" json:"serie"`
	Numero        string    `gorm:"type:varchar(10)" json:"numero"`
	MetodoPago    string    `gorm:"type:varchar(50);default:'EFECTIVO'" json:"metodo_pago"` // EFECTIVO, TARJETA, TRANSFERENCIA, YAPE, PLIN
	Subtotal      float64   `gorm:"type:numeric(12,4);default:0" json:"subtotal"`
	IGV           float64   `gorm:"type:numeric(12,4);default:0" json:"igv"`
	Total         float64   `gorm:"type:numeric(12,4);default:0" json:"total"`
	Estado        string    `gorm:"type:varchar(50);default:'completed'" json:"estado"` // completed, cancelled
}

// SaleItem detail
type SaleItem struct {
	BaseModel
	SaleID         uuid.UUID `gorm:"type:uuid;not null;index" json:"sale_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Cantidad       float64   `gorm:"type:numeric(12,4);not null" json:"cantidad"`
	PrecioUnitario float64   `gorm:"type:numeric(12,4);not null" json:"precio_unitario"` // Con IGV
	Subtotal       float64   `gorm:"type:numeric(12,4);default:0" json:"subtotal"`        // Sin IGV
	IGV            float64   `gorm:"type:numeric(12,4);default:0" json:"igv"`             // Monto IGV
	Descuento      float64   `gorm:"type:numeric(12,4);default:0" json:"descuento"`
	TipoAfectacion string    `gorm:"type:varchar(2);default:'10'" json:"tipo_afectacion"` // Catálogo 07 (10=Gravado)
	UnidadSUNAT    string    `gorm:"type:varchar(10);default:'NIU'" json:"unidad_sunat"`  // Catálogo 03 (NIU=Unidad)
}

// BillingConfig configurations for FacturaAPI
type BillingConfig struct {
	BaseModel
	CompanyID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	ApiURL            string    `gorm:"type:varchar(255)" json:"api_url"`
	ApiKey            string    `gorm:"type:varchar(255)" json:"api_key"`
	TenantUUID        string    `gorm:"type:varchar(255)" json:"tenant_uuid"`
	Modo              string    `gorm:"type:varchar(50);default:'dev'" json:"modo"` // dev, prod
	Estado            string    `gorm:"type:varchar(50);default:'active'" json:"estado"`
	SolUser           string    `gorm:"type:varchar(100)" json:"sol_user"`
	SolPass           string    `gorm:"type:varchar(100)" json:"sol_pass"`
	CertificadoBase64 string    `gorm:"type:text" json:"certificado_base64"`
	CertificadoPassword string   `gorm:"type:varchar(255)" json:"certificado_password"`
	ClientID          string    `gorm:"type:varchar(255)" json:"client_id"`
	ClientSecret      string    `gorm:"type:varchar(255)" json:"client_secret"`
	LogoBase64        string    `gorm:"type:text" json:"logo_base64"`
	EmisionDiferida   bool      `gorm:"type:boolean;default:false" json:"emision_diferida"`
	CorrelativoPadding *int     `gorm:"type:integer" json:"correlativo_padding"`
}

// ElectronicDocument billing record
type ElectronicDocument struct {
	BaseModel
	CompanyID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	SaleID       *uuid.UUID `gorm:"type:uuid;index" json:"sale_id,omitempty"`
	Sale         *Sale      `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	DocumentUUID *string    `gorm:"type:varchar(255);uniqueIndex" json:"document_uuid"`
	TipoDocumento string     `gorm:"type:varchar(2);not null" json:"tipo_documento"` // 01=Factura, 03=Boleta, etc.
	Serie        string     `gorm:"type:varchar(4);not null" json:"serie"`
	Numero       string     `gorm:"type:varchar(8);not null" json:"numero"`
	Estado       string     `gorm:"type:varchar(50);default:'pending'" json:"estado"` // pending, accepted, rejected, voided
	PdfURL       string     `gorm:"type:varchar(500)" json:"pdf_url"`
	XmlURL       string     `gorm:"type:varchar(500)" json:"xml_url"`
	CdrURL       string     `gorm:"type:varchar(500)" json:"cdr_url"`
	SunatResponse string    `gorm:"type:text" json:"sunat_response"`
	FacturacionError string `gorm:"type:text" json:"facturacion_error"`
	FacturacionPayload string `gorm:"type:text" json:"facturacion_payload"`
	Observaciones string    `gorm:"type:text" json:"observaciones"`
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

// Plan defines SaaS subscription packages
type Plan struct {
	BaseModel
	Nombre      string  `gorm:"type:varchar(100);not null" json:"nombre"` // Básico, Profesional, Empresarial
	Precio      float64 `gorm:"type:numeric(12,2);not null" json:"precio"`
	MaxBranches int     `gorm:"type:integer;default:1" json:"max_branches"`
	MaxUsers    int     `gorm:"type:integer;default:5" json:"max_users"`
	Modulos     string  `gorm:"type:text" json:"modulos"` // Comma-separated list: "core,veterinaria"
}

// Subscription represents SaaS licensing state per company
type Subscription struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	PlanID    uuid.UUID `gorm:"type:uuid;not null" json:"plan_id"`
	Plan      Plan      `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	Estado    string    `gorm:"type:varchar(50);default:'TRIAL'" json:"estado"` // TRIAL, ACTIVE, EXPIRED, SUSPENDED, CANCELLED
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Payment logs manual and gateways payments
type Payment struct {
	BaseModel
	CompanyID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	SubscriptionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"subscription_id"`
	Monto          float64    `gorm:"type:numeric(12,2);not null" json:"monto"`
	MetodoPago     string     `gorm:"type:varchar(50);not null" json:"metodo_pago"` // YAPE, PLIN, TRANSFERENCIA, DEPOSITO
	Referencia     string     `gorm:"type:varchar(255);uniqueIndex" json:"referencia"`
	Estado         string     `gorm:"type:varchar(50);default:'PENDING'" json:"estado"` // PENDING, APPROVED, REJECTED
	ComprobanteUrl string     `gorm:"type:varchar(500)" json:"comprobante_url"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	ApprovedBy     *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
}

// CompanyModule defines which domain modules are active per tenant
type CompanyModule struct {
	BaseModel
	CompanyID        uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_module" json:"company_id"`
	ModuleKey        string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_module" json:"module_key"` // e.g. "veterinaria", "vidrieria", "restaurante"
	IsActive         bool       `gorm:"type:boolean;default:true" json:"is_active"`
	EnabledAt        *time.Time `json:"enabled_at,omitempty"`
	DisabledAt       *time.Time `json:"disabled_at,omitempty"`
	EnabledByUserID  *uuid.UUID `gorm:"type:uuid" json:"enabled_by_user_id,omitempty"`
	DisabledByUserID *uuid.UUID `gorm:"type:uuid" json:"disabled_by_user_id,omitempty"`
}

// SaleItemDimension extends Core's SaleItem for physical dimensional measurements (M, M2, M3)
type SaleItemDimension struct {
	BaseModel
	SaleItemID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;index" json:"sale_item_id"`
	Alto           float64   `gorm:"type:numeric(12,4);default:0" json:"alto"`            // Height / Length (M, M2, M3)
	Ancho          float64   `gorm:"type:numeric(12,4);default:0" json:"ancho"`           // Width (M2, M3)
	Espesor        float64   `gorm:"type:numeric(12,4);default:0" json:"espesor"`         // Thickness / Depth (M3)
	CantidadPiezas int       `gorm:"type:integer;default:1" json:"cantidad_piezas"`
}

// CompanySetting stores configuration key-values for a company tenant
type CompanySetting struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_setting" json:"company_id"`
	Clave     string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_setting" json:"clave"` // e.g. "timezone", "currency", "receipt_footer"
	Valor     string    `gorm:"type:text" json:"valor"`
}

// SaasAuditLog logs actions performed by SuperAdmin on SaaS Billing/Modules layer
type SaasAuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`       // SuperAdmin user id
	CompanyID   uuid.UUID `gorm:"type:uuid;index" json:"company_id"`    // Target company
	Accion      string    `gorm:"type:varchar(100);not null" json:"accion"` // e.g. "APPROVE_PAYMENT", "TOGGLE_MODULE"
	Descripcion string    `gorm:"type:text" json:"descripcion"`
	IP          string    `gorm:"type:varchar(45)" json:"ip"`
	Fecha       time.Time `json:"fecha"`
}

func (s *SaasAuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Fecha.IsZero() {
		s.Fecha = time.Now()
	}
	return
}

// CashSession represents a cash register session (Apertura/Cierre)
type CashSession struct {
	BaseModel
	CompanyID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OpenedAt       time.Time  `json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	SaldoInicial   float64    `gorm:"type:numeric(12,2);default:0" json:"saldo_inicial"`
	TotalVentasEfe float64    `gorm:"type:numeric(12,2);default:0" json:"total_ventas_efe"` // Total cash sales
	TotalVentasOtr float64    `gorm:"type:numeric(12,2);default:0" json:"total_ventas_otr"` // Total non-cash sales
	TotalIngresos  float64    `gorm:"type:numeric(12,2);default:0" json:"total_ingresos"`
	TotalEgresos   float64    `gorm:"type:numeric(12,2);default:0" json:"total_egresos"`
	SaldoFinalReal float64    `gorm:"type:numeric(12,2);default:0" json:"saldo_final_real"` // What cashier actually counted
	Estado         string     `gorm:"type:varchar(20);default:'OPEN'" json:"estado"`        // OPEN, CLOSED
	Notas          string     `gorm:"type:text" json:"notas"`
}

// CashMovement represents manual cash entries or exits
type CashMovement struct {
	BaseModel
	CashSessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"cash_session_id"`
	Tipo          string    `gorm:"type:varchar(20);not null" json:"tipo"` // INGRESO, EGRESO
	Monto         float64   `gorm:"type:numeric(12,2);not null" json:"monto"`
	Motivo        string    `gorm:"type:varchar(255);not null" json:"motivo"`
}

// NotaCredito represents a Credit Note (Nota de Crédito) for modifying/annulling sales
type NotaCredito struct {
	BaseModel
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	SaleID      uuid.UUID `gorm:"type:uuid;not null;index" json:"sale_id"`
	Sale        Sale      `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	Serie       string    `gorm:"type:varchar(4);not null" json:"serie"`
	Numero      string    `gorm:"type:varchar(10);not null" json:"numero"`
	Motivo      string    `gorm:"type:varchar(255);not null" json:"motivo"`
	Total       float64   `gorm:"type:numeric(12,4);default:0" json:"total"`
	Estado      string    `gorm:"type:varchar(50);default:'completed'" json:"estado"` // completed, cancelled
}

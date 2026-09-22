package constants

import "time"

const (
	APIPrefix           = "/api/v1"
	HealthPath          = "/healthz"
	SuccessCode         = 0
	SuccessMessage      = "ok"
	DefaultPage         = 1
	DefaultPageSize     = 12
	MaxPageSize         = 100
	StatusInStock       = "in_stock"
	StatusOutOfStock    = "out_of_stock"
	StatusDiscontinued  = "discontinued"
	SupplierPending     = "pending"
	SupplierApproved    = "approved"
	SupplierRejected    = "rejected"
	RoleAdmin           = "admin"
	RoleSupplier        = "supplier"
	RoleUser            = "user"
	TrendDays30         = "30d"
	TrendDays90         = "90d"
	TrendDaysYear       = "1y"
	BudgetRoomLiving    = "living_room"
	BudgetRoomKitchen   = "kitchen"
	BudgetRoomBathroom  = "bathroom"
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDContextKey    = "user_id"
	RoleContextKey      = "role"
	DemoUserID          = "demo-user"
	HistorySeedDays     = 365
)

var DemoTokenLifetime = time.Hour * 24

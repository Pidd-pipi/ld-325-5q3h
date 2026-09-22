package model

import "gorm.io/gorm"

// PurchaseOrder is the aggregate root of a building-material split purchase.
// A rejected order is still persisted (status=failed) so its failure reason is
// queryable via the detail endpoint and idempotent retries return the same result.
type PurchaseOrder struct {
	gorm.Model
	UserID          string `gorm:"size:64;not null;uniqueIndex:idx_purchase_user_idempotency,priority:1"`
	IdempotencyKey  string `gorm:"size:64;not null;uniqueIndex:idx_purchase_user_idempotency,priority:2"`
	ProductID       uint   `gorm:"not null;index"`
	ProductName     string `gorm:"size:120;not null"`
	ProductUnit     string `gorm:"size:20"`
	TotalQuantity   int    `gorm:"not null"`
	Status          string `gorm:"size:20;not null;index"`
	TotalAmount     float64
	MaxDeliveryDays int
	FailureCode     string `gorm:"size:40"`
	FailureReason   string `gorm:"size:200"`
	Items           []PurchaseOrderItem `gorm:"foreignKey:OrderID"`
}

// PurchaseOrderItem is one supplier allocation line of a successful purchase order.
// Failed orders carry no items and reserve no stock.
type PurchaseOrderItem struct {
	gorm.Model
	OrderID      uint `gorm:"not null;index"`
	OfferID      uint `gorm:"not null;index"`
	SupplierID   uint `gorm:"not null"`
	SupplierName string `gorm:"size:120;not null"`
	UnitPrice       float64
	Quantity        int `gorm:"not null"`
	MOQ             int `gorm:"not null"`
	AvailableBefore int
	AvailableAfter  int
	DeliveryDays    int
	Freight      string `gorm:"size:120"`
	LineAmount   float64
}

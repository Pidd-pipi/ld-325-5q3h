package model

import "gorm.io/gorm"

type PurchaseOrder struct {
	gorm.Model
	UserID         string `gorm:"uniqueIndex:idx_purchase_orders_user_key;not null"`
	ProductID      uint   `gorm:"not null"`
	Product        Product
	Quantity       int
	Status         string `gorm:"index"`
	FailureReason  string
	TotalAmount    float64
	DeliveryDays   int
	IdempotencyKey string `gorm:"uniqueIndex:idx_purchase_orders_user_key;not null"`
	Items          []PurchaseOrderItem
}

type PurchaseOrderItem struct {
	gorm.Model
	PurchaseOrderID uint `gorm:"index;not null"`
	OfferID         uint `gorm:"not null"`
	SupplierID      uint `gorm:"not null"`
	Supplier        Supplier
	Quantity        int
	UnitPrice       float64
	Amount          float64
	DeliveryDays    int
}

package model

import "gorm.io/gorm"

type Offer struct {
	gorm.Model
	ProductID    uint
	SupplierID   uint
	Supplier     Supplier
	UnitPrice    float64
	MOQ          int
	Freight      string
	DeliveryDays int
	StockStatus  string
}

package model

import "gorm.io/gorm"

type Favorite struct {
	gorm.Model
	UserID    string
	ProductID uint
	Folder    string
	Product   Product
}
type PriceAlert struct {
	gorm.Model
	UserID      string
	ProductID   uint
	TargetPrice float64
	DropPercent float64
	Active      bool
}
type Budget struct {
	gorm.Model
	UserID   string
	RoomType string
	Area     float64
	Estimate float64
	Payload  string
}

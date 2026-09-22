package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Brand          string
	ModelNumber    string `gorm:"column:model_number" json:"Model"`
	Unit           string
	Thumbnail      string
	CategoryID     uint
	Category       Category
	SalesCount     int
	Rating         float64
	Offers         []Offer
	PriceHistories []PriceHistory
}

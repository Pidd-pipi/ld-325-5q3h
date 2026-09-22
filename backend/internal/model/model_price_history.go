package model

import (
	"gorm.io/gorm"
	"time"
)

type PriceHistory struct {
	gorm.Model
	ProductID  uint
	OfferID    uint
	Price      float64
	RecordedAt time.Time
}

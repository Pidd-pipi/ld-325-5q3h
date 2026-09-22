package model

import "gorm.io/gorm"

type Supplier struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex;not null"`
	Address       string
	Contact       string
	Rating        float64
	Status        string
	Qualification string
	Offers        []Offer
}

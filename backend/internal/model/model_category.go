package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex;not null"`
	ParentID  *uint
	SortOrder int
	Products  []Product
}

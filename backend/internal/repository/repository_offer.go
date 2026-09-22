package repository

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type OfferRepository interface {
	ListByProduct(uint) ([]model.Offer, error)
	UpdateStatus(uint, string) (model.Offer, error)
}
type offerRepository struct{ db *gorm.DB }

func NewOfferRepository(db *gorm.DB) OfferRepository { return &offerRepository{db} }
func (r *offerRepository) ListByProduct(id uint) ([]model.Offer, error) {
	var data []model.Offer
	if err := r.db.Preload("Supplier").Where("product_id = ?", id).Order("unit_price ASC").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("list offers: %w", err)
	}
	return data, nil
}
func (r *offerRepository) UpdateStatus(id uint, status string) (model.Offer, error) {
	var item model.Offer
	if err := r.db.First(&item, id).Error; err != nil {
		return item, fmt.Errorf("find offer: %w", err)
	}
	item.StockStatus = status
	if err := r.db.Save(&item).Error; err != nil {
		return item, fmt.Errorf("update offer: %w", err)
	}
	return item, nil
}

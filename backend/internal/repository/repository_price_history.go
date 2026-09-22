package repository

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
	"time"
)

type PriceHistoryRepository interface {
	List(uint, time.Time) ([]model.PriceHistory, error)
}
type priceHistoryRepository struct{ db *gorm.DB }

func NewPriceHistoryRepository(db *gorm.DB) PriceHistoryRepository {
	return &priceHistoryRepository{db}
}
func (r *priceHistoryRepository) List(productID uint, since time.Time) ([]model.PriceHistory, error) {
	var rows []model.PriceHistory
	if err := r.db.Where("product_id = ? AND recorded_at >= ?", productID, since).Order("recorded_at ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	return rows, nil
}

package repository

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type UserDataRepository interface {
	CreateFavorite(model.Favorite) (model.Favorite, error)
	ListFavorites(string) ([]model.Favorite, error)
	CreateAlert(model.PriceAlert) (model.PriceAlert, error)
	CreateBudget(model.Budget) (model.Budget, error)
}
type userDataRepository struct{ db *gorm.DB }

func NewUserDataRepository(db *gorm.DB) UserDataRepository { return &userDataRepository{db} }
func (r *userDataRepository) CreateFavorite(value model.Favorite) (model.Favorite, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create favorite: %w", err)
	}
	return value, nil
}
func (r *userDataRepository) ListFavorites(user string) ([]model.Favorite, error) {
	var values []model.Favorite
	if err := r.db.Preload("Product").Where("user_id = ?", user).Find(&values).Error; err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	return values, nil
}
func (r *userDataRepository) CreateAlert(value model.PriceAlert) (model.PriceAlert, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create alert: %w", err)
	}
	return value, nil
}
func (r *userDataRepository) CreateBudget(value model.Budget) (model.Budget, error) {
	if err := r.db.Create(&value).Error; err != nil {
		return value, fmt.Errorf("create budget: %w", err)
	}
	return value, nil
}

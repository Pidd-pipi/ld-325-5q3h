package repository

import (
	"fmt"
	"strings"

	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	List(string, string, string, int, int) ([]model.Product, int64, error)
	Get(uint) (model.Product, error)
	Compare([]uint) ([]model.Product, error)
}
type productRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository { return &productRepository{db: db} }
func (r *productRepository) List(query, category, sort string, page, pageSize int) ([]model.Product, int64, error) {
	statement := r.db.Model(&model.Product{}).Preload("Category").Preload("Offers.Supplier")
	if query != "" {
		term := "%" + strings.ToLower(query) + "%"
		statement = statement.Where("LOWER(products.name) LIKE ? OR LOWER(products.brand) LIKE ?", term, term)
	}
	if category != "" {
		statement = statement.Joins("JOIN categories ON categories.id = products.category_id").Where("categories.name = ?", category)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}
	order := "products.created_at DESC"
	if sort == "sales" {
		order = "products.sales_count DESC"
	}
	if sort == "rating" {
		order = "products.rating DESC"
	}
	if sort == "price" {
		order = "(SELECT MIN(unit_price) FROM offers WHERE offers.product_id = products.id AND offers.deleted_at IS NULL) ASC"
	}
	var products []model.Product
	err := statement.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	return products, total, nil
}
func (r *productRepository) Get(id uint) (model.Product, error) {
	var p model.Product
	err := r.db.Preload("Category").Preload("Offers.Supplier").First(&p, id).Error
	if err == gorm.ErrRecordNotFound {
		return p, apperrors.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}
func (r *productRepository) Compare(ids []uint) ([]model.Product, error) {
	var p []model.Product
	err := r.db.Preload("Category").Preload("Offers.Supplier").Find(&p, ids).Error
	if err != nil {
		return nil, fmt.Errorf("compare products: %w", err)
	}
	return p, nil
}

package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseOrderRepository interface {
	WithTx(func(PurchaseOrderRepository) error) error
	FindByIdempotencyKey(string, string) (model.PurchaseOrder, error)
	LockEligibleOffers(uint) ([]model.Offer, error)
	CreateOrder(*model.PurchaseOrder) error
	DecrementOfferStock(uint, int) error
	GetByID(uint, string) (model.PurchaseOrder, error)
	ListByUser(string) ([]model.PurchaseOrder, error)
}

type purchaseOrderRepository struct{ db *gorm.DB }

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db}
}

func (r *purchaseOrderRepository) WithTx(fn func(PurchaseOrderRepository) error) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&purchaseOrderRepository{db: tx})
	}); err != nil {
		return fmt.Errorf("purchase order transaction: %w", err)
	}
	return nil
}

func (r *purchaseOrderRepository) FindByIdempotencyKey(userID, key string) (model.PurchaseOrder, error) {
	var order model.PurchaseOrder
	err := r.db.Preload("Product").Preload("Items.Supplier").
		Where("user_id = ? AND idempotency_key = ?", userID, key).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return order, apperrors.ErrNotFound
	}
	if err != nil {
		return order, fmt.Errorf("find purchase order by idempotency key: %w", err)
	}
	return order, nil
}

// LockEligibleOffers reads the in-stock offers of approved suppliers ordered
// by allocation preference, including offers whose stock is exhausted so the
// service can distinguish "no eligible offer" from "stock exhausted".
// PostgreSQL locks the rows FOR UPDATE so concurrent orders serialize on the
// same stock; sqlite (unit tests) has no row locks.
func (r *purchaseOrderRepository) LockEligibleOffers(productID uint) ([]model.Offer, error) {
	var offers []model.Offer
	statement := r.db.Preload("Supplier").
		Joins("JOIN suppliers ON suppliers.id = offers.supplier_id AND suppliers.deleted_at IS NULL").
		Where("offers.product_id = ? AND offers.stock_status = ? AND suppliers.status = ?",
			productID, constants.StatusInStock, constants.SupplierApproved).
		Order("offers.unit_price ASC, offers.delivery_days ASC, offers.id ASC")
	if r.db.Dialector.Name() == constants.PostgresDialect {
		statement = statement.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := statement.Find(&offers).Error; err != nil {
		return nil, fmt.Errorf("lock allocatable offers: %w", err)
	}
	return offers, nil
}

func (r *purchaseOrderRepository) CreateOrder(order *model.PurchaseOrder) error {
	if err := r.db.Omit("Product", "Items.Supplier").Create(order).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("create purchase order: %w", apperrors.ErrDuplicateKey)
		}
		return fmt.Errorf("create purchase order: %w", err)
	}
	return nil
}

// DecrementOfferStock applies the guarded stock deduction for one allocation;
// a zero row count means the stock changed concurrently and the whole
// transaction must roll back.
func (r *purchaseOrderRepository) DecrementOfferStock(offerID uint, quantity int) error {
	result := r.db.Model(&model.Offer{}).
		Where("id = ? AND available_qty >= ?", offerID, quantity).
		Update("available_qty", gorm.Expr("available_qty - ?", quantity))
	if result.Error != nil {
		return fmt.Errorf("decrement offer stock: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("decrement offer stock of %d: %w", offerID, apperrors.ErrStockChanged)
	}
	return nil
}

func (r *purchaseOrderRepository) GetByID(id uint, userID string) (model.PurchaseOrder, error) {
	var order model.PurchaseOrder
	err := r.db.Preload("Product").Preload("Items.Supplier").
		Where("id = ? AND user_id = ?", id, userID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return order, apperrors.ErrNotFound
	}
	if err != nil {
		return order, fmt.Errorf("get purchase order: %w", err)
	}
	return order, nil
}

func (r *purchaseOrderRepository) ListByUser(userID string) ([]model.PurchaseOrder, error) {
	var orders []model.PurchaseOrder
	err := r.db.Preload("Product").Preload("Items.Supplier").
		Where("user_id = ?", userID).
		Order("created_at DESC").Limit(constants.PurchaseOrderListLimit).Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("list purchase orders: %w", err)
	}
	return orders, nil
}

func isUniqueViolation(err error) bool {
	message := err.Error()
	return strings.Contains(message, "duplicate key value") || strings.Contains(message, "UNIQUE constraint failed")
}

package repository

import (
	"fmt"
	"strings"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PurchasePlanner maps a lock-consistent set of eligible offers to the desired
// quantities (implemented by service.planAllocation). ok=false rejects the order.
type PurchasePlanner func(offers []model.Offer, totalQuantity int) (quantities map[uint]int, ok bool)

type PurchaseRepository interface {
	FindByIdempotency(userID, key string) (model.PurchaseOrder, bool, error)
	// Place performs allocation, stock occupation and order creation in one
	// transaction. A rejected order is persisted as failed and returned so the
	// caller can expose the failure reason and support idempotent replay.
	Place(userID, key string, productID uint, totalQuantity int, planner PurchasePlanner) (model.PurchaseOrder, bool, error)
	GetByUserAndID(userID string, id uint) (model.PurchaseOrder, error)
}

type purchaseRepository struct{ db *gorm.DB }

func NewPurchaseRepository(db *gorm.DB) PurchaseRepository { return &purchaseRepository{db} }

func (r *purchaseRepository) FindByIdempotency(userID, key string) (model.PurchaseOrder, bool, error) {
	var order model.PurchaseOrder
	err := r.db.Preload("Items").Where("user_id = ? AND idempotency_key = ?", userID, key).First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return model.PurchaseOrder{}, false, nil
	}
	if err != nil {
		return model.PurchaseOrder{}, false, fmt.Errorf("find purchase by idempotency key: %w", err)
	}
	return order, true, nil
}

func (r *purchaseRepository) GetByUserAndID(userID string, id uint) (model.PurchaseOrder, error) {
	var order model.PurchaseOrder
	err := r.db.Preload("Items").Where("user_id = ?", userID).First(&order, id).Error
	if err == gorm.ErrRecordNotFound {
		return order, apperrors.ErrNotFound
	}
	if err != nil {
		return order, fmt.Errorf("get purchase order: %w", err)
	}
	return order, nil
}

func (r *purchaseRepository) Place(userID, key string, productID uint, totalQuantity int, planner PurchasePlanner) (model.PurchaseOrder, bool, error) {
	var failed model.PurchaseOrder
	var rejected bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Fast path inside the same transaction: duplicate keys serialize here.
		var existing model.PurchaseOrder
		findErr := tx.Where("user_id = ? AND idempotency_key = ?", userID, key).First(&existing).Error
		if findErr == nil {
			if err := tx.Preload("Items").First(&existing, existing.ID).Error; err != nil {
				return fmt.Errorf("reload idempotent purchase: %w", err)
			}
			failed = existing
			rejected = existing.Status == constants.PurchaseStatusFailed
			return nil
		}
		if findErr != gorm.ErrRecordNotFound {
			return fmt.Errorf("lookup idempotent purchase: %w", findErr)
		}

		var product model.Product
		if err := tx.First(&product, productID).Error; err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotFound
		} else if err != nil {
			return fmt.Errorf("load product for purchase: %w", err)
		}

		var offers []model.Offer
		query := tx.Joins("JOIN suppliers ON suppliers.id = offers.supplier_id").
			Preload("Supplier").
			Where("offers.product_id = ? AND offers.stock_status = ? AND suppliers.status = ?",
				productID, constants.StatusInStock, constants.SupplierApproved).
			Order("offers.unit_price ASC, offers.id ASC")
		// Postgres: lock only the offer rows (not suppliers) for the duration of
		// the transaction so concurrent purchases serialize on supply.
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "offers"}})
		}
		if err := query.Find(&offers).Error; err != nil {
			return fmt.Errorf("lock eligible offers: %w", err)
		}

		quantities, ok := planner(offers, totalQuantity)
		order := model.PurchaseOrder{
			UserID:         userID,
			IdempotencyKey: key,
			ProductID:      productID,
			ProductName:    product.Name,
			ProductUnit:    product.Unit,
			TotalQuantity:  totalQuantity,
		}
		if !ok {
			rejected = true
			order.Status = constants.PurchaseStatusFailed
			order.FailureCode, order.FailureReason = describeFailure(offers, totalQuantity)
			if err := tx.Create(&order).Error; err != nil {
				return fmt.Errorf("persist rejected purchase: %w", err)
			}
			failed = order
			return nil
		}

		items := buildOrderItems(offers, quantities)
		if !occupyStock(tx, offers, quantities) {
			rejected = true
			order.Status = constants.PurchaseStatusFailed
			order.FailureCode = constants.PurchaseFailureStockChanged
			order.FailureReason = constants.PurchaseReasonQuantityChanged
			if err := tx.Create(&order).Error; err != nil {
				return fmt.Errorf("persist conflicted purchase: %w", err)
			}
			failed = order
			return nil
		}

		totalAmount := 0.0
		maxDelivery := 0
		for i := range items {
			totalAmount += items[i].LineAmount
			if items[i].DeliveryDays > maxDelivery {
				maxDelivery = items[i].DeliveryDays
			}
		}
		order.Status = constants.PurchaseStatusSucceeded
		order.TotalAmount = totalAmount
		order.MaxDeliveryDays = maxDelivery
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("create purchase order: %w", err)
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("create purchase items: %w", err)
			}
		}
		order.Items = items
		failed = order
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			order, _, lookupErr := r.FindByIdempotency(userID, key)
			if lookupErr != nil {
				return model.PurchaseOrder{}, false, lookupErr
			}
			return order, false, nil
		}
		return model.PurchaseOrder{}, false, err
	}
	return failed, rejected, nil
}

func buildOrderItems(offers []model.Offer, quantities map[uint]int) []model.PurchaseOrderItem {
	byID := make(map[uint]model.Offer, len(offers))
	for _, offer := range offers {
		byID[offer.ID] = offer
	}
	items := make([]model.PurchaseOrderItem, 0, len(quantities))
	for offerID, quantity := range quantities {
		if quantity <= 0 {
			continue
		}
		offer := byID[offerID]
		items = append(items, model.PurchaseOrderItem{
			OfferID:         offerID,
			SupplierID:      offer.SupplierID,
			SupplierName:    offer.Supplier.Name,
			UnitPrice:       offer.UnitPrice,
			Quantity:        quantity,
			MOQ:             offer.MOQ,
			AvailableBefore: offer.AvailableStock,
			AvailableAfter:  offer.AvailableStock - quantity,
			DeliveryDays:    offer.DeliveryDays,
			Freight:         offer.Freight,
			LineAmount:      float64(quantity) * offer.UnitPrice,
		})
	}
	return items
}

// occupyStock performs conditional per-offer decrements so a quantity changed
// between planning and writing aborts the whole order with no partial writes.
func occupyStock(tx *gorm.DB, offers []model.Offer, quantities map[uint]int) bool {
	for _, offer := range offers {
		quantity := quantities[offer.ID]
		if quantity <= 0 {
			continue
		}
		result := tx.Model(&model.Offer{}).
			Where("id = ? AND available_stock >= ?", offer.ID, quantity).
			UpdateColumn("available_stock", gorm.Expr("available_stock - ?", quantity))
		if result.Error != nil || result.RowsAffected != 1 {
			return false
		}
	}
	return true
}

func describeFailure(offers []model.Offer, totalQuantity int) (string, string) {
	if len(offers) == 0 {
		return constants.PurchaseFailureNoAvailable, constants.PurchaseReasonNoInStockOffer
	}
	available := 0
	for _, offer := range offers {
		if offer.AvailableStock >= offer.MOQ {
			available += offer.AvailableStock
		}
	}
	if available == 0 {
		return constants.PurchaseFailureNoAvailable, constants.PurchaseReasonBelowMOQ
	}
	if available < totalQuantity {
		return constants.PurchaseFailureInsufficient,
			fmt.Sprintf(constants.PurchaseReasonInsufficientTemplate, available, totalQuantity)
	}
	return constants.PurchaseFailureInsufficient, constants.PurchaseReasonBelowMOQ
}

func isUniqueViolation(err error) bool {
	if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
		return true
	}
	return false
}

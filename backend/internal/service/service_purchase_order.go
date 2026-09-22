package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type PurchaseOrderService struct {
	orders   repository.PurchaseOrderRepository
	products repository.ProductRepository
	logger   *slog.Logger
}

func NewPurchaseOrderService(orders repository.PurchaseOrderRepository, products repository.ProductRepository, logger *slog.Logger) *PurchaseOrderService {
	return &PurchaseOrderService{orders, products, logger}
}

// Place creates a purchase order for the requested quantity. The order row,
// its allocation items and the stock deductions commit in one transaction;
// when the allocatable stock cannot cover the total, the order is recorded as
// rejected and no stock changes. Re-submitting the same idempotency key
// replays the stored result.
func (s *PurchaseOrderService) Place(userID string, req dto.CreatePurchaseOrderRequest) (dto.PurchaseOrderView, error) {
	if existing, err := s.orders.FindByIdempotencyKey(userID, req.IdempotencyKey); err == nil {
		return dto.NewPurchaseOrderView(existing), nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return dto.PurchaseOrderView{}, fmt.Errorf("check idempotency key: %w", err)
	}
	product, err := s.products.Get(req.ProductID)
	if err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("load product for purchase: %w", err)
	}
	var placed model.PurchaseOrder
	err = s.orders.WithTx(func(tx repository.PurchaseOrderRepository) error {
		offers, err := tx.LockEligibleOffers(req.ProductID)
		if err != nil {
			return err
		}
		plan, reason := Allocate(offers, req.Quantity)
		order := model.PurchaseOrder{
			UserID:         userID,
			ProductID:      product.ID,
			Quantity:       req.Quantity,
			IdempotencyKey: req.IdempotencyKey,
		}
		if reason == constants.RejectNone {
			order.Status = constants.PurchaseOrderConfirmed
			order.Items = buildItems(plan)
			for _, item := range order.Items {
				order.TotalAmount += item.Amount
				order.DeliveryDays = max(order.DeliveryDays, item.DeliveryDays)
			}
		} else {
			order.Status = constants.PurchaseOrderRejected
			order.FailureReason = rejectMessage(reason, product, req.Quantity, offers)
		}
		if err := tx.CreateOrder(&order); err != nil {
			return err
		}
		for _, allocation := range plan {
			if err := tx.DecrementOfferStock(allocation.Offer.ID, allocation.Quantity); err != nil {
				return err
			}
		}
		placed = order
		return nil
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateKey) {
			return s.replay(userID, req.IdempotencyKey)
		}
		return dto.PurchaseOrderView{}, fmt.Errorf("place purchase order: %w", err)
	}
	placed.Product = product
	s.logger.Info("purchase order placed", "order_id", placed.ID, "status", placed.Status, "quantity", placed.Quantity)
	return dto.NewPurchaseOrderView(placed), nil
}

func (s *PurchaseOrderService) Get(id uint, userID string) (dto.PurchaseOrderView, error) {
	order, err := s.orders.GetByID(id, userID)
	if err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("get purchase order: %w", err)
	}
	return dto.NewPurchaseOrderView(order), nil
}

func (s *PurchaseOrderService) List(userID string) ([]dto.PurchaseOrderView, error) {
	orders, err := s.orders.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("list purchase orders: %w", err)
	}
	views := make([]dto.PurchaseOrderView, 0, len(orders))
	for _, order := range orders {
		views = append(views, dto.NewPurchaseOrderView(order))
	}
	return views, nil
}

// replay returns the stored result when a concurrent request with the same
// idempotency key committed first.
func (s *PurchaseOrderService) replay(userID, key string) (dto.PurchaseOrderView, error) {
	existing, err := s.orders.FindByIdempotencyKey(userID, key)
	if err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("replay idempotent purchase order: %w", err)
	}
	return dto.NewPurchaseOrderView(existing), nil
}

func buildItems(plan []Allocation) []model.PurchaseOrderItem {
	items := make([]model.PurchaseOrderItem, 0, len(plan))
	for _, allocation := range plan {
		items = append(items, model.PurchaseOrderItem{
			OfferID:      allocation.Offer.ID,
			SupplierID:   allocation.Offer.SupplierID,
			Supplier:     allocation.Offer.Supplier,
			Quantity:     allocation.Quantity,
			UnitPrice:    allocation.Offer.UnitPrice,
			Amount:       float64(allocation.Quantity) * allocation.Offer.UnitPrice,
			DeliveryDays: allocation.Offer.DeliveryDays,
		})
	}
	return items
}

func rejectMessage(reason string, product model.Product, quantity int, offers []model.Offer) string {
	switch reason {
	case constants.RejectNoAvailableOffers:
		return constants.RejectMessageNoOffers
	case constants.RejectInsufficientStock:
		available := 0
		for _, offer := range offers {
			available += offer.AvailableQty
		}
		return fmt.Sprintf(constants.RejectMessageInsufficient, available, product.Unit, quantity, product.Unit)
	default:
		return constants.RejectMessageMOQ
	}
}

package service

import (
	"fmt"
	"net/http"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"log/slog"
)

type PurchaseService struct {
	repo   repository.PurchaseRepository
	logger *slog.Logger
}

func NewPurchaseService(repo repository.PurchaseRepository, logger *slog.Logger) *PurchaseService {
	return &PurchaseService{repo: repo, logger: logger}
}

// Place creates a split purchase order. A rejected (insufficient supply)
// request is a business failure carrying the persisted order detail; the
// handler converts it to a 409 response while keeping idempotent replay intact.
func (s *PurchaseService) Place(userID string, req dto.CreatePurchaseOrderRequest) (model.PurchaseOrder, error) {
	existing, found, err := s.repo.FindByIdempotency(userID, req.IdempotencyKey)
	if err != nil {
		return model.PurchaseOrder{}, fmt.Errorf("lookup purchase idempotency: %w", err)
	}
	if found {
		return existing, nil
	}
	order, rejected, err := s.repo.Place(userID, req.IdempotencyKey, req.ProductID, req.TotalQuantity, adaptPlanner)
	if err != nil {
		return model.PurchaseOrder{}, fmt.Errorf("place purchase order: %w", err)
	}
	if rejected {
		result := toPurchaseResult(order)
		s.logger.Warn("purchase order rejected", "order_id", order.ID, "reason", order.FailureCode)
		return order, apperrors.NewBusinessError(
			constants.ErrorPurchaseRejected, http.StatusConflict, order.FailureReason, result, nil)
	}
	return order, nil
}

func (s *PurchaseService) Get(userID string, id uint) (model.PurchaseOrder, error) {
	return s.repo.GetByUserAndID(userID, id)
}

// adaptPlanner bridges the repository transaction to the pure allocation logic.
func adaptPlanner(offers []model.Offer, totalQuantity int) (map[uint]int, bool) {
	lines, ok := planAllocation(offers, totalQuantity)
	if !ok {
		return nil, false
	}
	quantities := make(map[uint]int, len(lines))
	for _, line := range lines {
		quantities[line.Offer.OfferID] = line.Quantity
	}
	return quantities, true
}

// ToPurchaseResult maps the persisted aggregate to the external detail payload.
func ToPurchaseResult(order model.PurchaseOrder) dto.PurchaseOrderResult { return toPurchaseResult(order) }

func toPurchaseResult(order model.PurchaseOrder) dto.PurchaseOrderResult {
	items := make([]dto.PurchaseAllocationItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.PurchaseAllocationItem{
			OfferID:         item.OfferID,
			SupplierID:      item.SupplierID,
			SupplierName:    item.SupplierName,
			UnitPrice:       item.UnitPrice,
			Quantity:        item.Quantity,
			MOQ:             item.MOQ,
			AvailableBefore: item.AvailableBefore,
			AvailableAfter:  item.AvailableAfter,
			DeliveryDays:    item.DeliveryDays,
			Freight:         item.Freight,
			LineAmount:      item.LineAmount,
		})
	}
	return dto.PurchaseOrderResult{
		ID:              order.ID,
		ProductID:       order.ProductID,
		ProductName:     order.ProductName,
		ProductUnit:     order.ProductUnit,
		TotalQuantity:   order.TotalQuantity,
		Status:          order.Status,
		TotalAmount:     order.TotalAmount,
		MaxDeliveryDays: order.MaxDeliveryDays,
		FailureCode:     order.FailureCode,
		FailureReason:   order.FailureReason,
		Items:           items,
		CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

package dto

import (
	"time"

	"github.com/blueship581/cybuildprice/backend/internal/model"
)

type CreatePurchaseOrderRequest struct {
	ProductID      uint   `json:"product_id" validate:"required,gt=0"`
	Quantity       int    `json:"quantity" validate:"required,gt=0,lte=100000"`
	IdempotencyKey string `json:"idempotency_key" validate:"required,min=8,max=64"`
}

type PurchaseOrderItemView struct {
	OfferID      uint    `json:"offer_id"`
	SupplierID   uint    `json:"supplier_id"`
	SupplierName string  `json:"supplier_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	Amount       float64 `json:"amount"`
	DeliveryDays int     `json:"delivery_days"`
}

type PurchaseOrderView struct {
	ID             uint                    `json:"id"`
	ProductID      uint                    `json:"product_id"`
	ProductName    string                  `json:"product_name"`
	Unit           string                  `json:"unit"`
	Quantity       int                     `json:"quantity"`
	Status         string                  `json:"status"`
	FailureReason  string                  `json:"failure_reason,omitempty"`
	TotalAmount    float64                 `json:"total_amount"`
	DeliveryDays   int                     `json:"delivery_days"`
	IdempotencyKey string                  `json:"idempotency_key"`
	CreatedAt      time.Time               `json:"created_at"`
	Items          []PurchaseOrderItemView `json:"items"`
}

func NewPurchaseOrderView(order model.PurchaseOrder) PurchaseOrderView {
	view := PurchaseOrderView{
		ID:             order.ID,
		ProductID:      order.ProductID,
		ProductName:    order.Product.Name,
		Unit:           order.Product.Unit,
		Quantity:       order.Quantity,
		Status:         order.Status,
		FailureReason:  order.FailureReason,
		TotalAmount:    order.TotalAmount,
		DeliveryDays:   order.DeliveryDays,
		IdempotencyKey: order.IdempotencyKey,
		CreatedAt:      order.CreatedAt,
		Items:          make([]PurchaseOrderItemView, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		view.Items = append(view.Items, PurchaseOrderItemView{
			OfferID:      item.OfferID,
			SupplierID:   item.SupplierID,
			SupplierName: item.Supplier.Name,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			Amount:       item.Amount,
			DeliveryDays: item.DeliveryDays,
		})
	}
	return view
}

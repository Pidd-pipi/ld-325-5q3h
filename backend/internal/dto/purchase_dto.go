package dto

type CreatePurchaseOrderRequest struct {
	ProductID      uint   `json:"product_id" validate:"required,gt=0"`
	TotalQuantity  int    `json:"total_quantity" validate:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" validate:"required,min=8,max=64"`
}

type PurchaseAllocationItem struct {
	OfferID             uint    `json:"offer_id"`
	SupplierID          uint    `json:"supplier_id"`
	SupplierName        string  `json:"supplier_name"`
	UnitPrice           float64 `json:"unit_price"`
	Quantity            int     `json:"quantity"`
	MOQ                 int     `json:"moq"`
	AvailableBefore     int     `json:"available_before"`
	AvailableAfter      int     `json:"available_after"`
	DeliveryDays        int     `json:"delivery_days"`
	Freight             string  `json:"freight"`
	LineAmount          float64 `json:"line_amount"`
}

type PurchaseOrderResult struct {
	ID              uint                     `json:"id"`
	ProductID       uint                     `json:"product_id"`
	ProductName     string                   `json:"product_name"`
	ProductUnit     string                   `json:"product_unit"`
	TotalQuantity   int                      `json:"total_quantity"`
	Status          string                   `json:"status"`
	TotalAmount     float64                  `json:"total_amount"`
	MaxDeliveryDays int                      `json:"max_delivery_days"`
	FailureCode     string                   `json:"failure_code,omitempty"`
	FailureReason   string                   `json:"failure_reason,omitempty"`
	Items           []PurchaseAllocationItem `json:"items"`
	CreatedAt       string                   `json:"created_at"`
}

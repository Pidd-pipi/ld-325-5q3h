package dto

type CreateAlertRequest struct {
	ProductID   uint    `json:"product_id" validate:"required,gt=0"`
	TargetPrice float64 `json:"target_price" validate:"required,gt=0"`
	DropPercent float64 `json:"drop_percent" validate:"gte=0,lte=100"`
}

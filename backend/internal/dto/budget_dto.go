package dto

type BudgetRequest struct {
	RoomType string  `json:"room_type" validate:"required,oneof=living_room kitchen bathroom"`
	Area     float64 `json:"area" validate:"required,gte=1,lte=1000"`
}

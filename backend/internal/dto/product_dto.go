package dto

type ProductQuery struct {
	Query    string `form:"q"`
	Category string `form:"category"`
	Sort     string `form:"sort"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
type ProductCompareRequest struct {
	IDs []uint `json:"ids" validate:"required,min=2,max=4,dive,gt=0"`
}

package dto

type UpdateSupplierStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending approved rejected"`
}
type UpdateOfferStatusRequest struct {
	StockStatus string `json:"stock_status" validate:"required,oneof=in_stock out_of_stock discontinued"`
}

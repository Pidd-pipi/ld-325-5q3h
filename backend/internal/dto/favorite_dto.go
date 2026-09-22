package dto

type CreateFavoriteRequest struct {
	ProductID uint   `json:"product_id" validate:"required,gt=0"`
	Folder    string `json:"folder" validate:"required,max=30"`
}

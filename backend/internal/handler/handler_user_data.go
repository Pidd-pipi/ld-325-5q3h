package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserDataHandler struct {
	service  *service.UserDataService
	validate *validator.Validate
}

func NewUserDataHandler(s *service.UserDataService, v *validator.Validate) *UserDataHandler {
	return &UserDataHandler{s, v}
}
func (h *UserDataHandler) CreateFavorite(c *gin.Context) {
	var req dto.CreateFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	row, err := h.service.Favorite(c.GetString("user_id"), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, row)
}
func (h *UserDataHandler) ListFavorites(c *gin.Context) {
	rows, err := h.service.Favorites(c.GetString("user_id"))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, rows)
}
func (h *UserDataHandler) CreateAlert(c *gin.Context) {
	var req dto.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	row, err := h.service.Alert(c.GetString("user_id"), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, row)
}
func (h *UserDataHandler) CreateBudget(c *gin.Context) {
	var req dto.BudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	row, err := h.service.Budget(c.GetString("user_id"), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, row)
}

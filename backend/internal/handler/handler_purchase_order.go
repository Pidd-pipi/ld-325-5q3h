package handler

import (
	"strconv"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PurchaseOrderHandler struct {
	service  *service.PurchaseOrderService
	validate *validator.Validate
}

func NewPurchaseOrderHandler(s *service.PurchaseOrderService, v *validator.Validate) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{s, v}
}

func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	view, err := h.service.Place(c.GetString(constants.UserIDContextKey), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, view)
}

func (h *PurchaseOrderHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(apperrors.ErrInvalidInput)
		return
	}
	view, err := h.service.Get(uint(id), c.GetString(constants.UserIDContextKey))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, view)
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	views, err := h.service.List(c.GetString(constants.UserIDContextKey))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, views)
}

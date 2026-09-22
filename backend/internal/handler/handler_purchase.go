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

type PurchaseHandler struct {
	service  *service.PurchaseService
	validate *validator.Validate
}

func NewPurchaseHandler(s *service.PurchaseService, v *validator.Validate) *PurchaseHandler {
	return &PurchaseHandler{service: s, validate: v}
}

// Create places a split purchase order. Rejected orders surface as 409 with the
// same envelope shape so clients can render the failure reason; the persisted
// order remains readable via GET and idempotent keys replay the stored result.
func (h *PurchaseHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	order, err := h.service.Place(c.GetString(constants.UserIDContextKey), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, service.ToPurchaseResult(order))
}

func (h *PurchaseHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(apperrors.ErrInvalidInput)
		return
	}
	order, err := h.service.Get(c.GetString(constants.UserIDContextKey), uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, service.ToPurchaseResult(order))
}

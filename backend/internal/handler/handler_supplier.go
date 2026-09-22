package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SupplierHandler struct {
	service  *service.SupplierService
	validate *validator.Validate
}

func NewSupplierHandler(s *service.SupplierService, v *validator.Validate) *SupplierHandler {
	return &SupplierHandler{s, v}
}
func (h *SupplierHandler) List(c *gin.Context) {
	rows, err := h.service.List()
	if err != nil {
		c.Error(err)
		return
	}
	success(c, rows)
}
func (h *SupplierHandler) UpdateStatus(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	var req dto.UpdateSupplierStatusRequest
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	row, err := h.service.UpdateStatus(path.ID, req.Status)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, row)
}

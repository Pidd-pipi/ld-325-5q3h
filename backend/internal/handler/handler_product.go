package handler

import (
	"strconv"

	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProductHandler struct {
	service  *service.ProductService
	validate *validator.Validate
}

func NewProductHandler(s *service.ProductService, v *validator.Validate) *ProductHandler {
	return &ProductHandler{s, v}
}
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(err)
		return
	}
	rows, total, page, pageSize, err := h.service.List(q)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, gin.H{"items": rows, "total": total, "page": page, "page_size": pageSize})
}
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(apperrors.ErrInvalidInput)
		return
	}
	data, err := h.service.Get(uint(id))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}
func (h *ProductHandler) Compare(c *gin.Context) {
	var req dto.ProductCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	data, err := h.service.Compare(req.IDs)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

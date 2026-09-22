package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TrendHandler struct{ service *service.PriceHistoryService }

func NewTrendHandler(s *service.PriceHistoryService) *TrendHandler { return &TrendHandler{s} }
func (h *TrendHandler) Get(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
	rangeValue := c.DefaultQuery("range", constants.TrendDays30)
	data, err := h.service.Trend(path.ID, rangeValue)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

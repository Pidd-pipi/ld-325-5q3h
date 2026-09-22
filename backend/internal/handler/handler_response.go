package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/gin-gonic/gin"
	"net/http"
)

func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, dto.Response{Code: constants.SuccessCode, Message: constants.SuccessMessage, Data: data})
}

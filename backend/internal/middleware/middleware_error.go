package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		last := c.Errors.Last()
		err := last.Err
		code, status, message := constants.ErrorInternal, http.StatusInternalServerError, "internal server error"
		if errors.Is(err, apperrors.ErrUnauthorized) {
			code, status, message = constants.ErrorUnauthorized, http.StatusUnauthorized, "unauthorized"
		} else if errors.Is(err, apperrors.ErrNotFound) {
			code, status, message = constants.ErrorNotFound, http.StatusNotFound, "resource not found"
		} else if isClientError(err) || last.Type == gin.ErrorTypeBind {
			code, status, message = constants.ErrorValidation, http.StatusBadRequest, "validation failed"
		}
		c.JSON(status, dto.Response{Code: code, Message: message})
	}
}

func isClientError(err error) bool {
	if errors.Is(err, apperrors.ErrInvalidInput) {
		return true
	}
	var validation validator.ValidationErrors
	if errors.As(err, &validation) {
		return true
	}
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return true
	}
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return true
	}
	var numError *strconv.NumError
	if errors.As(err, &numError) {
		return true
	}
	return false
}

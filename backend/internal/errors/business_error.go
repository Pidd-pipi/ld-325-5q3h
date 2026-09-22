package errors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid request input")
	ErrUnauthorized = errors.New("unauthorized")
)

type BusinessError struct {
	Code    int
	Status  int
	Message string
	Data    any
	Err     error
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }

// NewBusinessError builds a structured business error. Cause may be nil.
func NewBusinessError(code, status int, message string, data any, cause error) *BusinessError {
	return &BusinessError{Code: code, Status: status, Message: message, Data: data, Err: cause}
}

package errors

import (
	"fmt"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Error codes
const (
	CodeNotFound             = 404
	CodeBadRequest           = 400
	CodeUnauthorized         = 401
	CodeForbidden            = 403
	CodeInternalError         = 500
	CodeInsufficientFunds    = 40001
	CodeInsufficientInventory = 40002
	CodeInvalidInput         = 40003
)

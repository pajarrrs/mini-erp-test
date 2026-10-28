package apperror

import "net/http"

type Error struct {
	Status int `json:"-"`
	Code string `json:"code"`
	Message string `json:"message"`
	Fields map[string]string `json:"fields,omitempty"`
	SKU string `json:"sku,omitempty"`
	Available *int `json:"available,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func Validation(fields map[string]string) *Error {
	return &Error{
		Status: http.StatusBadRequest,
		Code: "VALIDATION_ERROR",
		Message: "Some fields are invalid",
		Fields: fields,
	}
}

func BadRequest(message string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: message}
}

func NotFound(message string) *Error {
	return &Error{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: message}
}

func InsufficientStock(sku string, available int) *Error {
	return &Error{
		Status: http.StatusConflict,
		Code: "INSUFFICIENT_STOCK",
		Message: "Not enough stock for " + sku,
		SKU: sku,
		Available: &available,
	}
}

func Conflict(code, message string) *Error {
	return &Error{Status: http.StatusConflict, Code: code, Message: message}
}

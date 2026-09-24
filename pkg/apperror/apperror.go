// Package apperror defines structured service errors with HTTP status codes.
package apperror

import "net/http"

// AppError is a structured service error with an HTTP status.
type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string { return e.Message }

// New builds an AppError with an explicit status, code, and message.
func New(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

// BadRequest builds a 400 AppError.
func BadRequest(msg string) *AppError {
	return New(http.StatusBadRequest, "bad_request", msg)
}

// Conflict builds a 409 AppError.
func Conflict(msg string) *AppError {
	return New(http.StatusConflict, "conflict", msg)
}

// Internal builds a generic 500 AppError without leaking internals.
func Internal() *AppError {
	return New(http.StatusInternalServerError, "internal", "internal server error")
}

// Unauthorized builds a 401 AppError with a generic message.
func Unauthorized(msg string) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", msg)
}

// NotFound builds a 404 AppError.
func NotFound(msg string) *AppError {
	return New(http.StatusNotFound, "not_found", msg)
}

// TooManyRequests builds a 429 AppError.
func TooManyRequests(msg string) *AppError {
	return New(http.StatusTooManyRequests, "rate_limited", msg)
}

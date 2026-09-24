// Package apperror is a service error that already knows its HTTP answer.
package apperror

import "net/http"

type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string { return e.Message }

func New(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

func BadRequest(msg string) *AppError {
	return New(http.StatusBadRequest, "bad_request", msg)
}

func Conflict(msg string) *AppError {
	return New(http.StatusConflict, "conflict", msg)
}

// Internal hides details; the real error goes to the log, not the client.
func Internal() *AppError {
	return New(http.StatusInternalServerError, "internal", "internal server error")
}

func Unauthorized(msg string) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", msg)
}

func NotFound(msg string) *AppError {
	return New(http.StatusNotFound, "not_found", msg)
}

func TooManyRequests(msg string) *AppError {
	return New(http.StatusTooManyRequests, "rate_limited", msg)
}

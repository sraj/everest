package apperror

import (
	"fmt"
	"net/http"
)

// AppError is a structured application error with a machine-readable kind
// and an HTTP status code. Return this from handlers — the error handler
// middleware will serialize it to JSON with the right status.
type AppError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func newAppError(status int, kind, message string, err error) *AppError {
	return &AppError{Status: status, Kind: kind, Message: message, Err: err}
}

func BadRequest(format string, args ...any) *AppError {
	return newAppError(http.StatusBadRequest, "bad_request", fmt.Sprintf(format, args...), nil)
}

func NotFound(format string, args ...any) *AppError {
	return newAppError(http.StatusNotFound, "not_found", fmt.Sprintf(format, args...), nil)
}

func Internal(format string, args ...any) *AppError {
	return newAppError(http.StatusInternalServerError, "internal", fmt.Sprintf(format, args...), nil)
}

func Unauthorized(format string, args ...any) *AppError {
	return newAppError(http.StatusUnauthorized, "unauthorized", fmt.Sprintf(format, args...), nil)
}

func Forbidden(format string, args ...any) *AppError {
	return newAppError(http.StatusForbidden, "forbidden", fmt.Sprintf(format, args...), nil)
}

func Conflict(format string, args ...any) *AppError {
	return newAppError(http.StatusConflict, "conflict", fmt.Sprintf(format, args...), nil)
}

// Wrap wraps an underlying error into an AppError. Use this
// when you want to preserve the original error for logging.
func Wrap(err error, status int, kind, format string, args ...any) *AppError {
	return newAppError(status, kind, fmt.Sprintf(format, args...), err)
}

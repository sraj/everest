// Package bind provides reusable Fiber binding and validation helpers.
// It has zero dependencies on internal/ packages and can be used by
// any transport (HTTP, gRPC gateway, CLI, workers).
package bind

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Error wraps a binding or validation error.
type Error struct {
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Body binds the request body to target and validates it.
// Use BodyParser for JSON/XML and FormParser for multipart.
func Body(c *fiber.Ctx, target any) error {
	if err := c.BodyParser(target); err != nil {
		return &Error{Message: "invalid request body", Err: err}
	}
	if err := validate.Struct(target); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return &Error{Message: "validation failed", Err: ve}
		}
		return &Error{Message: "validation failed", Err: err}
	}
	return nil
}

// Query binds query parameters to target and validates them.
func Query(c *fiber.Ctx, target any) error {
	if err := c.QueryParser(target); err != nil {
		return &Error{Message: "invalid query parameters", Err: err}
	}
	if err := validate.Struct(target); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return &Error{Message: "validation failed", Err: ve}
		}
		return &Error{Message: "validation failed", Err: err}
	}
	return nil
}

// IsValidation reports whether err wraps a validator.ValidationErrors.
func IsValidation(err error) bool {
	var ve validator.ValidationErrors
	return errors.As(err, &ve)
}

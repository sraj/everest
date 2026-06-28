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

// RegisterValidation registers a custom validation tag.
// Use this at package init() to add domain-specific rules.
//
//	func init() {
//	    bind.RegisterValidation("doctitle", func(fl validator.FieldLevel) bool {
//	        return fl.Field().String() != "Untitled Document"
//	    })
//	}
func RegisterValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// RegisterStructValidation registers a struct-level validation function.
// Use this for cross-field business rules.
//
//	func (r SignupRequest) Validate() error { ... }
//	bind.RegisterStructValidation(SignupRequest.Validate, SignupRequest{})
func RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) {
	validate.RegisterStructValidation(fn, types...)
}

// HumanizeError converts validator.ValidationErrors into human-readable messages.
func HumanizeError(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	msgs := make(map[string]string, len(ve))
	for _, fe := range ve {
		msgs[fe.Field()] = msgForTag(fe)
	}
	return msgs
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "eqfield":
		return fmt.Sprintf("must match %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("invalid (%s)", fe.Tag())
	}
}

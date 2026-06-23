package store

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested resource does not exist.
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s: not found", e.Resource, e.ID)
}

func (e ErrNotFound) Unwrap() error {
	return errors.New("not found")
}

// ErrConflict is returned when a resource already exists or a constraint is violated.
type ErrConflict struct {
	Resource string
	Field    string
	Err      error
}

func (e ErrConflict) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s: already exists: %v", e.Resource, e.Field, e.Err)
	}
	return fmt.Sprintf("%s %s: already exists", e.Resource, e.Field)
}

func (e ErrConflict) Unwrap() error {
	return e.Err
}

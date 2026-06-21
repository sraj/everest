package dbx

import (
	"errors"
	"testing"
)

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}

	if ErrNotFound.Error() == "" {
		t.Error("ErrNotFound should have a message")
	}
}

func TestErrNoRows(t *testing.T) {
	if ErrNoRows == nil {
		t.Error("ErrNoRows should not be nil")
	}

	if ErrNoRows.Error() == "" {
		t.Error("ErrNoRows should have a message")
	}
}

func TestErrNotFound_Is(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("errors.Is should recognize ErrNotFound")
	}
}

func TestErrNoRows_Is(t *testing.T) {
	if !errors.Is(ErrNoRows, ErrNoRows) {
		t.Error("errors.Is should recognize ErrNoRows")
	}
}

func TestErrNotFound_NotEqual_ErrNoRows(t *testing.T) {
	if errors.Is(ErrNotFound, ErrNoRows) {
		t.Error("ErrNotFound should not equal ErrNoRows")
	}
}

func TestErrNoRows_NotEqual_ErrNotFound(t *testing.T) {
	if errors.Is(ErrNoRows, ErrNotFound) {
		t.Error("ErrNoRows should not equal ErrNotFound")
	}
}

func TestErrNotFound_CustomError_Check(t *testing.T) {
	// Test that we can use errors.Is with ErrNotFound
	err := ErrNotFound
	if !errors.Is(err, ErrNotFound) {
		t.Error("should be able to detect ErrNotFound with errors.Is")
	}

	if errors.Is(err, ErrNoRows) {
		t.Error("ErrNotFound should not match ErrNoRows")
	}
}

func TestErrNoRows_CustomError_Check(t *testing.T) {
	// Test that we can use errors.Is with ErrNoRows
	err := ErrNoRows
	if !errors.Is(err, ErrNoRows) {
		t.Error("should be able to detect ErrNoRows with errors.Is")
	}

	if errors.Is(err, ErrNotFound) {
		t.Error("ErrNoRows should not match ErrNotFound")
	}
}

func TestErrors_Distinct_Sentinel_Values(t *testing.T) {
	// Verify that the errors are distinct sentinel values
	if ErrNotFound == ErrNoRows {
		t.Error("ErrNotFound and ErrNoRows should be different sentinel values")
	}

	if &ErrNotFound == &ErrNoRows {
		t.Error("ErrNotFound and ErrNoRows should have different addresses")
	}
}

func TestErrNotFound_ErrorMessage_Content(t *testing.T) {
	msg := ErrNotFound.Error()
	if msg != "not found" {
		t.Errorf("ErrNotFound message should be 'not found', got '%s'", msg)
	}
}

func TestErrNoRows_ErrorMessage_Content(t *testing.T) {
	msg := ErrNoRows.Error()
	if msg != "no rows affected" {
		t.Errorf("ErrNoRows message should be 'no rows affected', got '%s'", msg)
	}
}

func TestErrors_Usage_Pattern(t *testing.T) {
	// Simulate typical usage pattern
	testError := func(err error) string {
		if errors.Is(err, ErrNotFound) {
			return "not found"
		} else if errors.Is(err, ErrNoRows) {
			return "no rows affected"
		}
		return "other error"
	}

	if result := testError(ErrNotFound); result != "not found" {
		t.Errorf("expected 'not found', got '%s'", result)
	}

	if result := testError(ErrNoRows); result != "no rows affected" {
		t.Errorf("expected 'no rows affected', got '%s'", result)
	}
}

func TestErrors_AsNilInterface(t *testing.T) {
	// Errors should not be nil, can be compared directly
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
	if ErrNoRows == nil {
		t.Error("ErrNoRows should not be nil")
	}
}

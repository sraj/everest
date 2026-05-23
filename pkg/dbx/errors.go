package dbx

import "errors"

// Sentinel errors returned by the query builder execution methods.
// Callers use errors.Is() to distinguish them.
var (
	// ErrNotFound is returned when GetContext finds no matching row.
	ErrNotFound = errors.New("not found")

	// ErrNoRows is returned by ExecMustAffect when 0 rows were changed.
	ErrNoRows = errors.New("no rows affected")
)

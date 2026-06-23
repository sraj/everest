package dbx

import (
	"errors"

	"github.com/lib/pq"
)

// Sentinel errors returned by the query builder execution methods.
// Callers use errors.Is() to distinguish them.
var (
	ErrNotFound = errors.New("not found")
	ErrNoRows   = errors.New("no rows affected")
)

// IsUniqueViolation reports whether err is a PostgreSQL unique constraint violation (code 23505).
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// IsForeignKeyError reports whether err is a PostgreSQL foreign key violation (code 23503).
func IsForeignKeyError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

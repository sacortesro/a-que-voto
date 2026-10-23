package utils

import (
	"errors"

	"github.com/lib/pq"
)

// pgCheckViolation is the SQLSTATE Postgres uses when a CHECK constraint fails.
const pgCheckViolation = "23514"

// IsCheckViolation reports whether err came from a failed CHECK constraint,
// which means the input was rejected by the database, not that the server broke.
func IsCheckViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgCheckViolation
}

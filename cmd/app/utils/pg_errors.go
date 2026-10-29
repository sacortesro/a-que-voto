package utils

import (
	"errors"

	"github.com/lib/pq"
)

// pgCheckViolation is the SQLSTATE Postgres uses when a CHECK constraint fails.
const pgCheckViolation = "23514"

// pgUniqueViolation is the SQLSTATE Postgres uses when a CHECK constraint fails.
const pgUniqueViolation = "23505"

func IsCheckViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgCheckViolation
}

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgUniqueViolation
}

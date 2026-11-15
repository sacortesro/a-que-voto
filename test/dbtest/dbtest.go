// Package dbtest gives tests a real Postgres. Tests open the database named by
// AQV_TEST_DB_URL and run inside a transaction that is always rolled back.
package dbtest

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const envURL = "AQV_TEST_DB_URL"

// Open connects to the migrated test database. It skips the test when the
// variable is unset; a skip is not evidence, so the gate treats it as a failure.
func Open(t testing.TB) *sql.DB {
	t.Helper()

	dsn := os.Getenv(envURL)
	if dsn == "" {
		t.Skip(envURL + " not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// Begin starts a transaction acting as a fresh actor and rolls it back on cleanup.
func Begin(t testing.TB, db *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { tx.Rollback() })

	var actorID string
	err = tx.QueryRow(`INSERT INTO actor (kind, name) VALUES ('system', $1) RETURNING id::text`,
		"dbtest."+uuid.NewString()).Scan(&actorID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	if _, err := tx.Exec(`SELECT set_config('app.actor_id', $1, true)`, actorID); err != nil {
		t.Fatalf("set actor: %v", err)
	}
	return tx
}

// ConstraintName returns the constraint Postgres reports for err, or "".
func ConstraintName(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Constraint
	}
	return ""
}

// AssertConstraint fails unless err was rejected by the named constraint.
func AssertConstraint(t testing.TB, err error, name string) {
	t.Helper()

	if err == nil {
		t.Fatalf("want rejection by constraint %q, got no error", name)
	}
	if got := ConstraintName(err); got != name {
		t.Fatalf("rejected by constraint %q, want %q (err: %v)", got, name, err)
	}
}

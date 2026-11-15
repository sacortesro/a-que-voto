package dbtest_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/sacortesro/a-que-voto/test/dbtest"
)

func TestConstraintName(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"pq error", &pq.Error{Constraint: "bill_kind_check"}, "bill_kind_check"},
		{"wrapped pq error", fmt.Errorf("create: %w", &pq.Error{Constraint: "source_code_key"}), "source_code_key"},
		{"pq error without constraint", &pq.Error{Message: "boom"}, ""},
		{"plain error", errors.New("boom"), ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dbtest.ConstraintName(tt.err); got != tt.want {
				t.Fatalf("ConstraintName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBeginSetsActorAndRollsBack(t *testing.T) {
	db := dbtest.Open(t)
	code := "dbtest-" + t.Name()

	t.Run("insert inside tx uses the actor", func(t *testing.T) {
		tx := dbtest.Begin(t, db)
		var createdBy, actor string
		err := tx.QueryRow(`INSERT INTO source (code, base_url, priority) VALUES ($1, 'https://example.test', 1)
			RETURNING created_by::text, app_actor()::text`, code).Scan(&createdBy, &actor)
		if err != nil {
			t.Fatalf("insert source: %v", err)
		}
		if createdBy == "" || createdBy != actor {
			t.Fatalf("created_by = %q, app_actor = %q, want equal and non-empty", createdBy, actor)
		}
	})

	t.Run("previous tx left nothing behind", func(t *testing.T) {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM source WHERE code = $1`, code).Scan(&n); err != nil {
			t.Fatalf("count: %v", err)
		}
		if n != 0 {
			t.Fatalf("source rows = %d, want 0 (rolled back)", n)
		}
	})
}

func TestAssertConstraintOnRealViolation(t *testing.T) {
	db := dbtest.Open(t)
	tx := dbtest.Begin(t, db)

	_, err := tx.Exec(`INSERT INTO source (code, base_url, priority) VALUES ('dup', 'https://example.test', 1)`)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	_, err = tx.Exec(`INSERT INTO source (code, base_url, priority) VALUES ('dup', 'https://example.test', 1)`)

	dbtest.AssertConstraint(t, err, "source_code_key")
}

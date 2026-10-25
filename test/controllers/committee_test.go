package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var committeeColumns = []string{
	"id", "chamber", "kind", "name", "number", "created_at", "created_by", "updated_at", "updated_by",
}

const (
	committeeID   = "0196c2a8-7b1e-7000-8000-0000000000f1"
	committeeName = "Comisión Primera Constitucional Permanente"
)

func TestHandleCreateCommittee(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	insertStep := fakedb.Step{
		Query:   "INSERT INTO committee",
		Args:    []driver.Value{"senado", "constitucional", committeeName, int64(1)},
		Columns: committeeColumns,
		Rows: [][]driver.Value{{
			committeeID, "senado", "constitucional", committeeName, int64(1), now, actorID, now, actorID,
		}},
	}
	body := `{"chamber":"senado","kind":"constitucional","name":"` + committeeName + `","number":1}`

	t.Run("creates committee", func(t *testing.T) {
		api, fake := newAPIConfig(t, setActor(), insertStep)

		rec := invoke(t, api.HandleCreateCommittee, body, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Committee
		decode(t, rec, &got)
		if got.ID != uuid.MustParse(committeeID) {
			t.Errorf("id = %s, want %s", got.ID, committeeID)
		}
		if got.Name != committeeName || got.Number != 1 {
			t.Errorf("committee = %q #%d, want %q #1", got.Name, got.Number, committeeName)
		}
		if got.CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got.CreatedAt)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("creates committee without number", func(t *testing.T) {
		noNumber := insertStep
		noNumber.Args = []driver.Value{"camara", "accidental", "Comisión Accidental", nil}
		noNumber.Rows = [][]driver.Value{{
			committeeID, "camara", "accidental", "Comisión Accidental", nil, now, actorID, now, actorID,
		}}
		api, _ := newAPIConfig(t, setActor(), noNumber)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":"camara","kind":"accidental","name":"Comisión Accidental"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Committee
		decode(t, rec, &got)
		if got.Number != 0 {
			t.Errorf("number = %d, want 0", got.Number)
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects unknown chamber without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":"congreso","kind":"legal","name":"X"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"chamber not found"}`)
	})

	t.Run("rejects unknown kind without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":"senado","kind":"especial","name":"X"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee kind"}`)
	})

	t.Run("rejects empty name without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":"senado","kind":"legal","name":""}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee name"}`)
	})

	t.Run("rejects whitespace-only name without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateCommittee, `{"chamber":"senado","kind":"legal","name":"  \t "}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee name"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		failing := insertStep
		failing.Err = errors.New("unique violation")
		api, fake := newAPIConfig(t, setActor(), failing)

		rec := invoke(t, api.HandleCreateCommittee, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

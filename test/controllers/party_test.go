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

var partyColumns = []string{"id", "name", "acronym", "created_at", "created_by", "updated_at", "updated_by", "is_active"}

func TestHandleCreateParty(t *testing.T) {
	const partyID = "0196c2a8-7b1e-7000-8000-0000000000a1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("creates party with acronym", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO party",
				Args:    []driver.Value{"Movimiento Pacto Histórico", "PH"},
				Columns: partyColumns,
				Rows:    [][]driver.Value{{partyID, "Movimiento Pacto Histórico", "PH", now, actorID, now, actorID, true}},
			},
		)

		rec := invoke(t, api.HandleCreateParty, `{"name":"Movimiento Pacto Histórico","acronym":"PH"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Party
		decode(t, rec, &got)
		if got.ID != uuid.MustParse(partyID) {
			t.Errorf("id = %s, want %s", got.ID, partyID)
		}
		if got.Acronym != "PH" {
			t.Errorf("acronym = %q, want %q", got.Acronym, "PH")
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("creates party without acronym", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO party",
				Args:    []driver.Value{"Partido Verde", nil},
				Columns: partyColumns,
				Rows:    [][]driver.Value{{partyID, "Partido Verde", nil, now, actorID, now, actorID, true}},
			},
		)

		rec := invoke(t, api.HandleCreateParty, `{"name":"Partido Verde"}`, true)

		assertStatus(t, rec, http.StatusCreated)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		// No statements are scripted: the handler must stop before touching the database.
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateParty, `{"name":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("fails without actor in context", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateParty, `{"name":"Partido Verde"}`, false)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query: "INSERT INTO party",
				Err:   errors.New("duplicate key value violates unique constraint"),
			},
		)

		rec := invoke(t, api.HandleCreateParty, `{"name":"Partido Verde"}`, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
		if fake.Committed {
			t.Error("transaction committed despite the insert error")
		}
	})
}

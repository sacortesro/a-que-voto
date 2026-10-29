package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

func TestHandleGetParties(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns active parties ordered by name", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM party WHERE is_active = true ORDER BY name",
				Args:    []driver.Value{},
				Columns: partyColumns,
				Rows: [][]driver.Value{
					{"0196c2a8-7b1e-7000-8000-0000000000a1", "Movimiento Pacto Histórico", "PH", now, actorID, now, actorID, true},
					{"0196c2a8-7b1e-7000-8000-0000000000a2", "Partido Verde", nil, now, actorID, now, actorID, true},
				},
			},
		)

		rec := invoke(t, api.HandleGetParties, "", true)

		assertStatus(t, rec, http.StatusOK)
		var got []models.Party
		decode(t, rec, &got)
		if len(got) != 2 || got[0].Name != "Movimiento Pacto Histórico" || got[1].Name != "Partido Verde" {
			t.Fatalf("parties = %+v, want Movimiento Pacto Histórico then Partido Verde", got)
		}
		if got[0].Acronym != "PH" || got[1].Acronym != "" {
			t.Errorf("acronyms = %q, %q, want PH and empty", got[0].Acronym, got[1].Acronym)
		}
		if got[0].CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got[0].CreatedAt)
		}
	})

	t.Run("returns empty array when there are no active parties", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM party", Columns: partyColumns},
		)

		rec := invoke(t, api.HandleGetParties, "", true)

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM party", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleGetParties, "", true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

func TestHandleUpdateParty(t *testing.T) {
	const partyID = "0196c2a8-7b1e-7000-8000-0000000000a1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("updates party with trimmed name", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "UPDATE party",
				Args:    []driver.Value{partyID, "Partido Verde", "PV", false},
				Columns: partyColumns,
				Rows:    [][]driver.Value{{partyID, "Partido Verde", "PV", now, actorID, now, actorID, false}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"  Partido Verde  ","acronym":"PV","is_active":false}`)

		assertStatus(t, rec, http.StatusOK)
		var got models.Party
		decode(t, rec, &got)
		if got.Name != "Partido Verde" || got.Acronym != "PV" || got.IsActive {
			t.Errorf("party = %+v, want Partido Verde / PV / inactive", got)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("clears acronym when omitted", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "UPDATE party",
				Args:    []driver.Value{partyID, "Partido Verde", nil, true},
				Columns: partyColumns,
				Rows:    [][]driver.Value{{partyID, "Partido Verde", nil, now, actorID, now, actorID, true}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"Partido Verde","is_active":true}`)

		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdateParty, "abc", `{"name":"Partido Verde","is_active":true}`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid party id"}`)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects missing is_active", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"Partido Verde"}`)

		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("rejects blank name", func(t *testing.T) {
		for _, body := range []string{`{"is_active":true}`, `{"name":"   ","is_active":true}`} {
			api, _ := newAPIConfig(t)

			rec := invokeWithID(t, api.HandleUpdateParty, partyID, body)

			assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid name"}`)
		}
	})

	t.Run("returns 404 for unknown party", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "UPDATE party", Columns: partyColumns},
		)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"Partido Verde","is_active":true}`)

		assertBody(t, rec, http.StatusNotFound, `{"error":"party not found"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 409 when name already exists", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "UPDATE party", Err: &pq.Error{Code: "23505", Message: "party_name_key"}},
		)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"Partido Verde","is_active":true}`)

		assertBody(t, rec, http.StatusConflict, `{"error":"Name already exists"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "UPDATE party", Err: errors.New("connection refused")},
		)

		rec := invokeWithID(t, api.HandleUpdateParty, partyID, `{"name":"Partido Verde","is_active":true}`)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

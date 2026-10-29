package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var districtColumns = []string{
	"id", "chamber", "kind", "department", "name", "created_at", "created_by", "updated_at", "updated_by",
}

func TestHandleCreateDistrict(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("creates national district without department", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO district",
				Args:    []driver.Value{"senado", "estatuto_oposicion", nil, "Estatuto oposición"},
				Columns: districtColumns,
				Rows:    [][]driver.Value{{int64(4), "senado", "estatuto_oposicion", nil, "Estatuto oposición", now, actorID, now, actorID}},
			},
		)

		rec := invoke(t, api.HandleCreateDistrict, `{"chamber":"senado","kind":"estatuto_oposicion","department":null,"name":"Estatuto oposición"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.District
		decode(t, rec, &got)
		if got.ID != 4 {
			t.Errorf("id = %d, want 4", got.ID)
		}
		if got.Department != "" {
			t.Errorf("department = %q, want empty", got.Department)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("creates department district", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO district",
				Args:    []driver.Value{"senado", "territorial", "santander", "Circunscripción en Santander"},
				Columns: districtColumns,
				Rows:    [][]driver.Value{{int64(2), "senado", "territorial", "santander", "Circunscripción en Santander", now, actorID, now, actorID}},
			},
		)

		rec := invoke(t, api.HandleCreateDistrict, `{"chamber":"senado","kind":"territorial","department":"santander","name":"Circunscripción en Santander"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.District
		decode(t, rec, &got)
		if got.Department != "santander" {
			t.Errorf("department = %q, want santander", got.Department)
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateDistrict, `{"chamber":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query: "INSERT INTO district",
				Err:   errors.New("check constraint violated"),
			},
		)

		rec := invoke(t, api.HandleCreateDistrict, `{"chamber":"senado","kind":"nacional","name":"X"}`, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

func TestHandleListDistricts(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns districts ordered by id", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM district ORDER BY id",
				Args:    []driver.Value{},
				Columns: districtColumns,
				Rows: [][]driver.Value{
					{int64(1), "senado", "nacional", nil, "Nacional", now, actorID, now, actorID},
					{int64(2), "camara", "territorial", "Antioquia", "Antioquia", now, actorID, now, actorID},
				},
			},
		)

		rec := invoke(t, api.HandleListDistricts, "", true)

		assertStatus(t, rec, http.StatusOK)
		var got []models.District
		decode(t, rec, &got)
		if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
			t.Fatalf("districts = %+v, want ids 1 then 2", got)
		}
		if got[0].Department != "" || got[1].Department != "Antioquia" {
			t.Errorf("departments = %q, %q, want empty and Antioquia", got[0].Department, got[1].Department)
		}
		if got[0].CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got[0].CreatedAt)
		}
	})

	t.Run("returns empty array when there are no districts", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM district", Columns: districtColumns},
		)

		rec := invoke(t, api.HandleListDistricts, "", true)

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM district", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleListDistricts, "", true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

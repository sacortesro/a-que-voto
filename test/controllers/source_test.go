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

var sourceColumns = []string{
	"id", "code", "base_url", "priority", "created_at", "created_by", "updated_at", "updated_by",
}

func TestHandleCreateSource(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	body := `{"code":"camara","base_url":"https://www.camara.gov.co","priority":1}`

	t.Run("creates source", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			fakedb.Step{Query: "FROM source WHERE code", Args: []driver.Value{"camara"}, Columns: sourceColumns},
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO source",
				Args:    []driver.Value{"camara", "https://www.camara.gov.co", int64(1)},
				Columns: sourceColumns,
				Rows:    [][]driver.Value{{int64(1), "camara", "https://www.camara.gov.co", int64(1), now, actorID, now, actorID}},
			},
		)

		rec := invoke(t, api.HandleCreateSource, body, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Source
		decode(t, rec, &got)
		if got.ID != 1 || got.Code != "camara" {
			t.Errorf("source = %+v, want id 1 code camara", got)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("rejects duplicate code", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM source WHERE code",
				Columns: sourceColumns,
				Rows:    [][]driver.Value{{int64(1), "camara", "https://www.camara.gov.co", int64(1), now, actorID, now, actorID}},
			},
		)

		rec := invoke(t, api.HandleCreateSource, body, true)

		assertBody(t, rec, http.StatusConflict, `{"error":"code already exists"}`)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreateSource, `{"code":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("returns 500 when code lookup fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM source WHERE code", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleCreateSource, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

func TestHandleGetAllSources(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns sources ordered by id", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM source ORDER BY id",
				Columns: sourceColumns,
				Rows: [][]driver.Value{
					{int64(1), "camara", "https://www.camara.gov.co", int64(1), now, actorID, now, actorID},
					{int64(2), "congresovisible", "https://congresovisible.uniandes.edu.co", int64(3), now, actorID, now, actorID},
				},
			},
		)

		rec := invoke(t, api.HandleGetAllSources, "", true)

		assertStatus(t, rec, http.StatusOK)
		var got []models.Source
		decode(t, rec, &got)
		if len(got) != 2 || got[0].Code != "camara" || got[1].Code != "congresovisible" {
			t.Errorf("sources = %+v, want camara then congresovisible", got)
		}
	})

	t.Run("returns empty array when there are no sources", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM source", Columns: sourceColumns},
		)

		rec := invoke(t, api.HandleGetAllSources, "", true)

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM source", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleGetAllSources, "", true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

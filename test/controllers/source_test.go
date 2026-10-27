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

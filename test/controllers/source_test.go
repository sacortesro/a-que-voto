package controllers_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var sourceColumns = []string{
	"id", "code", "base_url", "priority", "created_at", "created_by", "updated_at", "updated_by",
}

// invokeWithID runs handler with {id} set the way chi sets it when routing.
func invokeWithID(t *testing.T, handler http.HandlerFunc, id, body string) *httptest.ResponseRecorder {
	t.Helper()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(auth.WithActor(ctx, testActor))

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
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

func TestHandleUpdateSource(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("updates base_url and priority", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "UPDATE source",
				Args:    []driver.Value{int64(3), "https://www.camara.gov.co/api", int64(2)},
				Columns: sourceColumns,
				Rows:    [][]driver.Value{{int64(3), "camara", "https://www.camara.gov.co/api", int64(2), now, actorID, now, actorID}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdateSource, "3", `{"base_url":"https://www.camara.gov.co/api/","priority":2}`)

		assertStatus(t, rec, http.StatusOK)
		var got models.Source
		decode(t, rec, &got)
		if got.ID != 3 || got.Priority != 2 {
			t.Errorf("source = %+v, want id 3 priority 2", got)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("ignores code in the body", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "UPDATE source",
				Args:    []driver.Value{int64(3), "https://www.camara.gov.co", int64(1)},
				Columns: sourceColumns,
				Rows:    [][]driver.Value{{int64(3), "camara", "https://www.camara.gov.co", int64(1), now, actorID, now, actorID}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdateSource, "3", `{"code":"renamed","base_url":"https://www.camara.gov.co","priority":1}`)

		assertStatus(t, rec, http.StatusOK)
		var got models.Source
		decode(t, rec, &got)
		if got.Code != "camara" {
			t.Errorf("code = %q, want camara", got.Code)
		}
	})

	t.Run("returns 404 for unknown source", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "UPDATE source", Columns: sourceColumns},
		)

		rec := invokeWithID(t, api.HandleUpdateSource, "99", `{"base_url":"https://example.org","priority":1}`)

		assertBody(t, rec, http.StatusNotFound, `{"error":"source not found"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "UPDATE source", Err: errors.New("connection refused")},
		)

		rec := invokeWithID(t, api.HandleUpdateSource, "3", `{"base_url":"https://example.org","priority":1}`)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		for _, id := range []string{"abc", "", "70000"} {
			api, _ := newAPIConfig(t)

			rec := invokeWithID(t, api.HandleUpdateSource, id, `{"base_url":"https://example.org","priority":1}`)

			assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid source id"}`)
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdateSource, "3", `{"base_url":`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects base_url that is not an https URL", func(t *testing.T) {
		for _, raw := range []string{"", "%zz", "localhost", "www.camara.gov.co", "http://www.camara.gov.co", "ftp://files.example.org", "https://"} {
			t.Run(raw, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeWithID(t, api.HandleUpdateSource, "3", `{"base_url":"`+raw+`","priority":1}`)

				assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid URL. base_url must be https URL"}`)
			})
		}
	})

	t.Run("rejects priority below 1", func(t *testing.T) {
		for _, body := range []string{
			`{"base_url":"https://example.org","priority":0}`,
			`{"base_url":"https://example.org","priority":-1}`,
			`{"base_url":"https://example.org"}`,
		} {
			t.Run(body, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeWithID(t, api.HandleUpdateSource, "3", body)

				assertBody(t, rec, http.StatusBadRequest, `{"error":"priority must be greater than 0"}`)
			})
		}
	})
}

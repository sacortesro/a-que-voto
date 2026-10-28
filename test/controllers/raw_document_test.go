package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var rawDocumentColumns = []string{
	"id", "source_id", "url", "content_sha256", "storage_key", "content_type", "fetched_at", "fetch_run_id", "created_at", "created_by",
}

func invokeWithQuery(t *testing.T, handler http.HandlerFunc, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(auth.WithActor(req.Context(), testActor))

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestHandleGetRawDocuments(t *testing.T) {
	newer := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	older := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	runID := "0196c2a8-7b1e-7000-8000-0000000000aa"

	t.Run("returns documents of the source newest first", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM raw_document WHERE source_id = $1 ORDER BY fetched_at DESC LIMIT 20",
				Args:    []driver.Value{int64(2)},
				Columns: rawDocumentColumns,
				Rows: [][]driver.Value{
					{"0196c2a8-7b1e-7000-8000-000000000011", int64(2), "https://example.org/a", []byte{0x01}, "raw/a", "text/html", newer, runID, newer, actorID},
					{"0196c2a8-7b1e-7000-8000-000000000012", int64(2), "https://example.org/b", []byte{0x02}, "raw/b", nil, older, nil, older, actorID},
				},
			},
		)

		rec := invokeWithQuery(t, api.HandleGetRawDocuments, "/raw_document?source_id=2")

		assertStatus(t, rec, http.StatusOK)
		var got []models.RawDocument
		decode(t, rec, &got)
		if len(got) != 2 {
			t.Fatalf("documents = %+v, want 2", got)
		}
		if got[0].URL != "https://example.org/a" || got[0].ContentType != "text/html" || got[0].FetchRunID.String() != runID {
			t.Errorf("first document = %+v", got[0])
		}
		if got[0].FetchedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("fetched_at = %q, want 2026-10-22T10:00:00Z", got[0].FetchedAt)
		}
		if got[1].ContentType != "" {
			t.Errorf("NULL content_type = %q, want empty", got[1].ContentType)
		}
	})

	t.Run("returns empty array when the source has no documents", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM raw_document", Columns: rawDocumentColumns},
		)

		rec := invokeWithQuery(t, api.HandleGetRawDocuments, "/raw_document?source_id=2")

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM raw_document", Err: errors.New("connection refused")},
		)

		rec := invokeWithQuery(t, api.HandleGetRawDocuments, "/raw_document?source_id=2")

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("rejects invalid source_id", func(t *testing.T) {
		for _, target := range []string{"/raw_document", "/raw_document?source_id=", "/raw_document?source_id=abc", "/raw_document?source_id=70000"} {
			t.Run(target, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeWithQuery(t, api.HandleGetRawDocuments, target)

				assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid source"}`)
			})
		}
	})
}

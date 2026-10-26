package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var stageColumns = []string{
	"slug", "label", "sort_order", "created_at", "created_by", "updated_at", "updated_by",
}

var documentKindColumns = []string{
	"slug", "label", "created_at", "created_by", "updated_at", "updated_by",
}

func TestHandleGetAllStages(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns stages in query order", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM stage ORDER BY sort_order",
				Columns: stageColumns,
				Rows: [][]driver.Value{
					{"radicado", "Radicado", int64(1), now, actorID, now, actorID},
					{"primer_debate", "Primer debate", int64(2), now, actorID, now, actorID},
				},
			},
		)

		rec := invoke(t, api.HandleGetAllStages, "", true)

		assertBody(t, rec, http.StatusOK,
			`[{"slug":"radicado","label":"Radicado","created_at":"2026-10-22T10:00:00Z","updated_at":"2026-10-22T10:00:00Z"},`+
				`{"slug":"primer_debate","label":"Primer debate","created_at":"2026-10-22T10:00:00Z","updated_at":"2026-10-22T10:00:00Z"}]`)
	})

	t.Run("returns empty array when there are no stages", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM stage", Columns: stageColumns},
		)

		rec := invoke(t, api.HandleGetAllStages, "", true)

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM stage", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleGetAllStages, "", true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

func TestHandleGetAllDocumentKinds(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns document kinds in query order", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM document_kind ORDER BY slug",
				Columns: documentKindColumns,
				Rows: [][]driver.Value{
					{"gaceta", "Gaceta", now, actorID, now, actorID},
					{"ponencia", "Ponencia", now, actorID, now, actorID},
				},
			},
		)

		rec := invoke(t, api.HandleGetAllDocumentKinds, "", true)

		assertBody(t, rec, http.StatusOK,
			`[{"slug":"gaceta","label":"Gaceta","created_at":"2026-10-22T10:00:00Z","updated_at":"2026-10-22T10:00:00Z"},`+
				`{"slug":"ponencia","label":"Ponencia","created_at":"2026-10-22T10:00:00Z","updated_at":"2026-10-22T10:00:00Z"}]`)
	})

	t.Run("returns empty array when there are no document kinds", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM document_kind", Columns: documentKindColumns},
		)

		rec := invoke(t, api.HandleGetAllDocumentKinds, "", true)

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM document_kind", Err: errors.New("connection refused")},
		)

		rec := invoke(t, api.HandleGetAllDocumentKinds, "", true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

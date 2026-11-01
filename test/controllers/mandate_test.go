package controllers_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var mandateColumns = []string{
	"id", "person_id", "term_id", "chamber", "district_id", "party_id", "start_date", "end_date",
	"replaces_mandate_id", "status", "created_at", "created_by", "updated_at", "updated_by",
}

const (
	mandatePersonID = "0196c2a8-7b1e-7000-8000-0000000000c1"
	mandatePartyID  = "0196c2a8-7b1e-7000-8000-0000000000d1"
	mandateID       = "0196c2a8-7b1e-7000-8000-0000000000e1"
	mandatePerson   = "Gustavo Francisco Petro Urrego"
	mandateTerm     = "2018 - 2022"
	mandateParty    = "Movimiento Pacto Histórico"
)

// mandateBody builds the request. department is a raw JSON value: null or "santander".
func mandateBody(department string) string {
	return fmt.Sprintf(`{
	"person": %q,
	"term": %q,
	"chamber": "senado",
	"district_kind": "estatuto_oposicion",
	"district_department": %s,
	"party": %q,
	"start_date": "2018-07-20",
	"end_date": "2022-07-19"
}`, mandatePerson, mandateTerm, department, mandateParty)
}

// lookupSteps scripts the four reads that run before the insert.
// Pass nil for districtRows to simulate a district that does not exist.
func lookupSteps(department driver.Value, districtRows [][]driver.Value) []fakedb.Step {
	return []fakedb.Step{
		{
			Query:   "FROM person",
			Args:    []driver.Value{mandatePerson},
			Columns: []string{"id"},
			Rows:    [][]driver.Value{{mandatePersonID}},
		},
		{
			Query:   "FROM term",
			Args:    []driver.Value{mandateTerm},
			Columns: []string{"id"},
			Rows:    [][]driver.Value{{int64(2)}},
		},
		{
			Query:   "FROM district",
			Args:    []driver.Value{"senado", "estatuto_oposicion", department},
			Columns: []string{"id"},
			Rows:    districtRows,
		},
		{
			Query:   "FROM party",
			Args:    []driver.Value{mandateParty},
			Columns: []string{"id"},
			Rows:    [][]driver.Value{{mandatePartyID}},
		},
	}
}

func TestHandlerCreateMandate(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	startDate := time.Date(2018, 7, 20, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 7, 19, 0, 0, 0, 0, time.UTC)

	insertStep := fakedb.Step{
		Query: "INSERT INTO mandate",
		// Args: person, term, chamber, district, party, start, end, replaces (nil), status.
		Args:    []driver.Value{mandatePersonID, int64(2), "senado", int64(4), mandatePartyID, startDate, endDate, nil, controllers.StatusActive},
		Columns: mandateColumns,
		Rows: [][]driver.Value{{
			mandateID, mandatePersonID, int64(2), "senado", int64(4), mandatePartyID,
			startDate, endDate, nil, controllers.StatusActive, now, actorID, now, actorID,
		}},
	}

	t.Run("creates mandate for a district without department", func(t *testing.T) {
		steps := append(lookupSteps(nil, [][]driver.Value{{int64(4)}}), setActor(), insertStep)
		api, fake := newAPIConfig(t, steps...)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Mandate
		decode(t, rec, &got)
		if got.ID != uuid.MustParse(mandateID) {
			t.Errorf("id = %s, want %s", got.ID, mandateID)
		}
		if got.DistrictID != 4 {
			t.Errorf("district_id = %d, want 4", got.DistrictID)
		}
		if got.StartDate != "2018-07-20" || got.EndDate != "2022-07-19" {
			t.Errorf("dates = %q - %q, want 2018-07-20 - 2022-07-19", got.StartDate, got.EndDate)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("creates open-ended mandate with empty end date", func(t *testing.T) {
		openEnded := insertStep
		openEnded.Args = []driver.Value{mandatePersonID, int64(2), "senado", int64(4), mandatePartyID, startDate, nil, nil, controllers.StatusActive}
		openEnded.Rows = [][]driver.Value{{
			mandateID, mandatePersonID, int64(2), "senado", int64(4), mandatePartyID,
			startDate, nil, nil, controllers.StatusActive, now, actorID, now, actorID,
		}}
		steps := append(lookupSteps(nil, [][]driver.Value{{int64(4)}}), setActor(), openEnded)
		api, _ := newAPIConfig(t, steps...)
		body := strings.Replace(mandateBody("null"), `"end_date": "2022-07-19"`, `"end_date": null`, 1)

		rec := invoke(t, api.HandlerCreateMandate, body, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Mandate
		decode(t, rec, &got)
		if got.EndDate != "" {
			t.Errorf("end_date = %q, want empty", got.EndDate)
		}
	})

	t.Run("rejects malformed start date", func(t *testing.T) {
		api, _ := newAPIConfig(t)
		body := strings.Replace(mandateBody("null"), `"2018-07-20"`, `"2018-07-20T00:00:00Z"`, 1)

		rec := invoke(t, api.HandlerCreateMandate, body, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"start_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rejects end date before start date", func(t *testing.T) {
		api, _ := newAPIConfig(t)
		body := strings.Replace(mandateBody("null"), `"2022-07-19"`, `"2017-07-19"`, 1)

		rec := invoke(t, api.HandlerCreateMandate, body, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"end_date must not be before start_date"}`)
	})

	t.Run("creates mandate for a district with department", func(t *testing.T) {
		steps := append(lookupSteps("santander", [][]driver.Value{{int64(4)}}), setActor(), insertStep)
		api, _ := newAPIConfig(t, steps...)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody(`"santander"`), true)

		assertStatus(t, rec, http.StatusCreated)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandlerCreateMandate, `{"person":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("reports unknown person", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{
			Query:   "FROM person",
			Columns: []string{"id"},
		})

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"person not found"}`)
	})

	t.Run("reports unknown term", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person", Columns: []string{"id"}, Rows: [][]driver.Value{{mandatePersonID}}},
			fakedb.Step{Query: "FROM term", Columns: []string{"id"}},
		)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"term not found"}`)
	})

	t.Run("rejects unknown chamber without touching the database", func(t *testing.T) {
		body := fmt.Sprintf(`{"person":%q,"term":%q,"chamber":"congreso","district_kind":"estatuto_oposicion","party":%q,"start_date":"2018-07-20"}`,
			mandatePerson, mandateTerm, mandateParty)
		// No statements are scripted: validation runs before any lookup.
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandlerCreateMandate, body, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"chamber not found"}`)
	})

	t.Run("reports missing district", func(t *testing.T) {
		api, _ := newAPIConfig(t, lookupSteps(nil, nil)[:3]...)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"district not found"}`)
	})

	t.Run("reports unknown party", func(t *testing.T) {
		steps := lookupSteps(nil, [][]driver.Value{{int64(4)}})
		steps[3].Rows = nil
		api, _ := newAPIConfig(t, steps[:4]...)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"party not found"}`)
	})

	t.Run("reports database failure during lookup as internal error", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{
			Query: "FROM person",
			Err:   errors.New("connection refused"),
		})

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		failing := insertStep
		failing.Err = errors.New("foreign key violation")
		steps := append(lookupSteps(nil, [][]driver.Value{{int64(4)}}), setActor(), failing)
		api, fake := newAPIConfig(t, steps...)

		rec := invoke(t, api.HandlerCreateMandate, mandateBody("null"), true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

// invokeListMandates calls the handler with {id} as the chi URL param and query as the raw query string.
func invokeListMandates(t *testing.T, handler http.HandlerFunc, id, query string) *httptest.ResponseRecorder {
	t.Helper()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	req := httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestHandlerGetMandatesByPerson(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	start2018 := time.Date(2018, 7, 20, 0, 0, 0, 0, time.UTC)
	end2022 := time.Date(2022, 7, 19, 0, 0, 0, 0, time.UTC)
	start2022 := time.Date(2022, 7, 20, 0, 0, 0, 0, time.UTC)
	const ongoingID = "0196c2a8-7b1e-7000-8000-0000000000e2"

	t.Run("returns overlapping mandates including an ongoing one", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{
			Query: "FROM mandate WHERE person_id = $1",
			Args: []driver.Value{
				mandatePersonID,
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			Columns: mandateColumns,
			Rows: [][]driver.Value{
				{mandateID, mandatePersonID, int64(2), "senado", int64(4), mandatePartyID,
					start2018, end2022, nil, controllers.StatusActive, now, actorID, now, actorID},
				{ongoingID, mandatePersonID, int64(3), "senado", int64(4), mandatePartyID,
					start2022, nil, nil, controllers.StatusActive, now, actorID, now, actorID},
			},
		})

		rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, mandatePersonID, "from_date=2020-01-01&to_date=2024-12-31")

		assertStatus(t, rec, http.StatusOK)
		var got []models.Mandate
		decode(t, rec, &got)
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0].ID != uuid.MustParse(mandateID) || got[0].EndDate != "2022-07-19" {
			t.Errorf("first mandate = %+v", got[0])
		}
		if got[1].ID != uuid.MustParse(ongoingID) || got[1].EndDate != "" {
			t.Errorf("ongoing mandate = %+v, want empty end_date", got[1])
		}
	})

	t.Run("defaults from_date to 2000-01-01 when omitted", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{
			Query:   "FROM mandate WHERE person_id = $1",
			Args:    []driver.Value{mandatePersonID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), fakedb.Any},
			Columns: mandateColumns,
		})

		rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, mandatePersonID, "")

		assertBody(t, rec, http.StatusOK, "[]")
	})

	t.Run("rejects invalid person id", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, "not-a-uuid", "")

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid person id"}`)
	})

	t.Run("rejects malformed dates", func(t *testing.T) {
		for _, key := range []string{"from_date", "to_date"} {
			t.Run(key, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, mandatePersonID, key+"=2020/01/01")

				assertBody(t, rec, http.StatusBadRequest, `{"error":"`+key+` must be a date of format YYYY-MM-DD"}`)
			})
		}
	})

	t.Run("rejects from_date after to_date without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, mandatePersonID, "from_date=2024-01-01&to_date=2020-01-01")

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid date range"}`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{
			Query: "FROM mandate WHERE person_id = $1",
			Err:   errors.New("connection refused"),
		})

		rec := invokeListMandates(t, api.HandlerGetMandatesByPerson, mandatePersonID, "")

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

package controllers_test

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var (
	termColumns        = []string{"id", "label", "start_date", "end_date", "created_at", "created_by", "updated_at", "updated_by"}
	legislatureColumns = []string{"id", "term_id", "label", "start_date", "end_date", "created_at", "created_by", "updated_at", "updated_by"}
	periodNow          = time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
)

// periodBody builds a term request from raw date strings.
func periodBody(start, end string) string {
	return fmt.Sprintf(`{"start_date":%q,"end_date":%q}`, start, end)
}

func termStep(label string) fakedb.Step {
	actorID := testActor.ID.String()
	return fakedb.Step{
		Query:   "INSERT INTO term",
		Args:    []driver.Value{label, fakedb.Any, fakedb.Any},
		Columns: termColumns,
		Rows: [][]driver.Value{{
			int64(2), label,
			time.Date(2018, 7, 20, 0, 0, 0, 0, time.UTC), time.Date(2022, 7, 19, 0, 0, 0, 0, time.UTC),
			periodNow, actorID, periodNow, actorID,
		}},
	}
}

func legislatureStep(label string, id int64) fakedb.Step {
	actorID := testActor.ID.String()
	return fakedb.Step{
		Query:   "INSERT INTO legislature",
		Args:    []driver.Value{int64(2), label, fakedb.Any, fakedb.Any},
		Columns: legislatureColumns,
		Rows:    [][]driver.Value{{id, int64(2), label, periodNow, periodNow, periodNow, actorID, periodNow, actorID}},
	}
}

func TestHandleCreatePeriod(t *testing.T) {
	t.Run("creates term with its four legislatures", func(t *testing.T) {
		steps := []fakedb.Step{setActor(), termStep("2018 - 2022")}
		for i, label := range []string{"2018 - 2019", "2019 - 2020", "2020 - 2021", "2021 - 2022"} {
			steps = append(steps, legislatureStep(label, int64(i+1)))
		}
		api, fake := newAPIConfig(t, steps...)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2018-07-20", "2022-07-19"), true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Term
		decode(t, rec, &got)
		if got.Label != "2018 - 2022" {
			t.Errorf("label = %q, want %q", got.Label, "2018 - 2022")
		}
		if got.StartDate != "2018-07-20" || got.EndDate != "2022-07-19" {
			t.Errorf("dates = %q - %q, want 2018-07-20 - 2022-07-19", got.StartDate, got.EndDate)
		}
		if got.CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got.CreatedAt)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("rolls back the term when a legislature fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			termStep("2018 - 2022"),
			legislatureStep("2018 - 2019", 1),
			fakedb.Step{Query: "INSERT INTO legislature", Err: errors.New("connection reset")},
		)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2018-07-20", "2022-07-19"), true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
		if fake.Committed {
			t.Error("transaction committed despite the legislature error")
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePeriod, `{"start_date":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects malformed start date", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2018/07/20", "2022-07-19"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"start_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rejects start date outside an election year", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2019-07-20", "2023-07-19"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"start_date must be 20 July of an election year"}`)
	})

	t.Run("rejects end date that is not 19 July", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2018-07-20", "2022-07-18"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"end_date must be 19 July"}`)
	})

	t.Run("rejects term that does not span four years", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePeriod, periodBody("2018-07-20", "2023-07-19"), true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"term must span 4 years, got 5"}`)
	})
}

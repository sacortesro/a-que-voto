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

var committeeMembershipColumns = []string{
	"mandate_id", "committee_id", "role", "start_date", "end_date",
	"created_at", "created_by", "updated_at", "updated_by",
}

func TestHandleAddPersonToCommittee(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 7, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	mandateStep := fakedb.Step{
		Query:   "FROM mandate",
		Args:    []driver.Value{mandateID},
		Columns: []string{"id"},
		Rows:    [][]driver.Value{{mandateID}},
	}
	committeeStep := fakedb.Step{
		Query:   "FROM committee",
		Args:    []driver.Value{committeeID},
		Columns: []string{"id"},
		Rows:    [][]driver.Value{{committeeID}},
	}
	noMembershipStep := fakedb.Step{
		Query:   "FROM committee_membership",
		Args:    []driver.Value{mandateID, committeeID, start},
		Columns: []string{"mandate_id", "committee_id", "start_date"},
	}
	insertStep := fakedb.Step{
		Query:   "INSERT INTO committee_membership",
		Args:    []driver.Value{mandateID, committeeID, "presidente", start, nil},
		Columns: committeeMembershipColumns,
		Rows: [][]driver.Value{{
			mandateID, committeeID, "presidente", start, nil, now, actorID, now, actorID,
		}},
	}
	body := `{"mandate_id":"` + mandateID + `","committee_id":"` + committeeID + `","role":"presidente","start_date":"2026-07-20"}`

	t.Run("adds membership without end_date", func(t *testing.T) {
		api, fake := newAPIConfig(t, mandateStep, committeeStep, noMembershipStep, setActor(), insertStep)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.CommitteeMembership
		decode(t, rec, &got)
		if got.MandateID != uuid.MustParse(mandateID) || got.CommitteeID != uuid.MustParse(committeeID) {
			t.Errorf("ids = %s/%s, want %s/%s", got.MandateID, got.CommitteeID, mandateID, committeeID)
		}
		if got.Role != "presidente" {
			t.Errorf("role = %q, want presidente", got.Role)
		}
		if got.StartDate != "2026-07-20" || got.EndDate != "" {
			t.Errorf("dates = %q..%q, want 2026-07-20..\"\"", got.StartDate, got.EndDate)
		}
		if got.CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got.CreatedAt)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("adds membership with end_date", func(t *testing.T) {
		withEnd := insertStep
		withEnd.Args = []driver.Value{mandateID, committeeID, "presidente", start, end}
		withEnd.Rows = [][]driver.Value{{
			mandateID, committeeID, "presidente", start, end, now, actorID, now, actorID,
		}}
		api, _ := newAPIConfig(t, mandateStep, committeeStep, noMembershipStep, setActor(), withEnd)

		rec := invoke(t, api.HandleAddPersonToCommittee, body[:len(body)-1]+`,"end_date":"2027-07-19"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.CommitteeMembership
		decode(t, rec, &got)
		if got.EndDate != "2027-07-19" {
			t.Errorf("end_date = %q, want 2027-07-19", got.EndDate)
		}
	})

	t.Run("accepts end_date equal to start_date", func(t *testing.T) {
		sameDay := insertStep
		sameDay.Args = []driver.Value{mandateID, committeeID, "presidente", start, start}
		sameDay.Rows = [][]driver.Value{{
			mandateID, committeeID, "presidente", start, start, now, actorID, now, actorID,
		}}
		api, _ := newAPIConfig(t, mandateStep, committeeStep, noMembershipStep, setActor(), sameDay)

		rec := invoke(t, api.HandleAddPersonToCommittee, body[:len(body)-1]+`,"end_date":"2026-07-20"}`, true)

		assertStatus(t, rec, http.StatusCreated)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"mandate_id":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects invalid role without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"role":"secretario","start_date":"2026-07-20"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee role"}`)
	})

	t.Run("rejects empty role without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"start_date":"2026-07-20"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee role"}`)
	})

	t.Run("rejects malformed start_date without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"role":"miembro","start_date":"20/07/2026"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"start_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rejects malformed end_date without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"role":"miembro","start_date":"2026-07-20","end_date":"2027-13-01"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"end_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rejects end_date before start_date without touching the database", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleAddPersonToCommittee, `{"role":"miembro","start_date":"2026-07-20","end_date":"2026-07-19"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"end_date must not be before start_date"}`)
	})

	t.Run("rejects unknown mandate", func(t *testing.T) {
		missing := mandateStep
		missing.Rows = nil
		api, _ := newAPIConfig(t, missing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"mandate not found"}`)
	})

	t.Run("rejects unknown committee", func(t *testing.T) {
		missing := committeeStep
		missing.Rows = nil
		api, _ := newAPIConfig(t, mandateStep, missing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"committee not found"}`)
	})

	t.Run("rejects duplicate membership", func(t *testing.T) {
		existing := noMembershipStep
		existing.Rows = [][]driver.Value{{mandateID, committeeID, start}}
		api, _ := newAPIConfig(t, mandateStep, committeeStep, existing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusConflict, `{"error":"committee membership already exists"}`)
	})

	t.Run("returns 500 when mandate lookup fails", func(t *testing.T) {
		failing := mandateStep
		failing.Err = errors.New("connection reset")
		api, _ := newAPIConfig(t, failing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("returns 500 when committee lookup fails", func(t *testing.T) {
		failing := committeeStep
		failing.Err = errors.New("connection reset")
		api, _ := newAPIConfig(t, mandateStep, failing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("returns 500 when membership lookup fails", func(t *testing.T) {
		failing := noMembershipStep
		failing.Err = errors.New("connection reset")
		api, _ := newAPIConfig(t, mandateStep, committeeStep, failing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		failing := insertStep
		failing.Err = errors.New("foreign key violation")
		api, fake := newAPIConfig(t, mandateStep, committeeStep, noMembershipStep, setActor(), failing)

		rec := invoke(t, api.HandleAddPersonToCommittee, body, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

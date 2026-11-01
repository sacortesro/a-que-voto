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

func TestHandleListCommitteeMembers(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 7, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	const otherMandateID = "0196c2a8-7b1e-7000-8000-0000000000e2"

	committeeStep := fakedb.Step{
		Query:   "FROM committee WHERE id = $1",
		Args:    []driver.Value{committeeID},
		Columns: []string{"id"},
		Rows:    [][]driver.Value{{committeeID}},
	}

	t.Run("returns members ordered by start date", func(t *testing.T) {
		api, _ := newAPIConfig(t, committeeStep, fakedb.Step{
			Query:   "FROM committee_membership WHERE committee_id = $1 ORDER BY start_date, mandate_id",
			Args:    []driver.Value{committeeID},
			Columns: committeeMembershipColumns,
			Rows: [][]driver.Value{
				{mandateID, committeeID, "presidente", start, end, now, actorID, now, actorID},
				{otherMandateID, committeeID, "miembro", start, nil, now, actorID, now, actorID},
			},
		})

		rec := invokeWithID(t, api.HandleListCommitteeMembers, committeeID, "")

		assertStatus(t, rec, http.StatusOK)
		var got []models.CommitteeMembership
		decode(t, rec, &got)
		if len(got) != 2 || got[0].MandateID != uuid.MustParse(mandateID) || got[1].MandateID != uuid.MustParse(otherMandateID) {
			t.Fatalf("members = %+v", got)
		}
		if got[0].Role != "presidente" || got[0].EndDate != "2027-07-19" || got[1].EndDate != "" {
			t.Errorf("members = %+v", got)
		}
	})

	t.Run("returns empty array when the committee has no members", func(t *testing.T) {
		api, _ := newAPIConfig(t, committeeStep, fakedb.Step{
			Query:   "FROM committee_membership",
			Columns: committeeMembershipColumns,
		})

		rec := invokeWithID(t, api.HandleListCommitteeMembers, committeeID, "")

		assertBody(t, rec, http.StatusOK, "[]")
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleListCommitteeMembers, "not-a-uuid", "")

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid committee id"}`)
	})

	t.Run("returns 404 when the committee does not exist", func(t *testing.T) {
		api, _ := newAPIConfig(t, fakedb.Step{Query: "FROM committee WHERE id = $1", Columns: []string{"id"}})

		rec := invokeWithID(t, api.HandleListCommitteeMembers, committeeID, "")

		assertBody(t, rec, http.StatusNotFound, `{"error":"committee not found"}`)
	})

	t.Run("returns 500 when the list query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t, committeeStep, fakedb.Step{
			Query: "FROM committee_membership",
			Err:   errors.New("connection refused"),
		})

		rec := invokeWithID(t, api.HandleListCommitteeMembers, committeeID, "")

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

func TestHandleUpdateCommitteeMembership(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 7, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	updateStep := fakedb.Step{
		Query:   "UPDATE committee_membership",
		Args:    []driver.Value{"presidente", end, mandateID, committeeID, start},
		Columns: committeeMembershipColumns,
		Rows:    [][]driver.Value{{mandateID, committeeID, "presidente", start, end, now, actorID, now, actorID}},
	}
	body := `{"mandate_id":"` + mandateID + `","start_date":"2026-07-20","role":"presidente","end_date":"2027-07-19"}`

	t.Run("updates role and end date", func(t *testing.T) {
		api, fake := newAPIConfig(t, setActor(), updateStep)

		rec := invokeWithID(t, api.HandleUpdateCommitteeMembership, committeeID, body)

		assertStatus(t, rec, http.StatusOK)
		var got models.CommitteeMembership
		decode(t, rec, &got)
		if got.MandateID != uuid.MustParse(mandateID) || got.CommitteeID != uuid.MustParse(committeeID) {
			t.Errorf("membership = %+v", got)
		}
		if got.Role != "presidente" || got.StartDate != "2026-07-20" || got.EndDate != "2027-07-19" {
			t.Errorf("membership = %+v, want presidente 2026-07-20 - 2027-07-19", got)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("clears end date when omitted", func(t *testing.T) {
		open := updateStep
		open.Args = []driver.Value{"miembro", nil, mandateID, committeeID, start}
		open.Rows = [][]driver.Value{{mandateID, committeeID, "miembro", start, nil, now, actorID, now, actorID}}
		api, _ := newAPIConfig(t, setActor(), open)

		rec := invokeWithID(t, api.HandleUpdateCommitteeMembership, committeeID,
			`{"mandate_id":"`+mandateID+`","start_date":"2026-07-20","role":"miembro"}`)

		assertStatus(t, rec, http.StatusOK)
		var got models.CommitteeMembership
		decode(t, rec, &got)
		if got.EndDate != "" {
			t.Errorf("end_date = %q, want empty", got.EndDate)
		}
	})

	withMandate := func(rest string) string {
		return `{"mandate_id":"` + mandateID + `",` + rest + `}`
	}
	badRequests := []struct {
		name, id, body, want string
	}{
		{"rejects invalid committee id", "not-a-uuid", body, `{"error":"invalid committee id"}`},
		{"rejects invalid JSON", committeeID, `{"mandate_id":`, `{"error":"invalid JSON body"}`},
		{"rejects unknown fields", committeeID, withMandate(`"start_date":"2026-07-20","role":"miembro","committee_id":"x"`), `{"error":"invalid JSON body"}`},
		{"rejects malformed mandate id", committeeID, `{"mandate_id":"nope","start_date":"2026-07-20","role":"miembro"}`, `{"error":"invalid JSON body"}`},
		{"rejects missing mandate id", committeeID, `{"start_date":"2026-07-20","role":"miembro"}`, `{"error":"mandate_id is required"}`},
		{"rejects malformed start date", committeeID, withMandate(`"start_date":"20/07/2026","role":"miembro"`), `{"error":"start_date must be formatted as YYYY-MM-DD"}`},
		{"rejects invalid role", committeeID, withMandate(`"start_date":"2026-07-20","role":"rey"`), `{"error":"invalid committee role"}`},
		{"rejects missing role", committeeID, withMandate(`"start_date":"2026-07-20"`), `{"error":"invalid committee role"}`},
		{"rejects malformed end date", committeeID, withMandate(`"start_date":"2026-07-20","role":"miembro","end_date":"mañana"`), `{"error":"end_date must be formatted as YYYY-MM-DD"}`},
		{"rejects end date before start date", committeeID, withMandate(`"start_date":"2026-07-20","role":"miembro","end_date":"2026-07-19"`), `{"error":"end_date must not be before start_date"}`},
	}
	for _, tc := range badRequests {
		t.Run(tc.name, func(t *testing.T) {
			api, _ := newAPIConfig(t)

			rec := invokeWithID(t, api.HandleUpdateCommitteeMembership, tc.id, tc.body)

			assertBody(t, rec, http.StatusBadRequest, tc.want)
		})
	}

	t.Run("returns 404 when the membership does not exist", func(t *testing.T) {
		api, fake := newAPIConfig(t, setActor(), fakedb.Step{Query: "UPDATE committee_membership", Columns: committeeMembershipColumns})

		rec := invokeWithID(t, api.HandleUpdateCommitteeMembership, committeeID, body)

		assertBody(t, rec, http.StatusNotFound, `{"error":"committee membership not found"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		api, fake := newAPIConfig(t, setActor(), fakedb.Step{Query: "UPDATE committee_membership", Err: errors.New("connection refused")})

		rec := invokeWithID(t, api.HandleUpdateCommitteeMembership, committeeID, body)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

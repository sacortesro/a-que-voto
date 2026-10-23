package controllers_test

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var personColumns = []string{
	"id", "full_name", "name_norm", "birth_date", "photo_url", "social_media",
	"merged_into_id", "created_at", "created_by", "updated_at", "updated_by", "id_congreso_visible",
}

const personBody = `{
	"full_name": "Gustavo Francisco Petro Urrego",
	"birth_date": "1960-04-19",
	"photo_url": "https://example.com/photo.jpg",
	"social_media": [{"platform": "twitter", "url": "https://x.com/petrogustavo"}]
}`

func TestHandleCreatePerson(t *testing.T) {
	const personID = "0196c2a8-7b1e-7000-8000-0000000000b1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	birthDate := time.Date(1960, 4, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	socialMedia := `[{"platform":"twitter","url":"https://x.com/petrogustavo"}]`

	t.Run("creates person", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO person",
				Args:    []driver.Value{"Gustavo Francisco Petro Urrego", birthDate, "https://example.com/photo.jpg", fakedb.Any, nil, nil},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Gustavo Francisco Petro Urrego", "gustavo francisco petro urrego",
					birthDate, "https://example.com/photo.jpg", []byte(socialMedia),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		rec := invoke(t, api.HandleCreatePerson, personBody, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Person
		decode(t, rec, &got)
		if got.ID != uuid.MustParse(personID) {
			t.Errorf("id = %s, want %s", got.ID, personID)
		}
		if got.NameNorm != "gustavo francisco petro urrego" {
			t.Errorf("name_norm = %q", got.NameNorm)
		}
		if got.BirthDate != "1960-04-19" {
			t.Errorf("birth date = %q, want 1960-04-19", got.BirthDate)
		}
		for _, field := range []string{`"birth_date":"1960-04-19"`, `"created_at":"2026-10-22T10:00:00Z"`} {
			if !strings.Contains(rec.Body.String(), field) {
				t.Errorf("body %s does not contain %s", rec.Body, field)
			}
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePerson, `{"full_name":`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects malformed birth date", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invoke(t, api.HandleCreatePerson, `{"full_name":"X","birth_date":"1960-04-19T00:00:00Z"}`, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"birth_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rolls back when insert fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query: "INSERT INTO person",
				Err:   errors.New("connection reset"),
			},
		)

		rec := invoke(t, api.HandleCreatePerson, personBody, true)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("defaults social media to an empty array when omitted", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "INSERT INTO person",
				Args:    []driver.Value{"Ana Pérez", nil, nil, []byte(`[]`), nil, nil},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Ana Pérez", "ana perez",
					nil, nil, []byte(`[]`),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		rec := invoke(t, api.HandleCreatePerson, `{"full_name":"Ana Pérez"}`, true)

		assertStatus(t, rec, http.StatusCreated)
		var got models.Person
		decode(t, rec, &got)
		if got.BirthDate != "" {
			t.Errorf("birth date = %q, want empty for NULL", got.BirthDate)
		}
	})

	t.Run("reports check constraint violation as bad input", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query: "INSERT INTO person",
				Err:   &pq.Error{Code: "23514", Message: "valid_social_media"},
			},
		)

		rec := invoke(t, api.HandleCreatePerson, personBody, true)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"social_media is invalid"}`)
	})
}

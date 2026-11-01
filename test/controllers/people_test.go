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

func TestHandleListPersons(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns persons ordered by name with default pagination", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "WHERE merged_into_id IS NULL ORDER BY name_norm, id LIMIT $1 OFFSET $2",
				Args:    []driver.Value{int64(50), int64(0)},
				Columns: personColumns,
				Rows: [][]driver.Value{
					{"0196c2a8-7b1e-7000-8000-0000000000b1", "Ana Pérez", "ana perez", nil, nil, []byte(`[]`), nil, now, actorID, now, actorID, nil},
					{"0196c2a8-7b1e-7000-8000-0000000000b2", "Juan Gómez", "juan gomez", nil, nil, []byte(`[]`), nil, now, actorID, now, actorID, nil},
				},
			},
		)

		rec := invokeWithQuery(t, api.HandleListPersons, "/people")

		assertStatus(t, rec, http.StatusOK)
		var got []models.Person
		decode(t, rec, &got)
		if len(got) != 2 || got[0].FullName != "Ana Pérez" || got[1].FullName != "Juan Gómez" {
			t.Fatalf("persons = %+v, want Ana Pérez then Juan Gómez", got)
		}
		if got[0].CreatedAt != "2026-10-22T10:00:00Z" {
			t.Errorf("created_at = %q, want 2026-10-22T10:00:00Z", got[0].CreatedAt)
		}
	})

	t.Run("applies rows and page as limit and offset", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person", Args: []driver.Value{int64(10), int64(20)}, Columns: personColumns},
		)

		rec := invokeWithQuery(t, api.HandleListPersons, "/people?rows=10&page=3")

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns empty array when there are no persons", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person", Columns: personColumns},
		)

		rec := invokeWithQuery(t, api.HandleListPersons, "/people")

		assertBody(t, rec, http.StatusOK, `[]`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person", Err: errors.New("connection refused")},
		)

		rec := invokeWithQuery(t, api.HandleListPersons, "/people")

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})

	t.Run("rejects invalid rows", func(t *testing.T) {
		for _, target := range []string{"/people?rows=0", "/people?rows=-1", "/people?rows=abc", "/people?rows=1001"} {
			t.Run(target, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeWithQuery(t, api.HandleListPersons, target)

				assertBody(t, rec, http.StatusBadRequest, `{"error":"rows must be an integer between 1 and 1000"}`)
			})
		}
	})

	t.Run("rejects invalid page", func(t *testing.T) {
		for _, target := range []string{"/people?page=0", "/people?page=abc", "/people?page=101"} {
			t.Run(target, func(t *testing.T) {
				api, _ := newAPIConfig(t)

				rec := invokeWithQuery(t, api.HandleListPersons, target)

				assertBody(t, rec, http.StatusBadRequest, `{"error":"page must be an integer between 1 and 100"}`)
			})
		}
	})
}

func TestHandleGetPerson(t *testing.T) {
	const personID = "0196c2a8-7b1e-7000-8000-0000000000b1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	birthDate := time.Date(1960, 4, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()

	t.Run("returns the person", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{
				Query:   "FROM person WHERE id = $1 AND merged_into_id IS NULL",
				Args:    []driver.Value{personID},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Ana Pérez", "ana perez", birthDate, nil, []byte(`[]`),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		rec := invokeWithID(t, api.HandleGetPerson, personID, "")

		assertStatus(t, rec, http.StatusOK)
		var got models.Person
		decode(t, rec, &got)
		if got.ID != uuid.MustParse(personID) || got.FullName != "Ana Pérez" {
			t.Errorf("person = %+v", got)
		}
		if got.BirthDate != "1960-04-19" {
			t.Errorf("birth_date = %q, want 1960-04-19", got.BirthDate)
		}
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleGetPerson, "not-a-uuid", "")

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid person id"}`)
	})

	t.Run("returns 404 when the person does not exist or was merged", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person WHERE id = $1", Columns: personColumns},
		)

		rec := invokeWithID(t, api.HandleGetPerson, personID, "")

		assertBody(t, rec, http.StatusNotFound, `{"error":"person not found"}`)
	})

	t.Run("returns 500 when query fails", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			fakedb.Step{Query: "FROM person WHERE id = $1", Err: errors.New("connection refused")},
		)

		rec := invokeWithID(t, api.HandleGetPerson, personID, "")

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
	})
}

func TestHandleUpdatePerson(t *testing.T) {
	const personID = "0196c2a8-7b1e-7000-8000-0000000000b1"
	const updateQuery = "UPDATE person"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	birthDate := time.Date(1960, 4, 19, 0, 0, 0, 0, time.UTC)
	actorID := testActor.ID.String()
	socialMedia := `[{"platform":"twitter","url":"https://x.com/petrogustavo"}]`

	t.Run("replaces person with trimmed full name", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   "WHERE id = $1 AND merged_into_id IS NULL RETURNING",
				Args:    []driver.Value{personID, "Gustavo Petro", birthDate, "https://example.com/photo.jpg", []byte(socialMedia), int64(42)},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Gustavo Petro", "gustavo petro",
					birthDate, "https://example.com/photo.jpg", []byte(socialMedia),
					nil, now, actorID, now, actorID, int64(42),
				}},
			},
		)

		body := `{"full_name":"  Gustavo Petro  ","birth_date":"1960-04-19","photo_url":"https://example.com/photo.jpg",` +
			`"social_media":` + socialMedia + `,"id_congreso_visible":42}`
		rec := invokeWithID(t, api.HandleUpdatePerson, personID, body)

		assertStatus(t, rec, http.StatusOK)
		var got models.Person
		decode(t, rec, &got)
		if got.FullName != "Gustavo Petro" || got.BirthDate != "1960-04-19" {
			t.Errorf("person = %+v", got)
		}
		if !strings.Contains(rec.Body.String(), `"id_congreso_visible":42`) {
			t.Errorf("body %s does not contain id_congreso_visible 42", rec.Body)
		}
		if !fake.Committed {
			t.Error("transaction was not committed")
		}
	})

	t.Run("clears optional fields and returns null id_congreso_visible", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   updateQuery,
				Args:    []driver.Value{personID, "Ana Pérez", nil, nil, []byte(`[]`), nil},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Ana Pérez", "ana perez", nil, nil, []byte(`[]`),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"Ana Pérez"}`)

		assertStatus(t, rec, http.StatusOK)
		if !strings.Contains(rec.Body.String(), `"id_congreso_visible":null`) {
			t.Errorf("body %s does not contain null id_congreso_visible", rec.Body)
		}
	})

	t.Run("defaults null social media to an empty array", func(t *testing.T) {
		api, _ := newAPIConfig(t,
			setActor(),
			fakedb.Step{
				Query:   updateQuery,
				Args:    []driver.Value{personID, "Ana Pérez", nil, nil, []byte(`[]`), nil},
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Ana Pérez", "ana perez", nil, nil, []byte(`[]`),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"Ana Pérez","social_media":null}`)

		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdatePerson, "not-a-uuid", `{"full_name":"Ana Pérez"}`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid person id"}`)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"invalid JSON body"}`)
	})

	t.Run("rejects blank full name", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"full_name":"   "}`, `{"full_name":"   ","birth_date":"bad"}`} {
			api, _ := newAPIConfig(t)

			rec := invokeWithID(t, api.HandleUpdatePerson, personID, body)

			assertBody(t, rec, http.StatusBadRequest, `{"error":"mandatory full_name"}`)
		}
	})

	t.Run("rejects malformed birth date", func(t *testing.T) {
		api, _ := newAPIConfig(t)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"X","birth_date":"1960-04-19T00:00:00Z"}`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"birth_date must be formatted as YYYY-MM-DD"}`)
	})

	t.Run("rejects non-positive id_congreso_visible", func(t *testing.T) {
		for _, body := range []string{`{"full_name":"X","id_congreso_visible":0}`, `{"full_name":"X","id_congreso_visible":-5}`} {
			api, _ := newAPIConfig(t)

			rec := invokeWithID(t, api.HandleUpdatePerson, personID, body)

			assertBody(t, rec, http.StatusBadRequest, `{"error":"id_congreso_visible must be greater than zero"}`)
		}
	})

	t.Run("reports check constraint violation as bad input", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: updateQuery, Err: &pq.Error{Code: "23514", Message: "valid_social_media"}},
		)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"X","social_media":{"a":1}}`)

		assertBody(t, rec, http.StatusBadRequest, `{"error":"social_media is invalid"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 404 when the person does not exist or was merged", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: "WHERE id = $1 AND merged_into_id IS NULL", Columns: personColumns},
		)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"X"}`)

		assertBody(t, rec, http.StatusNotFound, `{"error":"person not found"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		api, fake := newAPIConfig(t,
			setActor(),
			fakedb.Step{Query: updateQuery, Err: errors.New("connection refused")},
		)

		rec := invokeWithID(t, api.HandleUpdatePerson, personID, `{"full_name":"X"}`)

		assertBody(t, rec, http.StatusInternalServerError, `{"error":"internal server error"}`)
		if !fake.RolledBack {
			t.Error("transaction was not rolled back")
		}
	})
}

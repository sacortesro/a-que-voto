package routers_test

import (
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/routers"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/store"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var sourceColumns = []string{
	"id", "code", "base_url", "priority", "created_at", "created_by", "updated_at", "updated_by",
}

var (
	member = auth.Actor{ID: uuid.MustParse("0196c2a8-7b1e-7000-8000-000000000002"), Kind: "user", Name: "reader", Role: auth.RoleMember}
	admin  = auth.Actor{ID: uuid.MustParse("0196c2a8-7b1e-7000-8000-000000000003"), Kind: "user", Name: "admin", Role: auth.RoleAdmin}
)

func serve(t *testing.T, actor auth.Actor, method, path, body string, steps ...fakedb.Step) *httptest.ResponseRecorder {
	t.Helper()

	db, _ := fakedb.Open(t, steps...)
	router := routers.GetRouter(controllers.NewAPIConfig(store.New(db)), auth.Fixed(actor))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestSourceRoutesAuthorization(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	body := `{"code":"camara","base_url":"https://www.camara.gov.co","priority":1}`

	t.Run("member cannot create a source", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/source", body)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("member can list sources", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/source", "",
			fakedb.Step{Query: "FROM source", Columns: sourceColumns},
		)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body)
		}
	})

	t.Run("admin can create a source", func(t *testing.T) {
		actorID := admin.ID.String()
		rec := serve(t, admin, http.MethodPost, "/v1/source", body,
			fakedb.Step{Query: "FROM source WHERE code", Columns: sourceColumns},
			fakedb.Step{Query: "set_config", Args: []driver.Value{actorID}},
			fakedb.Step{
				Query:   "INSERT INTO source",
				Columns: sourceColumns,
				Rows:    [][]driver.Value{{int64(1), "camara", "https://www.camara.gov.co", int64(1), now, actorID, now, actorID}},
			},
		)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body)
		}
	})
}

func TestPartyRoutes(t *testing.T) {
	t.Run("member can list parties", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/party", "",
			fakedb.Step{Query: "FROM party WHERE is_active", Columns: []string{"id", "name", "acronym", "created_at", "created_by", "updated_at", "updated_by", "is_active"}},
		)

		if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
			t.Errorf("status = %d, body = %s, want 200 []", rec.Code, rec.Body)
		}
	})
}

func TestPartyRoutesAuthorization(t *testing.T) {
	const partyID = "0196c2a8-7b1e-7000-8000-0000000000a1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	body := `{"name":"Partido Verde","is_active":true}`

	t.Run("member cannot update a party", func(t *testing.T) {
		rec := serve(t, member, http.MethodPut, "/v1/party/"+partyID, body)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("member cannot create a party", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/party", body)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("admin can update a party", func(t *testing.T) {
		actorID := admin.ID.String()
		rec := serve(t, admin, http.MethodPut, "/v1/party/"+partyID, body,
			fakedb.Step{Query: "set_config", Args: []driver.Value{actorID}},
			fakedb.Step{
				Query:   "UPDATE party",
				Columns: []string{"id", "name", "acronym", "created_at", "created_by", "updated_at", "updated_by", "is_active"},
				Rows:    [][]driver.Value{{partyID, "Partido Verde", nil, now, actorID, now, actorID, true}},
			},
		)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body)
		}
	})
}

func TestDistrictRoutesAuthorization(t *testing.T) {
	t.Run("member can list districts", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/district", "",
			fakedb.Step{Query: "FROM district", Columns: []string{"id", "chamber", "kind", "department", "name", "created_at", "created_by", "updated_at", "updated_by"}},
		)

		if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
			t.Errorf("status = %d, body = %s, want 200 []", rec.Code, rec.Body)
		}
	})

	t.Run("member cannot create a district", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/district", `{"chamber":"senado","kind":"nacional","name":"Nacional"}`)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func TestPersonRoutesAuthorization(t *testing.T) {
	t.Run("member can list persons", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/people", "",
			fakedb.Step{Query: "FROM person", Columns: []string{"id", "full_name", "name_norm", "birth_date", "photo_url", "social_media", "merged_into_id", "created_at", "created_by", "updated_at", "updated_by", "id_congreso_visible"}},
		)

		if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
			t.Errorf("status = %d, body = %s, want 200 []", rec.Code, rec.Body)
		}
	})

	t.Run("member can get a person", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/people/0196c2a8-7b1e-7000-8000-0000000000b1", "",
			fakedb.Step{Query: "FROM person WHERE id = $1", Columns: []string{"id", "full_name", "name_norm", "birth_date", "photo_url", "social_media", "merged_into_id", "created_at", "created_by", "updated_at", "updated_by", "id_congreso_visible"}},
		)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body)
		}
	})

	t.Run("member cannot create a person", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/people", `{"full_name":"Ana Pérez"}`)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func TestPeopleRoutesAuthorization(t *testing.T) {
	const personID = "0196c2a8-7b1e-7000-8000-0000000000b1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	body := `{"full_name":"Ana Pérez"}`
	personColumns := []string{
		"id", "full_name", "name_norm", "birth_date", "photo_url", "social_media",
		"merged_into_id", "created_at", "created_by", "updated_at", "updated_by", "id_congreso_visible",
	}

	t.Run("member cannot update a person", func(t *testing.T) {
		rec := serve(t, member, http.MethodPut, "/v1/people/"+personID, body)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("admin can update a person", func(t *testing.T) {
		actorID := admin.ID.String()
		rec := serve(t, admin, http.MethodPut, "/v1/people/"+personID, body,
			fakedb.Step{Query: "set_config", Args: []driver.Value{actorID}},
			fakedb.Step{
				Query:   "UPDATE person",
				Columns: personColumns,
				Rows: [][]driver.Value{{
					personID, "Ana Pérez", "ana perez", nil, nil, []byte(`[]`),
					nil, now, actorID, now, actorID, nil,
				}},
			},
		)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body)
		}
	})
}

func TestMandateRoutesAuthorization(t *testing.T) {
	const personID = "0196c2a8-7b1e-7000-8000-0000000000b1"
	mandateColumns := []string{
		"id", "person_id", "term_id", "chamber", "district_id", "party_id", "start_date", "end_date",
		"replaces_mandate_id", "status", "created_at", "created_by", "updated_at", "updated_by",
	}

	t.Run("member can list a person's mandates", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/people/"+personID+"/mandates", "",
			fakedb.Step{Query: "FROM mandate WHERE person_id = $1", Columns: mandateColumns},
		)

		if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
			t.Errorf("status = %d, body = %s, want 200 []", rec.Code, rec.Body)
		}
	})

	t.Run("member cannot create a mandate", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/mandate", `{}`)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("admin reaches mandate creation", func(t *testing.T) {
		rec := serve(t, admin, http.MethodPost, "/v1/mandate", `{"chamber":"congreso"}`)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body)
		}
	})
}

func TestCommitteeRoutesAuthorization(t *testing.T) {
	t.Run("member can list committees", func(t *testing.T) {
		rec := serve(t, member, http.MethodGet, "/v1/committee", "",
			fakedb.Step{Query: "FROM committee", Columns: []string{"id", "chamber", "kind", "name", "number", "created_at", "created_by", "updated_at", "updated_by"}},
		)

		if rec.Code != http.StatusOK || rec.Body.String() != "[]" {
			t.Errorf("status = %d, body = %s, want 200 []", rec.Code, rec.Body)
		}
	})

	t.Run("member cannot create a committee", func(t *testing.T) {
		rec := serve(t, member, http.MethodPost, "/v1/committee", `{}`)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func TestCommitteeUpdateRoutesAuthorization(t *testing.T) {
	const committeeID = "0196c2a8-7b1e-7000-8000-0000000000f1"
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	body := `{"kind":"legal","number":3}`

	t.Run("member cannot update a committee", func(t *testing.T) {
		rec := serve(t, member, http.MethodPut, "/v1/committee/"+committeeID, body)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("admin can update a committee", func(t *testing.T) {
		actorID := admin.ID.String()
		rec := serve(t, admin, http.MethodPut, "/v1/committee/"+committeeID, body,
			fakedb.Step{Query: "set_config", Args: []driver.Value{actorID}},
			fakedb.Step{
				Query:   "UPDATE committee",
				Columns: []string{"id", "chamber", "kind", "name", "number", "created_at", "created_by", "updated_at", "updated_by"},
				Rows:    [][]driver.Value{{committeeID, "senado", "legal", "Comisión Primera", int64(3), now, actorID, now, actorID}},
			},
		)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body)
		}
	})
}

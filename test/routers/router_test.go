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

	t.Run("admin can create a source", func(t *testing.T) {
		actorID := admin.ID.String()
		rec := serve(t, admin, http.MethodPost, "/v1/source", body,
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

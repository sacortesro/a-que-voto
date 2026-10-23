package controllers_test

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/store"
	"github.com/sacortesro/a-que-voto/test/fakedb"
)

var testActor = auth.Actor{
	ID:   uuid.MustParse("0196c2a8-7b1e-7000-8000-000000000001"),
	Kind: "user",
	Name: "tester",
}

// newAPIConfig wires the real store on top of a scripted database.
func newAPIConfig(t *testing.T, steps ...fakedb.Step) (*controllers.APIConfig, *fakedb.Fake) {
	t.Helper()

	db, fake := fakedb.Open(t, steps...)
	return controllers.NewAPIConfig(store.New(db)), fake
}

// invoke posts body to handler. withActor puts testActor in the request context.
func invoke(t *testing.T, handler http.HandlerFunc, body string, withActor bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if withActor {
		req = req.WithContext(auth.WithActor(req.Context(), testActor))
	}

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// setActor is the statement store.Write runs before every write.
func setActor() fakedb.Step {
	return fakedb.Step{
		Query: "set_config",
		Args:  []driver.Value{testActor.ID.String()},
	}
}

// assertBody checks status and the exact body. Error bodies are compared exactly
// so that a second JSON document written after the first one is caught.
func assertBody(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d", rec.Code, wantStatus)
	}
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %s, want %s", got, wantBody)
	}
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, wantStatus, rec.Body)
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, rec.Body)
	}
}

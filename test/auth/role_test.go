package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/auth"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		actor      *auth.Actor
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "admin passes",
			actor:      &auth.Actor{ID: uuid.New(), Kind: "user", Name: "admin", Role: auth.RoleAdmin},
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
		{
			name:       "member is forbidden",
			actor:      &auth.Actor{ID: uuid.New(), Kind: "user", Name: "reader", Role: auth.RoleMember},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "actor without role is forbidden",
			actor:      &auth.Actor{ID: uuid.New(), Kind: "system", Name: "legacy"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing actor is unauthorized",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.actor != nil {
				req = req.WithContext(auth.WithActor(req.Context(), *tt.actor))
			}
			rec := httptest.NewRecorder()

			auth.RequireRole(auth.RoleAdmin)(next).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if called != tt.wantCalled {
				t.Errorf("next called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

func TestRequireRoleAfterMiddleware(t *testing.T) {
	member := auth.Actor{ID: uuid.New(), Kind: "user", Name: "reader", Role: auth.RoleMember}

	r := http.NewServeMux()
	r.Handle("/", auth.Middleware(auth.Fixed(member))(
		auth.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler ran for a member")
		})),
	))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

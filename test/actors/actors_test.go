package actors_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/store"

	_ "github.com/lib/pq"
)

// These tests need a migrated database. Set AQV_TEST_DB_URL to run them;
// otherwise they are skipped.
func testStore(t *testing.T) *store.Store {
	t.Helper()
	dsn := os.Getenv("AQV_TEST_DB_URL")
	if dsn == "" {
		t.Skip("AQV_TEST_DB_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.New(db)
}

func createToken(t *testing.T, s *store.Store, expiresAt time.Time, revoked bool) (auth.Actor, string) {
	t.Helper()
	ctx := context.Background()

	owner, err := actors.Ensure(ctx, s.Queries(), actors.KindSystem, "test.token."+uuid.NewString()[:8])
	if err != nil {
		t.Fatalf("ensure actor: %v", err)
	}

	raw := uuid.NewString()
	sum := sha256.Sum256([]byte(raw))

	err = s.Write(auth.WithActor(ctx, owner), func(q *database.Queries) error {
		tok, err := q.CreateToken(ctx, database.CreateTokenParams{
			ActorID:   owner.ID,
			TokenHash: sum[:],
			ExpiresAt: expiresAt,
		})
		if err != nil || !revoked {
			return err
		}
		return q.RevokeToken(ctx, tok.ID)
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	return owner, raw
}

func resolve(s *store.Store, raw string) (auth.Actor, error) {
	r := httptest.NewRequest("GET", "/", nil)
	if raw != "" {
		r.Header.Set("Authorization", "Bearer "+raw)
	}
	return actors.Token(s.Queries())(r)
}

func TestTokenResolvesActor(t *testing.T) {
	s := testStore(t)
	owner, raw := createToken(t, s, time.Now().Add(time.Hour), false)

	got, err := resolve(s, raw)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.ID != owner.ID {
		t.Fatalf("got actor %s, want %s", got.ID, owner.ID)
	}
}

func TestTokenRejectsInvalid(t *testing.T) {
	s := testStore(t)
	_, expired := createToken(t, s, time.Now().Add(-time.Hour), false)
	_, revoked := createToken(t, s, time.Now().Add(time.Hour), true)

	tests := map[string]string{
		"missing": "",
		"unknown": uuid.NewString(),
		"expired": expired,
		"revoked": revoked,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := resolve(s, raw)
			if !errors.Is(err, auth.ErrUnauthorized) {
				t.Fatalf("got %v, want ErrUnauthorized", err)
			}
		})
	}
}

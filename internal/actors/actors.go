// Package actors turns a name into an auth.Actor by reading or registering
// its row in the `actor` table. Only entry points (HTTP, jobs, maintenance)
// should call it; writers never pass an actor by hand.
package actors

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
)

// Kinds mirror the CHECK constraint on actor.kind.
const (
	KindUser   = "user"
	KindJob    = "job"
	KindSystem = "system"
)

// Ensure registers the actor if it does not exist yet and returns it.
// The actor table has no audit columns, so this needs no app.actor_id.
func Ensure(ctx context.Context, q *database.Queries, kind, name string) (auth.Actor, error) {
	row, err := q.UpsertActor(ctx, database.UpsertActorParams{Kind: kind, Name: name})
	if err != nil {
		return auth.Actor{}, fmt.Errorf("ensure actor %q: %w", name, err)
	}
	return toAuth(row), nil
}

func toAuth(row database.Actor) auth.Actor {
	return auth.Actor{ID: row.ID, Kind: row.Kind, Name: row.Name}
}

func Token(q *database.Queries) auth.Resolver {
	return func(r *http.Request) (auth.Actor, error) {
		key, err := auth.GetAPIKey(r.Header)
		if err != nil {
			return auth.Actor{}, auth.ErrUnauthorized
		}
		sum := sha256.Sum256([]byte(key))
		row, err := q.GetActorByTokenHash(r.Context(), sum[:])
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Actor{}, auth.ErrUnauthorized
		}
		if err != nil {
			return auth.Actor{}, fmt.Errorf("get actor by token: %w", err)
		}
		return toAuth(row), nil
	}
}

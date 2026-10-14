// Package actors turns a name into an auth.Actor by reading or registering
// its row in the `actor` table. Only entry points (HTTP, jobs, maintenance)
// should call it; writers never pass an actor by hand.
package actors

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
)

// Kinds mirror the CHECK constraint on actor.kind.
const (
	KindUser   = "user"
	KindJob    = "job"
	KindSystem = "system"
)

// Get loads an existing actor by ID. It fails if the row does not exist,
// so a bad DEV_ACTOR_ID is caught at startup instead of on the first write.
func Get(ctx context.Context, q *database.Queries, id uuid.UUID) (auth.Actor, error) {
	row, err := q.GetActorByID(ctx, id)
	if err != nil {
		return auth.Actor{}, fmt.Errorf("get actor %s: %w", id, err)
	}
	return toAuth(row), nil
}

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

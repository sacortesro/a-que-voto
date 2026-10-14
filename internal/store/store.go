package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
)

var ErrNoActor = errors.New("no actor in context")

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Queries() *database.Queries {
	return database.New(s.db)
}

func (s *Store) Write(ctx context.Context, fn func(q *database.Queries) error) error {
	actor, ok := auth.ActorFromContext(ctx)
	if !ok {
		return ErrNoActor
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT set_config('app.actor_id', $1, true)", actor.ID); err != nil {
		return fmt.Errorf("set actor: %w", err)
	}

	if err := fn(database.New(tx)); err != nil {
		return err
	}

	return tx.Commit()
}

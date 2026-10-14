package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type Actor struct {
	ID   uuid.UUID
	Kind string
	Name string
}

var ErrUnauthorized = errors.New("unauthorized")

type actorCtxKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorCtxKey{}).(Actor)
	return actor, ok
}

type Resolver func(r *http.Request) (Actor, error)

func Fixed(actor Actor) Resolver {
	return func(*http.Request) (Actor, error) {
		return actor, nil
	}
}

func Middleware(resolve Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, err := resolve(r)
			if err != nil {
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), actor)))
		})
	}
}

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/routers"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/store"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	port, err := requireEnv("PORT")
	if err != nil {
		return err
	}

	dbURL, err := requireEnv("DB_URL")
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}

	actorRaw, err := requireEnv("DEV_ACTOR_ID")
	if err != nil {
		return err
	}

	actorID, err := uuid.Parse(actorRaw)
	if err != nil {
		return fmt.Errorf("DEV_ACTOR_ID is not a valid uuid: %w", err)
	}

	ctx := context.Background()
	devActor, err := actors.Get(ctx, database.New(conn), actorID)
	if err != nil {
		return fmt.Errorf("DEV_ACTOR_ID: %w", err)
	}

	cfg := controllers.NewAPIConfig(store.New(conn))

	srv := &http.Server{
		Handler: routers.GetRouter(cfg, auth.Fixed(devActor)),
		Addr:    ":" + port,
	}

	log.Printf("Server starting on port %v", port)
	return srv.ListenAndServe()
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New(key + " environment variable not set")
	}
	return value, nil
}

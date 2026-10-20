package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/routers"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/config"
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
	port, err := config.RequireEnv("PORT")
	if err != nil {
		return err
	}

	dbURL, err := config.RequireEnv("DB_URL")
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}

	actorRaw, err := config.RequireEnv("DEV_ACTOR_ID")
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
		Handler:           routers.GetRouter(cfg, auth.Fixed(devActor)),
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on port %v", port)
	return srv.ListenAndServe()
}

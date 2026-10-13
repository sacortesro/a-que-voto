package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/routers"
	"github.com/sacortesro/a-que-voto/internal/database"

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

	// Built but not wired into the router yet.
	_ = controllers.NewAPIConfig(database.New(conn))

	srv := &http.Server{
		Handler: routers.GetRouter(),
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

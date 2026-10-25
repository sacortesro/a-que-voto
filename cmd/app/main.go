package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/cmd/app/routers"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/config"
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
	defer conn.Close()

	s := store.New(conn)
	cfg := controllers.NewAPIConfig(s)

	srv := &http.Server{
		Handler:           routers.GetRouter(cfg, actors.Token(s.Queries())),
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on port %v", port)
	return srv.ListenAndServe()
}

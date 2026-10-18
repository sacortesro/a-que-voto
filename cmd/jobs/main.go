// Command jobs runs one scraper job and exits. It is started by the scheduler
// (GitHub Actions today), so it must exit non-zero when the job fails.
//
//	jobs run <name>
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/internal/jobs"
	"github.com/sacortesro/a-que-voto/internal/store"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) != 2 || args[0] != "run" {
		return errors.New("usage: jobs run <name>")
	}
	name := args[1]

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return errors.New("DB_URL environment variable not set")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}
	defer conn.Close()

	s := store.New(conn)

	// Register every job here. Nothing is registered yet.
	registry, err := jobs.NewRegistry(
		scrapePersons{store: s, corporacion: "1"}, // scrape-representates
		scrapePersons{store: s, corporacion: "2"}, // scrape-senadores
	)
	if err != nil {
		return err
	}

	runner := jobs.NewRunner(s, registry)
	return runner.Run(context.Background(), name)
}

// auth -name <actor> [-kind system|user] [-days 90]
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/config"
	"github.com/sacortesro/a-que-voto/internal/database"
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
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	name := fs.String("name", "", "actor that will own the token")
	kind := fs.String("kind", actors.KindUser, "actor kind: system or user")
	days := fs.Int("days", 90, "days until the token expires")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return errors.New("usage: auth -name <actor> [-kind system|user] [-days 90]")
	}
	if *kind != actors.KindSystem && *kind != actors.KindUser {
		return fmt.Errorf("invalid kind %q", *kind)
	}
	if *days <= 0 {
		return errors.New("days must be greater than 0")
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
	ctx := context.Background()

	cli, err := actors.Ensure(ctx, s.Queries(), actors.KindSystem, "system:auth-cli")
	if err != nil {
		return err
	}
	ctx = auth.WithActor(ctx, cli)

	owner, err := actors.Ensure(ctx, s.Queries(), *kind, *name)
	if err != nil {
		return err
	}

	raw, hash, err := auth.GenerateToken(32)
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	var token database.ApiToken
	err = s.Write(ctx, func(q *database.Queries) error {
		var err error
		token, err = q.CreateToken(ctx, database.CreateTokenParams{
			ActorID:   owner.ID,
			TokenHash: hash,
			ExpiresAt: time.Now().AddDate(0, 0, *days),
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}

	fmt.Printf("actor:   %s (%s)\n", owner.Name, owner.ID)
	fmt.Printf("expires: %s\n", utils.FormatTimestamp(token.ExpiresAt))
	fmt.Printf("token:   %s\n", raw)
	return nil
}

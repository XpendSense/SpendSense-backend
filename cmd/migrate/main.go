package main

import (
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	databaseURL = strings.ReplaceAll(databaseURL, "postgresql://", "postgres://")

	m, err := migrate.New("file://internal/db/migrations", databaseURL)
	if err != nil {
		log.Fatalf("migration init: %v", err)
	}
	defer m.Close()

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration up: %v", err)
		}
	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration down: %v", err)
		}
	default:
		log.Fatalf("unknown direction %q — use 'up' or 'down'", direction)
	}

	log.Println("migrations: up to date")
}

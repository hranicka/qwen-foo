package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/hranicka/qwen-foo/config"
	"github.com/hranicka/qwen-foo/internal/migrations"
	foo "github.com/hranicka/qwen-foo/pkg"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: migrate [command]\nCommands: up, down, version, status\n")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	sourceDriver, err := migrations.NewDriver(foo.MigrationsFS)
	if err != nil {
		log.Fatalf("init migrations: %v", err)
	}

	dbURL := cfg.Database.URL + "&search_path=public"

	m, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer m.Close()

	pgxDriver, err := pgxv5.WithInstance(m, &pgxv5.Config{})
	if err != nil {
		log.Fatalf("init pgx driver: %v", err)
	}

	mgr, err := migrate.NewWithInstance("migrations", sourceDriver, "postgres", pgxDriver)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}

	switch os.Args[1] {
	case "up":
		if err := mgr.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrations applied successfully")
	case "down":
		if err := mgr.Down(); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("migration rolled back successfully")
	case "version":
		v, _, err := mgr.Version()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("current version: %d\n", v)
	case "status":
		v, d, err := mgr.Version()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("status: %v", err)
		}
		fmt.Printf("version: %d, dirty: %v\n", v, d)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/url"
	"testing"

	"github.com/hranicka/qwen-foo/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func dsn(dbName string, cfg *config.Database) string {
	if cfg.URL != "" {
		u, err := url.Parse(cfg.URL)
		if err != nil {
			return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, dbName)
		}
		u.Path = "/" + dbName
		return u.String()
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, dbName)
}

func newTestDB(t *testing.T) (pool *pgxpool.Pool, cleanup func()) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	suffix := fmt.Sprintf("%06x", rand.Uint32()%1000000)
	dbName := "test_" + suffix

	adminDSN := dsn("template1", &cfg.Database)
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	defer admin.Close()

	if _, err := admin.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'test_template' AND pid <> pg_backend_pid()"); err != nil {
		t.Fatalf("terminate connections: %v", err)
	}

	if _, err := admin.Exec(fmt.Sprintf(
		"CREATE DATABASE %s TEMPLATE test_template", dbName)); err != nil {
		t.Fatalf("create db: %v", err)
	}

	ctx := context.Background()
	testDSN := dsn(dbName, &cfg.Database)
	pool, err = pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	cleanup = func() {
		pool.Close()
		conn, err := sql.Open("pgx", adminDSN)
		if err == nil {
			defer conn.Close()
			conn.Exec(fmt.Sprintf(
				"DROP DATABASE %s WITH (FORCE)", dbName))
		}
	}

	return pool, cleanup
}

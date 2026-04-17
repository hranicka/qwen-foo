package server_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/hranicka/qwen-foo/config"
	"github.com/hranicka/qwen-foo/internal/db"
	"github.com/hranicka/qwen-foo/internal/migrations"
	"github.com/hranicka/qwen-foo/internal/server"
	foo "github.com/hranicka/qwen-foo/pkg"
)

var testSourceDriver, _ = migrations.NewDriver(foo.MigrationsFS)

func TestMain(m *testing.M) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	adminDSN := dsn("template1", &cfg.Database)
	adminDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer adminDB.Close()

	// Create template database
	adminDB.Exec("DROP DATABASE IF EXISTS test_template")
	adminDB.Exec("CREATE DATABASE test_template")

	// Apply migrations to template
	migDSN := dsn("test_template", &cfg.Database)
	migDB, err := sql.Open("pgx", migDSN)
	if err != nil {
		log.Fatalf("connect template: %v", err)
	}
	defer migDB.Close()

	pgxDriver, err := pgxv5.WithInstance(migDB, &pgxv5.Config{})
	if err != nil {
		log.Fatalf("init driver: %v", err)
	}

	mgr, err := migrate.NewWithInstance("migrations", testSourceDriver, "postgres", pgxDriver)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}

	if err := mgr.Up(); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestHello(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		status      int
		expectMsg   string
		expectCount int64
	}{
		{
			name:        "hello returns 200 with message and counter",
			path:        "/hello",
			status:      200,
			expectMsg:   "Hello",
			expectCount: 1,
		},
		{
			name:      "unknown path returns 404",
			path:      "/unknown",
			status:    404,
			expectMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			pool, teardown := newTestDB(t)
			defer teardown()

			schemaDB := db.NewFromPool(pool)
			srv := server.New(schemaDB)

			if err := schemaDB.Reset(ctx); err != nil {
				t.Fatalf("reset: %v", err)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.path, nil)
			if err != nil {
				t.Fatalf("request: %v", err)
			}

			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != tt.status {
				t.Errorf("status: got %d, want %d", rr.Code, tt.status)
			}

			if tt.status == 200 {
				var got map[string]any
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode: %v", err)
				}

				if got["message"] != tt.expectMsg {
					t.Errorf("message: got %q, want %q", got["message"], tt.expectMsg)
				}

				counterVal, ok := got["counter"].(float64)
				if !ok {
					t.Fatalf("counter not a number")
				}

				if int64(counterVal) != tt.expectCount {
					t.Errorf("counter: got %v, want %d", counterVal, tt.expectCount)
				}
			}
		})
	}
}

func TestHelloCounterIncrements(t *testing.T) {
	ctx := context.Background()

	pool, teardown := newTestDB(t)
	defer teardown()

	schemaDB := db.NewFromPool(pool)
	srv := server.New(schemaDB)

	if err := schemaDB.Reset(ctx); err != nil {
		t.Fatalf("reset: %v", err)
	}

	expected := []int64{1, 2, 3}
	for i, want := range expected {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/hello", nil)
		if err != nil {
			t.Fatalf("request: %v", err)
		}

		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("call %d: status: got %d, want %d", i+1, rr.Code, http.StatusOK)
		}

		var got map[string]any
		if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
			t.Fatalf("call %d: decode: %v", i+1, err)
		}

		counterVal, ok := got["counter"].(float64)
		if !ok {
			t.Fatalf("call %d: counter not a number", i+1)
		}

		if int64(counterVal) != want {
			t.Errorf("call %d: counter: got %v, want %d", i+1, counterVal, want)
		}
	}
}

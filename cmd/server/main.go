package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hranicka/qwen-foo/config"
	"github.com/hranicka/qwen-foo/internal/db"
	"github.com/hranicka/qwen-foo/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	pool, err := db.New(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Init(ctx); err != nil {
		slog.Error("failed to init counter", "err", err)
		os.Exit(1)
	}

	srv := server.New(pool)

	httpSrv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "addr", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped", "err", err)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
}

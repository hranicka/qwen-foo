package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	slog.Info("connected to postgres")
	return &Pool{pool: pool}, nil
}

func (p *Pool) Init(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO counter (id, value) VALUES (1, 0) ON CONFLICT (id) DO NOTHING`)
	return err
}

func NewFromPool(pool *pgxpool.Pool) *Pool {
	return &Pool{pool: pool}
}

func (p *Pool) Incr(ctx context.Context) (int64, error) {
	var val int64
	err := p.pool.QueryRow(ctx, `
		WITH upd AS (
			UPDATE counter SET value = value + 1 WHERE id = 1 RETURNING value
		)
		SELECT value FROM upd
	`).Scan(&val)
	return val, err
}

func (p *Pool) Reset(ctx context.Context) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM counter"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO counter (id, value) VALUES (1, 0)"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Pool) Close() {
	p.pool.Close()
}

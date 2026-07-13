package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a pgxpool connection pool with explicit bounded limits.
// Design Rule #2: Always bound connection pools to prevent database crash under load.
func Connect(ctx context.Context, databaseURL string, maxOpen, maxIdle int, maxLifetime time.Duration) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	cfg.MaxConns = int32(maxOpen)
	cfg.MinConns = int32(maxIdle)
	cfg.MaxConnLifetime = maxLifetime
	cfg.MaxConnIdleTime = time.Duration(maxIdle) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return pool, nil
}

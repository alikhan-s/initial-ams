package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database connection settings
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string // "disable" for local, "require" for prod
}

// New initializes a PostgreSQL connection pool with production settings
func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	// Parse the config string into a pgxpool.Config struct
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// --- Production Tuning ---
	// MaxConns: Max total connections in the pool.
	// If the app tries to get a connection and the pool is full, it will wait.
	poolConfig.MaxConns = 25

	// MinConns: Connections to keep open even when idle.
	poolConfig.MinConns = 5

	// MaxConnLifetime: Close connections after 1 hour to prevent stale connection issues.
	poolConfig.MaxConnLifetime = time.Hour

	// MaxConnIdleTime: Close idle connections after 30 minutes.
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection immediately
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return pool, nil
}

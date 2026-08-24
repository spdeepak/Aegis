package ratelimit

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/spdeepak/aegis/server/internal/config"
)

// PostgresStore keeps brute-force state in a shared Postgres table, so it is
// centralized across all pods. All operations are fail-open: a backend error
// allows the request rather than risk a self-inflicted denial of service.
type PostgresStore struct {
	queries Querier
	cfg     config.RateLimitConfig
}

func NewPostgresStore(pool *pgxpool.Pool, cfg config.RateLimitConfig) *PostgresStore {
	return &PostgresStore{queries: New(pool), cfg: cfg}
}

// durationToInterval converts a Go duration into the pgtype.Interval that the
// generated query expects for the window/lockout parameters.
func durationToInterval(d time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}
}

func (p *PostgresStore) IsLocked(ctx context.Context, key string) (bool, error) {
	locked, err := p.queries.IsLocked(ctx, key)
	if err != nil {
		slog.Warn("rate-limit store IsLocked error", "error", err)
		return false, nil // fail-open
	}
	return locked, nil
}

func (p *PostgresStore) RecordFailure(ctx context.Context, key string) (bool, error) {
	locked, err := p.queries.RecordAuthFailure(ctx, RecordAuthFailureParams{
		Key:         key,
		Window:      durationToInterval(p.cfg.Window),
		MaxAttempts: int32(p.cfg.MaxAttempts),
		Lockout:     durationToInterval(p.cfg.LockoutDuration),
	})
	if err != nil {
		slog.Warn("rate-limit store RecordFailure error", "error", err)
		return false, nil // fail-open
	}
	return locked, nil
}

func (p *PostgresStore) Reset(ctx context.Context, key string) error {
	if err := p.queries.ResetAuthFailure(ctx, key); err != nil {
		slog.Warn("rate-limit store Reset error", "error", err)
	}
	return nil // fail-open
}

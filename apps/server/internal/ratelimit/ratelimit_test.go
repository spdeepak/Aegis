package ratelimit

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spdeepak/aegis/server/internal/config"
)

func TestNormalizeIP_IPv4Unchanged(t *testing.T) {
	assert.Equal(t, "192.0.2.1", NormalizeIP("192.0.2.1"))
}

func TestNormalizeIP_IPv6CollapsedTo64(t *testing.T) {
	a := NormalizeIP("2001:db8::1")
	b := NormalizeIP("2001:db8::2")
	assert.Equal(t, "2001:db8::/64", a)
	assert.Equal(t, "2001:db8::/64", b)
	// different /64 prefixes must remain distinct
	c := NormalizeIP("2001:db9::1")
	assert.Equal(t, "2001:db9::/64", c)
	assert.NotEqual(t, a, c)
	// sanity: it really is a /64
	_, ipnet, err := net.ParseCIDR(a)
	require.NoError(t, err)
	ones, _ := ipnet.Mask.Size()
	assert.Equal(t, 64, ones)
}

func TestMemoryStore_LocksOutAndResets(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		MaxAttempts:     3,
		Window:          time.Minute,
		LockoutDuration: time.Minute,
	}
	s := NewMemoryStore(cfg)
	ctx := context.Background()
	key := "192.0.2.1"

	for i := 0; i < cfg.MaxAttempts-1; i++ {
		locked, err := s.RecordFailure(ctx, key)
		require.NoError(t, err)
		assert.False(t, locked, "should not be locked before threshold")
	}
	locked, err := s.RecordFailure(ctx, key)
	require.NoError(t, err)
	assert.True(t, locked)

	isLocked, err := s.IsLocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, isLocked)

	require.NoError(t, s.Reset(ctx, key))
	isLocked, err = s.IsLocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, isLocked)
}

func TestPostgresStore_Integration(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://admin:admin@localhost:5432/jwt_server?sslmode=disable")
	if err != nil || pool == nil {
		t.Skip("postgres not available; skipping integration test")
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skip("postgres not reachable; skipping integration test")
	}
	defer pool.Close()

	// Ensure the table exists (idempotent) so the test is self-contained.
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS auth_rate_limits
		(
			key          TEXT        NOT NULL PRIMARY KEY,
			failures     INT         NOT NULL DEFAULT 0,
			window_start TIMESTAMPTZ NOT NULL DEFAULT now(),
			locked_until TIMESTAMPTZ,
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		);`); err != nil {
		t.Fatalf("failed to ensure auth_rate_limits table: %v", err)
	}

	cfg := config.RateLimitConfig{
		Enabled:         true,
		MaxAttempts:     3,
		Window:          time.Minute,
		LockoutDuration: time.Minute,
	}
	s := NewPostgresStore(pool, cfg)
	key := "test:integration:" + time.Now().Format("20060102150405.000000000")

	// ensure clean state
	_ = s.Reset(ctx, key)

	for i := 0; i < cfg.MaxAttempts-1; i++ {
		locked, err := s.RecordFailure(ctx, key)
		require.NoError(t, err)
		assert.False(t, locked, "should not be locked before threshold")
	}
	locked, err := s.RecordFailure(ctx, key)
	require.NoError(t, err)
	assert.True(t, locked)

	isLocked, err := s.IsLocked(ctx, key)
	require.NoError(t, err)
	assert.True(t, isLocked)

	require.NoError(t, s.Reset(ctx, key))
	isLocked, err = s.IsLocked(ctx, key)
	require.NoError(t, err)
	assert.False(t, isLocked)
}

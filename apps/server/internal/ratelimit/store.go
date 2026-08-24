package ratelimit

import (
	"context"
	"log/slog"
	"net"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/spdeepak/aegis/server/internal/config"
)

// AuthLockoutsAccount counts account-level (IP+email) brute-force lockouts.
var AuthLockoutsAccount = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "go_jwt_server_auth_lockouts_account_total",
		Help: "Total number of account-level (client IP + email) brute-force lockouts",
	},
)

func init() {
	prometheus.MustRegister(AuthLockoutsAccount)
}

// Store is the pluggable backend for brute-force state. All methods are safe to
// call concurrently. Implementations must fail open: on any backend error they
// should allow the request rather than blocking it.
type Store interface {
	// IsLocked reports whether the given key is currently locked out.
	IsLocked(ctx context.Context, key string) (bool, error)
	// RecordFailure records an authentication failure for key and returns
	// whether the key is now locked out as a result.
	RecordFailure(ctx context.Context, key string) (locked bool, err error)
	// Reset clears any failure state for key (called after a success).
	Reset(ctx context.Context, key string) error
}

// NormalizeIP returns a stable, privacy/abuse-aware key for a client IP:
//   - IPv4 addresses are used as-is (/32)
//   - IPv6 addresses are collapsed to their /64 prefix, so SLAAC-style address
//     rotation (a single site effectively has a /64) cannot be used to evade
//     per-address rate limits.
func NormalizeIP(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ipStr
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask).String() + "/64"
}

// NewStore builds the configured Store. When disabled it returns a NoopStore.
// Backend selection: "redis" (if redis addr is configured, else falls back to
// postgres with a warning), "memory", or "postgres" (the default).
func NewStore(cfg config.RateLimitConfig, pool *pgxpool.Pool) (Store, error) {
	if !cfg.Enabled {
		return &NoopStore{}, nil
	}

	switch strings.ToLower(cfg.Backend) {
	case "redis":
		redisCfg := cfg
		if redisCfg.Redis.Addr == "" {
			// Default to the redis service name used by docker-compose and the
			// bundled k8s manifests, so operators only need to flip backend.
			redisCfg.Redis.Addr = "redis:6379"
			slog.Info("rate-limit backend 'redis' with no explicit addr; defaulting to redis:6379")
		}
		return NewRedisStore(redisCfg)
	case "memory":
		return NewMemoryStore(cfg), nil
	default:
		return NewPostgresStore(pool, cfg), nil
	}
}

// NoopStore is used when rate limiting is disabled.
type NoopStore struct{}

func (n *NoopStore) IsLocked(_ context.Context, _ string) (bool, error) { return false, nil }
func (n *NoopStore) RecordFailure(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (n *NoopStore) Reset(_ context.Context, _ string) error { return nil }

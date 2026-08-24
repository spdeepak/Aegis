package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/spdeepak/aegis/server/internal/config"
	"github.com/spdeepak/aegis/server/internal/error"
	"github.com/spdeepak/aegis/server/internal/ratelimit"
)

var authLockouts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "go_jwt_server_auth_lockouts_total",
		Help: "Total number of client IPs locked out due to repeated authentication failures",
	},
)

func init() {
	prometheus.MustRegister(authLockouts)
}

// RateLimiter enforces a per-client brute-force lockout based on the number of
// authentication failures (HTTP 401) observed for a normalized client IP. State
// lives in the configured ratelimit.Store (in-memory, Postgres, or Redis), so it
// can be centralized across pods. IPs are normalized (IPv6 collapsed to /64) and
// the effective IP is resolved only from trusted proxies, never from a
// client-supplied X-Forwarded-For.
type RateLimiter struct {
	cfg   config.RateLimitConfig
	store ratelimit.Store
}

// NewRateLimiter builds a RateLimiter from configuration and the shared store.
func NewRateLimiter(cfg config.RateLimitConfig, store ratelimit.Store) *RateLimiter {
	return &RateLimiter{cfg: cfg, store: store}
}

// Middleware returns a gin middleware that short-circuits requests from locked-out
// clients with 429 Too Many Requests, and otherwise counts authentication failures
// so that repeated failures trigger a temporary lockout.
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.cfg.Enabled {
			c.Next()
			return
		}

		key := ratelimit.NormalizeIP(c.ClientIP())
		ctx := c.Request.Context()

		// Fail-open: an error checking the lock must not block legitimate traffic.
		if locked, err := r.store.IsLocked(ctx, key); err == nil && locked {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, httperror.HttpError{
				Description: "Too many failed authentication attempts. Please try again later.",
				StatusCode:  http.StatusTooManyRequests,
			})
			return
		}

		c.Next()

		if c.Writer.Status() == http.StatusUnauthorized {
			if locked, _ := r.store.RecordFailure(ctx, key); locked {
				authLockouts.Inc()
			}
		} else {
			_ = r.store.Reset(ctx, key)
		}
	}
}

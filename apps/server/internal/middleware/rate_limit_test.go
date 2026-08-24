package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/spdeepak/aegis/server/internal/config"
	"github.com/spdeepak/aegis/server/internal/ratelimit"
)

func newTestLimiter(maxAttempts int, window, lockout time.Duration) *RateLimiter {
	cfg := config.RateLimitConfig{
		Enabled:         true,
		MaxAttempts:     maxAttempts,
		Window:          window,
		LockoutDuration: lockout,
	}
	return NewRateLimiter(cfg, ratelimit.NewMemoryStore(cfg))
}

func TestRateLimiter_LocksOutAfterMaxAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := newTestLimiter(3, time.Minute, time.Minute)

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/t", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_ResetsOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := newTestLimiter(2, time.Minute, time.Minute)

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ok", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/fail", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRateLimiter_DisabledIsNoOp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter(config.RateLimitConfig{Enabled: false}, ratelimit.NewMemoryStore(config.RateLimitConfig{}))

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/t", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

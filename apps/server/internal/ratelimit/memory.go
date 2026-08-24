package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/spdeepak/aegis/server/internal/config"
	pkgtime "github.com/spdeepak/aegis/server/pkg/time"
	"github.com/spdeepak/aegis/server/pkg/ttlcache"
)

type ipAuthState struct {
	failures    int64
	lockedUntil time.Time
}

// MemoryStore keeps brute-force state in an in-process cache. It is correct for
// a single instance (dev / single pod) but is NOT shared across pods; prefer
// PostgresStore or RedisStore in a multi-replica deployment.
type MemoryStore struct {
	cfg   config.RateLimitConfig
	cache *ttlcache.Cache
	mu    sync.Mutex
}

func NewMemoryStore(cfg config.RateLimitConfig) *MemoryStore {
	return &MemoryStore{
		cfg:   cfg,
		cache: ttlcache.New(nil),
	}
}

func (ms *MemoryStore) state(key string) *ipAuthState {
	if v, ok := ms.cache.Get(key); ok {
		if s, ok := v.(*ipAuthState); ok {
			return s
		}
	}
	return &ipAuthState{}
}

func (ms *MemoryStore) IsLocked(_ context.Context, key string) (bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	s := ms.state(key)
	return !s.lockedUntil.IsZero() && s.lockedUntil.After(pkgtime.Now()), nil
}

func (ms *MemoryStore) RecordFailure(_ context.Context, key string) (bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	s := ms.state(key)
	s.failures++
	now := pkgtime.Now()

	if s.failures >= int64(ms.cfg.MaxAttempts) {
		s.lockedUntil = now.Add(ms.cfg.LockoutDuration)
		ms.cache.Set(key, s, ms.cfg.LockoutDuration)
		return true, nil
	}
	ms.cache.Set(key, s, ms.cfg.Window)
	return false, nil
}

func (ms *MemoryStore) Reset(_ context.Context, key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.cache.Set(key, &ipAuthState{}, ms.cfg.Window)
	return nil
}

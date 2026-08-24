package ratelimit

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/spdeepak/aegis/server/internal/config"
)

const (
	rlCounterPrefix = "aegis:rl:"
	rlLockPrefix    = "aegis:rl:lock:"
)

// recordScript atomically increments the failure counter for a key and, once it
// reaches the threshold, sets a lock key with the configured TTL. Running this
// as a single Lua script keeps the increment + lockout decision atomic
var recordScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
if count >= tonumber(ARGV[2]) then
  redis.call('SET', KEYS[2], '1', 'PX', tonumber(ARGV[3]) * 1000)
  return 1
end
return 0
`)

// RedisStore keeps brute-force state in a shared Redis instance, so it is
// centralized across all pods. All operations are fail-open: a backend error
// allows the request rather than risk a self-inflicted denial of service.
type RedisStore struct {
	client *redis.Client
	cfg    config.RateLimitConfig
}

func NewRedisStore(cfg config.RateLimitConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	return &RedisStore{client: client, cfg: cfg}, nil
}

func (r *RedisStore) IsLocked(ctx context.Context, key string) (bool, error) {
	locked, err := r.client.Exists(ctx, rlLockPrefix+key).Result()
	if err != nil {
		slog.Warn("rate-limit redis IsLocked error", "error", err)
		return false, nil // fail-open
	}
	return locked == 1, nil
}

func (r *RedisStore) RecordFailure(ctx context.Context, key string) (bool, error) {
	counter := rlCounterPrefix + key
	lock := rlLockPrefix + key
	res, err := recordScript.Run(ctx, r.client,
		[]string{counter, lock},
		int(r.cfg.Window.Seconds()),
		r.cfg.MaxAttempts,
		int(r.cfg.LockoutDuration.Seconds()),
	).Int()
	if err != nil {
		slog.Warn("rate-limit redis RecordFailure error", "error", err)
		return false, nil // fail-open
	}
	return res == 1, nil
}

func (r *RedisStore) Reset(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, rlCounterPrefix+key, rlLockPrefix+key).Err(); err != nil {
		slog.Warn("rate-limit redis Reset error", "error", err)
	}
	return nil // fail-open
}

// Close releases the underlying Redis client.
func (r *RedisStore) Close() error {
	return r.client.Close()
}

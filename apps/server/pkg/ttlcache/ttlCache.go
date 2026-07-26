package ttlcache

import (
	"sync"
	"time"

	pkgtime "github.com/spdeepak/aegis/server/pkg/time"
)

type (
	item struct {
		Value      any
		Expiration time.Time
	}

	Cache struct {
		mu     sync.RWMutex
		items  map[string]item
		stopCh chan struct{}
	}
)

func New(cleanUpInterval *time.Duration) *Cache {
	c := &Cache{
		items:  make(map[string]item),
		stopCh: make(chan struct{}),
	}
	if cleanUpInterval != nil {
		go c.startCleanup(*cleanUpInterval)
	} else {
		go c.startCleanup(1 * time.Minute)
	}
	return c
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		Value:      value,
		Expiration: pkgtime.Now().UTC().Add(ttl),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if pkgtime.Now().UTC().After(item.Expiration) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	return item.Value, true
}

func (c *Cache) Exists(key string) bool {
	_, exists := c.Get(key)
	return exists
}

func (c *Cache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := pkgtime.Now().UTC()

			c.mu.Lock()
			for k, v := range c.items {
				if now.After(v.Expiration) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()

		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache) Close() {
	close(c.stopCh)
}

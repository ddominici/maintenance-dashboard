package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type item struct {
	Data      []byte
	ExpiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	c := &MemoryCache{items: map[string]item{}}
	if cleanupInterval > 0 {
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				c.cleanup()
			}
		}()
	}
	return c
}

func (c *MemoryCache) Get(key string, dest any) bool {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(it.ExpiresAt) {
		if ok {
			c.Delete(key)
		}
		return false
	}
	return json.Unmarshal(it.Data, dest) == nil
}

func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.items[key] = item{Data: b, ExpiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	c.items = map[string]item{}
	c.mu.Unlock()
}

func (c *MemoryCache) cleanup() {
	now := time.Now()
	c.mu.Lock()
	for k, v := range c.items {
		if now.After(v.ExpiresAt) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

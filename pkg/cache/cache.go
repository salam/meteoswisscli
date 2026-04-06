package cache

import (
	"sync"
	"time"
)

// Cache provides a simple in-memory HTTP response cache with TTL.
type Cache struct {
	entries map[string]*entry
	mu      sync.RWMutex
	ttl     time.Duration
}

type entry struct {
	data      []byte
	expiresAt time.Time
}

// New creates a cache with the given TTL for entries.
func New(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*entry),
		ttl:     ttl,
	}
}

// Get returns cached data for the key if it exists and hasn't expired.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.data, true
}

// Set stores data in the cache with the configured TTL.
func (c *Cache) Set(key string, data []byte) {
	c.mu.Lock()
	c.entries[key] = &entry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

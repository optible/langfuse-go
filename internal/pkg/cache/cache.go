package cache

import (
	"sync"
	"time"
)

// Entry represents a cached item with expiration
type Entry[T any] struct {
	Value      T
	ExpiresAt  time.Time
	IsFallback bool
}

// IsExpired returns true if the entry has expired
func (e *Entry[T]) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache is a thread-safe cache with TTL support
type Cache[T any] struct {
	entries map[string]*Entry[T]
	mu      sync.RWMutex
	ttl     time.Duration
}

// New creates a new cache with the specified TTL
func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		entries: make(map[string]*Entry[T]),
		ttl:     ttl,
	}
}

// Get retrieves an item from the cache.
// Returns the value, whether it was found, and whether it's expired (stale).
func (c *Cache[T]) Get(key string) (T, bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		var zero T
		return zero, false, false
	}

	return entry.Value, true, entry.IsExpired()
}

// Set stores an item in the cache
func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &Entry[T]{
		Value:      value,
		ExpiresAt:  time.Now().Add(c.ttl),
		IsFallback: false,
	}
}

// SetWithTTL stores an item in the cache with a custom TTL
func (c *Cache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &Entry[T]{
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		IsFallback: false,
	}
}

// Delete removes an item from the cache
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all items from the cache
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry[T])
}

// Cleanup removes expired entries from the cache
func (c *Cache[T]) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Size returns the number of items in the cache
func (c *Cache[T]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// SetTTL updates the default TTL for new entries
func (c *Cache[T]) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}

// GetTTL returns the default TTL
func (c *Cache[T]) GetTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.ttl
}

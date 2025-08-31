// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements a thread-safe, bounded LRU+TTL cache for detecting token replay attacks.
// filename: aeiou/replay_cache.go
// nlines: 79
// risk_rating: MEDIUM

package aeiou

import (
	"container/list"
	"sync"
	"time"
)

// cacheEntry is a value stored in the LRU cache.
type cacheEntry struct {
	jti         string
	expiresAt   time.Time
	listElement *list.Element
}

// ReplayCache is a thread-safe, LRU cache with TTL for detecting token replays.
// It is designed to be used per-SID by the host's loop controller.
type ReplayCache struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration
	items    map[string]*cacheEntry
	evict    *list.List
}

// NewReplayCache creates a new replay cache with the given capacity and TTL.
func NewReplayCache(capacity int, ttl time.Duration) *ReplayCache {
	return &ReplayCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*cacheEntry),
		evict:    list.New(),
	}
}

// CheckAndAdd checks if a JTI has been seen. If it has, it returns ErrTokenReplay.
// If not, it adds the JTI to the cache and returns nil. It is thread-safe.
func (c *ReplayCache) CheckAndAdd(jti string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the item exists and is still valid
	if entry, ok := c.items[jti]; ok {
		if time.Now().Before(entry.expiresAt) {
			return ErrTokenReplay
		}
		// If it exists but is expired, remove it so it can be replaced
		c.removeItem(entry)
	}

	// Prune the oldest item if at capacity
	if c.evict.Len() >= c.capacity {
		oldest := c.evict.Back()
		if oldest != nil {
			c.removeItem(oldest.Value.(*cacheEntry))
		}
	}

	// Add the new item
	entry := &cacheEntry{
		jti:       jti,
		expiresAt: time.Now().Add(c.ttl),
	}
	entry.listElement = c.evict.PushFront(entry)
	c.items[jti] = entry

	return nil
}

// removeItem is an internal helper to remove an entry.
// It assumes the lock is already held.
func (c *ReplayCache) removeItem(entry *cacheEntry) {
	delete(c.items, entry.jti)
	c.evict.Remove(entry.listElement)
}

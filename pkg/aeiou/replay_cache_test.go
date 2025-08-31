// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 replay cache.
// filename: aeiou/replay_cache_test.go
// nlines: 75
// risk_rating: LOW

package aeiou

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestReplayCache(t *testing.T) {
	t.Run("Detects replay", func(t *testing.T) {
		cache := NewReplayCache(10, 1*time.Minute)
		jti := "test-jti-1"

		err := cache.CheckAndAdd(jti)
		if err != nil {
			t.Fatalf("First CheckAndAdd failed unexpectedly: %v", err)
		}

		err = cache.CheckAndAdd(jti)
		if !errors.Is(err, ErrTokenReplay) {
			t.Fatalf("Expected ErrTokenReplay on second add, got %v", err)
		}
	})

	t.Run("LRU eviction works", func(t *testing.T) {
		capacity := 3
		cache := NewReplayCache(capacity, 1*time.Minute)

		// Fill the cache
		for i := 0; i < capacity; i++ {
			jti := fmt.Sprintf("jti-%d", i)
			if err := cache.CheckAndAdd(jti); err != nil {
				t.Fatalf("Failed to add initial set of JTIs: %v", err)
			}
		}

		// This should evict the first item ("jti-0")
		if err := cache.CheckAndAdd("jti-new"); err != nil {
			t.Fatalf("Failed to add new JTI to full cache: %v", err)
		}

		// Now, adding the evicted item again should succeed
		if err := cache.CheckAndAdd("jti-0"); err != nil {
			t.Fatalf("Adding evicted JTI should have succeeded, but failed: %v", err)
		}
	})

	t.Run("TTL expiration works", func(t *testing.T) {
		ttl := 10 * time.Millisecond
		cache := NewReplayCache(10, ttl)
		jti := "test-jti-ttl"

		if err := cache.CheckAndAdd(jti); err != nil {
			t.Fatalf("Failed to add JTI: %v", err)
		}

		// Wait for the item to expire
		time.Sleep(ttl + 5*time.Millisecond)

		// Adding it again should now succeed
		if err := cache.CheckAndAdd(jti); err != nil {
			t.Fatalf("Adding expired JTI should have succeeded, but failed: %v", err)
		}
	})
}

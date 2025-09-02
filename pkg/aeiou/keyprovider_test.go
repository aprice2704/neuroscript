// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the RotatingKeyProvider, including concurrency safety.
// filename: aeiou/keyprovider_test.go
// nlines: 63
// risk_rating: MEDIUM

package aeiou

import (
	"crypto/ed25519"
	"fmt"
	"sync"
	"testing"
)

func TestRotatingKeyProvider(t *testing.T) {
	kp := NewRotatingKeyProvider()
	pubKey1, _, _ := ed25519.GenerateKey(nil)
	kid1 := "key-1"

	t.Run("Add and get key", func(t *testing.T) {
		kp.Add(kid1, pubKey1)
		retrievedKey, err := kp.PublicKey(kid1)
		if err != nil {
			t.Fatalf("PublicKey() failed unexpectedly: %v", err)
		}
		if !retrievedKey.Equal(pubKey1) {
			t.Errorf("Retrieved key does not match the added key")
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, err := kp.PublicKey("non-existent-key")
		if err == nil {
			t.Fatal("Expected an error for a non-existent key, but got nil")
		}
	})

	t.Run("Concurrency test", func(t *testing.T) {
		kp := NewRotatingKeyProvider()
		var wg sync.WaitGroup
		numRoutines := 50

		// Writer goroutines
		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				kid := fmt.Sprintf("concurrent-key-%d", i)
				pub, _, _ := ed25519.GenerateKey(nil)
				kp.Add(kid, pub)
			}(i)
		}

		// Reader goroutines
		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				kid := fmt.Sprintf("concurrent-key-%d", i%10) // Intentionally cause contention
				// We don't care about the error, just that it doesn't panic.
				_, _ = kp.PublicKey(kid)
			}(i)
		}

		wg.Wait()
	})
}

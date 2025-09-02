// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements a thread-safe key provider to support hot-reloading of keys for rotation.
// filename: aeiou/keyprovider.go
// nlines: 39
// risk_rating: MEDIUM

package aeiou

import (
	"crypto/ed25519"
	"fmt"
	"sync"
)

// RotatingKeyProvider is a thread-safe implementation of KeyProvider that allows
// for adding new public keys at runtime, enabling key rotation.
type RotatingKeyProvider struct {
	mu   sync.RWMutex
	keys map[string]ed25519.PublicKey
}

// NewRotatingKeyProvider creates an empty key provider.
func NewRotatingKeyProvider() *RotatingKeyProvider {
	return &RotatingKeyProvider{
		keys: make(map[string]ed25519.PublicKey),
	}
}

// Add adds or replaces a public key in the provider's key set.
// This method is safe for concurrent use.
func (p *RotatingKeyProvider) Add(kid string, pubKey ed25519.PublicKey) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.keys[kid] = pubKey
}

// PublicKey retrieves a public key by its ID. It returns an error if the key
// is not found. This method is safe for concurrent use.
func (p *RotatingKeyProvider) PublicKey(kid string) (ed25519.PublicKey, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	key, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found for kid: %s", kid)
	}
	return key, nil
}

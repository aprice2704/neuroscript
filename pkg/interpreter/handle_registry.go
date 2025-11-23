// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Concrete implementation of the interfaces.HandleRegistry.
// Latest change: Fixed NewHandle to return wrapped lang.ErrInvalidArgument for empty kind, satisfying tests.
// filename: pkg/interpreter/handle_registry.go
// nlines: 83
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/uuid"
)

// handleEntry holds the registered object and its kind/type tag.
type handleEntry struct {
	Kind    string
	Payload any
}

// HandleRegistry manages the lifetime and lookup of opaque references (handles)
// to Go-side objects, scoped to a single interpreter instance.
type HandleRegistry struct {
	registry map[string]handleEntry
	mu       sync.RWMutex
}

// NewHandleRegistry creates an initialized HandleRegistry.
func NewHandleRegistry() *HandleRegistry {
	return &HandleRegistry{
		registry: make(map[string]handleEntry),
	}
}

// NewHandle implements interfaces.HandleRegistry.
func (r *HandleRegistry) NewHandle(payload any, kind string) (interfaces.HandleValue, error) {
	if kind == "" {
		return nil, fmt.Errorf("handle kind cannot be empty: %w", lang.ErrInvalidArgument)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Use an unguessable ID (UUID) for opaqueness.
	id := fmt.Sprintf("%s-%s", kind, uuid.New().String())

	r.registry[id] = handleEntry{
		Kind:    kind,
		Payload: payload,
	}

	// lang.NewHandleValue is used here to create the concrete NeuroScript value type.
	return lang.NewHandleValue(id, kind), nil
}

// GetHandle implements interfaces.HandleRegistry.
func (r *HandleRegistry) GetHandle(id string) (any, error) {
	if id == "" {
		return nil, lang.ErrInvalidArgument
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.registry[id]
	if !ok {
		return nil, lang.ErrHandleNotFound
	}
	return entry.Payload, nil
}

// DeleteHandle implements interfaces.HandleRegistry.
func (r *HandleRegistry) DeleteHandle(id string) error {
	if id == "" {
		return lang.ErrInvalidArgument
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.registry[id]; !ok {
		return lang.ErrHandleNotFound
	}

	delete(r.registry, id)
	return nil
}

// CheckKind is a helper to verify the kind tag before returning the payload.
// Note: This is an internal helper, not part of the interfaces.HandleRegistry.
func (r *HandleRegistry) CheckKind(id, expectedKind string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.registry[id]
	if !ok {
		return nil, lang.ErrHandleNotFound
	}
	if entry.Kind != expectedKind {
		return nil, lang.ErrHandleWrongType
	}
	return entry.Payload, nil
}

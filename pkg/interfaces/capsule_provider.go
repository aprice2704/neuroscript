// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the host-provided capsule service interface, using 'any' to avoid import cycles.
// filename: pkg/interfaces/capsule_provider.go
// nlines: 40
// risk_rating: LOW

package interfaces

import (
	"context"
	// No 'pkg/lang' import
)

// CapsuleProvider defines the interface for a service that the
// interpreter's built-in 'tool.capsule.*' tools will query.
// The host application (e.g., FDM) can provide a concrete
// implementation of this interface to route all capsule
// tool calls to its own central capsule service.
type CapsuleProvider interface {
	// Add parses and adds a new capsule from its text content.
	// Corresponds to 'tool.capsule.Add'.
	// Returned value should be a map[string]any of the added capsule.
	Add(ctx context.Context, capsuleContent string) (any, error)

	// GetLatest retrieves the latest version of a capsule by its logical name.
	// Corresponds to 'tool.capsule.GetLatest'.
	// Returned value should be a map[string]any of the capsule.
	GetLatest(ctx context.Context, name string) (any, error)

	// List returns all available capsule IDs.
	// Corresponds to 'tool.capsule.List'.
	// Returned value should be a []string.
	List(ctx context.Context) (any, error)

	// Read retrieves a capsule by its full ID ('name@version') or name.
	// Corresponds to 'tool.capsule.Read'.
	// Returned value should be a map[string]any of the capsule.
	Read(ctx context.Context, id string) (any, error)
}

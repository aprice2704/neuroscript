// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the interfaces for the ProviderRegistry.
// filename: pkg/interfaces/provider.go
// nlines: 22
// risk_rating: LOW

package interfaces

// ProviderRegistryReader provides read-only access to the provider registry.
type ProviderRegistryReader interface {
	// List returns the names of all registered providers.
	List() []string
	// Get retrieves a registered provider service by its logical name.
	// The returned value is the concrete provider (e.g., *google.Provider).
	Get(name string) (any, bool)
}

// ProviderRegistryAdmin provides administrative (write) access to the provider registry.
type ProviderRegistryAdmin interface {
	ProviderRegistryReader
	// Register adds a new provider service implementation.
	// The provider 'p' is accepted as 'any' to avoid import cycles.
	Register(name string, p any) error
	// Delete removes a provider.
	Delete(name string) bool
}

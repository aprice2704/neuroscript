// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the interfaces for the Account store.
// filename: pkg/interfaces/account.go
// nlines: 19
// risk_rating: LOW

package interfaces

// AccountReader provides read-only access to the account store.
type AccountReader interface {
	// List returns the names of all registered accounts.
	List() []string
	// Get retrieves a registered account by its logical name.
	// The returned value is the concrete account struct.
	Get(name string) (any, bool)
}

// AccountAdmin provides administrative (write) access to the account store.
type AccountAdmin interface {
	// Register adds a new account from a configuration map.
	Register(name string, cfg map[string]any) error
	// Delete removes an account.
	Delete(name string) bool
}

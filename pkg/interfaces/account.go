// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Corrects AccountAdmin to embed AccountReader.
// Latest change: AccountAdmin now embeds AccountReader.
// filename: pkg/interfaces/account.go
// nlines: 22
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
	AccountReader // <-- FIX: Embed the reader interface

	// Register adds a new account from a configuration map.
	Register(name string, cfg map[string]any) error
	// RegisterFromAccount adds a new account from a concrete struct.
	// This is for host injection and accepts 'any' to avoid import cycles.
	RegisterFromAccount(acc any) error
	// Delete removes an account.
	Delete(name string) bool
}

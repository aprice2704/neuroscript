// NeuroScript Version: 1
// File version: 6
// Purpose: Defines the interfaces for the opaque, type-checked handle system (HandleValue and HandleRegistry).
// Latest change: Added HandleKind() to HandleValue and kind string to HandleRegistry.NewHandle per spec.
// filename: pkg/interfaces/handles.go
// nlines: 40

package interfaces

// NeuroScriptType is the base type for all language value kinds.
// It is defined here to break the import cycle with the 'lang' package.
type NeuroScriptType string

// Value is the minimal NeuroScript Value interface definition required to embed
// in HandleValue without creating an import cycle with the 'lang' package.
type Value interface {
	Type() NeuroScriptType
	String() string
	IsTruthy() bool
}

// HandleValue is the interface for an opaque NeuroScript value representing a host object handle.
// It is local to a single interpreter instance and cannot be serialized or persisted.
type HandleValue interface {
	Value               // Embeds the minimal Value interface
	HandleID() string   // returns the internal canonical handle id string
	HandleKind() string // returns the type-tag of the underlying payload (e.g., "fsmeta", "ast")
}

// HandleRegistry defines the contract for a system that can register Go objects
// and return NeuroScript HandleValue objects for them. This allows NeuroScript
// to manage host-provided resources safely.
type HandleRegistry interface {
	// NewHandle stores an object and returns a unique, opaque HandleValue for it.
	// The 'kind' string is used for type-checking and identification in tools.
	NewHandle(payload any, kind string) (HandleValue, error)
	// GetHandle retrieves the object associated with a given handle ID.
	GetHandle(id string) (any, error)
	// DeleteHandle removes the handle and frees the associated object reference.
	DeleteHandle(id string) error
}

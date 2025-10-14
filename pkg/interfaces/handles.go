// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the interface for a handle manager.
// filename: pkg/interfaces/handles.go
// nlines: 15
// risk_rating: LOW

package interfaces

// HandleManager defines the contract for a system that can register Go objects
// and return string-based handles for them. This allows NeuroScript to manage
// host-provided resources without exposing the underlying Go pointers directly
// to the script.
type HandleManager interface {
	// RegisterHandle stores an object and returns a unique handle for it.
	RegisterHandle(obj interface{}, typePrefix string) (string, error)
	// GetHandleValue retrieves the object associated with a given handle.
	GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error)
}

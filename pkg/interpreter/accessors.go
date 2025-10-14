// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Provides public accessors for the internal interpreter's fields.
// filename: pkg/interpreter/accessors.go
// nlines: 21
// risk_rating: LOW

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// HostContext returns the interpreter's host context, providing a safe,
// read-only way for external packages like 'api' to access it.
func (i *Interpreter) HostContext() *HostContext {
	return i.hostContext
}

// Handles returns the interpreter's handle manager. Because the interpreter
// now implements the handle management methods directly, it can return itself
// to satisfy the interface.
func (i *Interpreter) Handles() interfaces.HandleManager {
	// The *Interpreter type now has RegisterHandle and GetHandleValue methods,
	// so it satisfies the interfaces.HandleManager interface.
	return i
}

// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Provides public accessors for the internal interpreter's fields, including the tool registry.
// Latest change: Updated Handles() to HandleRegistry() accessor.
// filename: pkg/interpreter/accessors.go
// nlines: 28
// risk_rating: LOW

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// HostContext returns the interpreter's host context, providing a safe,
// read-only way for external packages like 'api' to access it.
func (i *Interpreter) HostContext() *HostContext {
	return i.hostContext
}

// HandleRegistry returns the interpreter's handle registry.
// This method is required to satisfy the interfaces.Interpreter interface.
func (i *Interpreter) HandleRegistry() interfaces.HandleRegistry {
	return i.handleRegistry
}

// ToolRegistry returns the interpreter's tool registry.
// This method is required to satisfy the tool.Runtime interface.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.tools
}

// NeuroScript Version: 0.8.0
// File version: 67
// Purpose: Fixes AEIOU hook by calling SetPublicAPI on the embedded internal interpreter.
// filename: pkg/api/interpreter.go
// nlines: 133

package api

import (
	"context"
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"

	// "github.com/aprice2704/neuroscript/pkg/provider/google" // Removed
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Interpreter is a facade over the internal interpreter. It embeds the internal
// interpreter to provide access to its methods, and adds identity-awareness
// by implementing the ActorProvider interface.
type Interpreter struct {
	*interpreter.Interpreter
}

// Statically assert that *Interpreter satisfies the tool.Runtime and ActorProvider interfaces.
var _ tool.Runtime = (*Interpreter)(nil)
var _ interfaces.ActorProvider = (*Interpreter)(nil)
var _ tool.Wrapper = (*Interpreter)(nil) // Statically assert wrapper interface

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...Option) *Interpreter {
	// 1. Create the internal interpreter.
	internalInterp := interpreter.NewInterpreter(opts...)

	// 2. Create the public API wrapper by embedding the internal instance.
	i := &Interpreter{
		Interpreter: internalInterp,
	}

	// 3. FIX: Set the public API pointer on the internal interpreter.
	//    This allows internal components (like the 'ask' hook) to call
	//    back with the correct public, wrapped *api.Interpreter instance.
	i.Interpreter.SetPublicAPI(i) // <-- This was i.SetPublicAPI(i)

	// 4. Replace the internal interpreter's tool registry with one that
	//    is aware of the public API facade. This ensures that when tools
	//    are called, they receive the public *api.Interpreter as their
	//    tool.Runtime, not the internal *interpreter.Interpreter.
	//    This is crucial for tools that need to access API-level methods
	//    like Actor().
	publicToolRegistry := tool.NewToolRegistry(i)
	internalInterp.SetToolRegistry(publicToolRegistry)

	// 5. Re-register tools now that the public-aware registry is in place.
	//    This is necessary because NewInterpreter already registered them
	//    against the old, internal-only registry.
	internalInterp.RegisterStandardTools()

	return i
}

// Actor returns the actor associated with the interpreter's HostContext.
// This method is part of the interfaces.ActorProvider interface,
// which the public *api.Interpreter satisfies.
func (i *Interpreter) Actor() (interfaces.Actor, bool) {
	return i.Interpreter.Actor()
}

// GetTurnContext satisfies the TurnContextProvider interface.
func (i *Interpreter) GetTurnContext() context.Context {
	return i.Interpreter.GetTurnContext()
}

// Execute is a wrapper for the internal Execute to provide the public API's signature.
func (i *Interpreter) Execute(program *interfaces.Tree) (Value, error) {
	return i.Interpreter.Execute(program.Root.(*Program))
}

// Unwrap implements the tool.Wrapper interface. It returns the embedded
// internal interpreter, allowing internal tools to access it directly.
func (i *Interpreter) Unwrap() tool.Runtime {
	return i.Interpreter
}

// Unwrap converts a NeuroScript api.Value back into a standard Go `any` type.
func Unwrap(v Value) (any, error) {
	if val, ok := v.(lang.Value); ok {
		return lang.Unwrap(val), nil
	}
	return v, nil
}

// ParseLoopControl is deprecated and will be removed.
func ParseLoopControl(output string) (*LoopController, error) {
	return nil, errors.New("ParseLoopControl is deprecated; use the AEIOU v3 LoopController")
}

// KnownProcedures returns a map of all loaded procedures (func definitions).
// This provides a stable public API for introspecting the interpreter's state.
func (i *Interpreter) KnownProcedures() map[string]*ast.Procedure {
	return i.Interpreter.KnownProcedures()
}

// KnownEventHandlers returns a map of all loaded event handlers (on event definitions),
// keyed by the event name.
// This provides a stable public API for introscessing the interpreter's state.
func (i *Interpreter) KnownEventHandlers() map[string][]*ast.OnEventDecl {
	return i.Interpreter.KnownEventHandlers()
}

// CapsuleProvider returns the host-provided capsule service, if one was injected.
func (i *Interpreter) CapsuleProvider() interfaces.CapsuleProvider {
	return i.Interpreter.CapsuleProvider()
}

func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.Interpreter.ToolRegistry()
}

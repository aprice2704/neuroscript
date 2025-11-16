// NeuroScript Version: 0.8.0
// File version: 73
// Purpose: Removes redundant init() and public stub functions.
// Latest change: Commented out CapsuleProvider facade method to align with internal interpreter.
// filename: pkg/api/interpreter.go
// nlines: 281

package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter" // Corrected import
	"github.com/aprice2704/neuroscript/pkg/lang"
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
	i.Interpreter.SetPublicAPI(i)

	// 4. Replace the internal interpreter's tool registry with one that
	//    is aware of the public API facade.
	publicToolRegistry := tool.NewToolRegistry(i)
	internalInterp.SetToolRegistry(publicToolRegistry)

	// 5. Re-register tools now that the public-aware registry is in place.
	internalInterp.RegisterStandardTools()

	return i
}

// Actor returns the actor associated with the interpreter's HostContext.
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

// CapsuleProvider returns the host-provided capsule service, if one was injected.
// func (i *Interpreter) CapsuleProvider() interfaces.CapsuleProvider {
// 	// COMMENTED OUT: This was part of the old mechanism.
// 	// return i.Interpreter.CapsuleProvider()
// 	return nil
// }

func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.Interpreter.ToolRegistry()
}

// symbolProvider is a helper to safely get the SymbolProvider from the HostContext.
func (i *Interpreter) symbolProvider() interfaces.SymbolProvider {
	if i.HostContext() == nil || i.HostContext().ServiceRegistry == nil {
		return nil
	}
	regMap, ok := i.HostContext().ServiceRegistry.(map[string]any)
	if !ok {
		if logger := i.GetLogger(); logger != nil {
			logger.Error("symbolProvider: HostContext.ServiceRegistry is not a map[string]any", "type", fmt.Sprintf("%T", i.HostContext().ServiceRegistry))
		}
		return nil
	}
	provider, _ := regMap[interfaces.SymbolProviderKey].(interfaces.SymbolProvider)
	return provider
}

// ---
// --- Symbol Provenance API: Local ---
// ---
func (i *Interpreter) LocalProcedures() map[string]*ast.Procedure {
	return i.Interpreter.KnownProcedures()
}
func (i *Interpreter) LocalEventHandlers() map[string][]*ast.OnEventDecl {
	return i.Interpreter.KnownEventHandlers()
}
func (i *Interpreter) LocalGlobalConstants() map[string]lang.Value {
	return i.Interpreter.KnownGlobalConstants()
}

// ---
// --- Symbol Provenance API: Provider ---
// ---
func (i *Interpreter) ProviderProcedures() map[string]*ast.Procedure {
	provider := i.symbolProvider()
	if provider == nil {
		return nil
	}
	rawMap := provider.ListProcedures()
	if rawMap == nil {
		return nil
	}
	result := make(map[string]*ast.Procedure, len(rawMap))
	for name, procAny := range rawMap {
		if proc, ok := procAny.(*ast.Procedure); ok {
			result[name] = proc
		} else if procAny != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("SymbolProvider:ListProcedures", "err", fmt.Sprintf("invalid type for symbol '%s': expected *ast.Procedure, got %T", name, procAny))
			}
		}
	}
	return result
}
func (i *Interpreter) ProviderEventHandlers() map[string][]*ast.OnEventDecl {
	provider := i.symbolProvider()
	if provider == nil {
		return nil
	}
	rawMap := provider.ListEventHandlers()
	if rawMap == nil {
		return nil
	}
	result := make(map[string][]*ast.OnEventDecl, len(rawMap))
	for eventName, handlersAny := range rawMap {
		handlersList := make([]*ast.OnEventDecl, 0, len(handlersAny))
		valid := true
		for idx, hAny := range handlersAny {
			if h, ok := hAny.(*ast.OnEventDecl); ok && h != nil {
				handlersList = append(handlersList, h)
			} else {
				if logger := i.GetLogger(); logger != nil {
					logger.Error("SymbolProvider:ListEventHandlers", "err", fmt.Sprintf("invalid type for symbol '%s' at index %d: expected *ast.OnEventDecl, got %T", eventName, idx, hAny))
				}
				valid = false
				break
			}
		}
		if valid {
			result[eventName] = handlersList
		}
	}
	return result
}
func (i *Interpreter) ProviderGlobalConstants() map[string]lang.Value {
	provider := i.symbolProvider()
	if provider == nil {
		return nil
	}
	rawMap := provider.ListGlobalConstants()
	if rawMap == nil {
		return nil
	}
	result := make(map[string]lang.Value, len(rawMap))
	for name, valAny := range rawMap {
		if val, ok := valAny.(lang.Value); ok {
			result[name] = val
		} else if valAny != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("SymbolProvider:ListGlobalConstants", "err", fmt.Sprintf("invalid type for symbol '%s': expected lang.Value, got %T", name, valAny))
			}
		}
	}
	return result
}

// ---
// --- Symbol Provenance API: Known (Merged) ---
// ---
func (i *Interpreter) KnownProcedures() map[string]*ast.Procedure {
	known := i.ProviderProcedures()
	if known == nil {
		known = make(map[string]*ast.Procedure)
	}
	local := i.LocalProcedures()
	for name, proc := range local {
		known[name] = proc
	}
	return known
}
func (i *Interpreter) KnownEventHandlers() map[string][]*ast.OnEventDecl {
	known := i.ProviderEventHandlers()
	if known == nil {
		known = make(map[string][]*ast.OnEventDecl)
	}
	local := i.LocalEventHandlers()
	for name, handler := range local {
		known[name] = handler
	}
	return known
}
func (i *Interpreter) KnownGlobalConstants() map[string]lang.Value {
	known := i.ProviderGlobalConstants()
	if known == nil {
		known = make(map[string]lang.Value)
	}
	local := i.LocalGlobalConstants()
	for name, val := range local {
		known[name] = val
	}
	return known
}

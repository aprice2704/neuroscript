// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Fixes event handler collision check by using HandlerName instead of Named.
// filename: pkg/interpreter/interpreter_load.go
// nlines: 97

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// symbolProvider is a helper to safely get the SymbolProvider from the HostContext.
func (i *Interpreter) symbolProvider() interfaces.SymbolProvider {
	if i.hostContext == nil || i.hostContext.ServiceRegistry == nil {
		return nil
	}
	regMap, ok := i.hostContext.ServiceRegistry.(map[string]any)
	if !ok {
		if logger := i.Logger(); logger != nil {
			logger.Warn("HostContext.ServiceRegistry is not a map[string]any", "type", fmt.Sprintf("%T", i.hostContext.ServiceRegistry))
		}
		return nil
	}
	provider, _ := regMap[interfaces.SymbolProviderKey].(interfaces.SymbolProvider)
	return provider
}

func (i *Interpreter) Load(tree *interfaces.Tree) error {
	if tree == nil || tree.Root == nil {
		i.Logger().Warn("Load called with a nil program AST.")
		i.state.knownProcedures = make(map[string]*ast.Procedure)
		i.eventManager.eventHandlers = make(map[string][]*ast.OnEventDecl)
		i.state.commands = []*ast.CommandNode{}
		// Do not clear constants, as they are loaded separately
		return nil
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("interpreter.Load: expected root node of type *ast.Program, but got %T", tree.Root)
	}

	// Get the symbol provider for collision checking
	provider := i.symbolProvider()

	// Clear existing *script-loaded* definitions
	i.state.knownProcedures = make(map[string]*ast.Procedure)
	i.eventManager.eventHandlers = make(map[string][]*ast.OnEventDecl)
	i.state.commands = []*ast.CommandNode{}
	// Note: We do NOT clear globalConstants, as they are set via tools, not Load.

	i.state.variablesMu.RLock() // Lock for reading constants
	defer i.state.variablesMu.RUnlock()

	for name, proc := range program.Procedures {
		// "No Override" Rule: Check local procedures (redundant, but safe)
		if _, exists := i.state.knownProcedures[name]; exists {
			return fmt.Errorf("symbol '%s' is already defined in this script", name)
		}
		// "No Override" Rule: Check provider procedures
		if provider != nil {
			if _, exists := provider.GetProcedure(name); exists {
				return fmt.Errorf("symbol '%s' is provided by the host and cannot be overridden", name)
			}
		}
		// "No Override" Rule: Check local constants
		if _, exists := i.state.globalConstants[name]; exists {
			return fmt.Errorf("symbol '%s' is already defined as a global constant and cannot be used for a procedure", name)
		}
		// "No Override" Rule: Check provider constants
		if provider != nil {
			if valAny, exists := provider.GetGlobalConstant(name); exists {
				// Check for nil, as GetGlobalConstant returns 'any'
				if valAny != nil {
					return fmt.Errorf("symbol '%s' is provided by the host as a constant and cannot be used for a procedure", name)
				}
			}
		}
		i.state.knownProcedures[name] = proc
	}

	for _, eventDecl := range program.Events {
		// FIX: Check for collision on the handler's optional name (e.g., 'as my_handler')
		if eventDecl.HandlerName != "" {
			handlerName := eventDecl.HandlerName
			// "No Override" Rule: Check provider handlers
			if provider != nil {
				if handlers, exists := provider.GetEventHandlers(handlerName); exists {
					// Check for nil/empty, as GetEventHandlers returns '[]any'
					if len(handlers) > 0 {
						return fmt.Errorf("named event handler '%s' is provided by the host and cannot be overridden", handlerName)
					}
				}
			}
		}

		// Register (which checks for local collisions)
		if err := i.eventManager.register(eventDecl, i); err != nil {
			// register already checks for local collisions, so we just pass the error
			return fmt.Errorf("failed to register event handler: %w", err)
		}
	}

	if program.Commands != nil {
		i.state.commands = program.Commands
	}
	return nil
}

// KnownGlobalConstants returns the interpreter's local map of global constants.
// This is used by the API for the "pre-charge" persistence workflow.
func (i *Interpreter) KnownGlobalConstants() map[string]lang.Value {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()

	if i.state.globalConstants == nil {
		// This map should be initialized in NewInterpreter, but we safe-guard here
		return make(map[string]lang.Value)
	}

	// Create a copy to avoid external mutation
	copied := make(map[string]lang.Value, len(i.state.globalConstants))
	for k, v := range i.state.globalConstants {
		copied[k] = v
	}
	return copied
}

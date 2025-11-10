// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Fixes "must condition" failure by returning 'true' from def_global_const.
// filename: pkg/interpreter/interpreter_tools.go
// nlines: 118

package interpreter

import (
	"crypto/ed25519"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// RegisterStandardTools registers the built-in toolsets.
func (i *Interpreter) RegisterStandardTools() {
	if i.tools == nil {
		i.Logger().Warn("RegisterStandardTools called with a nil tool registry. Skipping.")
		return
	}
	if !i.skipStdTools {
		if err := tool.RegisterGlobalToolsets(i.tools); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register global toolsets: %v", err))
		}
	}
	// This function is assumed to exist in another file in this package
	if err := registerDebugTools(i.tools); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register debug tools: %v", err))
	}
	// ADDED: Register the new symbol tools (e.g., def_global_const)
	if err := i.registerSymbolTools(i.tools); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register symbol tools: %v", err))
	}
	_, transientPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to generate transient private key for AEIOU tool: %v", err))
	}
	i.transientPrivateKey = transientPrivateKey
}

// registerSymbolTools registers the tools for defining symbols (e.g., global constants)
func (i *Interpreter) registerSymbolTools(registry tool.ToolRegistry) error {
	toolSpec := tool.ToolSpec{
		Name:  "def_global_const",
		Group: "ns",
		Args: []tool.ArgSpec{
			{Name: "name", Type: "string"},
			{Name: "value", Type: "any"},
		},
		Description: "Defines a global, immutable constant. Privileged operation.",
	}

	toolFunc := func(rt tool.Runtime, args []any) (any, error) {
		var interp *Interpreter
		if i, ok := rt.(*Interpreter); ok {
			interp = i
		} else if wrapper, ok := rt.(tool.Wrapper); ok {
			if i, ok := wrapper.Unwrap().(*Interpreter); ok {
				interp = i
			}
		}

		if interp == nil {
			return nil, fmt.Errorf("internal error: def_global_const tool received invalid runtime type %T", rt)
		}

		root := interp.rootInterpreter()

		// 1. Enforce privileged context
		if root.GetExecPolicy() == nil || root.GetExecPolicy().Context != policy.ContextConfig {
			return nil, lang.NewRuntimeError(lang.ErrorCodePolicy,
				"tool.ns.def_global_const can only be called from a privileged (config) context", lang.ErrPolicyViolation)
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("arg 0 'name' must be a string, got %T", args[0])
		}
		value, err := lang.Wrap(args[1])
		if err != nil {
			return nil, fmt.Errorf("failed to wrap arg 1 'value': %w", err)
		}

		// 2. "No Override" Collision Check
		root.state.variablesMu.Lock()
		defer root.state.variablesMu.Unlock()

		// Check local constants
		if _, exists := root.state.globalConstants[name]; exists {
			return nil, fmt.Errorf("symbol '%s' is already defined as a global constant", name)
		}
		// Check local procedures
		if _, exists := root.state.knownProcedures[name]; exists {
			return nil, fmt.Errorf("symbol '%s' is already defined as a procedure", name)
		}
		// Check local event handlers (less likely, but for completeness)
		if _, exists := root.eventManager.eventHandlers[name]; exists {
			return nil, fmt.Errorf("symbol '%s' is already defined as an event name", name)
		}

		// Check provider symbols
		if provider := root.symbolProvider(); provider != nil {
			if _, exists := provider.GetGlobalConstant(name); exists {
				return nil, fmt.Errorf("symbol '%s' is provided by the host and cannot be overridden", name)
			}
			if _, exists := provider.GetProcedure(name); exists {
				return nil, fmt.Errorf("symbol '%s' is provided by the host and cannot be overridden", name)
			}
			if _, exists := provider.GetEventHandlers(name); exists {
				return nil, fmt.Errorf("symbol '%s' is provided by the host and cannot be overridden", name)
			}
		}

		// 3. Define the constant
		root.state.globalConstants[name] = value
		// FIX: Return 'true' on success to satisfy 'must' statements.
		return true, nil
	}

	_, err := registry.RegisterTool(tool.ToolImplementation{
		Spec: toolSpec,
		Func: toolFunc,
	})
	return err
}

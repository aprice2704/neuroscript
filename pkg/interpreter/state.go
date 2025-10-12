// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Enforces read-only access to global variables from non-root scopes.
// filename: pkg/interpreter/state.go
// nlines: 85
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/generative-ai-go/genai"
)

func (i *Interpreter) SandboxDir() string { return i.state.sandboxDir }

func (i *Interpreter) FileAPI() interfaces.FileAPI {
	if i.hostContext == nil || i.hostContext.FileAPI == nil {
		panic("FATAL: Interpreter has no FileAPI configured in its HostContext.")
	}
	return i.hostContext.FileAPI
}

// SetVariable is the internal method for setting a variable with a lang.Value.
// It prevents modification of global variables from non-root scopes.
func (i *Interpreter) SetVariable(name string, value lang.Value) error {
	if name == "" {
		return errors.New("variable name cannot be empty")
	}

	// Enforce read-only globals. A non-root interpreter cannot change a global.
	root := i.rootInterpreter()
	if _, isGlobal := root.state.globalVarNames[name]; isGlobal && i != root {
		return lang.NewRuntimeError(lang.ErrorCodeWriteViolation,
			fmt.Sprintf("cannot modify global variable '%s' from a nested scope", name), nil)
	}

	i.state.variablesMu.Lock()
	defer i.state.variablesMu.Unlock()
	if i.state.variables == nil {
		i.state.variables = make(map[string]lang.Value)
	}
	i.state.variables[name] = value
	return nil
}

// GetVariable is the internal method for retrieving a variable as a lang.Value.
// It correctly looks up from local scope first, then to the root's globals.
func (i *Interpreter) GetVariable(name string) (lang.Value, bool) {
	i.state.variablesMu.RLock()
	val, exists := i.state.variables[name]
	i.state.variablesMu.RUnlock()

	if exists {
		return val, true
	}

	// If not in local scope, check the root's globals.
	root := i.rootInterpreter()
	if i != root {
		root.state.variablesMu.RLock()
		defer root.state.variablesMu.RUnlock()
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			val, exists = root.state.variables[name]
			return val, exists
		}
	}

	return nil, false
}

// GetVar satisfies the tool.Runtime interface.
func (i *Interpreter) GetVar(name string) (any, bool) {
	val, exists := i.GetVariable(name)
	if !exists {
		return nil, false
	}
	return lang.Unwrap(val), true
}

// SetVar satisfies the tool.Runtime interface.
func (i *Interpreter) SetVar(name string, val any) {
	wrappedVal, err := lang.Wrap(val)
	if err != nil {
		i.Logger().Error("Failed to set variable from tool runtime", "variable", name, "error", err)
		return
	}
	// Note: We deliberately ignore the error here because the tool.Runtime interface
	// does not allow returning one. The error (e.g., writing to a global) will be logged.
	if err := i.SetVariable(name, wrappedVal); err != nil {
		i.Logger().Error("Failed to set variable from tool runtime", "variable", name, "error", err)
	}
}

func (i *Interpreter) GenAIClient() *genai.Client {
	// This needs to be re-evaluated. The genai.Client should likely be part of the HostContext.
	// For now, returning nil to satisfy the interface.
	return nil
}

func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.state.vectorIndex == nil {
		i.state.vectorIndex = make(map[string][]float32)
	}
	return i.state.vectorIndex
}

func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.state.vectorIndex = vi }

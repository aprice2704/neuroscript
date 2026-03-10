// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 17
// :: description: Updates GetVariable to safely coerce custom primitive types returned by the host.
// :: latestChange: Added reflection-based coercion for custom string/int types (like nodes.FieldName) when lang.Wrap fails.
// :: filename: pkg/interpreter/state.go
// :: serialization: go

package interpreter

import (
	"errors"
	"fmt"
	"reflect"

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
// It searches in order: local variables, root global variables, root global constants,
// and finally the symbol provider's constants.
func (i *Interpreter) GetVariable(name string) (lang.Value, bool) {
	// 1. Check local scope
	i.state.variablesMu.RLock()
	val, exists := i.state.variables[name]
	i.state.variablesMu.RUnlock()
	if exists {
		return val, true
	}

	root := i.rootInterpreter()

	// Lock root to check its globals and constants
	root.state.variablesMu.RLock()

	if i != root {
		// 2. Check root's global variables (only if we are a fork)
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			val, exists = root.state.variables[name]
			if exists {
				root.state.variablesMu.RUnlock()
				return val, true
			}
		}
	}

	// 3. Check root's global constants
	val, exists = root.state.globalConstants[name]
	root.state.variablesMu.RUnlock() // Release before calling out to provider!

	if exists {
		return val, true
	}

	// 4. Check Symbol Provider's constants (only root has provider)
	if provider := root.symbolProvider(); provider != nil {
		valAny, exists := provider.GetGlobalConstant(name)
		if exists {
			// If it's already a lang.Value, use it
			if val, ok := valAny.(lang.Value); ok {
				return val, true
			}
			// Try standard wrap
			if wrapped, err := lang.Wrap(valAny); err == nil {
				return wrapped, true
			} else {
				// THE FIX: Coerce custom primitive types (like nodes.FieldName) if wrap fails
				rt := reflect.TypeOf(valAny)
				if rt != nil {
					switch rt.Kind() {
					case reflect.String:
						return lang.StringValue{Value: reflect.ValueOf(valAny).String()}, true
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						return lang.NumberValue{Value: float64(reflect.ValueOf(valAny).Int())}, true
					case reflect.Float32, reflect.Float64:
						return lang.NumberValue{Value: reflect.ValueOf(valAny).Float()}, true
					case reflect.Bool:
						return lang.BoolValue{Value: reflect.ValueOf(valAny).Bool()}, true
					}
				}
				root.Logger().Warn("Failed to wrap global constant and could not coerce", "name", name, "type", fmt.Sprintf("%T", valAny), "error", err)
			}
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

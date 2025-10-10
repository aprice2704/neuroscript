// NeuroScript Version: 0.8.0
// File version: 14
// Purpose: FEAT: Adds a public SetMaxLoopIterations method to allow configuration from tests and external packages.
// filename: pkg/interpreter/interpreter_state.go
// nlines: 90
// risk_rating: HIGH

package interpreter

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
)

func (i *Interpreter) SandboxDir() string { return i.state.sandboxDir }

func (i *Interpreter) Logger() interfaces.Logger {
	if i.parcel != nil && i.parcel.Logger() != nil {
		return i.parcel.Logger()
	}
	// Fallback to a no-op logger to guarantee a non-nil return and prevent panics.
	return logging.NewNoOpLogger()
}

func (i *Interpreter) FileAPI() interfaces.FileAPI {
	if i.fileAPI == nil {
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}

// SetMaxLoopIterations sets the safety limit for the number of iterations in a loop.
func (i *Interpreter) SetMaxLoopIterations(max int) {
	i.state.maxLoopIterations = max
}

// SetVariable is the internal method for setting a variable with a lang.Value.
func (i *Interpreter) SetVariable(name string, value lang.Value) error {
	i.state.variablesMu.Lock()
	defer i.state.variablesMu.Unlock()
	if i.state.variables == nil {
		i.state.variables = make(map[string]lang.Value)
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.state.variables[name] = value
	return nil
}

// GetVariable is the internal method for retrieving a variable as a lang.Value.
// It now falls back to the parcel's global variables.
func (i *Interpreter) GetVariable(name string) (lang.Value, bool) {
	// Check local scope first
	i.state.variablesMu.RLock()
	val, exists := i.state.variables[name]
	i.state.variablesMu.RUnlock()
	if exists {
		return val, true
	}

	// Fallback to globals from the parcel
	if i.parcel != nil {
		globals := i.parcel.Globals()
		if globals != nil {
			if globalVal, ok := globals[name]; ok {
				// We need to wrap it to lang.Value
				wrapped, err := lang.Wrap(globalVal)
				if err == nil {
					return wrapped, true
				}
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
		// Log the error, as the interface doesn't allow returning one.
		i.Logger().Error("Failed to set variable from tool runtime", "variable", name, "error", err)
		return
	}
	i.SetVariable(name, wrappedVal)
}

func (i *Interpreter) GenAIClient() *genai.Client {
	if i.aiWorker == nil {
		return nil
	}
	return i.aiWorker.Client()
}

func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.state.vectorIndex == nil {
		i.state.vectorIndex = make(map[string][]float32)
	}
	return i.state.vectorIndex
}

func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.state.vectorIndex = vi }

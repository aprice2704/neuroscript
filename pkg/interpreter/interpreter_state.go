// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Removed noisy debug Printf from SetVariable.
// filename: pkg/interpreter/interpreter_state.go
// nlines: 85
// risk_rating: MEDIUM

package interpreter

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/generative-ai-go/genai"
)

func (i *Interpreter) SandboxDir() string { return i.state.sandboxDir }

func (i *Interpreter) Logger() interfaces.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

func (i *Interpreter) FileAPI() interfaces.FileAPI {
	if i.fileAPI == nil {
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
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
func (i *Interpreter) GetVariable(name string) (lang.Value, bool) {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	if i.state.variables == nil {
		return nil, false
	}
	val, exists := i.state.variables[name]
	return val, exists
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

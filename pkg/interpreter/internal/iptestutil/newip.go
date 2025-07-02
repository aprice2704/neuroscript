// filename: pkg/interpreter/internal/iptestutil/newip.go
package iptestutil

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// NewTestInterpreter creates a new interpreter instance suitable for testing.
func NewTestInterpreter(t *testing.T, initialVars map[string]Value, lastResult Value) (*neurogo.Interpreter, error) {
	t.Helper()
	testLogger := NewTestLogger(t)

	noOpLLMClient, err := NewLLMClient("", "", testLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create NoOpLLMClient: %w", err)
	}
	sandboxDir := t.TempDir()

	interp, err := NewInterpreter(testLogger, noOpLLMClient, sandboxDir, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create test interpreter: %w", err)
	}

	if initialVars != nil {
		for k, v := range initialVars {
			if err := interp.SetVariable(k, v); err != nil {
				return nil, fmt.Errorf("failed to set initial variable %q: %w", k, err)
			}
		}
	}

	if lastResult != nil {
		interp.lastCallResult = lastResult
	}

	if err := tool.RegisterCoreTools(interp); err != nil {
		return nil, fmt.Errorf("failed to register core tools for test interpreter: %w", err)
	}

	if err := interp.SetSandboxDir(sandboxDir); err != nil {
		return nil, fmt.Errorf("failed to set sandbox dir for test interpreter: %w", err)
	}

	return interp, nil
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter.
func NewDefaultTestInterpreter(t *testing.T) (*neurogo.Interpreter, error) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Added WithoutStandardTools option for isolated testing.
// filename: pkg/interpreter/options.go
// nlines: 54
// risk_rating: LOW

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// InterpreterOption defines a function signature for configuring an Interpreter.
type InterpreterOption func(*Interpreter)

// WithoutStandardTools is an option that prevents the automatic registration
// of the standard tool library. This is useful for creating a lightweight or
// highly-sandboxed interpreter, especially for testing individual tools.
func WithoutStandardTools() InterpreterOption {
	return func(i *Interpreter) {
		i.skipStdTools = true
	}
}

// --- Functional Options ---

func WithLogger(logger interfaces.Logger) InterpreterOption {
	return func(i *Interpreter) {
		i.logger = logger
	}
}

func WithLLMClient(client interfaces.LLMClient) InterpreterOption {
	return func(i *Interpreter) {
		i.aiWorker = client
	}
}

func WithSandboxDir(path string) InterpreterOption {
	return func(i *Interpreter) {
		i.setSandboxDir(path)
	}
}

func WithStdout(w io.Writer) InterpreterOption {
	return func(i *Interpreter) {
		i.stdout = w
	}
}

func WithStdin(r io.Reader) InterpreterOption {
	return func(i *Interpreter) {
		i.stdin = r
	}
}

func WithStderr(w io.Writer) InterpreterOption {
	return func(i *Interpreter) {
		i.stderr = w
	}
}

// WithInitialGlobals sets the initial global variables.
func WithInitialGlobals(globals map[string]interface{}) InterpreterOption {
	return func(i *Interpreter) {
		for key, val := range globals {
			if err := i.SetInitialVariable(key, val); err != nil {
				i.logger.Error("Failed to set initial global variable", "key", key, "error", err)
			}
		}
	}
}

// NeuroScript Version: 0.7.1
// File version: 6
// Purpose: Adds WithCapsuleRegistry option to support host-provided capsules.
// filename: pkg/interpreter/interpreter_options.go
// nlines: 75
// risk_rating: LOW

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/policy"
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
		i.SetSandboxDir(path)
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

// WithGlobals sets the initial global variables.
func WithGlobals(globals map[string]interface{}) InterpreterOption {
	return func(i *Interpreter) {
		for key, val := range globals {
			if err := i.SetInitialVariable(key, val); err != nil {
				i.logger.Error("Failed to set initial global variable", "key", key, "error", err)
			}
		}
	}
}

// WithExecPolicy applies a runtime execution policy to the interpreter.
func WithExecPolicy(policy *policy.ExecPolicy) InterpreterOption {
	return func(i *Interpreter) {
		i.ExecPolicy = policy
	}
}

// WithCapsuleRegistry creates an interpreter option that adds a custom
// capsule registry to the interpreter's store.
func WithCapsuleRegistry(registry *capsule.Registry) InterpreterOption {
	return func(i *Interpreter) {
		if i.capsuleStore != nil {
			i.capsuleStore.Add(registry)
		}
	}
}

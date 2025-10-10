// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Updates options to configure the RunnerParcel.
// filename: pkg/interpreter/interpreter_options.go
// nlines: 100
// risk_rating: MEDIUM

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
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

// WithParcel sets the entire runner parcel for the interpreter.
func WithParcel(p contract.RunnerParcel) InterpreterOption {
	return func(i *Interpreter) {
		i.parcel = p
	}
}

func WithLogger(logger interfaces.Logger) InterpreterOption {
	return func(i *Interpreter) {
		if i.parcel == nil {
			i.parcel = contract.NewParcel(nil, nil, logger, nil)
		} else {
			i.parcel = i.parcel.Fork(func(m *contract.ParcelMut) {
				m.Logger = logger
			})
		}
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

// WithGlobals sets the initial global variables on the parcel.
func WithGlobals(globals map[string]interface{}) InterpreterOption {
	return func(i *Interpreter) {
		// This is a bit tricky now. We need to create a new parcel with these globals.
		// For now, let's just log an error if the parcel is already set.
		if i.parcel != nil && len(i.parcel.Globals()) > 0 {
			if i.parcel.Logger() != nil {
				i.parcel.Logger().Error("WithGlobals should be used before other options that create a parcel.")
			}
			return
		}
		var logger interfaces.Logger
		if i.parcel != nil {
			logger = i.parcel.Logger()
		}
		i.parcel = contract.NewParcel(nil, nil, logger, globals)
	}
}

// WithExecPolicy applies a runtime execution policy to the interpreter's parcel.
func WithExecPolicy(policy *interfaces.ExecPolicy) InterpreterOption {
	return func(i *Interpreter) {
		if i.parcel == nil {
			i.parcel = contract.NewParcel(nil, policy, nil, nil)
		} else {
			i.parcel = i.parcel.Fork(func(m *contract.ParcelMut) {
				m.Policy = policy
			})
		}
	}
}

// WithCapsuleRegistry creates an interpreter option that adds a custom
// capsule registry to the interpreter's store for read-only access.
func WithCapsuleRegistry(registry *capsule.Registry) InterpreterOption {
	return func(i *Interpreter) {
		if i.capsuleStore != nil {
			i.capsuleStore.Add(registry)
		}
	}
}

// WithCapsuleAdminRegistry provides a writable capsule registry to the interpreter.
// This is for trusted, configuration contexts where scripts need to persist new capsules.
func WithCapsuleAdminRegistry(registry *capsule.Registry) InterpreterOption {
	return func(i *Interpreter) {
		i.adminCapsuleRegistry = registry
	}
}

// WithEventHandlerErrorCallback registers a function to be called when a runtime
// error occurs during the execution of an 'on event' handler.
func WithEventHandlerErrorCallback(f func(eventName, source string, err *lang.RuntimeError)) InterpreterOption {
	return func(i *Interpreter) {
		i.eventHandlerErrorCallback = f
	}
}

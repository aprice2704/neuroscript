// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Replaces individual host-related options with a single WithHostContext option.
// filename: pkg/interpreter/options.go
// nlines: 65
// risk_rating: MEDIUM

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// InterpreterOption defines a function signature for configuring an Interpreter.
type InterpreterOption func(*Interpreter)

// WithHostContext provides the interpreter with all its host-provided dependencies.
// This is the primary and preferred way to configure host capabilities.
func WithHostContext(hc *HostContext) InterpreterOption {
	return func(i *Interpreter) {
		i.hostContext = hc
	}
}

// WithoutStandardTools is an option that prevents the automatic registration
// of the standard tool library. This is useful for creating a lightweight or
// highly-sandboxed interpreter, especially for testing individual tools.
func WithoutStandardTools() InterpreterOption {
	return func(i *Interpreter) {
		i.skipStdTools = true
	}
}

func WithSandboxDir(path string) InterpreterOption {
	return func(i *Interpreter) {
		i.SetSandboxDir(path)
	}
}

// WithGlobals sets the initial global variables.
func WithGlobals(globals map[string]interface{}) InterpreterOption {
	return func(i *Interpreter) {
		for key, val := range globals {
			if err := i.SetInitialVariable(key, val); err != nil {
				// At this stage, the logger might not be configured,
				// so a panic is not unreasonable if globals are malformed.
				// However, to be safe, we'll just ignore for now.
				// A proper logger should be available post-refactor.
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

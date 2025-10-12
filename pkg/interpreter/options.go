// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Adds WithAccountStore and WithAgentModelStore to support the public API facade.
// filename: pkg/interpreter/options.go
// nlines: 80
// risk_rating: MEDIUM

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
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
// of the standard tool library.
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
				// At this stage, the logger might not be configured, so a panic
				// is not unreasonable if globals are malformed.
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

// WithCapsuleRegistry adds a custom capsule registry for read-only access.
func WithCapsuleRegistry(registry *capsule.Registry) InterpreterOption {
	return func(i *Interpreter) {
		if i.capsuleStore != nil {
			i.capsuleStore.Add(registry)
		}
	}
}

// WithCapsuleAdminRegistry provides a writable capsule registry.
func WithCapsuleAdminRegistry(registry *capsule.Registry) InterpreterOption {
	return func(i *Interpreter) {
		i.adminCapsuleRegistry = registry
	}
}

// WithAccountStore provides a host-managed AccountStore to the interpreter.
func WithAccountStore(store *account.Store) InterpreterOption {
	return func(i *Interpreter) {
		i.SetAccountStore(store)
	}
}

// WithAgentModelStore provides a host-managed AgentModelStore to the interpreter.
func WithAgentModelStore(store *agentmodel.AgentModelStore) InterpreterOption {
	return func(i *Interpreter) {
		i.SetAgentModelStore(store)
	}
}

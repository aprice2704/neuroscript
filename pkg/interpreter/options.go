// NeuroScript Version: 0.8.0
// File version: 21
// Purpose: Adds WithAccountAdmin and WithAgentModelAdmin for facade injection.
// Latest change: Removed ALL capsule options except WithCapsuleStore.
// filename: pkg/interpreter/options.go
// nlines: 106

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
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

// WithActor provides the interpreter with an identity for the current execution context.
// It modifies the HostContext.
func WithActor(actor interfaces.Actor) InterpreterOption {
	return func(i *Interpreter) {
		if i.hostContext != nil {
			i.hostContext.Actor = actor
		}
	}
}

// WithParser injects a pre-configured ParserAPI instance into the interpreter.
func WithParser(p *parser.ParserAPI) InterpreterOption {
	return func(i *Interpreter) {
		i.parser = p
	}
}

// WithASTBuilder injects a pre-configured ASTBuilder instance into the interpreter.
func WithASTBuilder(b *parser.ASTBuilder) InterpreterOption {
	return func(i *Interpreter) {
		i.astBuilder = b
	}
}

// WithAITranscriptWriter sets the writer for the AI conversation transcript.
// This is a convenience option that modifies the HostContext. It should be used
// after WithHostContext.
func WithAITranscriptWriter(w io.Writer) InterpreterOption {
	return func(i *Interpreter) {
		if i.hostContext != nil {
			i.hostContext.AITranscript = w
		}
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

// WithCapsuleStore provides a complete, layered capsule store to the interpreter.
// This store will be used for ALL capsule operations (read, write, list).
// The store's first registry (index 0) will be used for writes.
func WithCapsuleStore(store *capsule.Store) InterpreterOption {
	return func(i *Interpreter) {
		i.capsuleStore = store
	}
}

// WithCapsuleRegistry -- REMOVED.
// func WithCapsuleRegistry(registry *capsule.Registry) InterpreterOption {
// }

// WithCapsuleAdminRegistry -- REMOVED.
// func WithCapsuleAdminRegistry(registry *capsule.Registry) InterpreterOption {
// }

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

// WithProviderRegistry provides a host-managed ProviderRegistry to the interpreter.
func WithProviderRegistry(registry *provider.Registry) InterpreterOption {
	return func(iThis *Interpreter) {
		iThis.SetProviderRegistry(registry)
	}
}

// WithCapsuleProvider -- REMOVED.
// func WithCapsuleProvider(provider interfaces.CapsuleProvider) InterpreterOption {
// }

// --- CORRECTED FACADE OPTIONS ---

// WithAccountAdmin injects a host-provided implementation of the account store
// that satisfies the interfaces.AccountAdmin interface.
func WithAccountAdmin(admin interfaces.AccountAdmin) InterpreterOption {
	return func(i *Interpreter) {
		i.accountAdmin = admin
	}
}

// WithAgentModelAdmin injects a host-provided implementation of the model store
// that satisfies the interfaces.AgentModelAdmin interface.
func WithAgentModelAdmin(admin interfaces.AgentModelAdmin) InterpreterOption {
	return func(i *Interpreter) {
		i.agentModelAdmin = admin
	}
}

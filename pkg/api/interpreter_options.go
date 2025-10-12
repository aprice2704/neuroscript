// NeuroScript Version: 0.8.0
// File version: 52
// Purpose: Adds the missing WithSandboxDir option to correctly configure the interpreter's filesystem sandbox.
// filename: pkg/api/interpreter_options.go
// nlines: 107
// risk_rating: MEDIUM

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// WithHostContext provides the interpreter with its connection to the host.
// This is the primary and mandatory option for configuration. The provided
// HostContext must be created using the NewHostContextBuilder.
func WithHostContext(hc *HostContext) Option {
	// This function acts as an adapter, converting the public api.HostContext
	// to the internal interpreter.HostContext that the interpreter core expects.
	internalHC := &interpreter.HostContext{
		Logger:                    hc.logger,
		Emitter:                   hc.emitter,
		AITranscript:              hc.aiTranscript,
		Stdout:                    hc.stdout,
		Stdin:                     hc.stdin,
		Stderr:                    hc.stderr,
		EmitFunc:                  hc.emitFunc,
		WhisperFunc:               hc.whisperFunc,
		EventHandlerErrorCallback: hc.eventHandlerErrorCallback,
	}
	return interpreter.WithHostContext(internalHC)
}

// WithGlobals creates an interpreter option to set initial global variables.
func WithGlobals(globals map[string]any) Option {
	return interpreter.WithGlobals(globals)
}

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		if _, err := i.ToolRegistry().RegisterTool(t); err != nil {
			// A logger might not be set yet, so we can't reliably log.
			// This should be caught during development or testing.
		}
	}
}

// WithCapsuleRegistry adds a custom capsule registry to the interpreter's store.
func WithCapsuleRegistry(registry *CapsuleRegistry) Option {
	return interpreter.WithCapsuleRegistry(registry)
}

// WithCapsuleAdminRegistry provides a writable capsule registry to the interpreter.
func WithCapsuleAdminRegistry(registry *AdminCapsuleRegistry) Option {
	return interpreter.WithCapsuleAdminRegistry(registry)
}

// WithAccountStore provides a host-managed AccountStore to the interpreter.
func WithAccountStore(store *AccountStore) Option {
	return interpreter.WithAccountStore(store)
}

// WithAgentModelStore provides a host-managed AgentModelStore to the interpreter.
func WithAgentModelStore(store *AgentModelStore) Option {
	return interpreter.WithAgentModelStore(store)
}

// WithExecPolicy applies a runtime execution policy to the interpreter.
func WithExecPolicy(policy *ExecPolicy) Option {
	return interpreter.WithExecPolicy(policy)
}

// WithSandboxDir sets the root directory for all filesystem operations.
func WithSandboxDir(path string) Option {
	return interpreter.WithSandboxDir(path)
}

// WithInterpreter creates an option to reuse the internal state of an existing
// interpreter. This is for advanced use cases and should be used with caution.
func WithInterpreter(existing *Interpreter) Option {
	return func(i *interpreter.Interpreter) {
		if existing != nil && existing.internal != nil {
			*i = *existing.internal
		}
	}
}

// RegisterCriticalErrorHandler allows the host application to override the default
// panic behavior for critical errors.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError)) {
	lang.RegisterCriticalHandler(h)
}

// MakeToolFullName creates a correctly formatted, fully-qualified tool name.
func MakeToolFullName(group, name string) types.FullName {
	return types.MakeFullName(group, name)
}

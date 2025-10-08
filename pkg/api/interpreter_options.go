// NeuroScript Version: 0.7.4
// File version: 45
// Purpose: FIX: Corrected WithInterpreter to use a safe pointer assignment instead of copying a lock.
// filename: pkg/api/interpreter_options.go
// nlines: 139
// risk_rating: MEDIUM

package api

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// WithSandboxDir creates an interpreter option that sets the secure root directory
// for all subsequent file operations.
func WithSandboxDir(path string) interpreter.InterpreterOption {
	return interpreter.WithSandboxDir(path)
}

// WithStdout creates an interpreter option to set the standard output writer.
func WithStdout(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStdout(w)
}

// WithStderr creates an interpreter option to set the standard error writer.
func WithStderr(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStderr(w)
}

// WithLogger creates an interpreter option to provide a custom logger.
func WithLogger(logger Logger) Option {
	return interpreter.WithLogger(logger)
}

// WithGlobals creates an interpreter option to set initial global variables.
func WithGlobals(globals map[string]any) Option {
	return interpreter.WithGlobals(globals)
}

// WithAITranscript provides a writer to log the full, composed prompts sent to AI providers.
func WithAITranscript(w io.Writer) interpreter.InterpreterOption {
	return func(i *interpreter.Interpreter) {
		i.SetAITranscript(w)
	}
}

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		if _, err := i.ToolRegistry().RegisterTool(t); err != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.Spec.Name, "error", err)
			}
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

// WithEmitter creates an interpreter option to set a custom LLM event emitter.
func WithEmitter(emitter Emitter) Option {
	return func(i *interpreter.Interpreter) {
		i.SetEmitter(emitter)
	}
}

// WithEmitFunc creates an interpreter option to set a custom emit handler.
func WithEmitFunc(f func(Value)) Option {
	return func(i *interpreter.Interpreter) {
		i.SetEmitFunc(func(v lang.Value) {
			f(v)
		})
	}
}

// WithEventHandlerErrorCallback creates an interpreter option to set a custom callback for event handler errors.
func WithEventHandlerErrorCallback(f func(eventName, source string, err *RuntimeError)) Option {
	return interpreter.WithEventHandlerErrorCallback(f)
}

// WithAccountStore provides a host-managed AccountStore to the interpreter.
func WithAccountStore(store *AccountStore) Option {
	return func(i *interpreter.Interpreter) {
		if store != nil {
			i.SetAccountStore(store)
		}
	}
}

// WithAgentModelStore provides a host-managed AgentModelStore to the interpreter.
func WithAgentModelStore(store *AgentModelStore) Option {
	return func(i *interpreter.Interpreter) {
		if store != nil {
			i.SetAgentModelStore(store)
		}
	}
}

// RegisterCriticalErrorHandler allows the host application to override the default panic behavior.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError)) {
	lang.RegisterCriticalHandler(h)
}

// MakeToolFullName creates a correctly formatted, fully-qualified tool name.
func MakeToolFullName(group, name string) types.FullName {
	return types.MakeFullName(group, name)
}

// WithExecPolicy applies a runtime execution policy to the interpreter.
func WithExecPolicy(policy *ExecPolicy) Option {
	return interpreter.WithExecPolicy(policy)
}

// WithInterpreter creates an option to reuse the internal state of an existing interpreter.
func WithInterpreter(existing *Interpreter) Option {
	return func(i *interpreter.Interpreter) {
		if existing != nil && existing.internal != nil {
			// FIX: Use a safe pointer assignment instead of copying the struct value.
			i = existing.internal
		}
	}
}

// WithRuntime creates an interpreter option to set a custom runtime context.
func WithRuntime(rt Runtime) Option {
	return func(i *interpreter.Interpreter) {
		i.SetRuntime(rt)
	}
}

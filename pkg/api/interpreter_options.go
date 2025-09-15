// NeuroScript Version: 0.7.2
// File version: 36
// Purpose: With* for interpreter
// filename: pkg/api/reexport.go

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

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
// This allows host applications to layer in their own documentation for read-only access.
func WithCapsuleRegistry(registry *CapsuleRegistry) Option {
	return interpreter.WithCapsuleRegistry(registry)
}

// WithCapsuleAdminRegistry provides a writable capsule registry to the interpreter.
// This is for trusted, configuration contexts where scripts need to persist new capsules.
func WithCapsuleAdminRegistry(registry *CapsuleRegistry) Option {
	return interpreter.WithCapsuleAdminRegistry(registry)
}

// WithEmitter creates an interpreter option to set a custom LLM event emitter.
// The provided emitter will be called for every LLM interaction.
func WithEmitter(emitter Emitter) Option {
	return func(i *interpreter.Interpreter) {
		i.SetEmitter(emitter)
	}
}

// WithEmitFunc creates an interpreter option to set a custom emit handler.
func WithEmitFunc(f func(Value)) Option {
	return func(i *interpreter.Interpreter) {
		// We wrap the api.Value in the function signature to avoid exposing lang.Value.
		i.SetEmitFunc(func(v lang.Value) {
			f(v)
		})
	}
}

// WithEventHandlerErrorCallback creates an interpreter option to set a custom
// callback for handling errors that occur within event handlers.
func WithEventHandlerErrorCallback(f func(eventName, source string, err *RuntimeError)) Option {
	return interpreter.WithEventHandlerErrorCallback(f)
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

// WithExecPolicy applies a runtime execution policy to the interpreter.
func WithExecPolicy(policy *ExecPolicy) Option {
	return interpreter.WithExecPolicy(policy)
}

// WithInterpreter creates an option to reuse the internal state of an existing
// interpreter. This is useful for the host-managed ask-loop pattern.
func WithInterpreter(existing *Interpreter) Option {
	return func(i *interpreter.Interpreter) {
		if existing != nil && existing.internal != nil {
			*i = *existing.internal
		}
	}
}

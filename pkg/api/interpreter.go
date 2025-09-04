// NeuroScript Version: 0.7.0
// File version: 29
// Purpose: Implemented the critical fix for state persistence by ensuring Run() operates directly on the persistent internal interpreter, not a clone.
// filename: pkg/api/interpreter.go
// nlines: 165
// risk_rating: HIGH
package api

import (
	"errors"
	"io"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/google"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Interpreter is a facade over the internal interpreter, providing a stable,
// high-level API for embedding NeuroScript.
type Interpreter struct {
	internal *interpreter.Interpreter
}

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...Option) *Interpreter {
	i := interpreter.NewInterpreter(opts...)
	if i.ExecPolicy == nil {
		i.ExecPolicy = &policy.ExecPolicy{
			Context: policy.ContextNormal,
			Allow:   []string{},
		}
	}
	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)
	return &Interpreter{internal: i}
}

// ... (With... options are unchanged) ...
func WithSandboxDir(path string) interpreter.InterpreterOption {
	return interpreter.WithSandboxDir(path)
}
func WithStdout(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStdout(w)
}
func WithStderr(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStderr(w)
}
func WithLogger(logger Logger) Option {
	return interpreter.WithLogger(logger)
}
func WithGlobals(globals map[string]any) Option {
	return interpreter.WithGlobals(globals)
}

// WithAITranscript provides a writer to log the full, composed prompts sent to AI providers.
func WithAITranscript(w io.Writer) interpreter.InterpreterOption {
	return func(i *interpreter.Interpreter) {
		i.SetAITranscript(w)
	}
}

// SetSandboxDir sets the secure root directory for all subsequent file operations
// for this interpreter instance.
func (i *Interpreter) SetSandboxDir(path string) {
	i.internal.SetSandboxDir(path)
}

// SetStdout sets the standard output writer for the interpreter instance.
func (i *Interpreter) SetStdout(w io.Writer) {
	i.internal.SetStdout(w)
}

// SetStderr sets the standard error writer for the interpreter instance.
func (i *Interpreter) SetStderr(w io.Writer) {
	i.internal.SetStderr(w)
}

// SetEmitFunc sets a custom handler for the 'emit' statement.
func (i *Interpreter) SetEmitFunc(f func(Value)) {
	i.internal.SetEmitFunc(func(v lang.Value) {
		f(v)
	})
}

// RegisterProvider allows the host application to register a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p AIProvider) {
	i.internal.RegisterProvider(name, p)
}

// RegisterAgentModel allows the host application to register an AgentModel configuration.
// It now correctly accepts a map of native Go types.
func (i *Interpreter) RegisterAgentModel(name string, config map[string]any) error {
	return i.internal.AgentModelsAdmin().Register(types.AgentModelName(name), config)
}

// Load injects a parsed program into the interpreter via the interface.
func (i *Interpreter) Load(tree *interfaces.Tree) error {
	return i.internal.Load(tree)
}

// ExecuteCommands runs any unnamed 'command' blocks found in the loaded program.
func (i *Interpreter) ExecuteCommands() (Value, error) {
	return i.internal.ExecuteCommands()
}

// Run calls a specific, named procedure from the loaded program.
// It now executes directly on the persistent interpreter instance to maintain state.
func (i *Interpreter) Run(procName string, args ...lang.Value) (Value, error) {
	result, err := i.internal.Run(procName, args...)
	return result, err
}

// EmitEvent sends an event into an event-sink script.
func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.internal.EmitEvent(eventName, source, payload)
}

// ToolRegistry returns the tool registry associated with the interpreter.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.internal.ToolRegistry()
}

// Unwrap converts a NeuroScript api.Value back into a standard Go `any` type.
func Unwrap(v Value) (any, error) {
	if val, ok := v.(lang.Value); ok {
		return lang.Unwrap(val), nil
	}
	return v, nil
}

// ParseLoopControl is deprecated and will be removed. Use the re-exported
// aeiou.LoopController for V3-compliant loop management.
func ParseLoopControl(output string) (*LoopControl, error) {
	return nil, errors.New("ParseLoopControl is deprecated; use the AEIOU v3 LoopController")
}

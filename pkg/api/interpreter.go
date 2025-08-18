// NeuroScript Version: 0.6.0
// File version: 16
// Purpose: Exposes the internal tool registry via a public ToolRegistry() method for use in tools like the LSP.
// filename: pkg/api/interpreter.go
// nlines: 112
// risk_rating: LOW

package api

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"

	// Import default providers to make them available for registration.
	"github.com/aprice2704/neuroscript/pkg/api/providers/google"
)

// Interpreter is a facade over the internal interpreter, providing a stable,
// high-level API for embedding NeuroScript.
type Interpreter struct {
	internal *interpreter.Interpreter
}

// New creates a new, persistent NeuroScript interpreter instance.
// It now automatically registers a default set of AI providers.
func New(opts ...Option) *Interpreter {
	// The internal NewInterpreter now handles all setup, including applying
	// options and registering all standard tools in the correct order.
	i := interpreter.NewInterpreter(opts...)

	// Automatically register the default providers for this new instance.
	// For now, this is just the Google provider.
	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	return &Interpreter{internal: i}
}

// WithSandboxDir returns an option to set the secure directory for file operations.
// This is a mandatory setting for any interpreter that will interact with the filesystem.
func WithSandboxDir(path string) interpreter.InterpreterOption {
	return interpreter.WithSandboxDir(path)
}

// WithStdout returns an option to set the standard output for the interpreter.
func WithStdout(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStdout(w)
}

// WithStderr returns an option to set the standard error for the interpreter.
func WithStderr(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStderr(w)
}

// WithLogger creates an interpreter option to set a custom logger.
func WithLogger(logger Logger) Option {
	return interpreter.WithLogger(logger)
}

// SetStdout sets the standard output writer for the interpreter instance.
func (i *Interpreter) SetStdout(w io.Writer) {
	i.internal.SetStdout(w)
}

// SetStderr sets the standard error writer for the interpreter instance.
func (i *Interpreter) SetStderr(w io.Writer) {
	i.internal.SetStderr(w)
}

// RegisterProvider allows the host application to register a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p AIProvider) {
	i.internal.RegisterProvider(name, p)
}

// Load injects a verified and parsed program into the interpreter's memory.
func (i *Interpreter) Load(p *ast.Program) error {
	return i.internal.Load(p)
}

// ExecuteCommands runs any unnamed 'command' blocks found in the loaded program.
func (i *Interpreter) ExecuteCommands() (Value, error) {
	return i.internal.ExecuteCommands()
}

// Run calls a specific, named procedure from the loaded program.
func (i *Interpreter) Run(procName string, args ...lang.Value) (Value, error) {
	result, err := i.internal.Run(procName, args...)
	return result, err
}

// EmitEvent sends an event into an event-sink script.
func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.internal.EmitEvent(eventName, source, payload)
}

// THE FIX IS HERE: Expose the tool registry via a getter.
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

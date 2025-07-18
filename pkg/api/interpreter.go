// NeuroScript Version: 0.6.0
// File version: 5
// Purpose: Adds debug logging to the public API facade to trace the return value as it crosses the package boundary.
// filename: pkg/api/interpreter.go
// nlines: 55
// risk_rating: MEDIUM

package api

import (
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Interpreter is a facade over the internal interpreter, providing a stable,
// high-level API for embedding NeuroScript.
type Interpreter struct {
	internal *interpreter.Interpreter
}

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...interpreter.InterpreterOption) *Interpreter {
	i := interpreter.NewInterpreter(opts...)
	return &Interpreter{internal: i}
}

// WithStdout returns an option to set the standard output for the interpreter.
func WithStdout(w io.Writer) interpreter.InterpreterOption {
	return interpreter.WithStdout(w)
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
	// =========================================================================
	fmt.Printf(">>>> [DEBUG] api.Interpreter.Run: Calling internal interpreter for '%s'\n", procName)
	// =========================================================================
	result, err := i.internal.Run(procName, args...)
	// =========================================================================
	fmt.Printf(">>>> [DEBUG] api.Interpreter.Run: Value RECEIVED from internal interpreter is: %#v\n", result)
	// =========================================================================
	return result, err
}

// EmitEvent sends an event into an event-sink script.
func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.internal.EmitEvent(eventName, source, payload)
}

// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: CRITICAL: Removed all public methods that exposed direct access to shared stores (Accounts, Models, Capsules, Tools), forcing consumers to use the ax/contract model.
// filename: pkg/api/interpreter_methods.go
// nlines: 95
// risk_rating: HIGH

package api

import (
	"context"
	"io"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// SetRuntime allows the host application to set a custom runtime context.
func (i *Interpreter) SetRuntime(rt Runtime) {
	i.internal.SetRuntime(rt)
	i.runtime = rt
}

// SetTurnContext sets the context for the current AEIOU turn.
func (i *Interpreter) SetTurnContext(ctx context.Context) {
	i.internal.SetTurnContext(ctx)
}

// SetSandboxDir sets the secure root directory for file operations.
func (i *Interpreter) SetSandboxDir(path string) {
	i.internal.SetSandboxDir(path)
}

// SetStdout sets the standard output writer.
func (i *Interpreter) SetStdout(w io.Writer) {
	i.internal.SetStdout(w)
}

// SetStderr sets the standard error writer.
func (i *Interpreter) SetStderr(w io.Writer) {
	i.internal.SetStderr(w)
}

// SetEmitFunc sets a custom handler for the 'emit' statement.
func (i *Interpreter) SetEmitFunc(f func(Value)) {
	i.internal.SetEmitFunc(func(v lang.Value) {
		f(v)
	})
}

// RegisterProvider registers a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p AIProvider) {
	i.internal.RegisterProvider(name, p)
}

// Load injects a parsed program into the interpreter.
func (i *Interpreter) Load(tree *interfaces.Tree) error {
	return i.internal.Load(tree)
}

// AppendScript merges procedures and event handlers from a new AST.
func (i *Interpreter) AppendScript(tree *interfaces.Tree) error {
	return i.internal.AppendScript(tree)
}

// Execute runs the top-level 'command' blocks.
func (i *Interpreter) Execute() (Value, error) {
	return i.internal.ExecuteCommands()
}

// ExecuteCommands runs any unnamed 'command' blocks.
func (i *Interpreter) ExecuteCommands() (Value, error) {
	t, err := i.internal.ExecuteCommands()
	return t, err
}

// Run calls a specific, named procedure.
func (i *Interpreter) Run(procName string, args ...lang.Value) (Value, error) {
	result, err := i.internal.Run(procName, args...)
	return result, err
}

// EmitEvent sends an event into the interpreter.
func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.internal.EmitEvent(eventName, source, payload)
}

// Clone creates a new interpreter that inherits the parent's full state,
// including procedures and variables.
func (i *Interpreter) Clone() *Interpreter {
	clonedInternal := i.internal.Clone()
	newFacade := &Interpreter{
		internal: clonedInternal,
	}

	if hr, ok := i.runtime.(*hostRuntime); ok {
		newFacade.runtime = &hostRuntime{
			Runtime: clonedInternal,
			id:      hr.id,
		}
	} else {
		newFacade.runtime = clonedInternal
	}

	clonedInternal.SetRuntime(newFacade.runtime)
	return newFacade
}

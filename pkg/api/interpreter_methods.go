// NeuroScript Version: 0.7.4
// File version: 8
// Purpose: FIX: Added extensive debug logging to Clone and SetRuntime to trace identity propagation.
// filename: pkg/api/interpreter_methods.go
// nlines: 162
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Accounts returns the account store associated with the interpreter.
func (i *Interpreter) Accounts() *AccountStore {
	return i.internal.AccountStore()
}

// AgentModels returns the agent model store associated with the interpreter.
func (i *Interpreter) AgentModels() *AgentModelStore {
	return i.internal.AgentModelStore()
}

// SetRuntime allows the host application to set a custom runtime context.
func (i *Interpreter) SetRuntime(rt Runtime) {
	fmt.Printf("[API SetRuntime] Setting runtime on facade %p. Type: %T\n", i, rt)
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

// RegisterAgentModel registers an AgentModel configuration.
func (i *Interpreter) RegisterAgentModel(name string, config map[string]any) error {
	return i.internal.AgentModelsAdmin().Register(types.AgentModelName(name), config)
}

// CapsuleStore returns the interpreter's layered capsule store for read-only operations.
func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.internal.CapsuleStore()
}

// CapsuleRegistryForAdmin returns the administrative capsule registry.
func (i *Interpreter) CapsuleRegistryForAdmin() *AdminCapsuleRegistry {
	return i.internal.CapsuleRegistryForAdmin()
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
	return i.internal.ExecuteCommands()
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

// ToolRegistry returns the tool registry associated with the interpreter.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.internal.ToolRegistry()
}

// Clone creates a new interpreter that inherits the parent's full state,
// including procedures and variables.
func (i *Interpreter) Clone() *Interpreter {
	fmt.Printf("[API CLONE START] Parent facade %p runtime is type: %T\n", i, i.runtime)
	clonedInternal := i.internal.Clone()
	newFacade := &Interpreter{
		internal: clonedInternal,
	}

	if hr, ok := i.runtime.(*hostRuntime); ok {
		fmt.Printf("[API CLONE hostRuntime] Parent has hostRuntime. ID: %s. Wrapping new internal clone.\n", hr.id.DID())
		newFacade.runtime = &hostRuntime{
			Runtime: clonedInternal,
			id:      hr.id,
		}
	} else {
		fmt.Printf("[API CLONE non-hostRuntime] Parent has non-host runtime.\n")
		newFacade.runtime = clonedInternal
	}

	clonedInternal.SetRuntime(newFacade.runtime)
	fmt.Printf("[API CLONE END] Cloned facade %p runtime is now: %T\n", newFacade, newFacade.runtime)
	return newFacade
}

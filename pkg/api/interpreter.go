// NeuroScript Version: 0.8.0
// File version: 52
// Purpose: Corrected to use public accessors for internal components and added missing import.
// filename: pkg/api/interpreter.go
// nlines: 130
// risk_rating: HIGH

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces" // <-- Added missing import
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
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
// All host dependencies (I/O, logging, etc.) must be provided via Option funcs.
func New(opts ...Option) *Interpreter {
	i := interpreter.NewInterpreter(opts...)

	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	return &Interpreter{internal: i}
}

// HasEmitFunc returns true if a custom emit handler has been set on the interpreter.
func (i *Interpreter) HasEmitFunc() bool {
	return i.internal.HasEmitFunc()
}

// RegisterProvider allows the host application to register a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p AIProvider) {
	i.internal.RegisterProvider(name, p)
}

// RegisterAgentModel allows the host application to register an AgentModel configuration.
// It accepts a map of native Go types.
func (i *Interpreter) RegisterAgentModel(name string, config map[string]any) error {
	return i.internal.AgentModelsAdmin().Register(types.AgentModelName(name), config)
}

// CapsuleStore returns the interpreter's layered capsule store for read-only operations.
func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.internal.CapsuleStore()
}

// CapsuleRegistryForAdmin returns the interpreter's administrative capsule registry.
func (i *Interpreter) CapsuleRegistryForAdmin() *AdminCapsuleRegistry {
	return i.internal.CapsuleRegistryForAdmin()
}

// Load injects a parsed program into the interpreter via the interface.
func (i *Interpreter) Load(tree *interfaces.Tree) error {
	return i.internal.Load(tree)
}

// AppendScript merges procedures and event handlers from a new program AST
// into the interpreter's existing state without clearing previous definitions.
func (i *Interpreter) AppendScript(tree *interfaces.Tree) error {
	return i.internal.AppendScript(tree)
}

// Execute runs any unnamed 'command' blocks found in the loaded program.
func (i *Interpreter) Execute(program *interfaces.Tree) (Value, error) {
	return i.internal.Execute(program.Root.(*Program))
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

// ToolRegistry returns the tool registry associated with the interpreter.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.internal.ToolRegistry()
}

// HostContext returns the interpreter's host context.
// FIX: Correctly calls the public HostContext() method on the internal instance.
func (i *Interpreter) HostContext() *HostContext {
	return i.internal.HostContext()
}

// Handles returns the interpreter's handle manager.
// FIX: Correctly calls the public Handles() method on the internal instance.
func (i *Interpreter) Handles() interfaces.HandleManager {
	return i.internal.Handles()
}

// Unwrap converts a NeuroScript api.Value back into a standard Go `any` type.
func Unwrap(v Value) (any, error) {
	if val, ok := v.(lang.Value); ok {
		return lang.Unwrap(val), nil
	}
	return v, nil
}

// ParseLoopControl is deprecated and will be removed.
func ParseLoopControl(output string) (*LoopController, error) {
	return nil, errors.New("ParseLoopControl is deprecated; use the AEIOU v3 LoopController")
}

// GetVariable retrieves a variable from the interpreter's current state.
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	val, exists := i.internal.GetVariable(name)
	return val, exists
}

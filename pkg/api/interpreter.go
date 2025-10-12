// NeuroScript Version: 0.8.0
// File version: 48
// Purpose: Adds debug output to the New() function to trace the state of the ExecPolicy during interpreter creation.
// filename: pkg/api/interpreter.go
// nlines: 121
// risk_rating: HIGH

package api

import (
	"errors"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/capsule"
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
// All host dependencies (I/O, logging, etc.) must be provided via Option funcs.
func New(opts ...Option) *Interpreter {
	// --- DEBUG ---
	fmt.Fprintf(os.Stderr, "[DEBUG] api.New: Called with %d options.\n", len(opts))
	// --- END DEBUG ---

	i := interpreter.NewInterpreter(opts...)

	// --- DEBUG ---
	if i.ExecPolicy != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] api.New: After NewInterpreter(), ExecPolicy is PRESENT. Context: %v, Allow count: %d\n", i.ExecPolicy.Context, len(i.ExecPolicy.Allow))
		if len(i.ExecPolicy.Allow) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] api.New: Allowed tools include: %v\n", i.ExecPolicy.Allow[0])
		}
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] api.New: After NewInterpreter(), ExecPolicy is NIL.\n")
	}
	// --- END DEBUG ---

	if i.ExecPolicy == nil {
		// --- DEBUG ---
		fmt.Fprintf(os.Stderr, "[DEBUG] api.New: ExecPolicy was nil, creating a new default policy.\n")
		// --- END DEBUG ---
		i.ExecPolicy = &policy.ExecPolicy{
			Context: policy.ContextNormal,
			Allow:   []string{},
		}
	}

	// TODO: Provider registration should likely be an Option as well.
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

// CapsuleRegistryForAdmin returns the interpreter's administrative capsule registry,
// which is required for privileged tools that add or modify capsules.
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

// GetVariable retrieves a variable from the interpreter's current state.
// It returns the value and a boolean indicating if the variable was found.
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	val, exists := i.internal.GetVariable(name)
	return val, exists
}

// NeuroScript Version: 0.7.2
// File version: 42
// Purpose: Removes With... option helpers, which are now consolidated in interpreter_options.go.
// filename: pkg/api/interpreter.go
// nlines: 158
// risk_rating: HIGH
package api

import (
	"errors"
	"io"

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
func New(opts ...Option) *Interpreter {
	// --- DEBUG ---
	//fmt.Printf("[DEBUG] api.New() called with %d options.\n", len(opts))
	i := interpreter.NewInterpreter(opts...)
	if i.ExecPolicy == nil {
		i.ExecPolicy = &policy.ExecPolicy{
			Context: policy.ContextNormal,
			Allow:   []string{},
		}
	}
	// The internal interpreter is now expected to initialize its own capsule store.
	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	// --- DEBUG ---
	// if i.CapsuleRegistryForAdmin() != nil {
	// 	fmt.Println("[DEBUG] api.New(): Admin registry is PRESENT after initialization.")
	// } else {
	// 	fmt.Println("[DEBUG] api.New(): Admin registry is NIL after initialization.")
	// }
	return &Interpreter{internal: i}
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

// HasEmitFunc returns true if a custom emit handler has been set on the interpreter.
func (i *Interpreter) HasEmitFunc() bool {
	return i.internal.HasEmitFunc()
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

// CapsuleStore returns the interpreter's layered capsule store for read-only operations.
func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.internal.CapsuleStore()
}

// CapsuleRegistryForAdmin returns the interpreter's administrative capsule registry,
// which is required for privileged tools that add or modify capsules.
func (i *Interpreter) CapsuleRegistryForAdmin() *AdminCapsuleRegistry {
	// --- DEBUG ---
	reg := i.internal.CapsuleRegistryForAdmin()
	// if reg != nil {
	// 	fmt.Println("[DEBUG] api.Interpreter.CapsuleRegistryForAdmin(): Returning PRESENT admin registry.")
	// } else {
	// 	fmt.Println("[DEBUG] api.Interpreter.CapsuleRegistryForAdmin(): Returning NIL admin registry.")
	// }
	return reg
}

// Load injects a parsed program into the interpreter via the interface.
func (i *Interpreter) Load(tree *interfaces.Tree) error {
	return i.internal.Load(tree)
}

// AppendScript merges procedures and event handlers from a new program AST
// into the interpreter's existing state without clearing previous definitions.
// It returns an error if a procedure being added already exists.
func (i *Interpreter) AppendScript(tree *interfaces.Tree) error {
	return i.internal.AppendScript(tree)
}

// Execute runs the top-level 'command' blocks that have been loaded into the
// interpreter's state via Load() or AppendScript(). This is the correct method
// for executing the entry point of a multi-file program.
func (i *Interpreter) Execute() (Value, error) {
	return i.internal.ExecuteCommands()
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

// GetVariable retrieves a variable from the interpreter's current state.
// It returns the value and a boolean indicating if the variable was found.
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	val, exists := i.internal.GetVariable(name)
	return val, exists
}

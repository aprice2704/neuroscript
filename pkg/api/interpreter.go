// NeuroScript Version: 0.8.0
// File version: 58
// Purpose: Adds a public SetTurnContext method to the wrapper, allowing hosts to set ephemeral context for an execution.
// filename: pkg/api/interpreter.go
// nlines: 132
// risk_rating: HIGH

package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider/google"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Interpreter is a facade over the internal interpreter. It embeds the internal
// interpreter to provide access to its methods, and adds identity-awareness
// by implementing the ActorProvider interface.
type Interpreter struct {
	*interpreter.Interpreter
}

// Statically assert that *Interpreter satisfies the tool.Runtime and ActorProvider interfaces.
var _ tool.Runtime = (*Interpreter)(nil)
var _ interfaces.ActorProvider = (*Interpreter)(nil)

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...Option) *Interpreter {
	fmt.Println("[DEBUG] api.New: Starting interpreter creation.")

	// 1. Create the internal interpreter.
	internalInterp := interpreter.NewInterpreter(opts...)
	fmt.Printf("[DEBUG] api.New: Internal interpreter created with ID: %s\n", internalInterp.ID())

	// 2. Create the public API wrapper by embedding the internal one.
	publicInterp := &Interpreter{Interpreter: internalInterp}
	fmt.Printf("[DEBUG] api.New: Public API wrapper created for internal interpreter %s.\n", internalInterp.ID())

	// 3. Create the tool registry, passing the PUBLIC wrapper as the runtime.
	registry := tool.NewToolRegistry(publicInterp)
	fmt.Println("[DEBUG] api.New: Tool registry created with public interpreter as context.")

	// 4. Inject the correctly-contextualized registry back into the internal interpreter.
	internalInterp.SetToolRegistry(registry)
	fmt.Println("[DEBUG] api.New: Tool registry injected into internal interpreter.")

	// 5. Now, register standard tools on the new registry.
	internalInterp.RegisterStandardTools()
	fmt.Println("[DEBUG] api.New: Standard tools registered.")

	googleProvider := google.New()
	internalInterp.RegisterProvider("google", googleProvider)
	fmt.Println("[DEBUG] api.New: Google provider registered. Initialization complete.")

	return publicInterp
}

// Actor returns the actor identity associated with the interpreter's HostContext.
// This makes the public Interpreter identity-aware.
func (i *Interpreter) Actor() (interfaces.Actor, bool) {
	if i.HostContext() == nil || i.HostContext().Actor == nil {
		return nil, false
	}
	actor := i.HostContext().Actor
	return actor, true
}

// SetTurnContext allows the host to set the ephemeral context for a single execution.
// This is the public wrapper for the now-exported internal method.
func (i *Interpreter) SetTurnContext(ctx context.Context) {
	i.Interpreter.SetTurnContext(ctx)
}

// RegisterAgentModel overrides the embedded method to provide the correct public
// API signature, accepting a map of native Go types.
func (i *Interpreter) RegisterAgentModel(name string, config map[string]any) error {
	return i.AgentModelsAdmin().Register(types.AgentModelName(name), config)
}

// GetVar retrieves a variable from the interpreter's current state.
// This overrides the embedded GetVariable to provide the public API's signature.
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	val, exists := i.Interpreter.GetVariable(name)
	return val, exists
}

// Run calls a specific, named procedure from the loaded program.
// This overrides the embedded Run to provide the public API's signature.
func (i *Interpreter) Run(procName string, args ...lang.Value) (Value, error) {
	result, err := i.Interpreter.Run(procName, args...)
	return result, err
}

// Execute runs any unnamed 'command' blocks found in the loaded program.
// This overrides the embedded Execute to provide the public API's signature.
func (i *Interpreter) Execute(program *interfaces.Tree) (Value, error) {
	return i.Interpreter.Execute(program.Root.(*Program))
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

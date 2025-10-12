// NeuroScript Version: 0.8.0
// File version: 76
// Purpose: Reinstated the aiWorker field to the struct to hold the root LLM client.
// filename: pkg/interpreter/interpreter.go
// nlines: 200
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/uuid"
)

// DefaultSelfHandle is the internal handle for the default whisper buffer.
const DefaultSelfHandle = "default_self_buffer"

// Interpreter holds the state for a NeuroScript runtime environment.
type Interpreter struct {
	id                   string
	hostContext          *HostContext
	state                *interpreterState
	tools                tool.ToolRegistry
	eventManager         *EventManager
	aiWorker             interfaces.LLMClient // The root LLM client.
	shouldExit           bool
	exitCode             int
	returnValue          lang.Value
	lastCallResult       lang.Value
	maxLoopIterations    int
	bufferManager        *BufferManager
	objectCache          map[string]interface{}
	objectCacheMu        sync.Mutex
	skipStdTools         bool
	modelStore           *agentmodel.AgentModelStore
	ExecPolicy           *policy.ExecPolicy
	root                 *Interpreter
	turnCtx              context.Context
	transientPrivateKey  ed25519.PrivateKey
	accountStore         *account.Store
	capsuleStore         *capsule.Store
	adminCapsuleRegistry *capsule.Registry

	cloneRegistry   []*Interpreter
	cloneRegistryMu sync.Mutex
}

// SetAccountStore replaces the interpreter's default account store with a host-provided one.
func (i *Interpreter) SetAccountStore(store *account.Store) {
	i.rootInterpreter().accountStore = store
}

// SetAgentModelStore replaces the interpreter's default agent model store.
func (i *Interpreter) SetAgentModelStore(store *agentmodel.AgentModelStore) {
	i.rootInterpreter().modelStore = store
}

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		id:                fmt.Sprintf("interp-%s", uuid.NewString()[:8]),
		state:             newInterpreterState(),
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
		bufferManager:     NewBufferManager(),
		objectCache:       make(map[string]interface{}),
		turnCtx:           context.Background(),
		capsuleStore:      capsule.NewStore(capsule.DefaultRegistry()),
		cloneRegistry:     make([]*Interpreter, 0),
	}
	i.tools = tool.NewToolRegistry(i)
	i.root = i // A root's root is itself.
	i.modelStore = agentmodel.NewAgentModelStore()
	i.accountStore = account.NewStore()

	for _, opt := range opts {
		opt(i)
	}

	if i.hostContext == nil {
		panic("FATAL: NewInterpreter called without WithHostContext. A HostContext is mandatory.")
	}
	if i.hostContext.Logger == nil {
		panic("FATAL: HostContext.Logger cannot be nil.")
	}
	if i.hostContext.WhisperFunc == nil {
		i.hostContext.WhisperFunc = i.defaultWhisperFunc
	}
	i.bufferManager.Create(DefaultSelfHandle)

	if !i.skipStdTools {
		if err := tool.RegisterGlobalToolsets(i.tools); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register global toolsets: %v", err))
		}
	}
	if err := registerDebugTools(i.tools); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register debug tools: %v", err))
	}
	_, transientPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to generate transient private key for AEIOU tool: %v", err))
	}
	i.transientPrivateKey = transientPrivateKey
	primaryMinter, err := aeiou.NewMagicMinter(i.transientPrivateKey)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to create AEIOU primary minter: %v", err))
	}
	magicTool := aeiou.NewMagicTool(primaryMinter, nil)
	if err := registerAeiouTools(i.tools, magicTool); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register AEIOU tools: %v", err))
	}
	i.SetInitialVariable("self", lang.StringValue{Value: DefaultSelfHandle})
	return i
}

// CapsuleStore returns the interpreter's layered capsule store.
func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.capsuleStore
}

// RunProcedure is the public entry point for running a procedure.
func (i *Interpreter) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	result, err := i.runProcedure(procName, args...)
	if err == nil {
		i.lastCallResult = result
	}
	return result, err
}

func (i *Interpreter) SetInitialVariable(name string, value any) error {
	wrappedValue, err := lang.Wrap(value)
	if err != nil {
		return fmt.Errorf("failed to wrap initial variable '%s': %w", name, err)
	}
	i.state.setVariable(name, wrappedValue)
	i.state.globalVarNames[name] = true
	return nil
}

func (i *Interpreter) Load(tree *interfaces.Tree) error {
	if tree == nil || tree.Root == nil {
		i.Logger().Warn("Load called with a nil program AST.")
		i.state.knownProcedures = make(map[string]*ast.Procedure)
		i.eventManager.eventHandlers = make(map[string][]*ast.OnEventDecl)
		i.state.commands = []*ast.CommandNode{}
		return nil
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("interpreter.Load: expected root node of type *ast.Program, but got %T", tree.Root)
	}
	i.state.knownProcedures = make(map[string]*ast.Procedure)
	i.eventManager.eventHandlers = make(map[string][]*ast.OnEventDecl)
	i.state.commands = []*ast.CommandNode{}
	for name, proc := range program.Procedures {
		i.state.knownProcedures[name] = proc
	}
	for _, eventDecl := range program.Events {
		if err := i.eventManager.register(eventDecl, i); err != nil {
			return fmt.Errorf("failed to register event handler: %w", err)
		}
	}
	if program.Commands != nil {
		i.state.commands = program.Commands
	}
	return nil
}

// SetSandboxDir sets the secure root directory for file operations.
func (i *Interpreter) SetSandboxDir(path string) {
	i.state.sandboxDir = path
}

// GetGrantSet returns the currently active capability grant set for policy enforcement.
func (i *Interpreter) GetGrantSet() *capability.GrantSet {
	if i.ExecPolicy == nil {
		return &capability.GrantSet{}
	}
	return &i.ExecPolicy.Grants
}

func (i *Interpreter) rootInterpreter() *Interpreter {
	root := i
	for root.root != root { // Correct loop condition for finding the root
		root = root.root
	}
	return root
}

func (i *Interpreter) Accounts() interfaces.AccountReader {
	return account.NewReader(i.rootInterpreter().accountStore)
}

func (i *Interpreter) AccountsAdmin() interfaces.AccountAdmin {
	return account.NewAdmin(i.rootInterpreter().accountStore, i.ExecPolicy)
}

func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	return agentmodel.NewAgentModelReader(i.rootInterpreter().modelStore)
}

func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	return agentmodel.NewAgentModelAdmin(i.rootInterpreter().modelStore, i.ExecPolicy)
}

func (i *Interpreter) CapsuleRegistryForAdmin() *capsule.Registry {
	return i.rootInterpreter().adminCapsuleRegistry
}

// GetExecPolicy satisfies the policygate.Runtime interface.
func (i *Interpreter) GetExecPolicy() *policy.ExecPolicy {
	return i.ExecPolicy
}

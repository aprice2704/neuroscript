// NeuroScript Version: 0.8.0
// File version: 77
// Purpose: Removes the obsolete logger and ExecPolicy fields, fully committing to the RunnerParcel as the source of truth.
// filename: pkg/interpreter/interpreter.go
// nlines: 163
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/uuid"
)

// DefaultSelfHandle is the internal handle for the default whisper buffer.
const DefaultSelfHandle = "default_self_buffer"

// Interpreter holds the state for a NeuroScript runtime environment.
type Interpreter struct {
	id                        string // Unique ID for this interpreter instance
	fileAPI                   interfaces.FileAPI
	state                     *interpreterState
	tools                     tool.ToolRegistry
	runtime                   tool.Runtime // The runtime context passed to tools.
	eventManager              *EventManager
	evaluate                  *evaluation
	aiWorker                  interfaces.LLMClient
	shouldExit                bool
	exitCode                  int
	returnValue               lang.Value
	lastCallResult            lang.Value
	stdout                    io.Writer
	stdin                     io.Reader
	stderr                    io.Writer
	maxLoopIterations         int
	bufferManager             *BufferManager
	objectCache               map[string]interface{}
	objectCacheMu             sync.Mutex
	llmclient                 interfaces.LLMClient
	skipStdTools              bool
	modelStore                *agentmodel.AgentModelStore
	root                      *Interpreter
	customEmitFunc            func(lang.Value)
	customWhisperFunc         func(handle, data lang.Value)
	turnCtx                   context.Context
	aiTranscript              io.Writer
	transientPrivateKey       ed25519.PrivateKey
	accountStore              *account.Store
	capsuleStore              *capsule.Store
	eventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)
	emitter                   interfaces.Emitter // The LLM telemetry emitter.

	adminCapsuleRegistry *capsule.Registry // The writable registry for config scripts.

	// --- Runner Parcel ---
	parcel contract.RunnerParcel

	// --- Clone Tracking for Debugging ---
	cloneRegistry   []*Interpreter
	cloneRegistryMu sync.Mutex
}

// Compile-time check to ensure Interpreter satisfies ParcelProvider.
var _ contract.ParcelProvider = (*Interpreter)(nil)

func (i *Interpreter) GetParcel() contract.RunnerParcel  { return i.parcel }
func (i *Interpreter) SetParcel(p contract.RunnerParcel) { i.parcel = p }

// AccountStore returns the account store associated with the interpreter.
func (i *Interpreter) AccountStore() *account.Store {
	return i.rootInterpreter().accountStore
}

// AgentModelStore returns the agent model store associated with the interpreter.
func (i *Interpreter) AgentModelStore() *agentmodel.AgentModelStore {
	return i.rootInterpreter().modelStore
}

// SetEmitter sets the LLM telemetry emitter for the interpreter.
func (i *Interpreter) SetEmitter(e interfaces.Emitter) {
	i.emitter = e
}

// SetAITranscript sets the writer for logging AI prompts.
func (i *Interpreter) SetAITranscript(w io.Writer) {
	i.aiTranscript = w
}

// SetAccountStore replaces the interpreter's default account store with a host-provided one.
func (i *Interpreter) SetAccountStore(store *account.Store) {
	if i.root != nil {
		i.root.SetAccountStore(store)
		return
	}
	i.accountStore = store
}

// SetAgentModelStore replaces the interpreter's default agent model store with a host-provided one.
func (i *Interpreter) SetAgentModelStore(store *agentmodel.AgentModelStore) {
	if i.root != nil {
		i.root.SetAgentModelStore(store)
		return
	}
	i.modelStore = store
}

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		id:                fmt.Sprintf("interp-%s", uuid.NewString()[:8]), // Assign a unique ID
		state:             newInterpreterState(),
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
		stdout:            os.Stdout,
		stdin:             os.Stdin,
		stderr:            os.Stderr,
		bufferManager:     NewBufferManager(),
		objectCache:       make(map[string]interface{}),
		turnCtx:           context.Background(),
		capsuleStore:      capsule.NewStore(capsule.DefaultRegistry()),
		cloneRegistry:     make([]*Interpreter, 0), // Initialize the registry
	}
	i.evaluate = &evaluation{i: i}
	i.tools = tool.NewToolRegistry(i)
	i.runtime = i // By default, the interpreter is its own runtime.
	i.root = nil  // This is the root interpreter
	i.modelStore = agentmodel.NewAgentModelStore()
	i.accountStore = account.NewStore()

	i.bufferManager.Create(DefaultSelfHandle)
	i.customWhisperFunc = i.defaultWhisperFunc

	// Apply options before parcel creation to get logger and policy
	for _, opt := range opts {
		opt(i)
	}

	// Create the initial parcel if one wasn't provided via options
	if i.parcel == nil {
		i.parcel = contract.NewParcel(nil, nil, logging.NewNoOpLogger(), nil)
	}

	if !i.skipStdTools {
		if err := tool.RegisterGlobalToolsets(i.tools); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register global toolsets: %v", err))
		}
	}

	// Register debug tools.
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

func (i *Interpreter) EvaluateExpression(node ast.Expression) (lang.Value, error) {
	return i.evaluate.Expression(node)
}

func (i *Interpreter) Run(procName string, args ...lang.Value) (lang.Value, error) {
	result, err := i.RunProcedure(procName, args...)
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
	// Globals are now managed by the parcel.
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
	if i.parcel == nil || i.parcel.Policy() == nil {
		return &capability.GrantSet{}
	}
	return &i.parcel.Policy().Grants
}

func (i *Interpreter) rootInterpreter() *Interpreter {
	root := i
	for root.root != nil {
		root = root.root
	}
	return root
}

func (i *Interpreter) Accounts() interfaces.AccountReader {
	root := i.rootInterpreter()
	return account.NewReader(root.accountStore)
}

func (i *Interpreter) AccountsAdmin() interfaces.AccountAdmin {
	root := i.rootInterpreter()
	var execPolicy *interfaces.ExecPolicy
	if i.parcel != nil {
		execPolicy = i.parcel.Policy()
	}
	return account.NewAdmin(root.accountStore, execPolicy)
}

func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	root := i.rootInterpreter()
	return agentmodel.NewAgentModelReader(root.modelStore)
}

func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	root := i.rootInterpreter()
	var execPolicy *interfaces.ExecPolicy
	if i.parcel != nil {
		execPolicy = i.parcel.Policy()
	}
	return agentmodel.NewAgentModelAdmin(root.modelStore, execPolicy)
}

// CapsuleRegistryForAdmin returns the interpreter's administrative capsule registry.
// This is intended for use by privileged tools in a configuration context.
func (i *Interpreter) CapsuleRegistryForAdmin() *capsule.Registry {
	root := i.rootInterpreter()
	return root.adminCapsuleRegistry
}

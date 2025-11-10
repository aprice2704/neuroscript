// NeuroScript Version: 0.8.0
// File version: 95
// Purpose: Core Interpreter struct and New/Set methods. Split from original file.
// filename: pkg/interpreter/interpreter.go
// nlines: 219
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"

	// "github.com/aprice2704/neuroscript/pkg/ast" // No longer needed here
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/uuid"
)

// Statically assert that the concrete Interpreter type satisfies the tool.Runtime interface.
var _ tool.Runtime = (*Interpreter)(nil)

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
	providerRegistry     *provider.Registry
	ExecPolicy           *policy.ExecPolicy
	root                 *Interpreter
	turnCtx              context.Context
	transientPrivateKey  ed25519.PrivateKey
	accountStore         *account.Store
	capsuleStore         *capsule.Store
	adminCapsuleRegistry *capsule.Registry
	capsuleProvider      interfaces.CapsuleProvider // ADDED
	parser               *parser.ParserAPI
	astBuilder           *parser.ASTBuilder

	cloneRegistry   []*Interpreter
	cloneRegistryMu sync.Mutex

	PublicAPI tool.Runtime
}

// ID returns the unique identifier for this interpreter instance.
func (i *Interpreter) ID() string {
	return i.id
}

// Parser returns the interpreter's configured parser instance.
func (i *Interpreter) Parser() *parser.ParserAPI {
	return i.parser
}

// ASTBuilder returns the interpreter's configured AST builder instance.
func (i *Interpreter) ASTBuilder() *parser.ASTBuilder {
	return i.astBuilder
}

// SetToolRegistry allows the public API wrapper to replace the tool registry.
func (i *Interpreter) SetToolRegistry(r tool.ToolRegistry) {
	i.tools = r
}

// SetPublicAPI allows the public API wrapper to set a pointer to itself.
// This is used by internal components (like the 'ask' hook) to ensure
// they pass the public, wrapped interpreter to external services.
func (i *Interpreter) SetPublicAPI(publicAPI tool.Runtime) {
	i.PublicAPI = publicAPI
}

// SetAccountStore replaces the interpreter's default account store with a host-provided one.
func (i *Interpreter) SetAccountStore(store *account.Store) {
	i.rootInterpreter().accountStore = store
}

// SetAgentModelStore replaces the interpreter's default agent model store.
func (i *Interpreter) SetAgentModelStore(store *agentmodel.AgentModelStore) {
	i.rootInterpreter().modelStore = store
}

// SetProviderRegistry replaces the interpreter's default provider registry.
func (i *Interpreter) SetProviderRegistry(registry *provider.Registry) {
	i.rootInterpreter().providerRegistry = registry
}

// SetCapsuleProvider replaces the interpreter's default capsule logic with a host-provided one.
func (i *Interpreter) SetCapsuleProvider(provider interfaces.CapsuleProvider) {
	i.rootInterpreter().capsuleProvider = provider
}

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		id:                fmt.Sprintf("interp-%s", uuid.NewString()[:8]),
		state:             newInterpreterState(), // This now initializes globalConstants
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
		bufferManager:     NewBufferManager(),
		objectCache:       make(map[string]interface{}),
		turnCtx:           context.Background(),
		capsuleStore:      capsule.NewStore(capsule.DefaultRegistry()),
		cloneRegistry:     make([]*Interpreter, 0),
	}
	// Note: globalConstants map is initialized inside newInterpreterState() in state.go

	i.tools = tool.NewToolRegistry(i)

	i.root = i // A root's root is itself.
	i.modelStore = agentmodel.NewAgentModelStore()
	i.accountStore = account.NewStore()
	i.providerRegistry = provider.NewRegistry()

	for _, opt := range opts {
		opt(i)
	}

	if i.hostContext == nil {
		panic("FATAL: NewInterpreter called without WithHostContext. A HostContext is mandatory.")
	}
	if i.hostContext.Logger == nil {
		panic("FATAL: HostContext.Logger cannot be nil.")
	}
	if i.parser == nil {
		i.parser = parser.NewParserAPI(i.hostContext.Logger)
	}
	if i.astBuilder == nil {
		i.astBuilder = parser.NewASTBuilder(i.hostContext.Logger)
	}
	if i.ExecPolicy == nil {
		i.ExecPolicy = policy.NewBuilder(policy.ContextNormal).Build()
	}

	i.astBuilder.SetEventHandlerCallback(i.RegisterEventHandler)

	if i.hostContext.WhisperFunc == nil {
		i.hostContext.WhisperFunc = i.defaultWhisperFunc
	}
	i.bufferManager.Create(DefaultSelfHandle)

	i.RegisterStandardTools() // This is now in interpreter_tools.go

	i.SetInitialVariable("self", lang.StringValue{Value: DefaultSelfHandle})
	return i
}

// CapsuleStore returns the interpreter's layered capsule store.
func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.capsuleStore
}

// CapsuleProvider returns the host-provided capsule service, if one was injected.
func (i *Interpreter) CapsuleProvider() interfaces.CapsuleProvider {
	return i.rootInterpreter().capsuleProvider
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
	i.state.setGlobalVariable(name, wrappedValue)
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
	for root.root != root {
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

// Actor returns the actor identity associated with the interpreter's HostContext.
// This method makes the internal interpreter satisfy the interfaces.ActorProvider interface.
func (i *Interpreter) Actor() (interfaces.Actor, bool) {
	if i.hostContext == nil || i.hostContext.Actor == nil {
		return nil, false
	}
	return i.hostContext.Actor, true
}

// ForkSandboxed creates a new, sandboxed interpreter for executing
// an 'ask' loop action, as required by api.ExecuteSandboxedAST.
// It calls the unexported fork().
func (i *Interpreter) ForkSandboxed() (*Interpreter, error) {
	// This must be called on the root interpreter.
	root := i.rootInterpreter()

	// FIX: Call the internal fork() with no arguments, as per clone.go.
	clone := root.fork()

	// FIX: Return the clone and a nil error to match the API signature
	// and fix the 'not enough return values' error.
	return clone, nil
}

// SetHostContext allows replacing the interpreter's HostContext.
// This is used by sandboxing APIs like ExecuteSandboxedAST to
// inject an I/O-capturing context onto a fork.
// This FIXES the 'SetHostContext undefined' error.
func (i *Interpreter) SetHostContext(hc *HostContext) {
	i.hostContext = hc
}

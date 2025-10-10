// NeuroScript Version: 0.8.0
// File version: 85
// Purpose: FIX: Moves SetEmitter and SetAITranscript to interpreter_api.go to consolidate the public API.
// filename: pkg/interpreter/interpreter.go
// nlines: 162
// risk_rating: MEDIUM

package interpreter

import (
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/ax"
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
	runtime                   tool.Runtime // The runtime context passed to tools.
	evaluate                  *evaluation
	aiWorker                  interfaces.LLMClient
	shouldExit                bool
	exitCode                  int
	returnValue               lang.Value
	lastCallResult            lang.Value
	stdout                    io.Writer
	stdin                     io.Reader
	stderr                    io.Writer
	bufferManager             *BufferManager
	objectCache               map[string]interface{}
	objectCacheMu             sync.Mutex
	llmclient                 interfaces.LLMClient
	root                      *Interpreter
	customEmitFunc            func(lang.Value)
	customWhisperFunc         func(handle, data lang.Value)
	aiTranscript              io.Writer
	transientPrivateKey       ed25519.PrivateKey
	eventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)
	emitter                   interfaces.Emitter // The LLM telemetry emitter.
	eventManager              *EventManager
	skipStdTools              bool

	// --- AX Contracts ---
	parcel   contract.RunnerParcel
	catalogs contract.SharedCatalogs

	// --- Clone Tracking for Debugging ---
	cloneRegistry   []*Interpreter
	cloneRegistryMu sync.Mutex
}

// Compile-time check to ensure Interpreter satisfies ParcelProvider.
var _ contract.ParcelProvider = (*Interpreter)(nil)
var _ tool.Runtime = (*Interpreter)(nil)

func (i *Interpreter) GetParcel() contract.RunnerParcel  { return i.parcel }
func (i *Interpreter) SetParcel(p contract.RunnerParcel) { i.parcel = p }
func (i *Interpreter) Identity() ax.ID {
	if i.parcel != nil {
		return i.parcel.Identity()
	}
	return nil
}

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		id:            fmt.Sprintf("interp-%s", uuid.NewString()[:8]), // Assign a unique ID
		state:         newInterpreterState(),
		stdout:        os.Stdout,
		stdin:         os.Stdin,
		stderr:        os.Stderr,
		bufferManager: NewBufferManager(),
		objectCache:   make(map[string]interface{}),
		cloneRegistry: make([]*Interpreter, 0),
		eventManager:  newEventManager(),
	}
	i.evaluate = &evaluation{i: i}
	i.runtime = i // By default, the interpreter is its own runtime.
	i.root = nil  // This is the root interpreter

	i.bufferManager.Create(DefaultSelfHandle)
	i.customWhisperFunc = i.defaultWhisperFunc

	// Apply options before parcel and catalog creation.
	for _, opt := range opts {
		opt(i)
	}

	// Create the initial parcel if one wasn't provided via options.
	if i.parcel == nil {
		i.parcel = contract.NewParcel(nil, nil, logging.NewNoOpLogger(), nil)
	}

	// Create and populate the shared catalogs if not provided via options.
	if i.catalogs == nil {
		isRoot := true
		i.catalogs = newSharedCatalogs(i, isRoot, i.skipStdTools)
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

	if err := registerAeiouTools(i.catalogs.Tools().(tool.ToolRegistry), magicTool); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register AEIOU tools: %v", err))
	}

	i.SetInitialVariable("self", lang.StringValue{Value: DefaultSelfHandle})

	return i
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
	return i.rootInterpreter().catalogs.Accounts()
}

func (i *Interpreter) AccountsAdmin() interfaces.AccountAdmin {
	if sc, ok := i.rootInterpreter().catalogs.(*sharedCatalogs); ok {
		return account.NewAdmin(sc.accounts, i.parcel.Policy())
	}
	i.Logger().Error("FATAL: Could not get AccountAdmin; catalog is not the correct type")
	return nil
}

func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	return i.rootInterpreter().catalogs.AgentModels()
}

func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	if sc, ok := i.rootInterpreter().catalogs.(*sharedCatalogs); ok {
		return agentmodel.NewAgentModelAdmin(sc.agentModels, i.parcel.Policy())
	}
	i.Logger().Error("FATAL: Could not get AgentModelAdmin; catalog is not the correct type")
	return nil
}

func (i *Interpreter) CapsuleStore() *capsule.Store {
	return i.rootInterpreter().catalogs.Capsules()
}

func (i *Interpreter) CapsuleRegistryForAdmin() *capsule.Registry {
	if cs := i.rootInterpreter().catalogs.Capsules(); cs != nil {
		return cs.Registry(0)
	}
	return nil
}

func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	if i.catalogs == nil || i.catalogs.Tools() == nil {
		return nil
	}
	tr, _ := i.catalogs.Tools().(tool.ToolRegistry)
	return tr
}

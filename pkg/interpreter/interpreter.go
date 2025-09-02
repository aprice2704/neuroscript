// NeuroScript Version: 0.7.0
// File version: 58
// Purpose: Corrected the Load method to gracefully handle nil trees, fixing a test panic.
// filename: pkg/interpreter/interpreter.go
// nlines: 160
// risk_rating: MEDIUM

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
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// DefaultSelfHandle is the internal handle for the default whisper buffer.
const DefaultSelfHandle = "default_self_buffer"

// Interpreter holds the state for a NeuroScript runtime environment.
type Interpreter struct {
	logger              interfaces.Logger
	fileAPI             interfaces.FileAPI
	state               *interpreterState
	tools               tool.ToolRegistry
	eventManager        *EventManager
	evaluate            *evaluation
	aiWorker            interfaces.LLMClient
	shouldExit          bool
	exitCode            int
	returnValue         lang.Value
	lastCallResult      lang.Value
	stdout              io.Writer
	stdin               io.Reader
	stderr              io.Writer
	maxLoopIterations   int
	bufferManager       *BufferManager
	objectCache         map[string]interface{}
	objectCacheMu       sync.Mutex
	llmclient           interfaces.LLMClient
	skipStdTools        bool
	modelStore          *agentmodel.AgentModelStore
	ExecPolicy          *policy.ExecPolicy
	root                *Interpreter
	customEmitFunc      func(lang.Value)
	customWhisperFunc   func(handle, data lang.Value)
	turnCtx             context.Context
	aiTranscript        io.Writer
	transientPrivateKey ed25519.PrivateKey
	accountStore        *account.Store
}

// SetAITranscript sets the writer for logging AI prompts.
func (i *Interpreter) SetAITranscript(w io.Writer) {
	i.aiTranscript = w
}

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		state:             newInterpreterState(),
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
		logger:            logging.NewNoOpLogger(),
		stdout:            os.Stdout,
		stdin:             os.Stdin,
		stderr:            os.Stderr,
		bufferManager:     NewBufferManager(),
		objectCache:       make(map[string]interface{}),
		turnCtx:           context.Background(),
	}
	i.evaluate = &evaluation{i: i}
	i.tools = tool.NewToolRegistry(i)
	i.root = nil
	i.modelStore = agentmodel.NewAgentModelStore()
	i.accountStore = account.NewStore()

	i.bufferManager.Create(DefaultSelfHandle)
	i.customWhisperFunc = i.defaultWhisperFunc

	for _, opt := range opts {
		opt(i)
	}

	if !i.skipStdTools {
		if err := tool.RegisterExtendedTools(i.tools); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register extended tools: %v", err))
		}
	}

	_, transientPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to generate transient private key for AEIOU tool: %v", err))
	}
	i.transientPrivateKey = transientPrivateKey // Store the key

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
	i.state.globalVarNames[name] = true
	return nil
}

func (i *Interpreter) Load(tree *interfaces.Tree) error {
	if tree == nil || tree.Root == nil {
		i.logger.Warn("Load called with a nil program AST.")
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

// Accounts provides a read-only view of the account store.
func (i *Interpreter) Accounts() interfaces.AccountReader {
	return account.NewReader(i.accountStore)
}

// AccountsAdmin provides a policy-gated administrative view of the account store.
func (i *Interpreter) AccountsAdmin() interfaces.AccountAdmin {
	return account.NewAdmin(i.accountStore, i.ExecPolicy)
}

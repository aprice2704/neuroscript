// NeuroScript Version: 0.7.0
// File version: 48
// Purpose: Corrected the clone method to properly propagate custom emit and whisper functions, fixing multiple test failures.
// filename: pkg/interpreter/interpreter.go
// nlines: 190
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

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
	logger            interfaces.Logger
	fileAPI           interfaces.FileAPI
	state             *interpreterState
	tools             tool.ToolRegistry
	eventManager      *EventManager
	evaluate          *evaluation
	aiWorker          interfaces.LLMClient
	shouldExit        bool
	exitCode          int
	returnValue       lang.Value
	lastCallResult    lang.Value
	stdout            io.Writer
	stdin             io.Reader
	stderr            io.Writer
	maxLoopIterations int
	bufferManager     *BufferManager
	objectCache       map[string]interface{}
	objectCacheMu     sync.Mutex
	llmclient         interfaces.LLMClient
	skipStdTools      bool
	modelStore        *agentmodel.AgentModelStore
	ExecPolicy        *policy.ExecPolicy
	root              *Interpreter
	customEmitFunc    func(lang.Value)
	customWhisperFunc func(handle, data lang.Value)
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
	}
	i.evaluate = &evaluation{i: i}
	i.tools = tool.NewToolRegistry(i)
	i.root = nil
	i.modelStore = agentmodel.NewAgentModelStore()

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

	i.SetInitialVariable("self", lang.StringValue{Value: DefaultSelfHandle})

	return i
}

// defaultWhisperFunc is the built-in whisper implementation.
func (i *Interpreter) defaultWhisperFunc(handle, data lang.Value) {
	i.bufferManager.Write(handle.String(), data.String()+"\n")
}

// clone creates a new interpreter instance for sandboxing.
func (i *Interpreter) clone() *Interpreter {
	clone := NewInterpreter(
		WithLogger(i.logger),
		WithStdout(i.stdout),
		WithStdin(i.stdin),
		WithStderr(i.stderr),
		WithSandboxDir(i.state.sandboxDir),
	)
	clone.tools = i.tools
	clone.ExecPolicy = i.ExecPolicy
	clone.modelStore = i.modelStore

	// Propagate custom handlers to the clone.
	clone.customEmitFunc = i.customEmitFunc
	clone.customWhisperFunc = i.customWhisperFunc

	rootInterpreter := i
	if i.root != nil {
		rootInterpreter = i.root
	}
	clone.root = rootInterpreter

	clone.state.knownProcedures = i.state.knownProcedures

	rootInterpreter.state.variablesMu.RLock()
	defer rootInterpreter.state.variablesMu.RUnlock()

	for name := range rootInterpreter.state.globalVarNames {
		if val, ok := rootInterpreter.state.variables[name]; ok {
			clone.SetVariable(name, val)
			clone.state.globalVarNames[name] = true
		}
	}

	return clone
}

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
func (i *Interpreter) AddProcedure(proc ast.Procedure) error {
	if i.state.knownProcedures == nil {
		i.state.knownProcedures = make(map[string]*ast.Procedure)
	}
	if proc.Name() == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.state.knownProcedures[proc.Name()]; exists {
		return fmt.Errorf("%w: '%s'", lang.ErrProcedureExists, proc.Name())
	}
	i.state.knownProcedures[proc.Name()] = &proc
	return nil
}

func (i *Interpreter) GetAllVariables() (map[string]lang.Value, error) {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	clone := make(map[string]lang.Value)
	for k, v := range i.state.variables {
		clone[k] = v
	}
	return clone, nil
}

func (i *Interpreter) EvaluateExpression(node ast.Expression) (lang.Value, error) {
	return i.evaluate.Expression(node)
}

func (i *Interpreter) LoadAndRun(program *ast.Program, mainProcName string, args ...lang.Value) (lang.Value, error) {
	if err := i.Load(program); err != nil {
		return nil, fmt.Errorf("failed to load program: %w", err)
	}
	return i.Run(mainProcName, args...)
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

func (i *Interpreter) Load(program *ast.Program) error {
	if program == nil {
		i.logger.Warn("Load called with a nil program AST.")
		i.state.knownProcedures = make(map[string]*ast.Procedure)
		i.eventManager.eventHandlers = make(map[string][]*ast.OnEventDecl)
		i.state.commands = []*ast.CommandNode{}
		return nil
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

func (i *Interpreter) setSandboxDir(path string) {
	i.state.sandboxDir = path
}

// GetGrantSet returns the currently active capability grant set for policy enforcement.
func (i *Interpreter) GetGrantSet() *capability.GrantSet {
	if i.ExecPolicy == nil {
		return &capability.GrantSet{}
	}
	return &i.ExecPolicy.Grants
}

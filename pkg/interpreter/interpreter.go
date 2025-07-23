// NeuroScript Version: 0.5.2
// File version: 21
// Purpose: Removes the package-level default sandbox in favor of explicit configuration per interpreter.
// filename: pkg/interpreter/interpreter.go
// nlines: 319
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// interpreterState holds the non-exported state of the interpreter.
type interpreterState struct {
	variables         map[string]lang.Value
	variablesMu       sync.RWMutex
	knownProcedures   map[string]*ast.Procedure
	commands          []*ast.CommandNode
	stackFrames       []string
	currentProcName   string
	errorHandlerStack [][]*ast.Step
	sandboxDir        string
	vectorIndex       map[string][]float32
	globalVarNames    map[string]bool
}

// EventManager handles event subscriptions and emissions.
type EventManager struct {
	eventHandlers   map[string][]*ast.OnEventDecl
	eventHandlersMu sync.RWMutex
}

// Interpreter holds the state for a NeuroScript runtime environment.
type Interpreter struct {
	logger             interfaces.Logger
	fileAPI            interfaces.FileAPI
	state              *interpreterState
	tools              tool.ToolRegistry
	eventManager       *EventManager
	evaluate           *evaluation
	aiWorker           interfaces.LLMClient
	shouldExit         bool
	exitCode           int
	returnValue        lang.Value
	lastCallResult     lang.Value
	stdout             io.Writer
	stdin              io.Reader
	stderr             io.Writer
	maxLoopIterations  int
	ToolCallTimestamps map[string][]time.Time
	rateLimitCount     int
	rateLimitDuration  time.Duration
	externalHandler    interface{}
	objectCache        map[string]interface{}
	objectCacheMu      sync.RWMutex // Mutex for the handle cache
	llmclient          interfaces.LLMClient
}

func (i *Interpreter) NTools() (ntools int) {
	return i.tools.NTools()
}

func (i *Interpreter) LLM() interfaces.LLMClient {
	return i.llmclient
}

// InterpreterOption defines a function signature for configuring an Interpreter.
type InterpreterOption func(*Interpreter)

// --- Functional Options ---

func WithLogger(logger interfaces.Logger) InterpreterOption {
	return func(i *Interpreter) {
		i.logger = logger
	}
}

func WithLLMClient(client interfaces.LLMClient) InterpreterOption {
	return func(i *Interpreter) {
		i.aiWorker = client
	}
}

func WithSandboxDir(path string) InterpreterOption {
	return func(i *Interpreter) {
		i.setSandboxDir(path)
	}
}

func WithStdout(w io.Writer) InterpreterOption {
	return func(i *Interpreter) {
		i.stdout = w
	}
}

func WithStdin(r io.Reader) InterpreterOption {
	return func(i *Interpreter) {
		i.stdin = r
	}
}

func WithStderr(w io.Writer) InterpreterOption {
	return func(i *Interpreter) {
		i.stderr = w
	}
}

// WithInitialGlobals sets the initial global variables.
func WithInitialGlobals(globals map[string]interface{}) InterpreterOption {
	return func(i *Interpreter) {
		for key, val := range globals {
			if err := i.SetInitialVariable(key, val); err != nil {
				i.logger.Error("Failed to set initial global variable", "key", key, "error", err)
			}
		}
	}
}

// --- interpreterState Methods ---

func newInterpreterState() *interpreterState {
	return &interpreterState{
		variables:       make(map[string]lang.Value),
		knownProcedures: make(map[string]*ast.Procedure),
		commands:        []*ast.CommandNode{},
		stackFrames:     []string{},
		globalVarNames:  make(map[string]bool),
	}
}

func (s *interpreterState) setVariable(name string, value lang.Value) {
	s.variablesMu.Lock()
	defer s.variablesMu.Unlock()
	if s.variables == nil {
		s.variables = make(map[string]lang.Value)
	}
	s.variables[name] = value
}

// --- EventManager Methods ---

func newEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[string][]*ast.OnEventDecl),
	}
}

func (em *EventManager) register(decl *ast.OnEventDecl, i *Interpreter) error {
	em.eventHandlersMu.Lock()
	defer em.eventHandlersMu.Unlock()

	eventName, err := i.evaluate.Expression(decl.EventNameExpr)
	if err != nil {
		return lang.WrapErrorWithPosition(err, decl.EventNameExpr.GetPos(), "evaluating event name expression")
	}

	eventNameStr, _ := lang.ToString(eventName)
	em.eventHandlers[eventNameStr] = append(em.eventHandlers[eventNameStr], decl)
	return nil
}

// --- Interpreter Methods ---

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		state:             newInterpreterState(),
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
		logger:            logging.NewNoOpLogger(),
		stdout:            os.Stdout,
		stdin:             os.Stdin,
		stderr:            os.Stderr,
		objectCache:       make(map[string]interface{}),
	}
	i.evaluate = &evaluation{i: i}
	i.tools = tool.NewToolRegistry(i)
	if err := tool.RegisterExtendedTools(i.tools); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register extended tools: %v", err))
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
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

// KnownProcedures returns the map of known procedures.
func (i *Interpreter) KnownProcedures() map[string]*ast.Procedure {
	if i.state.knownProcedures == nil {
		return make(map[string]*ast.Procedure)
	}
	return i.state.knownProcedures
}

// ToolRegistry returns the interpreter's tool registry.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.tools
}

// CloneForEventHandler creates a sandboxed clone for event handling.
func (i *Interpreter) CloneForEventHandler() *Interpreter {
	clone := NewInterpreter(WithLogger(i.logger), WithStdout(i.stdout), WithSandboxDir(i.state.sandboxDir))
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	for name := range i.state.globalVarNames {
		if val, ok := i.state.variables[name]; ok {
			clone.SetInitialVariable(name, val)
		}
	}
	for name, proc := range i.state.knownProcedures {
		clone.state.knownProcedures[name] = proc
	}
	return clone
}

func (i *Interpreter) GetLogger() interfaces.Logger {
	return i.logger
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
	fmt.Printf(">>>> [DEBUG] interpreter.Run: Value being RETURNED to API FACADE is: %#v\n", result)
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

func (i *Interpreter) SetLastResult(v lang.Value) {
	i.lastCallResult = v
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

func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	clone := NewInterpreter(WithLogger(i.logger), WithStdout(i.stdout))
	for k, v := range i.state.knownProcedures {
		clone.state.knownProcedures[k] = v
	}
	return clone
}

func (i *Interpreter) setSandboxDir(path string) {
	i.state.sandboxDir = path
}

func (i *Interpreter) RegisterEvent(decl *ast.OnEventDecl) error {
	return i.eventManager.register(decl, i)
}

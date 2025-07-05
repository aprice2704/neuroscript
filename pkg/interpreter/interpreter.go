// filename: pkg/interpreter/interpreter.go
package interpreter

import (
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
	llmclient          interfaces.LLMClient
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
		i.SetSandboxDir(path)
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
			// Error handling can be added here if wrapping fails
			wrappedVal, _ := lang.Wrap(val)
			i.state.setVariable(key, wrappedVal)
		}
	}
}

// WithInitialIncludes sets the initial include paths.
func WithInitialIncludes(includes []string) InterpreterOption {
	return func(i *Interpreter) {
		// Assuming you have a field for includes, e.g., i.state.includes
		// i.state.includes = includes
	}
}

// --- interpreterState Methods ---

func newInterpreterState() *interpreterState {
	return &interpreterState{
		variables:       make(map[string]lang.Value),
		knownProcedures: make(map[string]*ast.Procedure),
		commands:        []*ast.CommandNode{},
		stackFrames:     []string{},
	}
}

func (s *interpreterState) getProcedure(name string) *ast.Procedure {
	if s.knownProcedures == nil {
		return nil
	}
	return s.knownProcedures[name]
}

func (s *interpreterState) setProcedure(name string, proc *ast.Procedure) {
	if s.knownProcedures == nil {
		s.knownProcedures = make(map[string]*ast.Procedure)
	}
	s.knownProcedures[name] = proc
}

func (s *interpreterState) clearProcedures() {
	s.knownProcedures = make(map[string]*ast.Procedure)
}

func (s *interpreterState) addCommand(cmd *ast.CommandNode) {
	s.commands = append(s.commands, cmd)
}

func (s *interpreterState) clearCommands() {
	s.commands = []*ast.CommandNode{}
}

func (s *interpreterState) pushStackFrame(name string) {
	s.stackFrames = append(s.stackFrames, name)
}

func (s *interpreterState) popStackFrame() {
	if len(s.stackFrames) > 0 {
		s.stackFrames = s.stackFrames[:len(s.stackFrames)-1]
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

func (s *interpreterState) Errorf(code lang.ErrorCode, format string, args ...interface{}) error {
	return lang.NewRuntimeError(code, fmt.Sprintf(format, args...), nil)
}

// --- EventManager Methods ---

func newEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[string][]*ast.OnEventDecl),
	}
}

func (em *EventManager) clear() {
	em.eventHandlers = make(map[string][]*ast.OnEventDecl)
}

func (em *EventManager) register(decl *ast.OnEventDecl, i *Interpreter) error {
	return nil
}

// --- Interpreter Methods ---

func NewInterpreter(opts ...InterpreterOption) *Interpreter {
	i := &Interpreter{
		state:             newInterpreterState(),
		eventManager:      newEventManager(),
		maxLoopIterations: 1000,
	}
	i.logger = logging.NewNoOpLogger()
	i.evaluate = &evaluation{i: i}
	i.stdout = os.Stdout
	i.stdin = os.Stdin
	i.stderr = os.Stderr

	for _, opt := range opts {
		opt(i)
	}

	i.tools = NewToolRegistry(i)
	RegisterCoreTools(i.tools)

	return i
}

// GetLogger satisfies the tool.Runtime interface.
func (i *Interpreter) GetLogger() interfaces.Logger {
	return i.logger
}

func (i *Interpreter) GetAllVariables() (map[string]lang.Value, error) {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	// Return a copy to prevent race conditions if the caller modifies the map.
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
	if i.state.getProcedure(procName) == nil {
		return &lang.NilValue{}, fmt.Errorf("procedure '%s' not found", procName)
	}

	result, err := i.callProcedure(procName, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *Interpreter) callProcedure(name string, args ...lang.Value) (lang.Value, error) {
	proc := i.state.getProcedure(name)
	if proc == nil {
		return nil, i.state.Errorf(lang.ErrorCodeProcNotFound, "procedure '%s' not found", name)
	}

	i.state.pushStackFrame(name)
	defer i.state.popStackFrame()

	_, _, _, err := i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		return nil, err
	}

	return i.returnValue, nil
}

func (i *Interpreter) IsRunningInTestMode() bool {
	return os.Getenv("GO_TEST_MODE") == "1"
}

func NewForTesting() *Interpreter {
	i := NewInterpreter(nil)
	i.logger = logging.NewNoOpLogger()
	return i
}

func (i *Interpreter) SetInitialVariable(name string, value any) error {
	wrappedValue, err := lang.Wrap(value)
	if err != nil {
		return fmt.Errorf("failed to wrap initial variable '%s': %w", name, err)
	}
	i.state.setVariable(name, wrappedValue)
	return nil
}

func (i *Interpreter) SetLastResult(v lang.Value) {
	i.lastCallResult = v
}

func (i *Interpreter) Load(program *ast.Program) error {
	if program == nil {
		return fmt.Errorf("cannot load a nil program")
	}

	i.state.clearProcedures()
	i.state.clearCommands()
	i.returnValue = &lang.NilValue{}

	for name, proc := range program.Procedures {
		i.state.setProcedure(name, proc)
	}

	for _, cmd := range program.Commands {
		i.state.addCommand(cmd)
	}

	if i.eventManager != nil {
		i.eventManager.clear()
		for _, eventDecl := range program.Events {
			if err := i.eventManager.register(eventDecl, i); err != nil {
				return fmt.Errorf("failed to register event handler: %w", err)
			}
		}
	}

	return nil
}

func RegisterCoreTools(registry tool.ToolRegistry) {
}

func NewToolRegistry(i *Interpreter) tool.ToolRegistry {
	return nil
}

func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	clone := *i
	clone.state = newInterpreterState()
	for k, v := range i.state.knownProcedures {
		clone.state.setProcedure(k, v)
	}
	return &clone
}

func (i *Interpreter) SetSandboxDir(path string) {
	i.state.sandboxDir = path
}

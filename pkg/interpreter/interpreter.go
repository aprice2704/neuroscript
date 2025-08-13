// NeuroScript Version: 0.6.0
// File version: 38.0.0
// Purpose: Initializes the AgentModelStore in the constructor to prevent nil pointer panics when using agent model tools.
// filename: pkg/interpreter/interpreter.go
// nlines: 255
// risk_rating: HIGH

package interpreter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/runtime"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

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
	ToolCallTimestamps map[string]interface{}
	rateLimitCount     int
	rateLimitDuration  interface{}
	externalHandler    interface{}
	objectCache        map[string]interface{}
	objectCacheMu      interface{}
	llmclient          interfaces.LLMClient
	skipStdTools       bool
	modelStore         *runtime.AgentModelStore
	ExecPolicy         *runtime.ExecPolicy
	root               *Interpreter // Points to the top-level interpreter instance.
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
		objectCache:       make(map[string]interface{}),
	}
	i.evaluate = &evaluation{i: i}
	i.tools = tool.NewToolRegistry(i)
	i.root = nil // The new interpreter is its own root.
	i.modelStore = runtime.NewAgentModelStore()

	for _, opt := range opts {
		opt(i)
	}

	if !i.skipStdTools {
		if err := tool.RegisterExtendedTools(i.tools); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register extended tools: %v", err))
		}
	}

	return i
}

// clone creates a new interpreter instance for sandboxing.
// It inherits shared state like tools and agent models via the root pointer,
// but gets its own variable scope to prevent state leakage.
func (i *Interpreter) clone() *Interpreter {
	clone := NewInterpreter(
		WithLogger(i.logger),
		WithStdout(i.stdout),
		WithStdin(i.stdin),
		WithStderr(i.stderr),
		WithSandboxDir(i.state.sandboxDir),
	)
	// Share the same tool registry, execution policy, and agent model store.
	clone.tools = i.tools
	clone.ExecPolicy = i.ExecPolicy
	clone.modelStore = i.modelStore

	// Point clone's root to the original interpreter's root, or to the
	// original if it is the root. This ensures access to shared state
	// like AgentModels and providers.
	if i.root != nil {
		clone.root = i.root
	} else {
		clone.root = i
	}

	// The clone gets its own variable map, but inherits known procedures.
	clone.state.knownProcedures = i.state.knownProcedures

	return clone
}

func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	return runtime.NewAgentModelReader(i.modelStore)
}
func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	return runtime.NewAgentModelAdmin(i.modelStore, i.ExecPolicy)
}

// PromptUser satisfies the tool.Runtime interface.
func (i *Interpreter) PromptUser(prompt string) (string, error) {
	if _, err := fmt.Fprint(i.Stdout(), prompt+" "); err != nil {
		return "", fmt.Errorf("failed to write prompt to stdout: %w", err)
	}
	reader := bufio.NewReader(i.Stdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}
	return strings.TrimSpace(response), nil
}

// RegisterProvider allows the host application to register a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p provider.AIProvider) {
	// Delegate to root if this is a clone
	if i.root != nil {
		i.root.RegisterProvider(name, p)
		return
	}
	i.state.providersMu.Lock()
	defer i.state.providersMu.Unlock()
	i.state.providers[name] = p
}

// GetProvider retrieves a registered AIProvider by name.
func (i *Interpreter) GetProvider(name string) (provider.AIProvider, bool) {
	if i.root != nil {
		return i.root.GetProvider(name)
	}
	i.state.providersMu.RLock()
	defer i.state.providersMu.RUnlock()
	p, found := i.state.providers[name]
	return p, found
}

// NTools returns the number of registered tools.
func (i *Interpreter) NTools() (ntools int) {
	return i.tools.NTools()
}

// LLM returns the configured LLM client.
func (i *Interpreter) LLM() interfaces.LLMClient {
	return i.llmclient
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
	clone := i.clone() // Use the centralized, corrected clone method.

	// Copy global variables for read-only access, as per the spec.
	// We lock the parent's variables for reading.
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()

	for name := range i.state.globalVarNames {
		if val, ok := i.state.variables[name]; ok {
			// This sets the variable on the clone's independent variable map.
			clone.SetVariable(name, val)
		}
	}
	return clone
}

// CloneWithNewVariables creates a clone with a fresh set of variables for procedure calls.
func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	return i.clone() // The corrected clone method already creates a fresh variable map.
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

func (i *Interpreter) setSandboxDir(path string) {
	i.state.sandboxDir = path
}

func (i *Interpreter) RegisterEvent(decl *ast.OnEventDecl) error {
	return i.eventManager.register(decl, i)
}

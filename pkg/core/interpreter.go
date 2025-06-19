// NeuroScript Version: 0.3.1
// File version: 9
// Purpose: Added a new public ExecuteProc method to run a loaded procedure by name.
// filename: pkg/core/interpreter.go
// nlines: 295+
// risk_rating: HIGH
package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
)

var GrammarVersion string
var AppVersion string

type Interpreter struct {
	variables          map[string]Value
	variablesMu        sync.RWMutex
	knownProcedures    map[string]*Procedure
	lastCallResult     Value
	vectorIndex        map[string][]float32
	embeddingDim       int
	currentProcName    string
	sandboxDir         string
	LibPaths           []string
	stdout             io.Writer
	externalHandler    ToolHandler
	toolRegistry       *ToolRegistryImpl
	logger             interfaces.Logger
	objectCache        map[string]interface{}
	llmClient          interfaces.LLMClient
	fileAPI            *FileAPI
	aiWorkerManager    *AIWorkerManager
	ToolCallTimestamps map[string][]time.Time
	rateLimitCount     int
	rateLimitDuration  time.Duration
	maxLoopIterations  int
	eventHandlers      map[string][]*OnEventDecl
	eventHandlersMu    sync.RWMutex
}

// ExecuteProc finds and executes a procedure that has already been loaded into the
// interpreter by name, returning the final unwrapped result.
func (i *Interpreter) ExecuteProc(procName string) (interface{}, error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		// As a fallback for simple test scripts, if only one procedure is loaded, run it.
		if len(i.knownProcedures) == 1 {
			for _, p := range i.knownProcedures {
				proc = p
				break
			}
		} else {
			return nil, NewRuntimeError(ErrorCodeProcNotFound, fmt.Sprintf("procedure '%s' not found", procName), ErrProcedureNotFound)
		}
	}

	i.currentProcName = proc.Name
	finalResult, wasReturn, _, err := i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		return nil, err
	}

	var returnValue Value
	if wasReturn {
		returnValue = finalResult
	} else {
		// If the function ends without an explicit return, the result is the
		// result of the last executed statement.
		returnValue = i.lastCallResult
	}

	return Unwrap(returnValue), nil
}

// LoadProgram registers all procedures and event handlers from a parsed Program AST.
func (i *Interpreter) LoadProgram(prog *Program) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]*Procedure)
	}
	for name, proc := range prog.Procedures {
		if _, exists := i.knownProcedures[name]; exists {
			return fmt.Errorf("procedure '%s' already exists", name)
		}
		i.knownProcedures[name] = proc
	}
	if i.eventHandlers == nil {
		i.eventHandlers = make(map[string][]*OnEventDecl)
	}
	for _, ev := range prog.Events {
		nameLit, ok := ev.EventNameExpr.(*StringLiteralNode)
		if !ok {
			// This check enforces that event names must be static strings at load time.
			return NewRuntimeError(ErrorCodeType, "event name must be a static string literal", nil).WithPosition(ev.Pos)
		}
		eventName := nameLit.Value

		i.eventHandlersMu.Lock()
		i.eventHandlers[eventName] = append(i.eventHandlers[eventName], ev)
		i.eventHandlersMu.Unlock()
	}
	return nil
}

// ... (The rest of the file is unchanged) ...
const handleSeparator = "::"

func NewInterpreter(logger interfaces.Logger, llmClient interfaces.LLMClient, sandboxDir string, initialVars map[string]interface{}, libPaths []string) (*Interpreter, error) {
	effectiveLogger := logger
	if effectiveLogger == nil {
		if !IsRunningInTestMode() {
			log.Fatalf("FATAL: Critical error: No logger is active, and we are not in test mode. Exiting.")
		}
		effectiveLogger = &coreNoOpLogger{}
	}
	if GrammarVersion != "" {
		effectiveLogger.Infof("NeuroScript Grammar Version: %s", GrammarVersion)
	}
	if AppVersion != "" {
		effectiveLogger.Infof("NeuroScript App Version: %s", AppVersion)
	}
	effectiveLLMClient := llmClient
	if effectiveLLMClient == nil {
		effectiveLogger.Warn("NewInterpreter: nil LLMClient provided. Initializing with a NoOp LLMClient.")
		effectiveLLMClient, _ = NewLLMClient("", "", effectiveLogger)
	}
	cleanSandboxDir := "."
	if sandboxDir != "" {
		absPath, err := filepath.Abs(sandboxDir)
		if err != nil {
			return nil, fmt.Errorf("invalid sandbox directory '%s': %w", sandboxDir, err)
		}
		cleanSandboxDir = filepath.Clean(absPath)
	}
	fileAPI := NewFileAPI(cleanSandboxDir, effectiveLogger)
	vars := make(map[string]Value)
	mustWrapAndSet := func(k string, v interface{}) {
		wrapped, err := Wrap(v)
		if err != nil {
			panic(fmt.Sprintf("FATAL: could not wrap internal variable %s: %v", k, err))
		}
		vars[k] = wrapped
	}
	mustWrapAndSet("NEUROSCRIPT_DEVELOP_PROMPT", prompts.PromptDevelop)
	mustWrapAndSet("NEUROSCRIPT_EXECUTE_PROMPT", prompts.PromptExecute)
	mustWrapAndSet("TYPE_STRING", string(TypeString))
	mustWrapAndSet("TYPE_NUMBER", string(TypeNumber))
	mustWrapAndSet("TYPE_BOOLEAN", string(TypeBoolean))
	mustWrapAndSet("TYPE_LIST", string(TypeList))
	mustWrapAndSet("TYPE_MAP", string(TypeMap))
	mustWrapAndSet("TYPE_NIL", string(TypeNil))
	mustWrapAndSet("TYPE_FUNCTION", string(TypeFunction))
	mustWrapAndSet("TYPE_TOOL", string(TypeTool))
	mustWrapAndSet("TYPE_ERROR", string(TypeError))
	mustWrapAndSet("TYPE_UNKNOWN", string(TypeUnknown))
	if initialVars != nil {
		for k, v := range initialVars {
			wrappedVal, err := Wrap(v)
			if err != nil {
				return nil, fmt.Errorf("failed to wrap initial variable '%s': %w", k, err)
			}
			vars[k] = wrappedVal
		}
	}
	interp := &Interpreter{
		variables:          vars,
		variablesMu:        sync.RWMutex{},
		knownProcedures:    make(map[string]*Procedure),
		lastCallResult:     NilValue{},
		vectorIndex:        make(map[string][]float32),
		embeddingDim:       16,
		currentProcName:    "",
		sandboxDir:         cleanSandboxDir,
		LibPaths:           libPaths,
		stdout:             os.Stdout,
		externalHandler:    nil,
		toolRegistry:       nil,
		logger:             effectiveLogger,
		objectCache:        make(map[string]interface{}),
		llmClient:          effectiveLLMClient,
		fileAPI:            fileAPI,
		aiWorkerManager:    nil,
		ToolCallTimestamps: make(map[string][]time.Time),
		rateLimitCount:     10,
		rateLimitDuration:  time.Minute,
		maxLoopIterations:  1000000,
		eventHandlers:      make(map[string][]*OnEventDecl),
	}
	interp.toolRegistry = NewToolRegistry(interp)
	if err := RegisterCoreTools(interp); err != nil {
		return nil, fmt.Errorf("FATAL: failed to register core tools: %w", err)
	}
	return interp, nil
}
func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	i.variablesMu.RLock()
	newVars := make(map[string]Value, len(i.variables))
	for k, v := range i.variables {
		newVars[k] = v
	}
	i.variablesMu.RUnlock()
	i.eventHandlersMu.RLock()
	newHandlers := make(map[string][]*OnEventDecl, len(i.eventHandlers))
	for k, v := range i.eventHandlers {
		newHandlers[k] = v
	}
	i.eventHandlersMu.RUnlock()
	clone := &Interpreter{
		knownProcedures:    i.knownProcedures,
		lastCallResult:     i.lastCallResult,
		vectorIndex:        i.vectorIndex,
		embeddingDim:       i.embeddingDim,
		sandboxDir:         i.sandboxDir,
		LibPaths:           i.LibPaths,
		stdout:             i.stdout,
		externalHandler:    i.externalHandler,
		toolRegistry:       i.toolRegistry,
		logger:             i.logger,
		objectCache:        i.objectCache,
		llmClient:          i.llmClient,
		fileAPI:            i.fileAPI,
		aiWorkerManager:    i.aiWorkerManager,
		ToolCallTimestamps: i.ToolCallTimestamps,
		rateLimitCount:     i.rateLimitCount,
		rateLimitDuration:  i.rateLimitDuration,
		maxLoopIterations:  i.maxLoopIterations,
		variables:          newVars,
		variablesMu:        sync.RWMutex{},
		eventHandlers:      newHandlers,
		eventHandlersMu:    sync.RWMutex{},
	}
	return clone
}
func (i *Interpreter) SetStdout(writer io.Writer) {
	if writer == nil {
		i.logger.Warn("Attempted to set nil stdout writer on interpreter, using os.Stdout as fallback.")
		i.stdout = os.Stdout
		return
	}
	i.stdout = writer
}
func (i *Interpreter) Stdout() io.Writer {
	if i.stdout == nil {
		return os.Stdout
	}
	return i.stdout
}
func (i *Interpreter) SetAIWorkerManager(manager *AIWorkerManager) {
	i.aiWorkerManager = manager
}
func (i *Interpreter) AIWorkerManager() *AIWorkerManager {
	return i.aiWorkerManager
}
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }
func (i *Interpreter) Logger() interfaces.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}
func (i *Interpreter) FileAPI() *FileAPI {
	if i.fileAPI == nil {
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}
func (i *Interpreter) SetSandboxDir(newSandboxDir string) error {
	var cleanNewSandboxDir string
	if newSandboxDir == "" {
		cleanNewSandboxDir = "."
	} else {
		absPath, err := filepath.Abs(newSandboxDir)
		if err != nil {
			return fmt.Errorf("invalid sandbox directory '%s': %w", newSandboxDir, err)
		}
		cleanNewSandboxDir = filepath.Clean(absPath)
	}
	if i.sandboxDir != cleanNewSandboxDir {
		i.sandboxDir = cleanNewSandboxDir
		i.fileAPI = NewFileAPI(i.sandboxDir, i.logger)
	}
	return nil
}
func (i *Interpreter) SetVariable(name string, value Value) error {
	i.variablesMu.Lock()
	defer i.variablesMu.Unlock()
	if i.variables == nil {
		i.variables = make(map[string]Value)
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	return nil
}
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	i.variablesMu.RLock()
	defer i.variablesMu.RUnlock()
	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		return nil
	}
	return i.llmClient.Client()
}
func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }

var _ ToolRegistry = (*Interpreter)(nil)

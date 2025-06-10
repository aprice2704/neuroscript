// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Core interpreter struct definition, constructor, and basic state management. Added mutex for variable safety.
// filename: pkg/core/interpreter.go
// nlines: 140+
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

// Injected version of grammar we are using, from NeuroScript.g4
var GrammarVersion string

// Injected version of grammar we are using, from NeuroScript.g4
var AppVersion string

type Interpreter struct {
	variables          map[string]interface{}
	variablesMu        sync.RWMutex // Mutex to protect concurrent access to the variables map.
	knownProcedures    map[string]*Procedure
	lastCallResult     interface{}
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

// LoadProgram registers all procedures and event handlers from a parsed Program AST.
// This is the new primary method for loading a script into the interpreter.
func (i *Interpreter) LoadProgram(prog *Program) error {
	// Register Procedures
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]*Procedure)
	}
	for name, proc := range prog.Procedures {
		if _, exists := i.knownProcedures[name]; exists {
			return fmt.Errorf("procedure '%s' already exists", name)
		}
		i.knownProcedures[name] = proc
	}

	// Register Event Handlers
	if i.eventHandlers == nil {
		i.eventHandlers = make(map[string][]*OnEventDecl)
	}
	for _, ev := range prog.Events {
		nameLit, ok := ev.EventNameExpr.(*StringLiteralNode)
		if !ok {
			return NewRuntimeError(ErrorCodeType, "event name must be a static string literal", nil).WithPosition(ev.Pos)
		}
		eventName := nameLit.Value

		i.eventHandlersMu.Lock()
		i.eventHandlers[eventName] = append(i.eventHandlers[eventName], ev)
		i.eventHandlersMu.Unlock()
	}
	return nil
}

// --- Constants ---
const handleSeparator = "::"

// --- Constructor ---

func NewInterpreter(logger interfaces.Logger, llmClient interfaces.LLMClient, sandboxDir string, initialVars map[string]interface{}, libPaths []string) (*Interpreter, error) {

	effectiveLogger := logger

	if effectiveLogger == nil {
		if IsRunningInTestMode() {
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

	vars := make(map[string]interface{})
	vars["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	vars["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute
	vars["TYPE_STRING"] = string(TypeString)
	vars["TYPE_NUMBER"] = string(TypeNumber)
	vars["TYPE_BOOLEAN"] = string(TypeBoolean)
	vars["TYPE_LIST"] = string(TypeList)
	vars["TYPE_MAP"] = string(TypeMap)
	vars["TYPE_NIL"] = string(TypeNil)
	vars["TYPE_FUNCTION"] = string(TypeFunction)
	vars["TYPE_TOOL"] = string(TypeTool)
	vars["TYPE_ERROR"] = string(TypeError)
	vars["TYPE_UNKNOWN"] = string(TypeUnknown)

	if initialVars != nil {
		for k, v := range initialVars {
			vars[k] = v
		}
	}

	interp := &Interpreter{
		variables:          vars,
		variablesMu:        sync.RWMutex{},
		knownProcedures:    make(map[string]*Procedure),
		lastCallResult:     nil,
		vectorIndex:        make(map[string][]float32),
		embeddingDim:       16,
		currentProcName:    "",
		sandboxDir:         cleanSandboxDir,
		LibPaths:           libPaths,
		stdout:             os.Stdout,
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

func (i *Interpreter) SetVariable(name string, value interface{}) error {
	i.variablesMu.Lock()
	defer i.variablesMu.Unlock()
	if i.variables == nil {
		i.variables = make(map[string]interface{})
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	return nil
}

func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
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

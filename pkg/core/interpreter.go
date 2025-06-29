// NeuroScript Version: 0.4.2
// File version: 14
// Purpose: Contains the core Interpreter struct definition and its constructor/cloner. Refactored from a larger file.
// filename: pkg/core/interpreter.go
// nlines: 149
// risk_rating: LOW
package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

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
	stderr             io.Writer
	stdin              io.Reader
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

	// A stack to hold lists of error handlers for each active procedure call.
	errorHandlerStack [][]*Step
}

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
	for k, v := range initialVars {
		wrappedVal, err := Wrap(v)
		if err != nil {
			return nil, fmt.Errorf("failed to wrap initial variable '%s': %w", k, err)
		}
		vars[k] = wrappedVal
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
		stderr:             os.Stderr,
		stdin:              os.Stdin,
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
		errorHandlerStack:  make([][]*Step, 0),
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
		stderr:             i.stderr,
		stdin:              i.stdin,
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
		errorHandlerStack:  make([][]*Step, 0),
	}
	return clone
}

var _ ToolRegistry = (*Interpreter)(nil)

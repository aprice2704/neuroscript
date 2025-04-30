// filename: pkg/core/interpreter_new.go
package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings" // Keep strings import

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
)

// Interpreter holds the state of a running NeuroScript program.
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure // Assumes Procedure is defined (likely in ast.go)
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string // Track current proc for semantic checks
	sandboxDir      string

	toolRegistry *ToolRegistry
	logger       logging.Logger
	objectCache  map[string]interface{} // Cache for handle objects
	llmClient    LLMClient              // Store the interface value directly

	// Note: No Frame/CallStack here based on user's provided code structure.
	// Error state is handled via return values from executeSteps.
}

// --- Constants ---
const handleSeparator = "::"

// --- Constructor ---

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger logging.Logger, llmClient LLMClient) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = &coreNoOpLogger{}
	}

	effectiveLLMClient := llmClient
	if effectiveLLMClient == nil {
		effectiveLogger.Warn("Interpreter created with nil LLMClient, using internal NoOpLLMClient.")
		effectiveLLMClient = NewNoOpLLMClient(effectiveLogger)
	}

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16,
		// CORRECTED: NewToolRegistry takes no arguments
		toolRegistry: NewToolRegistry(),
		logger:       effectiveLogger,
		objectCache:  make(map[string]interface{}),
		llmClient:    effectiveLLMClient,
		sandboxDir:   ".",
	}

	// Initialize default prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// Register core tools
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization", "error", err)
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("Core tools registered successfully.")
	}
	return interp
}

// --- Getters / Setters ---
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }
func (i *Interpreter) Logger() logging.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

func (i *Interpreter) SetVariable(name string, value interface{}) error {
	if i.variables == nil {
		i.variables = make(map[string]interface{})
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	i.Logger().Debug("Set variable", "name", name, "value", value, "type", fmt.Sprintf("%T", value))
	return nil
}

func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}

func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		i.Logger().Warn("ToolRegistry accessed before initialization, creating new one.")
		// CORRECTED: NewToolRegistry takes no arguments
		i.toolRegistry = NewToolRegistry()
	}
	return i.toolRegistry
}

// GenAIClient attempts to retrieve the underlying *genai.Client if the
// current llmClient implementation holds one. Returns nil otherwise.
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		i.Logger().Warn("GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	// Placeholder: Requires concreteLLMClient to expose the *genai.Client
	// Option 1: Via method
	// type genAIProvider interface { Client() *genai.Client }
	// if provider, ok := i.llmClient.(genAIProvider); ok {
	//     return provider.Client()
	// }
	// Option 2: Via field (less ideal)
	// if concrete, ok := i.llmClient.(*concreteLLMClient); ok {
	//     // return concrete.GenAI // Requires exported field GenAI *genai.Client
	// }
	i.Logger().Warn("GenAIClient() called, but the configured LLMClient does not provide a *genai.Client via type assertion.")
	return nil
}

// AddProcedure adds a parsed procedure definition.
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if proc.Metadata == nil {
		proc.Metadata = make(map[string]string)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	i.Logger().Debug("Added procedure to known procedures.", "name", proc.Name)
	return nil
}

func (i *Interpreter) KnownProcedures() map[string]Procedure {
	if i.knownProcedures == nil {
		return make(map[string]Procedure)
	}
	return i.knownProcedures
}

func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }

// --- Handle Management ---
func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", errors.New("handle type prefix cannot be empty")
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("handle type prefix '%s' cannot contain separator '%s'", typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
	}
	handleID := uuid.NewString()
	fullHandle := fmt.Sprintf("%s%s%s", typePrefix, handleSeparator, handleID)
	i.objectCache[fullHandle] = obj
	i.Logger().Debug("Registered handle", "handle", fullHandle, "type", typePrefix)
	return fullHandle, nil
}
func (i *Interpreter) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	if expectedTypePrefix == "" {
		return nil, errors.New("expected handle type prefix cannot be empty")
	}
	if handle == "" {
		return nil, errors.New("handle cannot be empty")
	}
	prefixWithSeparator := expectedTypePrefix + handleSeparator
	if !strings.HasPrefix(handle, prefixWithSeparator) {
		parts := strings.SplitN(handle, handleSeparator, 2)
		actualPrefix := "(invalid format)"
		if len(parts) > 0 {
			actualPrefix = parts[0]
		}
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s' (full handle: '%s')", ErrCacheObjectWrongType, expectedTypePrefix, actualPrefix, handle)
	}
	if i.objectCache == nil {
		i.Logger().Error("GetHandleValue called but objectCache is nil.")
		return nil, errors.New("internal error: object cache is not initialized")
	}
	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s'", ErrCacheObjectNotFound, handle)
	}
	i.Logger().Debug("Retrieved handle", "handle", handle, "expected_type", expectedTypePrefix)
	return obj, nil
}
func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
		i.Logger().Debug("Removed handle", "handle", handle)
	}
	return found
}

// --- Main Execution Entry Point ---
func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
	originalProcName := i.currentProcName
	i.Logger().Info("Running procedure", "name", procName, "arg_count", len(args))
	i.currentProcName = procName

	defer func() {
		i.currentProcName = originalProcName
		logArgs := []any{"proc_name", procName, "restored_proc_name", i.currentProcName, "result", result, "result_type", fmt.Sprintf("%T", result), "error", err}
		i.Logger().Info("Finished procedure.", logArgs...)
	}()

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		i.Logger().Error("Procedure not found", "name", procName)
		return nil, err
	}

	// Argument Handling
	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments ('needs'), but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		i.Logger().Error("Argument mismatch", "proc_name", procName, "required", numRequired, "provided", numProvided)
		return nil, err
	}
	if numProvided > numTotalParams {
		i.Logger().Warn("Procedure called with extra arguments.", "proc_name", procName, "provided", numProvided, "defined", numTotalParams)
	}

	// Scope Management
	procScope := make(map[string]interface{})
	for k, v := range i.variables {
		procScope[k] = v
	}
	originalScope := i.variables
	i.variables = procScope
	defer func() {
		i.variables = originalScope
		i.Logger().Debug("Restored variable scope.", "proc_name", procName)
	}()

	// Assign Args to Scope
	i.Logger().Debug("Assigning required parameters", "count", numRequired, "proc_name", procName)
	for idx, paramName := range proc.RequiredParams {
		if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
			i.Logger().Error("Failed setting required proc arg", "param_name", paramName, "error", setErr)
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
		i.Logger().Debug("Assigned required parameter", "name", paramName)
	}
	i.Logger().Debug("Assigning optional parameters", "count", numOptional, "proc_name", procName)
	for idx, paramName := range proc.OptionalParams {
		providedArgIndex := numRequired + idx
		valueToSet := interface{}(nil)
		argProvided := false
		if providedArgIndex < numProvided {
			valueToSet = args[providedArgIndex]
			argProvided = true
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			i.Logger().Error("Failed setting optional proc arg", "param_name", paramName, "error", setErr)
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		}
		i.Logger().Debug("Assigned optional parameter", "name", paramName, "provided", argProvided)
	}

	// Execute steps
	result, _, _, err = i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		return nil, err
	}

	// Return Count Validation
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	if result != nil {
		resultValue := reflect.ValueOf(result)
		kind := resultValue.Kind()
		if kind == reflect.Ptr || kind == reflect.Interface {
			if !resultValue.IsNil() {
				resultValue = resultValue.Elem()
				kind = resultValue.Kind()
			} else {
				kind = reflect.Invalid
			}
		}
		if kind == reflect.Slice {
			actualReturnCount = resultValue.Len()
		} else if resultValue.IsValid() {
			actualReturnCount = 1
		}
	}

	if actualReturnCount != expectedReturnCount {
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but returned %d", ErrArgumentMismatch, procName, expectedReturnCount, actualReturnCount)
		i.Logger().Error("Return count mismatch", "proc_name", procName, "expected", expectedReturnCount, "actual", actualReturnCount)
		return nil, err
	}
	i.Logger().Debug("Return count validated", "proc_name", procName, "count", actualReturnCount)

	i.lastCallResult = result
	return result, nil
}

// --- Assumed Definitions (for reference) ---
// type Procedure struct { Name string; RequiredParams []string; OptionalParams []string; ReturnVarNames []string; Steps []Step; Metadata map[string]string }
// type Step interface { GetPos() *Position } // Base interface for all steps
// var ErrProcedureNotFound = errors.New("procedure not found")
// var ErrArgumentMismatch = errors.New("argument count mismatch")
// var ErrCacheObjectNotFound = errors.New("object not found in handle cache")
// var ErrCacheObjectWrongType = errors.New("cached object has wrong type prefix")
// type RuntimeError struct { Code ErrorCode; Message string; Pos *Position }
// type ErrorCode int
// func RegisterCoreTools(registry *ToolRegistry) error { panic("not implemented") }
// func NewToolRegistry() *ToolRegistry { panic("not implemented") }
// func (i *Interpreter) executeSteps(steps []Step, isInErrorHandler bool, activeError *RuntimeError) (result interface{}, stopExecution bool, returned bool, err error) { panic("not implemented") }
// type FileAPI struct { sandboxRoot string; logger logging.Logger }
// func NewFileAPI(sandboxRoot string, logger logging.Logger) *FileAPI { return &FileAPI{sandboxRoot, logger} }
// type SecurityPolicy interface { IsAllowed(op SecurityOperation, target string) bool }
// func DefaultSecurityPolicy() SecurityPolicy { panic("not implemented") }
// type SecurityOperation string

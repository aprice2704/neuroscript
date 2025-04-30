// filename: pkg/core/interpreter.go
// No changes needed from the version in the previous response.
// It already uses `*FileAPI` type and `NewFileAPI(...)` constructor.
// It already has the `FileAPI()` getter method.

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	// FileAPI and NewFileAPI are now defined in file_api.go in this package.
)

// Interpreter holds the state of a running NeuroScript program.
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure // Assumes Procedure is defined (likely in ast.go)
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string // Track current proc for semantic checks
	sandboxDir      string // The initial path given for the sandbox

	toolRegistry *ToolRegistry
	logger       logging.Logger
	objectCache  map[string]interface{} // Cache for handle objects
	llmClient    LLMClient              // Store the interface value directly
	fileAPI      *FileAPI               // The FileAPI instance using the defined type
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

	// Default sandbox directory (can be overridden later if needed)
	defaultSandbox := "."

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default, might be overridden by LLM
		toolRegistry:    NewToolRegistry(),
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       effectiveLLMClient,
		sandboxDir:      defaultSandbox, // Store the original setting
		fileAPI:         nil,            // Initialize as nil first
	}

	// Initialize FileAPI *after* logger and sandboxDir are set
	// Use the new exported constructor from file_api.go
	// Pass the originally configured sandboxDir (NewFileAPI handles defaults/abs path)
	interp.fileAPI = NewFileAPI(interp.sandboxDir, interp.logger) // <<< USE EXPORTED CONSTRUCTOR

	// Initialize default prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// Register core tools
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization", "error", err)
		// Using panic here because core tools failing to register is likely unrecoverable
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("Core tools registered successfully.")
	}
	return interp
}

// --- Getters / Setters ---

// SandboxDir returns the originally configured sandbox directory path.
// Note: The actual sandboxing is handled by FileAPI using its internal absolute path.
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }

func (i *Interpreter) Logger() logging.Logger {
	if i.logger == nil {
		// This should be caught by NewInterpreter, but defensive check.
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

// FileAPI returns the initialized FileAPI instance.
// This provides safe access to the unexported fileAPI field.
func (i *Interpreter) FileAPI() *FileAPI {
	if i.fileAPI == nil {
		// This should be caught by NewInterpreter, but defensive check.
		i.Logger().Error("FATAL: Interpreter.FileAPI() called but fileAPI field is nil!")
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}

// SetSandboxDir updates the sandbox directory AND re-initializes the FileAPI.
// Use with caution after initialization, as it changes the security boundary.
func (i *Interpreter) SetSandboxDir(newSandboxDir string) {
	i.Logger().Warn("Interpreter sandbox directory is being changed post-initialization.", "old", i.sandboxDir, "new", newSandboxDir)
	i.sandboxDir = newSandboxDir
	// Re-initialize FileAPI with the new path
	i.fileAPI = NewFileAPI(i.sandboxDir, i.logger)
	i.Logger().Info("FileAPI re-initialized with new sandbox directory.", "path", i.fileAPI.sandboxRoot)
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
	client := i.llmClient.Client() // Use the interface method
	if client == nil {
		// This is not necessarily an error, could be a different LLM type
		i.Logger().Debug("GenAIClient() called, but the configured LLMClient implementation does not provide a *genai.Client.")
	}
	return client
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

	// Use recover for panics during execution (e.g., nil pointer dereference)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during procedure '%s': %v", procName, r)
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r)
			// Potentially log stack trace here if needed
		}
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
		valueToSet := interface{}(nil) // Default to nil if not provided
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
	// Ensure executeSteps propagates errors correctly
	result, _, _, err = i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		// Error already logged within executeSteps or its called functions
		return nil, err // Return the error directly
	}

	// Return Count Validation
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	var finalResult interface{} // Variable to hold the result in the expected format (single value or slice)

	if result != nil {
		resultValue := reflect.ValueOf(result)
		kind := resultValue.Kind()
		// Handle pointers/interfaces correctly
		if kind == reflect.Ptr || kind == reflect.Interface {
			if !resultValue.IsNil() {
				resultValue = resultValue.Elem()
				kind = resultValue.Kind()
			} else {
				kind = reflect.Invalid // Treat nil pointer/interface as invalid for counting
			}
		}

		if kind == reflect.Slice {
			actualReturnCount = resultValue.Len()
			// If expected 1, but got a slice, maybe wrap it? For now, strict check.
			// Keep the result as the slice for multi-return
			finalResult = result
		} else if resultValue.IsValid() {
			actualReturnCount = 1
			// If expected > 1, this is an error. If expected 1, this is correct.
			// Store the single value.
			finalResult = result
		}
		// If kind is Invalid (e.g., nil interface/pointer), actualReturnCount remains 0.
	}

	// Strict check: number of returns must match declaration exactly
	if actualReturnCount != expectedReturnCount {
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but returned %d", ErrArgumentMismatch, procName, expectedReturnCount, actualReturnCount)
		i.Logger().Error("Return count mismatch", "proc_name", procName, "expected", expectedReturnCount, "actual", actualReturnCount)
		return nil, err
	}
	i.Logger().Debug("Return count validated", "proc_name", procName, "count", actualReturnCount)

	// If we reached here, the return count is correct.
	i.lastCallResult = finalResult // Store the validated result
	return finalResult, nil        // Return the validated result and nil error
}

// filename: pkg/core/interpreter.go
// Last Modified: 2025-05-02 20:25:00 PM PDT // Fix GetHandleValue error wrapping
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
// It initializes with CORE tools by default. Use SetToolRegistry to override.
func NewInterpreter(logger logging.Logger, llmClient LLMClient) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = &coreNoOpLogger{}
		// Optionally log this default assignment if a logger *was* available
		// effectiveLogger.Warn("NewInterpreter: nil logger provided, using internal NoOpLogger.")
	}

	effectiveLLMClient := llmClient
	if effectiveLLMClient == nil {
		effectiveLogger.Warn("NewInterpreter: nil LLMClient provided, using internal NoOpLLMClient.")
		effectiveLLMClient = NewNoOpLLMClient(effectiveLogger)
	}

	// Default sandbox directory (can be overridden later if needed)
	defaultSandbox := "."
	// Initialize FileAPI early as other parts might depend on it
	fileAPI := NewFileAPI(defaultSandbox, effectiveLogger) // Pass logger

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16,                // Default, might be overridden by LLM
		toolRegistry:    NewToolRegistry(), // Create default registry first
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       effectiveLLMClient,
		sandboxDir:      defaultSandbox, // Store the original setting
		fileAPI:         fileAPI,        // Assign initialized FileAPI
	}

	// Initialize default prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// Register core tools into the default registry
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization", "error", err)
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("Core tools registered successfully.")
	}
	return interp
}

// --- Getters / Setters ---

// SandboxDir returns the originally configured sandbox directory path.
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }

func (i *Interpreter) Logger() logging.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

// FileAPI returns the initialized FileAPI instance.
func (i *Interpreter) FileAPI() *FileAPI {
	if i.fileAPI == nil {
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}

// SetSandboxDir updates the sandbox directory AND re-initializes the FileAPI.
func (i *Interpreter) SetSandboxDir(newSandboxDir string) { // REMOVED error return
	i.Logger().Warn("Interpreter sandbox directory is being changed post-initialization.", "old", i.sandboxDir, "new", newSandboxDir)
	i.sandboxDir = newSandboxDir
	// Re-initialize FileAPI with the new path
	i.fileAPI = NewFileAPI(i.sandboxDir, i.logger)
	i.Logger().Info("FileAPI re-initialized with new sandbox directory.", "path", i.fileAPI.sandboxRoot)
	// REMOVED: return nil
}

// SetToolRegistry replaces the interpreter's current tool registry.
// This is useful for test setups that need to inject a registry
// containing extended tools after the interpreter is created.
func (i *Interpreter) SetToolRegistry(registry *ToolRegistry) {
	if registry == nil {
		i.logger.Error("Attempted to set a nil tool registry. Ignoring.")
		return
	}
	i.logger.Info("Replacing interpreter's tool registry.")
	i.toolRegistry = registry
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
		i.Logger().Error("ToolRegistry accessed but is nil!") // Should not happen with constructor logic
		panic("FATAL: Interpreter toolRegistry is nil")
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

// RegisterHandle stores an object and returns a unique handle string (TypePrefix::UUID).
func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", fmt.Errorf("%w: handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("%w: handle type prefix '%s' cannot contain separator '%s'", ErrInvalidArgument, typePrefix, handleSeparator)
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

// GetHandleValue retrieves the value associated with a handle string, validating format and type prefix.
func (i *Interpreter) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	if expectedTypePrefix == "" {
		return nil, fmt.Errorf("%w: expected handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", ErrInvalidArgument)
	}

	parts := strings.SplitN(handle, handleSeparator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("%w: invalid handle format: expected 'Type%sUUID', got '%s'", ErrInvalidArgument, handleSeparator, handle)
	}
	actualPrefix := parts[0]

	if actualPrefix != expectedTypePrefix {
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s' (full handle: '%s')", ErrHandleWrongType, expectedTypePrefix, actualPrefix, handle)
	}

	if i.objectCache == nil {
		i.Logger().Error("GetHandleValue called but objectCache is nil.")
		return nil, fmt.Errorf("%w: internal error: object cache is not initialized", ErrInternalTool)
	}

	obj, found := i.objectCache[handle]
	if !found {
		// Return specific error for not found
		return nil, fmt.Errorf("%w: handle '%s'", ErrNotFound, handle)
	}

	i.Logger().Debug("Retrieved handle", "handle", handle, "expected_type", expectedTypePrefix)
	return obj, nil
}

// RemoveHandle removes an object from the handle cache. Returns true if found and removed.
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
// (RunProcedure and helpers remain the same)
func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
	originalProcName := i.currentProcName
	i.Logger().Info("Running procedure", "name", procName, "arg_count", len(args))
	i.currentProcName = procName
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during procedure '%s': %v", procName, r)
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r)
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
	result, _, _, err = i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		// Wrap error if it occurred during step execution
		return nil, fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
	}
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	var finalResult interface{}
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
			finalResult = result
		} else if resultValue.IsValid() {
			actualReturnCount = 1
			finalResult = result
		}
	}
	if actualReturnCount != expectedReturnCount {
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but returned %d", ErrArgumentMismatch, procName, expectedReturnCount, actualReturnCount)
		i.Logger().Error("Return count mismatch", "proc_name", procName, "expected", expectedReturnCount, "actual", actualReturnCount)
		return nil, err
	}
	i.Logger().Debug("Return count validated", "proc_name", procName, "count", actualReturnCount)
	i.lastCallResult = finalResult
	return finalResult, nil
}

// --- Internal NoOp Implementations (if not provided externally by core) ---

var _ logging.Logger = (*coreNoOpLogger)(nil) // Ensure it implements the interface

var _ LLMClient = (*coreNoOpLLMClient)(nil) // Ensure it implements the interface

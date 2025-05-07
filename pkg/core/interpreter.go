// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Fix NewNoOpLLMClient call in NewInterpreter.
// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai" // Used by GenAIClient() method
	"github.com/google/uuid"
)

// Interpreter holds the state of a running NeuroScript program.
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	sandboxDir      string

	toolRegistry    *ToolRegistry
	logger          logging.Logger
	objectCache     map[string]interface{}
	llmClient       LLMClient // Should be initialized by NewInterpreter
	fileAPI         *FileAPI
	aiWorkerManager *AIWorkerManager
}

// --- Constants ---
const handleSeparator = "::"

// --- Constructor ---

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger logging.Logger, llmClient LLMClient, sandboxDir string, initialVars map[string]interface{}) (*Interpreter, error) {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = &coreNoOpLogger{} // From utils.go
		// No log message here as logger itself is NoOp
	}

	effectiveLLMClient := llmClient
	if effectiveLLMClient == nil {
		effectiveLogger.Warn("NewInterpreter: nil LLMClient provided. Initializing with a NoOp LLMClient via NewLLMClient factory.")
		// Parameters for NewLLMClient: apiKey, apiHost, modelID, logger, enabled
		// For a NoOp client, apiKey, apiHost, and modelID can be empty as they won't be used.
		effectiveLLMClient = NewLLMClient("", "", "", effectiveLogger, false) // <<< FIXED
	}

	cleanSandboxDir := "."
	if sandboxDir != "" {
		absPath, err := filepath.Abs(sandboxDir)
		if err != nil {
			effectiveLogger.Errorf("Failed to get absolute path for provided sandbox directory: %v (path: %s)", err, sandboxDir)
			return nil, fmt.Errorf("invalid sandbox directory '%s': %w", sandboxDir, err)
		}
		cleanSandboxDir = filepath.Clean(absPath)
		effectiveLogger.Infof("Interpreter sandbox directory set to: %s", cleanSandboxDir)
	} else {
		effectiveLogger.Warn("No sandbox directory provided, using default '.'")
	}

	fileAPI := NewFileAPI(cleanSandboxDir, effectiveLogger)

	vars := make(map[string]interface{})
	vars["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	vars["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute
	if initialVars != nil {
		for k, v := range initialVars {
			vars[k] = v
		}
	}

	interp := &Interpreter{
		variables:       vars,
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16,
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       effectiveLLMClient, // Assign the resolved LLM client
		sandboxDir:      cleanSandboxDir,
		fileAPI:         fileAPI,
		// aiWorkerManager is initialized by RegisterAIWorkerTools if needed
	}

	interp.toolRegistry = NewToolRegistry(interp)

	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Errorf("FATAL: Failed to register core tools during interpreter initialization: %v", err)
		return nil, fmt.Errorf("FATAL: failed to register core tools: %w", err)
	}
	effectiveLogger.Debug("Core tools registered successfully.")

	// AI Worker tools are registered separately, typically by the application
	// or if a specific toolset registration function is called.
	// For example, RegisterAIWorkerTools(interp) could be called here if always needed.

	return interp, nil
}

// --- Getters / Setters ---
// (Remaining methods like SetAIWorkerManager, SandboxDir, Logger, FileAPI, etc. unchanged from version 0.0.3)
func (i *Interpreter) SetAIWorkerManager(manager *AIWorkerManager) {
	i.aiWorkerManager = manager
}
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }

func (i *Interpreter) Logger() logging.Logger {
	if i.logger == nil {
		// This should not happen if NewInterpreter ensures a logger.
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}
func (i *Interpreter) FileAPI() *FileAPI {
	if i.fileAPI == nil {
		// This should not happen if NewInterpreter ensures a FileAPI.
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}

func (i *Interpreter) SetSandboxDir(newSandboxDir string) error {
	i.Logger().Debug("Attempting to set new sandbox directory.", "new_path", newSandboxDir)
	var cleanNewSandboxDir string
	if newSandboxDir == "" {
		cleanNewSandboxDir = "."
		i.Logger().Warn("SetSandboxDir called with empty path, using default '.'")
	} else {
		absPath, err := filepath.Abs(newSandboxDir)
		if err != nil {
			i.Logger().Error("Failed to get absolute path for new sandbox directory", "path", newSandboxDir, "error", err)
			return fmt.Errorf("invalid sandbox directory '%s': %w", newSandboxDir, err)
		}
		cleanNewSandboxDir = filepath.Clean(absPath)
	}
	if i.sandboxDir != cleanNewSandboxDir {
		i.Logger().Info("Interpreter sandbox directory changed.", "old", i.sandboxDir, "new", cleanNewSandboxDir)
		i.sandboxDir = cleanNewSandboxDir
		// Re-initialize FileAPI with the new sandbox directory
		i.fileAPI = NewFileAPI(i.sandboxDir, i.logger) // Pass logger
		i.Logger().Info("FileAPI re-initialized with new sandbox directory.", "path", i.fileAPI.sandboxRoot)
	} else {
		i.Logger().Debug("New sandbox directory is the same as the current one. No change made.", "path", cleanNewSandboxDir)
	}
	return nil
}
func (i *Interpreter) SetToolRegistry(registry *ToolRegistry) {
	if registry == nil {
		i.logger.Error("Attempted to set a nil tool registry. Ignoring.")
		return
	}
	if registry.interpreter != i {
		// This indicates a potentially problematic setup, reassigning interpreter.
		i.logger.Warn("Setting tool registry that belongs to a different interpreter instance. Re-assigning interpreter pointer.")
		registry.interpreter = i // Ensure registry points back to this interpreter
	}
	i.logger.Info("Replacing interpreter's tool registry.")
	i.toolRegistry = registry
}

func (i *Interpreter) SetVariable(name string, value interface{}) error {
	if i.variables == nil {
		// This should be initialized in NewInterpreter, but defensive.
		i.variables = make(map[string]interface{})
		i.Logger().Warn("Interpreter variables map was nil, re-initialized.")
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	i.Logger().Debug("Set variable", "name", name, "type", fmt.Sprintf("%T", value))
	return nil
}

func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
	if i.variables == nil {
		i.Logger().Warn("Interpreter variables map is nil during GetVariable.")
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}

func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		// This should be initialized in NewInterpreter.
		i.Logger().Error("ToolRegistry accessed but is nil!")
		panic("FATAL: Interpreter toolRegistry is nil") // Or return an error
	}
	return i.toolRegistry
}

// GenAIClient provides access to the underlying *genai.Client if available.
// Returns nil if the configured LLMClient does not expose a *genai.Client
// or if LLMClient itself is nil.
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		i.Logger().Warn("GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	// Delegate to the LLMClient's Client() method
	client := i.llmClient.Client()
	if client == nil {
		i.Logger().Debug("GenAIClient() called, but the configured LLMClient implementation does not provide a *genai.Client (it might be a NoOp or a different provider type).")
	}
	return client
}
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
		i.Logger().Warn("Interpreter knownProcedures map was nil, re-initialized.")
	}
	if proc.Metadata == nil { // Ensure metadata map exists
		proc.Metadata = make(map[string]string)
	}
	if proc.Name == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("%w: '%s'", ErrProcedureExists, proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	i.Logger().Debug("Added procedure definition.", "name", proc.Name)
	return nil
}

func (i *Interpreter) KnownProcedures() map[string]Procedure {
	if i.knownProcedures == nil {
		return make(map[string]Procedure) // Return empty map if nil
	}
	return i.knownProcedures
}

func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32) // Initialize if nil
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }

// --- Handle Management ---
func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", fmt.Errorf("%w: handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("%w: handle type prefix '%s' cannot contain separator '%s'", ErrInvalidArgument, typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
		i.Logger().Warn("Interpreter objectCache was nil, re-initialized.")
	}
	handleIDPart := uuid.NewString() // Just the UUID part
	fullHandle := fmt.Sprintf("%s%s%s", typePrefix, handleSeparator, handleIDPart)
	i.objectCache[fullHandle] = obj
	i.Logger().Debug("Registered handle", "handle", fullHandle, "type", typePrefix)
	return fullHandle, nil
}

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
		// Check if the expected prefix is a more generic one, e.g. "Handle" vs "FileHandle"
		// For now, require exact match or a clear policy.
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s' (full handle: '%s')", ErrHandleWrongType, expectedTypePrefix, actualPrefix, handle)
	}

	if i.objectCache == nil {
		i.Logger().Error("GetHandleValue called but objectCache is nil.")
		return nil, fmt.Errorf("%w: internal error: object cache is not initialized", ErrInternal)
	}
	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s' (prefix '%s')", ErrHandleNotFound, handle, expectedTypePrefix)
	}
	i.Logger().Debug("Retrieved handle", "handle", handle, "expected_type", expectedTypePrefix)
	return obj, nil
}

func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		i.Logger().Warn("RemoveHandle called but objectCache is nil.")
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
		if r := recover(); r != nil {
			// Ensure err is a RuntimeError
			err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("panic occurred during procedure '%s': %v", procName, r), errors.New("panic"))
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r, "error", err)
			result = nil // Ensure result is nil on panic
		}
		i.currentProcName = originalProcName // Restore original proc name
		logArgsMap := map[string]interface{}{
			"proc_name":          procName,
			"restored_proc_name": i.currentProcName,
			"result_type":        fmt.Sprintf("%T", result), // Get type of result
			"error":              err,                       // Log the error object itself
		}
		// Avoid logging potentially large result values directly.
		i.Logger().Info("Finished procedure.", "details", logArgsMap)
	}()

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		i.Logger().Error("Procedure definition not found", "name", procName)
		return nil, err
	}

	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments ('needs'), but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		i.Logger().Error("Argument count mismatch (too few)", "proc_name", procName, "required", numRequired, "provided", numProvided)
		return nil, err
	}
	if numProvided > numTotalParams && !proc.Variadic { // Only warn if not variadic
		i.Logger().Warn("Procedure called with extra arguments.", "proc_name", procName, "provided", numProvided, "defined_max", numTotalParams)
		// Do not error out, just ignore extra args if not variadic.
	}

	// Scope management
	procScope := make(map[string]interface{})
	if i.variables != nil { // Copy global/parent scope
		for k, v := range i.variables {
			procScope[k] = v
		}
	}
	originalScope := i.variables
	i.variables = procScope
	defer func() {
		i.variables = originalScope // Restore parent scope
		i.Logger().Debug("Restored parent variable scope.", "proc_name", procName)
	}()

	i.Logger().Debug("Assigning required parameters", "count", numRequired, "proc_name", procName)
	for idx, paramName := range proc.RequiredParams {
		if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
			i.Logger().Error("Failed setting required proc arg", "param_name", paramName, "error", setErr)
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
	}

	i.Logger().Debug("Assigning optional parameters", "count", numOptional, "proc_name", procName)
	for idx, paramSpec := range proc.OptionalParams { // Assuming OptionalParams is []ParamSpec
		paramName := paramSpec.Name
		valueToSet := paramSpec.DefaultValue // Use default value from ParamSpec
		argProvided := false
		providedArgIndex := numRequired + idx
		if providedArgIndex < numProvided {
			valueToSet = args[providedArgIndex]
			argProvided = true
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			i.Logger().Error("Failed setting optional proc arg", "param_name", paramName, "error", setErr)
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		}
		i.Logger().Debug("Assigned optional parameter", "name", paramName, "provided", argProvided, "value_type", fmt.Sprintf("%T", valueToSet))
	}

	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numRequired+numOptional {
		variadicArgs := []interface{}{}
		startIndex := numRequired + numOptional
		if startIndex < numProvided {
			variadicArgs = append(variadicArgs, args[startIndex:]...)
		}
		if setErr := i.SetVariable(proc.VariadicParamName, variadicArgs); setErr != nil {
			i.Logger().Error("Failed setting variadic proc arg", "param_name", proc.VariadicParamName, "error", setErr)
			return nil, fmt.Errorf("failed to set variadic parameter '%s': %w", proc.VariadicParamName, setErr)
		}
		i.Logger().Debug("Assigned variadic parameter", "name", proc.VariadicParamName, "num_args", len(variadicArgs))
	}

	// Execute steps
	result, _, _, err = i.executeSteps(proc.Steps, false, nil) // procCtx is nil for top-level proc call
	if err != nil {
		// If it's already a RuntimeError, don't re-wrap it unless adding more context.
		if _, ok := err.(*RuntimeError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
	}

	// Return value handling
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	var finalResult interface{} // This will hold the result to be returned

	if result != nil {
		resultValue := reflect.ValueOf(result)
		kind := resultValue.Kind()

		// Dereference if pointer or interface to get actual kind for slice check
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
			finalResult = result // Return the slice itself
		} else if resultValue.IsValid() { // Check if it's a valid non-nil single value
			actualReturnCount = 1
			finalResult = result
		}
		// If result was nil, actualReturnCount remains 0, finalResult remains nil
	}
	// If result is nil, actualReturnCount is 0.

	if actualReturnCount != expectedReturnCount {
		// Special case: if 0 expected and 0 actual (result was nil), it's okay.
		if !(expectedReturnCount == 0 && actualReturnCount == 0) {
			err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but evaluation yielded %d", ErrReturnMismatch, procName, expectedReturnCount, actualReturnCount)
			i.Logger().Error("Return count mismatch", "proc_name", procName, "expected", expectedReturnCount, "actual", actualReturnCount)
			return nil, err
		}
	}

	i.Logger().Debug("Return count validated", "proc_name", procName, "count", actualReturnCount)
	i.lastCallResult = finalResult // Store the result that matches expected count (or nil if 0 expected)
	return finalResult, nil
}

// (coreNoOpLogger and coreInternalNoOpLLMClient are defined in utils.go and llm.go respectively)
// Ensure var _ declarations are correct after types are defined.
// var _ logging.Logger = (*coreNoOpLogger)(nil) // This belongs in utils.go
// var _ LLMClient = (*coreInternalNoOpLLMClient)(nil) // This belongs in llm.go

// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Define ErrProcedureExists, ErrReturnMismatch
// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"path/filepath" // Added import
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
func NewInterpreter(logger logging.Logger, llmClient LLMClient, sandboxDir string, initialVars map[string]interface{}) (*Interpreter, error) {
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

	// Validate and clean sandboxDir
	cleanSandboxDir := "." // Default
	if sandboxDir != "" {
		absPath, err := filepath.Abs(sandboxDir)
		if err != nil {
			effectiveLogger.Error("Failed to get absolute path for provided sandbox directory", "path", sandboxDir, "error", err)
			return nil, fmt.Errorf("invalid sandbox directory '%s': %w", sandboxDir, err)
		}
		cleanSandboxDir = filepath.Clean(absPath)
		effectiveLogger.Info("Interpreter sandbox directory set", "path", cleanSandboxDir)
	} else {
		effectiveLogger.Warn("No sandbox directory provided, using default '.'")
	}

	// Initialize FileAPI early as other parts might depend on it
	fileAPI := NewFileAPI(cleanSandboxDir, effectiveLogger) // Pass logger

	// Initialize variables map, start with defaults then merge initialVars
	vars := make(map[string]interface{})
	vars["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	vars["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute
	if initialVars != nil {
		for k, v := range initialVars {
			vars[k] = v
		}
	}

	interp := &Interpreter{
		variables:       vars, // Use initialized map
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default, might be overridden by LLM
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       effectiveLLMClient,
		sandboxDir:      cleanSandboxDir, // Store the final clean path
		fileAPI:         fileAPI,         // Assign initialized FileAPI
	}

	// Create the tool registry, passing the interpreter instance
	interp.toolRegistry = NewToolRegistry(interp) // Pass interp

	// Register core tools using the init-based mechanism
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization", "error", err)
		return nil, fmt.Errorf("FATAL: failed to register core tools: %w", err)
	} else {
		effectiveLogger.Debug("Core tools registered successfully.")
	}

	return interp, nil // Return the interpreter instance and nil error
}

// --- Getters / Setters ---

func (i *Interpreter) SandboxDir() string { return i.sandboxDir }
func (i *Interpreter) Logger() logging.Logger {
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
	// ... (implementation unchanged from previous correction) ...
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
		i.fileAPI = NewFileAPI(i.sandboxDir, i.logger)
		i.Logger().Info("FileAPI re-initialized with new sandbox directory.", "path", i.fileAPI.sandboxRoot)
	} else {
		i.Logger().Debug("New sandbox directory is the same as the current one. No change made.", "path", cleanNewSandboxDir)
	}
	return nil
}
func (i *Interpreter) SetToolRegistry(registry *ToolRegistry) {
	// ... (implementation unchanged from previous correction) ...
	if registry == nil {
		i.logger.Error("Attempted to set a nil tool registry. Ignoring.")
		return
	}
	if registry.interpreter != i {
		i.logger.Warn("Setting tool registry that belongs to a different interpreter instance. Re-assigning.")
		registry.interpreter = i
	}
	i.logger.Info("Replacing interpreter's tool registry.")
	i.toolRegistry = registry
}
func (i *Interpreter) SetVariable(name string, value interface{}) error {
	// ... (implementation unchanged from previous correction) ...
	if i.variables == nil {
		i.variables = make(map[string]interface{})
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	i.Logger().Debug("Set variable", "name", name, "type", fmt.Sprintf("%T", value))
	return nil
}
func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
	// ... (implementation unchanged from previous correction) ...
	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	// ... (implementation unchanged from previous correction) ...
	if i.toolRegistry == nil {
		i.Logger().Error("ToolRegistry accessed but is nil!")
		panic("FATAL: Interpreter toolRegistry is nil")
	}
	return i.toolRegistry
}
func (i *Interpreter) GenAIClient() *genai.Client {
	// ... (implementation unchanged from previous correction) ...
	if i.llmClient == nil {
		i.Logger().Warn("GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	client := i.llmClient.Client()
	if client == nil {
		i.Logger().Debug("GenAIClient() called, but the configured LLMClient implementation does not provide a *genai.Client.")
	}
	return client
}
func (i *Interpreter) AddProcedure(proc Procedure) error {
	// ... uses ErrProcedureExists ... (implementation unchanged from previous correction) ...
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if proc.Metadata == nil {
		proc.Metadata = make(map[string]string)
	}
	if proc.Name == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("%w: '%s'", ErrProcedureExists, proc.Name) // Use defined error
	}
	i.knownProcedures[proc.Name] = proc
	i.Logger().Debug("Added procedure definition.", "name", proc.Name)
	return nil
}
func (i *Interpreter) KnownProcedures() map[string]Procedure {
	// ... (implementation unchanged from previous correction) ...
	if i.knownProcedures == nil {
		return make(map[string]Procedure)
	}
	return i.knownProcedures
}
func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	// ... (implementation unchanged from previous correction) ...
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }

// --- Handle Management ---
// ... (RegisterHandle, GetHandleValue, RemoveHandle implementations unchanged from previous correction) ...
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
		return nil, fmt.Errorf("%w: internal error: object cache is not initialized", ErrInternal)
	} // Use ErrInternal
	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s' (prefix '%s')", ErrHandleNotFound, handle, expectedTypePrefix)
	} // Use ErrHandleNotFound
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
	// ... uses ErrProcedureNotFound, ErrArgumentMismatch, ErrReturnMismatch ...
	// ... (implementation unchanged from previous correction) ...
	originalProcName := i.currentProcName
	i.Logger().Info("Running procedure", "name", procName, "arg_count", len(args))
	i.currentProcName = procName
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during procedure '%s': %v", procName, r)
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r)
			result = nil
		}
		i.currentProcName = originalProcName
		logArgs := []any{"proc_name", procName, "restored_proc_name", i.currentProcName, "result_type", fmt.Sprintf("%T", result), "error", err}
		i.Logger().Info("Finished procedure.", logArgs...)
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
	if numProvided > numTotalParams {
		i.Logger().Warn("Procedure called with extra arguments.", "proc_name", procName, "provided", numProvided, "defined_max", numTotalParams)
	}
	procScope := make(map[string]interface{})
	if i.variables != nil {
		for k, v := range i.variables {
			procScope[k] = v
		}
	}
	originalScope := i.variables
	i.variables = procScope
	defer func() {
		i.variables = originalScope
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
		var runtimeErr *RuntimeError
		if errors.As(err, &runtimeErr) {
			return nil, err
		}
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
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but evaluation yielded %d", ErrReturnMismatch, procName, expectedReturnCount, actualReturnCount)
		i.Logger().Error("Return count mismatch", "proc_name", procName, "expected", expectedReturnCount, "actual", actualReturnCount)
		return nil, err
	} // Use defined error
	i.Logger().Debug("Return count validated", "proc_name", procName, "count", actualReturnCount)
	i.lastCallResult = finalResult
	return finalResult, nil
}

// --- Internal NoOp Implementations (if not provided externally by core) ---
var _ logging.Logger = (*coreNoOpLogger)(nil)
var _ LLMClient = (*coreNoOpLLMClient)(nil)

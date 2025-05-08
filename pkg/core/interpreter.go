// NeuroScript Version: 0.3.1
// File version: 0.0.15 // Correct usage of ErrorCodeArgMismatch and ErrorCodeToolExecutionFailed.
// nlines: 442
// risk_rating: HIGH
// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time" // Needed for ExecuteTool rate limiting

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
	LibPaths        []string

	toolRegistry    *toolRegistryImpl // Concrete struct from tools_registry.go
	logger          logging.Logger
	objectCache     map[string]interface{}
	llmClient       LLMClient
	fileAPI         *FileAPI
	aiWorkerManager *AIWorkerManager

	// Rate limiting for tool execution
	toolCallTimestamps map[string][]time.Time
	rateLimitCount     int
	rateLimitDuration  time.Duration
}

// --- Constants ---
const handleSeparator = "::"

// --- Constructor ---

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger logging.Logger, llmClient LLMClient, sandboxDir string, initialVars map[string]interface{}, libPaths []string) (*Interpreter, error) {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = &coreNoOpLogger{}
	}

	effectiveLLMClient := llmClient
	if effectiveLLMClient == nil {
		effectiveLogger.Warn("NewInterpreter: nil LLMClient provided. Initializing with a NoOp LLMClient.")
		effectiveLLMClient = NewLLMClient("", "", "", effectiveLogger, false)
	}

	cleanSandboxDir := "."
	if sandboxDir != "" {
		absPath, err := filepath.Abs(sandboxDir)
		if err != nil {
			effectiveLogger.Errorf("Failed to get absolute path for sandbox directory: %v (path: %s)", err, sandboxDir)
			return nil, fmt.Errorf("invalid sandbox directory '%s': %w", sandboxDir, err)
		}
		cleanSandboxDir = filepath.Clean(absPath)
		effectiveLogger.Infof("Interpreter sandbox directory set to: %s", cleanSandboxDir)
	} else {
		effectiveLogger.Warn("No sandbox directory provided, using default '.'")
	}

	fileAPI, _ := NewFileAPI(cleanSandboxDir, effectiveLogger)

	vars := make(map[string]interface{})
	vars["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	vars["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute
	if initialVars != nil {
		for k, v := range initialVars {
			vars[k] = v
		}
	}

	effectiveLibPaths := libPaths
	if effectiveLibPaths == nil {
		effectiveLibPaths = []string{}
	}

	interp := &Interpreter{
		variables:          vars,
		knownProcedures:    make(map[string]Procedure),
		vectorIndex:        make(map[string][]float32),
		embeddingDim:       16,
		logger:             effectiveLogger,
		objectCache:        make(map[string]interface{}),
		llmClient:          effectiveLLMClient,
		sandboxDir:         cleanSandboxDir,
		fileAPI:            fileAPI,
		LibPaths:           effectiveLibPaths,
		toolCallTimestamps: make(map[string][]time.Time), // Initialize rate limit map
		rateLimitCount:     10,                           // Default: 10 calls
		rateLimitDuration:  time.Minute,                  // Default: per minute
	}

	interp.toolRegistry = NewToolRegistry(interp) // NewToolRegistry returns *toolRegistryImpl

	// RegisterCoreTools expects an argument that satisfies the ToolRegistry INTERFACE.
	// The *Interpreter instance (interp) itself implements this interface.
	if err := RegisterCoreTools(interp); err != nil {
		effectiveLogger.Errorf("FATAL: Failed to register core tools during interpreter initialization: %v", err)
		return nil, fmt.Errorf("FATAL: failed to register core tools: %w", err)
	}
	effectiveLogger.Debug("Core tools registered successfully.")

	return interp, nil
}

// --- ToolRegistry Interface Compliance ---

// ToolRegistry returns the interpreter itself, as *Interpreter implements the ToolRegistry interface.
func (i *Interpreter) ToolRegistry() ToolRegistry {
	return i
}

// RegisterTool delegates to the internal toolRegistry field.
func (i *Interpreter) RegisterTool(impl ToolImplementation) error {
	if i.toolRegistry == nil {
		i.logger.Error("RegisterTool called on interpreter with nil internal toolRegistry field.")
		return errors.New("internal error: interpreter's tool registry field is not initialized")
	}
	return i.toolRegistry.RegisterTool(impl)
}

// GetTool delegates to the internal toolRegistry field.
func (i *Interpreter) GetTool(name string) (ToolImplementation, bool) {
	if i.toolRegistry == nil {
		i.logger.Error("GetTool called on interpreter with nil internal toolRegistry field.")
		return ToolImplementation{}, false
	}
	return i.toolRegistry.GetTool(name)
}

// ListTools delegates to the internal toolRegistry field.
func (i *Interpreter) ListTools() []ToolSpec {
	if i.toolRegistry == nil {
		i.logger.Error("ListTools called on interpreter with nil internal toolRegistry field.")
		return []ToolSpec{}
	}
	return i.toolRegistry.ListTools()
}

// ExecuteTool retrieves and executes a registered tool by name.
// It handles argument validation, rate limiting, and execution.
// This method makes *Interpreter satisfy the ToolRegistry interface.
func (i *Interpreter) ExecuteTool(toolName string, args map[string]interface{}) (interface{}, error) {
	i.logger.Debug("Attempting to execute tool", "tool_name", toolName, "args_count", len(args))

	// --- Rate Limiting Check ---
	now := time.Now()
	if i.rateLimitCount > 0 && i.rateLimitDuration > 0 {
		timestamps := i.toolCallTimestamps[toolName]
		// Remove timestamps older than the duration
		validTimestamps := []time.Time{}
		cutoff := now.Add(-i.rateLimitDuration)
		for _, ts := range timestamps {
			if ts.After(cutoff) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		// Check if limit exceeded
		if len(validTimestamps) >= i.rateLimitCount {
			err := NewRuntimeError(ErrorCodeRateLimited,
				fmt.Sprintf("tool '%s' rate limit exceeded (%d calls per %s)", toolName, i.rateLimitCount, i.rateLimitDuration.String()),
				ErrRateLimited)
			i.logger.Warn("Tool execution rate limited", "tool_name", toolName, "limit", i.rateLimitCount, "duration", i.rateLimitDuration)
			return nil, err
		}
		// Add current timestamp and update map
		validTimestamps = append(validTimestamps, now)
		i.toolCallTimestamps[toolName] = validTimestamps
	}
	// --- End Rate Limiting Check ---

	impl, found := i.GetTool(toolName) // Use the GetTool method which delegates
	if !found {
		i.logger.Error("Tool not found during execution attempt", "tool_name", toolName)
		return nil, NewRuntimeError(ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), ErrToolNotFound)
	}

	// Validate provided arguments against the tool's specification
	validatedArgs := make([]interface{}, len(impl.Spec.Args))
	providedArgsSet := make(map[string]bool)
	for k := range args {
		providedArgsSet[k] = true
	}

	for idx, argSpec := range impl.Spec.Args {
		value, provided := args[argSpec.Name]
		if !provided {
			if argSpec.Required {
				i.logger.Error("Required argument missing for tool", "tool_name", toolName, "arg_name", argSpec.Name)
				// <<< FIXED: Use ErrorCodeArgMismatch from errors.go
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': missing required argument '%s'", toolName, argSpec.Name), ErrArgumentMismatch)
			}
			// Use default value if optional and not provided (defaults are typically nil/zero for now)
			// If default values were specified in ArgSpec, we'd use them here.
			validatedArgs[idx] = nil // Or appropriate zero value based on argSpec.Type
			i.logger.Debug("Optional argument not provided, using default (nil)", "tool_name", toolName, "arg_name", argSpec.Name)
		} else {
			// Basic type checking (can be expanded)
			// This is simplified; a more robust check would use reflect and handle type conversions
			// err := checkType(argSpec.Type, value)
			// if err != nil {
			//     i.logger.Error("Argument type mismatch for tool", "tool_name", toolName, "arg_name", argSpec.Name, "expected_type", argSpec.Type, "actual_type", fmt.Sprintf("%T", value), "error", err)
			// 	   return nil, NewRuntimeError(ErrorCodeTypeMismatch, fmt.Sprintf("tool '%s': argument '%s' type mismatch: %v", toolName, argSpec.Name, err), ErrTypeMismatch)
			// }
			validatedArgs[idx] = value
			delete(providedArgsSet, argSpec.Name) // Mark as used
		}
	}

	// Check for extraneous arguments
	if len(providedArgsSet) > 0 {
		extraArgs := []string{}
		for name := range providedArgsSet {
			extraArgs = append(extraArgs, name)
		}
		i.logger.Warn("Extraneous arguments provided to tool", "tool_name", toolName, "extra_args", strings.Join(extraArgs, ", "))
		// Decide whether to error out or just ignore extra args. Ignoring for now.
		// return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': extraneous arguments provided: %s", toolName, strings.Join(extraArgs, ", ")), ErrArgumentMismatch)
	}

	// Execute the tool function
	i.logger.Info("Executing tool function", "tool_name", toolName)
	result, err := impl.Func(i, validatedArgs) // Pass interpreter and validated args
	if err != nil {
		// Wrap error if it's not already a RuntimeError
		if _, ok := err.(*RuntimeError); !ok {
			i.logger.Error("Tool execution failed with non-runtime error", "tool_name", toolName, "error", err)
			// <<< FIXED: Use ErrorCodeToolExecutionFailed from errors.go (added in previous step)
			return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed: %v", toolName, err), err) // Wrap original error
		}
		// It's already a RuntimeError, return as is
		i.logger.Error("Tool execution failed", "tool_name", toolName, "error", err)
		return nil, err
	}

	i.logger.Info("Tool execution successful", "tool_name", toolName, "result_type", fmt.Sprintf("%T", result))
	return result, nil
}

// --- Getters / Setters (existing methods) ---

func (i *Interpreter) SetAIWorkerManager(manager *AIWorkerManager) {
	i.aiWorkerManager = manager
}

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
		i.fileAPI, _ = NewFileAPI(i.sandboxDir, i.logger)
		i.Logger().Info("FileAPI re-initialized with new sandbox directory.", "path", i.fileAPI.sandboxRoot)
	} else {
		i.Logger().Debug("New sandbox directory is the same as the current one. No change made.", "path", cleanNewSandboxDir)
	}
	return nil
}

func (i *Interpreter) SetInternalToolRegistry(registry *toolRegistryImpl) {
	if registry == nil {
		i.logger.Error("Attempted to set a nil internal toolRegistryImpl. Ignoring.")
		return
	}
	if registry.interpreter != i {
		i.logger.Warn("Setting internal toolRegistryImpl that points to a different interpreter. Re-assigning its interpreter pointer.")
		registry.interpreter = i
	}
	i.logger.Info("Replacing interpreter's internal toolRegistryImpl.")
	i.toolRegistry = registry
}

func (i *Interpreter) SetVariable(name string, value interface{}) error {
	if i.variables == nil {
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

func (i *Interpreter) InternalToolRegistry() *toolRegistryImpl {
	if i.toolRegistry == nil {
		i.Logger().Error("InternalToolRegistry (*toolRegistryImpl) accessed but is nil!")
		panic("FATAL: Interpreter's internal toolRegistry field is nil")
	}
	return i.toolRegistry
}

func (i *Interpreter) GenAIClient() *genai.Client {
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
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
		i.Logger().Warn("Interpreter knownProcedures map was nil, re-initialized.")
	}
	if proc.Metadata == nil {
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
		return "", fmt.Errorf("%w: handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("%w: handle type prefix '%s' cannot contain separator '%s'", ErrInvalidArgument, typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
		i.Logger().Warn("Interpreter objectCache was nil, re-initialized.")
	}
	handleIDPart := uuid.NewString()
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
			err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("panic occurred during procedure '%s': %v", procName, r), errors.New("panic"))
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r, "error", err)
			result = nil
		}
		i.currentProcName = originalProcName
		logArgsMap := map[string]interface{}{
			"proc_name":          procName,
			"restored_proc_name": i.currentProcName,
			"result_type":        fmt.Sprintf("%T", result),
			"error":              err,
		}
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
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments, but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		i.Logger().Error("Argument count mismatch (too few)", "proc_name", procName, "required", numRequired, "provided", numProvided)
		return nil, err
	}
	if numProvided > numTotalParams && !proc.Variadic {
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

	for idx, paramName := range proc.RequiredParams {
		if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
	}

	for idx, paramSpec := range proc.OptionalParams {
		paramName := paramSpec.Name
		valueToSet := paramSpec.DefaultValue
		if (numRequired + idx) < numProvided {
			valueToSet = args[numRequired+idx]
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		}
	}

	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgs := args[numTotalParams:]
		if setErr := i.SetVariable(proc.VariadicParamName, variadicArgs); setErr != nil {
			return nil, fmt.Errorf("failed to set variadic parameter '%s': %w", proc.VariadicParamName, setErr)
		}
	}

	result, _, _, err = i.executeSteps(proc.Steps, false, nil) // Assuming executeSteps is defined elsewhere
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			err = fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
		}
		return nil, err
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
		if !(expectedReturnCount == 0 && actualReturnCount == 0) {
			err = fmt.Errorf("%w: procedure '%s' expected %d return values, but yielded %d", ErrReturnMismatch, procName, expectedReturnCount, actualReturnCount)
			return nil, err
		}
	}
	i.lastCallResult = finalResult
	return finalResult, nil
}

// Compile-time check to ensure *Interpreter implements the ToolRegistry interface.
// The ToolRegistry interface is defined in tools_types.go.
var _ ToolRegistry = (*Interpreter)(nil)

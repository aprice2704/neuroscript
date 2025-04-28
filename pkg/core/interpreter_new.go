// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings" // Keep strings import

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
)

// Interpreter holds the state of a running NeuroScript program.
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string // Track current proc for semantic checks
	sandboxDir      string

	toolRegistry *ToolRegistry
	logger       interfaces.Logger
	objectCache  map[string]interface{} // Cache for handle objects
	llmClient    *LLMClient
	modelName    string
	// Note: No Frame/CallStack here based on user's provided code structure.
	// Error state is handled via return values from executeSteps.
}

// --- Constants ---
const handleSeparator = "::"

// --- Constructor ---
func NewInterpreter(logger interfaces.Logger, llmClient *LLMClient) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		panic("FATAL: Interpreter requires a non-nil logger dependency")
	}
	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default, consider making configurable
		toolRegistry:    NewToolRegistry(),
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       llmClient,
		modelName:       "gemini-1.5-flash-latest", // Default model
		sandboxDir:      ".",                       // Default sandbox directory
	}
	if llmClient != nil && llmClient.modelName != "" {
		interp.modelName = llmClient.modelName
	}
	// Initialize default prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// Register core tools
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization: %v", err)
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("[INFO INTERP] Core tools registered successfully.")
	}
	return interp
}

// --- Getters / Setters ---
func (i *Interpreter) SandboxDir() string { return i.sandboxDir }
func (i *Interpreter) Logger() interfaces.Logger {
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
	// Read-Only Check Placeholder (requires isInHandler context)
	// if isInHandlerContext && (name == "err_code" || name == "err_msg") {
	//     return fmt.Errorf("cannot assign to read-only variable '%s' within on_error handler", name)
	// }
	i.variables[name] = value
	i.Logger().Debug("[DEBUG-INTERP] Set variable '%s' = %v (%T)", name, value, value)
	return nil
}

func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
	// Handler Variable Injection Placeholder (requires isInHandler context & activeError)
	// if isInHandlerContext && name == "err_code" { return activeError.Code, true }
	// if isInHandlerContext && name == "err_msg" { return activeError.Message, true }

	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}

func (i *Interpreter) SetModelName(name string) error {
	if name == "" {
		return errors.New("model name cannot be empty")
	}
	i.modelName = name
	i.Logger().Info("[INFO INTERP] Interpreter model name set to: %s", name)
	return nil
}

func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		i.Logger().Warn("[WARN INTERP] ToolRegistry accessed before initialization, creating new one.")
		i.toolRegistry = NewToolRegistry()
	}
	return i.toolRegistry
}

func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		i.Logger().Warn("[WARN INTERP] GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	return i.llmClient.Client()
}

func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	// Ensure metadata map exists if needed by Procedure struct
	if proc.Metadata == nil {
		proc.Metadata = make(map[string]string)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	i.Logger().Debug("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
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
	i.Logger().Info("[DEBUG-INTERP] Registered handle '%s' for type '%s'", fullHandle, typePrefix)
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
		// Assumes ErrCacheObjectWrongType is defined in errors.go
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s' (full handle: '%s')", ErrCacheObjectWrongType, expectedTypePrefix, actualPrefix, handle)
	}
	if i.objectCache == nil {
		i.Logger().Error("[ERROR INTERP] GetHandleValue called but objectCache is nil.")
		return nil, errors.New("internal error: object cache is not initialized")
	}
	obj, found := i.objectCache[handle]
	if !found {
		// Assumes ErrCacheObjectNotFound is defined in errors.go
		return nil, fmt.Errorf("%w: handle '%s'", ErrCacheObjectNotFound, handle)
	}
	i.Logger().Info("[DEBUG-INTERP] Retrieved handle '%s' with expected type '%s'", handle, expectedTypePrefix)
	return obj, nil
}
func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
		i.Logger().Info("[DEBUG-INTERP] Removed handle '%s'", handle)
	}
	return found
}

// --- Main Execution Entry Point ---
// RunProcedure sets up the context and starts execution for a specific procedure.
func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
	originalProcName := i.currentProcName // Store previous proc name
	i.Logger().Info("[DEBUG-INTERP] Running procedure '%s' with %d args provided.", procName, len(args))
	i.currentProcName = procName // Set current proc name for context/checks

	// Defer setting currentProcName back and logging final result/error
	defer func() {
		i.currentProcName = originalProcName
		i.Logger().Info("[DEBUG-INTERP] Finished procedure '%s'. Restored currentProcName to '%s'. Final Result: %v (%T), Err: %v", procName, i.currentProcName, result, result, err)
	}()

	proc, exists := i.knownProcedures[procName]
	if !exists {
		// Assumes ErrProcedureNotFound is defined in errors.go
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		i.Logger().Error("[ERROR] %v", err)
		return nil, err
	}

	// --- Argument Handling ---
	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		// Assumes ErrArgumentMismatch is defined in errors.go
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments ('needs'), but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		i.Logger().Error("[ERROR] %v", err)
		return nil, err
	}
	if numProvided > numTotalParams {
		i.Logger().Warn("[WARN INTERP] Procedure '%s' called with %d args, but only %d (required + optional) are defined. Extra args ignored.", procName, numProvided, numTotalParams)
	}
	// --- End Argument Handling ---

	// --- Scope Management ---
	procScope := make(map[string]interface{})
	for k, v := range i.variables { // Inherit outer scope
		procScope[k] = v
	}
	originalScope := i.variables // Save outer scope
	i.variables = procScope      // Activate new scope
	defer func() {
		i.variables = originalScope // Restore outer scope on exit
		i.Logger().Debug("[DEBUG-INTERP] Restored variable scope after '%s' finished.", procName)
	}()
	// --- End Scope Management ---

	// --- Assign Args to Scope ---
	i.Logger().Debug("[DEBUG-INTERP] Assigning %d required params for '%s'", numRequired, procName)
	for idx, paramName := range proc.RequiredParams {
		if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
			i.Logger().Error("[ERROR INTERP] Failed setting required proc arg '%s': %v", paramName, setErr)
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
		i.Logger().Debug("[DEBUG-INTERP]   Assigned required '%s'", paramName)
	}
	i.Logger().Debug("[DEBUG-INTERP] Assigning up to %d optional params for '%s'", numOptional, procName)
	for idx, paramName := range proc.OptionalParams {
		providedArgIndex := numRequired + idx
		valueToSet := interface{}(nil) // Default to nil
		if providedArgIndex < numProvided {
			valueToSet = args[providedArgIndex]
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			i.Logger().Error("[ERROR INTERP] Failed setting optional proc arg '%s': %v", paramName, setErr)
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		}
		logMsg := fmt.Sprintf("Optional param '%s' set to nil (not provided)", paramName)
		if providedArgIndex < numProvided {
			logMsg = fmt.Sprintf("Assigned optional '%s' (from provided arg %d)", paramName, providedArgIndex)
		}
		i.Logger().Debug("[DEBUG-INTERP]   %s", logMsg)
	}
	// --- End Assign Args ---

	// Execute the procedure steps (call moved to interpreter_exec.go)
	result, _, _, err = i.executeSteps(proc.Steps, false, nil) // Initial call: not in handler, no active error
	if err != nil {
		return nil, err // Error from execution
	}

	// --- Return Count Validation ---
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	if result != nil {
		resultValue := reflect.ValueOf(result)
		// Use reflect.Indirect if result could be a pointer
		kind := resultValue.Kind()
		if kind == reflect.Ptr || kind == reflect.Interface {
			resultValue = resultValue.Elem()
			kind = resultValue.Kind()
		}
		if kind == reflect.Slice {
			actualReturnCount = resultValue.Len()
		} else if resultValue.IsValid() { // Check if it's non-nil and valid before counting as 1
			actualReturnCount = 1
		}
	} // else actualReturnCount remains 0

	if actualReturnCount != expectedReturnCount {
		// Assumes ErrArgumentMismatch is defined elsewhere
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but returned %d", ErrArgumentMismatch, procName, expectedReturnCount, actualReturnCount)
		i.Logger().Error("[ERROR] %v", err)
		return nil, err
	}
	i.Logger().Debug("[DEBUG-INTERP] Return count validated for '%s': Expected %d, Got %d.", procName, expectedReturnCount, actualReturnCount)
	// --- End Return Count Validation ---

	i.lastCallResult = result
	return result, nil
}

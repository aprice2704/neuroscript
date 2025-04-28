// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"reflect" // FIX: Needed for reflect.ValueOf
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
)

// --- Interpreter (Struct definition unchanged) ---
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
}

// --- (SandboxDir, Logger, Variable Mgmt, Model Name, Tool Registry, LLM Client, Vector Index, Handle Cache unchanged) ---
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
	i.variables[name] = value
	i.logger.Debug("[DEBUG-INTERP] Set variable '%s' = %v (%T)", name, value, value)
	return nil
}
func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
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
	i.logger.Info("[INFO INTERP] Interpreter model name set to: %s", name)
	return nil
}
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		i.logger.Warn("[WARN INTERP] ToolRegistry accessed before initialization, creating new one.")
		i.toolRegistry = NewToolRegistry()
	}
	return i.toolRegistry
}
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		i.logger.Warn("[WARN INTERP] GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	return i.llmClient.Client()
}
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	i.logger.Debug("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
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
	i.logger.Info("[DEBUG-INTERP] Registered handle '%s' for type '%s'", fullHandle, typePrefix)
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
		i.logger.Error("[ERROR INTERP] GetHandleValue called but objectCache is nil.")
		return nil, errors.New("internal error: object cache is not initialized")
	}
	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s'", ErrCacheObjectNotFound, handle)
	}
	i.logger.Info("[DEBUG-INTERP] Retrieved handle '%s' with expected type '%s'", handle, expectedTypePrefix)
	return obj, nil
}
func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
		i.logger.Info("[DEBUG-INTERP] Removed handle '%s'", handle)
	}
	return found
}

const handleSeparator = "::" // Needs to be defined

// --- Constructor (Unchanged) ---
func NewInterpreter(logger interfaces.Logger, llmClient *LLMClient) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		panic("FATAL: Interpreter requires a non-nil logger dependency")
	}
	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16,
		toolRegistry:    NewToolRegistry(),
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		llmClient:       llmClient,
		modelName:       "gemini-1.5-flash-latest",
		sandboxDir:      ".",
	}
	if llmClient != nil && llmClient.modelName != "" {
		interp.modelName = llmClient.modelName
	}
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization: %v", err)
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("[INFO INTERP] Core tools registered successfully.")
	}
	return interp
}

// --- Execution Logic ---

// FIX: Updated RunProcedure semantic check logic
func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
	originalProcName := i.currentProcName // Store previous proc name
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Running procedure '%s' with %d args provided.", procName, len(args))
	}
	i.currentProcName = procName // Set current proc name for context/checks

	// Defer setting currentProcName back and logging final result/error
	defer func() {
		i.currentProcName = originalProcName
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP] Finished procedure '%s'. Restored currentProcName to '%s'. Final Result: %v (%T), Err: %v", procName, i.currentProcName, result, result, err)
		}
	}()

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		}
		return nil, err
	}

	// Argument Handling (Simplified - assuming correct types for now)
	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments ('needs'), but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		}
		return nil, err
	}
	if numProvided > numTotalParams {
		i.logger.Warn("[WARN INTERP] Procedure '%s' called with %d args, but only %d (required + optional) are defined. Extra args ignored.", procName, numProvided, numTotalParams)
	}

	// Setup Scope
	originalVars := make(map[string]interface{})
	for k, v := range i.variables {
		originalVars[k] = v
	}
	defer func() {
		i.variables = originalVars
		if i.logger != nil {
			i.logger.Debug("-INTERP]   Restored variable scope after '%s' finished.", procName)
		}
	}()

	// Assign Args
	if i.logger != nil {
		i.logger.Debug("-INTERP]   Assigning %d required params for '%s'", numRequired, procName)
	}
	for idx, paramName := range proc.RequiredParams {
		if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
			if i.logger != nil {
				i.logger.Error("[ERROR INTERP] Failed setting required proc arg '%s': %v", paramName, setErr)
			}
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		} else {
			if i.logger != nil {
				i.logger.Debug("-INTERP]     Assigned required '%s'", paramName)
			}
		}
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]   Assigning up to %d optional params for '%s'", numOptional, procName)
	}
	for idx, paramName := range proc.OptionalParams {
		providedArgIndex := numRequired + idx
		valueToSet := interface{}(nil)
		if providedArgIndex < numProvided {
			valueToSet = args[providedArgIndex]
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			if i.logger != nil {
				i.logger.Error("[ERROR INTERP] Failed setting optional proc arg '%s': %v", paramName, setErr)
			}
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		} else {
			if providedArgIndex < numProvided {
				if i.logger != nil {
					i.logger.Debug("-INTERP]     Assigned optional '%s' (from provided arg %d)", paramName, providedArgIndex)
				}
			} else {
				if i.logger != nil {
					i.logger.Debug("-INTERP]     Optional param '%s' not provided, set to nil", paramName)
				}
			}
		}
	}

	// Execute the procedure steps
	// 'wasReturn' is only relevant *within* executeSteps to stop processing early.
	// We analyze the final 'result' here to check against the signature.
	result, _, err = i.executeSteps(proc.Steps)
	if err != nil {
		// Error from execution, return it directly
		return nil, err
	}

	// --- Semantic Check: Validate Return Count ---
	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0

	if result != nil {
		resultValue := reflect.ValueOf(result)
		// Check if the result from executeReturn is a slice (indicating multiple returns)
		if resultValue.Kind() == reflect.Slice {
			actualReturnCount = resultValue.Len() // Count items in the slice
		} else {
			// If not a slice, it's a single return value
			actualReturnCount = 1
		}
	} else {
		// If result is nil, it means either 'return;' was called or the func ended naturally.
		actualReturnCount = 0
	}

	// Now compare expected vs actual count
	if actualReturnCount != expectedReturnCount {
		err = fmt.Errorf("%w: procedure '%s' expected %d return values (declared in 'returns'), but returned %d", ErrArgumentMismatch, procName, expectedReturnCount, actualReturnCount)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		}
		return nil, err // Return error if counts don't match
	}

	if i.logger != nil {
		i.logger.Debug("-INTERP]   Return count validated for '%s': Expected %d, Got %d.", procName, expectedReturnCount, actualReturnCount)
	}
	// --- End Semantic Check ---

	// Store the result (which might be a slice) in lastCallResult
	i.lastCallResult = result
	// Return the result (single value or slice) and nil error
	return result, nil
}

// executeSteps iterates through and executes steps, handling control flow.
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Executing %d steps...", len(steps))
	}
	for stepNum, step := range steps {
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]   Step %d: Type=%s, Target=%s", stepNum+1, strings.ToUpper(step.Type), step.Target)
		}

		switch strings.ToLower(step.Type) {
		case "set":
			err = i.executeSet(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
		case "call":
			callResult, callErr := i.executeCall(step, stepNum)
			if callErr != nil {
				return nil, false, fmt.Errorf("step %d (%s %s): %w", stepNum+1, step.Type, step.Target, callErr)
			}
			result = callResult // Update potential return value
		case "return":
			result, wasReturn, err = i.executeReturn(step, stepNum)
			// Note: executeReturn now returns the actual value(s) in 'result'
			// and 'wasReturn' is always true if no error occurred during evaluation.
			if err != nil {
				return nil, true, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
			// If executeReturn succeeded, we immediately signal return upwards
			return result, true, nil // Propagate return value(s) and wasReturn=true
		case "emit":
			err = i.executeEmit(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
		case "if":
			ifResult, ifReturned, ifErr := i.executeIf(step, stepNum)
			if ifErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, ifErr)
			}
			if ifReturned {
				return ifResult, true, nil
			}
			result = ifResult
		case "while":
			whileResult, whileReturned, whileErr := i.executeWhile(step, stepNum)
			if whileErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, whileErr)
			}
			if whileReturned {
				return whileResult, true, nil
			}
			result = whileResult
		case "for":
			forResult, forReturned, forErr := i.executeFor(step, stepNum)
			if forErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, forErr)
			}
			if forReturned {
				return forResult, true, nil
			}
			result = forResult
		case "must", "mustbe":
			err = i.executeMust(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
		case "try":
			tryResult, tryReturned, tryErr := i.executeTryCatch(step, stepNum)
			if tryErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, tryErr)
			}
			if tryReturned {
				return tryResult, true, nil
			}
			result = tryResult
		case "fail":
			failErr := i.executeFail(step, stepNum)
			return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, failErr)
		default:
			err = fmt.Errorf("%w: step %d: unknown step type '%s'", ErrUnknownKeyword, stepNum+1, step.Type)
			if i.logger != nil {
				i.logger.Error("[ERROR] %v", err)
			}
			return nil, false, err
		}
	}

	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Finished executing steps block normally.")
	}
	// If the loop finishes without an explicit return, wasReturn remains false.
	// 'result' holds the value from the last successful CALL or block execution, or nil otherwise.
	return result, false, nil
}

// executeBlock (Unchanged)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil {
			if i.logger != nil {
				i.logger.Debug("[DEBUG-INTERP] >> Entering empty block execution for %s (parent step %d)", blockType, parentStepNum+1)
			}
			return nil, false, nil
		}
		err = fmt.Errorf("step %d (%s): invalid block format - expected []Step, got %T", parentStepNum+1, blockType, blockValue)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		}
		return nil, false, err
	}
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] >> Entering block execution for %s (parent step %d, %d steps)", blockType, parentStepNum+1, len(steps))
	}
	result, wasReturn, err = i.executeSteps(steps) // Recursive call
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] << Exiting block execution for %s (parent step %d), wasReturn: %v, err: %v", blockType, parentStepNum+1, wasReturn, err)
	}
	return result, wasReturn, err
}

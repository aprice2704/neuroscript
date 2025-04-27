// filename: pkg/core/interpreter.go
package core

import (
	// Added for llmClient.CallLLM call
	"errors"
	"fmt"
	"strings" // Added for handle prefix checking

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid" // Added for handle generation
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	sandboxDir      string // Expected to be an absolute, clean path

	toolRegistry *ToolRegistry

	logger interfaces.Logger

	objectCache map[string]interface{} // Cache for handle objects

	llmClient *LLMClient // Renamed from genaiClient
	modelName string     // Model name specifically for interpreter use (if different from client default)
}

// SandboxDir returns the interpreter's sandbox directory.
func (i *Interpreter) SandboxDir() string {
	return i.sandboxDir
}

// Logger returns the interpreter's logger instance.
func (i *Interpreter) Logger() interfaces.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

// --- Variable Management ---

// SetVariable sets a variable in the interpreter's scope.
func (i *Interpreter) SetVariable(name string, value interface{}) error {
	if i.variables == nil {
		i.variables = make(map[string]interface{})
		i.logger.Info("[WARN INTERP] variables map was nil during SetVariable, initialized.")
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	// Maybe add check for reserved words if needed
	i.variables[name] = value
	i.logger.Debug("[DEBUG-INTERP] Set variable '%s' = %v (%T)", name, value, value)
	return nil
}

// GetVariable retrieves a variable. Returns value and true if found, nil and false otherwise.
func (i *Interpreter) GetVariable(name string) (interface{}, bool) {
	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}

// --- Model Name ---
func (i *Interpreter) SetModelName(name string) error {
	if name == "" {
		return errors.New("model name cannot be empty")
	}
	i.modelName = name
	i.logger.Info("[INFO INTERP] Interpreter model name set to: %s (Note: LLMClient might use its own default)", name)
	return nil
}

// --- Tool Registry ---
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		i.logger.Warn("[WARN INTERP] ToolRegistry accessed before initialization, creating new one.")
		i.toolRegistry = NewToolRegistry() // Initialize if nil
	}
	return i.toolRegistry
}

// --- LLM Client Access ---
// GenAIClient provides access to the underlying genai.Client.
// Necessary for tools that interact directly with the File API, etc.
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		i.logger.Warn("[WARN INTERP] GenAIClient() called but internal LLMClient is nil.")
		return nil
	}
	return i.llmClient.Client() // Assumes LLMClient has a Client() method
}

// --- Procedure Management ---
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
		return make(map[string]Procedure) // Return empty map if not initialized
	}
	return i.knownProcedures
}

// --- Vector Index ---
func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }

// --- Handle Cache / Object Management ---

const handleSeparator = "::" // Used to separate type prefix from UUID

// RegisterHandle stores an object and returns a typed handle string (prefix::uuid).
func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", errors.New("handle type prefix cannot be empty")
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("handle type prefix '%s' cannot contain separator '%s'", typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
		i.logger.Warn("[WARN INTERP] objectCache was nil during RegisterHandle, initialized.")
	}

	handleID := uuid.NewString()
	fullHandle := fmt.Sprintf("%s%s%s", typePrefix, handleSeparator, handleID)
	i.objectCache[fullHandle] = obj
	i.logger.Info("[DEBUG-INTERP] Registered handle '%s' for type '%s'", fullHandle, typePrefix)
	return fullHandle, nil
}

// GetHandleValue retrieves an object using its handle, checking the type prefix.
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
		// Wrap error with ErrCacheObjectWrongType
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s' (full handle: '%s')", ErrCacheObjectWrongType, expectedTypePrefix, actualPrefix, handle)
	}

	if i.objectCache == nil {
		i.logger.Error("[ERROR INTERP] GetHandleValue called but objectCache is nil.")
		return nil, errors.New("internal error: object cache is not initialized")
	}

	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s'", ErrCacheObjectNotFound, handle) // Wrap sentinel
	}
	i.logger.Info("[DEBUG-INTERP] Retrieved handle '%s' with expected type '%s'", handle, expectedTypePrefix)
	return obj, nil
}

// RemoveHandle deletes an object from the cache. Returns true if found and removed.
func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false // Nothing to remove
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
		i.logger.Info("[DEBUG-INTERP] Removed handle '%s'", handle)
	}
	return found
}

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
		embeddingDim:    16,                // Default embedding dimension
		toolRegistry:    NewToolRegistry(), // Initialize registry here
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}), // Initialize cache map
		llmClient:       llmClient,
		modelName:       "gemini-1.5-flash-latest", // Sensible default
		sandboxDir:      ".",                       // Default sandbox to current directory
	}
	// Override default model if provided via LLMClient
	if llmClient != nil && llmClient.modelName != "" {
		interp.modelName = llmClient.modelName
	}

	// Add built-in variables
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// Register core tools immediately upon creation
	if err := RegisterCoreTools(interp.toolRegistry); err != nil {
		effectiveLogger.Error("FATAL: Failed to register core tools during interpreter initialization: %v", err)
		panic(fmt.Sprintf("FATAL: Failed to register core tools: %v", err))
	} else {
		effectiveLogger.Info("[INFO INTERP] Core tools registered successfully.")
	}

	return interp
}

// --- Execution Logic ---

// RunProcedure finds and executes a defined procedure.
// UPDATED: Uses proc.RequiredParams and proc.OptionalParams
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Running procedure '%s' with %d args provided: %v", procName, len(args), args)
	}
	i.currentProcName = procName // Keep track for debugging/context

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		} // Use Error level
		return nil, err
	}

	// --- Argument Handling (v0.2.0) ---
	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments ('needs'), but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		if i.logger != nil {
			i.logger.Error("[ERROR] %v", err)
		} // Use Error level
		return nil, err
	}
	if numProvided > numTotalParams {
		i.logger.Warn("[WARN INTERP] Procedure '%s' called with %d args, but only %d (required + optional) are defined. Extra args ignored.", procName, numProvided, numTotalParams)
	}

	// Assign required arguments first
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
				i.logger.Debug("-INTERP]     Assigned required '%s' = %q", paramName, args[idx])
			}
		}
	}

	// Assign optional arguments if provided
	if i.logger != nil {
		i.logger.Debug("-INTERP]   Assigning %d optional params for '%s' (provided: %d)", numOptional, procName, numProvided-numRequired)
	}
	for idx, paramName := range proc.OptionalParams {
		providedArgIndex := numRequired + idx
		valueToSet := interface{}(nil) // Default to nil if not provided
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
					i.logger.Debug("-INTERP]     Assigned optional '%s' = %q (from provided arg %d)", paramName, args[providedArgIndex], providedArgIndex)
				}
			} else {
				if i.logger != nil {
					i.logger.Debug("-INTERP]     Optional param '%s' not provided, set to nil", paramName)
				}
			}
		}
	}
	// --- End Argument Handling ---

	// Execute the procedure steps
	result, _, err = i.executeSteps(proc.Steps) // Ignore wasReturn at the top level procedure call
	if err != nil {
		// Error is already contextualized by executeSteps/sub-calls
		return nil, err
	}

	// TODO: Validate return values against proc.ReturnVarNames if needed

	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Procedure '%s' finished, result: %v (%T)", procName, result, result)
	}
	i.lastCallResult = result // Store last result even if it's from a procedure
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

		switch strings.ToLower(step.Type) { // Use lowercase for switch
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
			if err != nil {
				return nil, true, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			} // Return error but signal return intent
			if wasReturn {
				return result, true, nil
			} // Propagate return
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
			} // Propagate return
			result = ifResult // Update potential return value (usually nil unless block returned)
		case "while":
			whileResult, whileReturned, whileErr := i.executeWhile(step, stepNum)
			if whileErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, whileErr)
			}
			if whileReturned {
				return whileResult, true, nil
			} // Propagate return
			result = whileResult
		case "for":
			forResult, forReturned, forErr := i.executeFor(step, stepNum)
			if forErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, forErr)
			}
			if forReturned {
				return forResult, true, nil
			} // Propagate return
			result = forResult
		case "must", "mustbe": // Combined MUST/MUSTBE execution
			err = i.executeMust(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			} // Must failure is an execution error
		case "try":
			tryResult, tryReturned, tryErr := i.executeTryCatch(step, stepNum)
			if tryErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, tryErr)
			} // Propagate error from try/catch/finally
			if tryReturned {
				return tryResult, true, nil
			} // Propagate return from try/catch/finally
			result = tryResult
		case "fail": // Use lowercase "fail" here to match grammar/AST builder
			failErr := i.executeFail(step, stepNum) // Call the newly added method
			// executeFail returns the specific error to propagate
			return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, failErr)
		default:
			err = fmt.Errorf("%w: step %d: unknown step type '%s'", ErrUnknownKeyword, stepNum+1, step.Type)
			if i.logger != nil {
				i.logger.Error("[ERROR] %v", err)
			}
			return nil, false, err
		}
	} // End step loop

	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Finished executing steps block normally.")
	}
	// Finished block without error or explicit RETURN
	return result, false, nil
}

// executeBlock executes steps within a block context (IF/WHILE/FOR/TRY/CATCH/FINALLY).
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		// Check if it's nil (e.g., empty ELSE or FINALLY) - this is okay
		if blockValue == nil {
			if i.logger != nil {
				i.logger.Debug("[DEBUG-INTERP] >> Entering empty block execution for %s (parent step %d)", blockType, parentStepNum+1)
			}
			return nil, false, nil
		}
		// Otherwise, it's an invalid type
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

// executeFail handles the FAIL statement.
func (i *Interpreter) executeFail(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP]      Executing FAIL (Step %d)", stepNum+1)
	}
	valueNode := step.Value
	var failMessage string

	if valueNode != nil {
		// Evaluate fail message expression and check for errors
		evaluatedValue, evalErr := i.evaluateExpression(valueNode) // Depth 0
		if evalErr != nil {
			// If message evaluation fails, return that error, wrapped
			return fmt.Errorf("evaluating FAIL message: %w", evalErr)
		}
		failMessage = fmt.Sprintf("%v", evaluatedValue) // Convert evaluated value to string
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]        FAIL evaluated message: %q", failMessage)
		}
	} else {
		failMessage = "FAIL statement encountered" // Default message
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]        FAIL with no message (using default)")
		}
	}

	// Return a specific error type or just a general error, wrapped
	return fmt.Errorf("%w: %s", ErrFailStatement, failMessage)
}

// --- Assume other execute* methods (Set, Call, Return, Emit, If, While, For, Must, TryCatch) exist in separate files ---
// --- Need to ensure evaluateExpression and helper functions are defined elsewhere ---

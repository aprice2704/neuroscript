// filename: pkg/core/interpreter.go
// UPDATED: Add SetVariable method
package core

import (
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
	sandboxDir      string

	toolRegistry *ToolRegistry

	logger interfaces.Logger

	objectCache map[string]interface{} // Cache for handle objects
	handleTypes map[string]string

	llmClient *LLMClient // Renamed from genaiClient
	modelName string
}

func (i *Interpreter) SandboxDir() string {
	return i.sandboxDir
}

func (i *Interpreter) Logger() interfaces.Logger {
	return i.logger
}

// --- Variable Management ---

// SetVariable sets a variable in the interpreter's scope.
func (i *Interpreter) SetVariable(name string, value interface{}) error {
	if i.variables == nil {
		// This shouldn't happen if initialized correctly, but safeguard
		i.variables = make(map[string]interface{})
		if i.logger != nil {
			i.logger.Info("[WARN INTERP] variables map was nil during SetVariable, initialized.")
		}
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	// Add any other validation for variable names if needed (e.g., reserved words)
	i.variables[name] = value
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Set variable '%s' = %v (%T)", name, value, value)
	}
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
	if i.logger != nil { // Added nil check for safety
		i.logger.Info("[INFO INTERP] Interpreter model name set to: %s (Note: LLMClient might use its own)", name)
	}
	return nil
}

// --- Tool Registry ---
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		if i.logger != nil {
			i.logger.Info("[WARN INTERP] ToolRegistry accessed before initialization, creating new one.")
		}
		i.toolRegistry = NewToolRegistry() // Initialize if nil
	}
	return i.toolRegistry
}

// --- LLM Client Access ---
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		if i.logger != nil {
			i.logger.Info("[WARN INTERP] GenAIClient() called but internal LLMClient is nil.")
		}
		return nil
	}
	return i.llmClient.Client()
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
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
	}
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

const handleSeparator = "::"

func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", errors.New("handle type prefix cannot be empty")
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("handle type prefix '%s' cannot contain separator '%s'", typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
		if i.logger != nil {
			i.logger.Info("[WARN INTERP] objectCache was nil during RegisterHandle, initialized.")
		}
	}

	handleID := uuid.NewString()
	fullHandle := fmt.Sprintf("%s%s%s", typePrefix, handleSeparator, handleID)
	i.objectCache[fullHandle] = obj

	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Registered handle '%s' for type '%s'", fullHandle, typePrefix)
	}
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
		return nil, fmt.Errorf("invalid handle type: expected prefix '%s', got '%s' (full handle: '%s')", expectedTypePrefix, actualPrefix, handle)
	}

	if i.objectCache == nil {
		if i.logger != nil {
			i.logger.Info("[ERROR INTERP] GetHandleValue called but objectCache is nil.")
		}
		return nil, errors.New("internal error: object cache is not initialized")
	}

	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("handle not found: '%s'", handle)
	}
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Retrieved handle '%s' with expected type '%s'", handle, expectedTypePrefix)
	}
	return obj, nil
}

func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false // Nothing to remove
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP] Removed handle '%s'", handle)
		}
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
		embeddingDim:    16, // Default or configure?
		toolRegistry:    NewToolRegistry(),
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}), // Initialize map
		llmClient:       llmClient,
		modelName:       "gemini-1.5-pro-latest", // Default, may be overridden
	}
	if llmClient != nil && llmClient.modelName != "" {
		interp.modelName = llmClient.modelName
	}

	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	return interp
}

// --- Execution Logic ---

func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Running procedure '%s' with args: %v", procName, args)
	}
	i.currentProcName = procName

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("procedure '%s' not found", procName)
		if i.logger != nil {
			i.logger.Info("[ERROR] %v", err)
		}
		return nil, err
	}

	// Argument Handling
	if len(args) != len(proc.Params) {
		if i.logger != nil {
			i.logger.Info("[WARN INTERP] Procedure '%s' called with %d args, expected %d based on docstring params.", procName, len(args), len(proc.Params))
		}
	}
	// Assign arguments to variables
	for idx, paramName := range proc.Params {
		if idx < len(args) {
			if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
				// Log or handle error setting variable? Usually shouldn't fail here.
				if i.logger != nil {
					i.logger.Info("[ERROR INTERP] Failed setting proc arg %s: %v", paramName, setErr)
				}
			}
		} else {
			if setErr := i.SetVariable(paramName, nil); setErr != nil {
				if i.logger != nil {
					i.logger.Info("[ERROR INTERP] Failed setting proc arg %s to nil: %v", paramName, setErr)
				}
			}
		}
	}

	result, _, err = i.executeSteps(proc.Steps) // Ignore wasReturn at top level
	if err != nil {
		return nil, err
	}

	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Procedure '%s' finished, result: %v (%T)", procName, result, result)
	}
	i.lastCallResult = result
	return result, nil
}

func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Executing %d steps...", len(steps))
	}
	for stepNum, step := range steps {
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]   Step %d: Type=%s, Target=%s", stepNum+1, step.Type, step.Target)
		}

		switch step.Type {
		case "SET":
			err = i.executeSet(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
		case "CALL":
			callResult, callErr := i.executeCall(step, stepNum)
			if callErr != nil {
				return nil, false, fmt.Errorf("step %d (%s %s): %w", stepNum+1, step.Type, step.Target, callErr)
			}
			result = callResult
		case "RETURN":
			result, wasReturn, err = i.executeReturn(step, stepNum)
			if wasReturn {
				return result, true, err
			}
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s) internal error: %w", stepNum+1, step.Type, err)
			}
		case "IF":
			ifResult, ifReturned, ifErr := i.executeIf(step, stepNum)
			if ifErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, ifErr)
			}
			if ifReturned {
				return ifResult, true, nil
			}
			result = ifResult
		case "WHILE":
			whileResult, whileReturned, whileErr := i.executeWhile(step, stepNum)
			if whileErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, whileErr)
			}
			if whileReturned {
				return whileResult, true, nil
			}
			result = whileResult
		case "FOR":
			forResult, forReturned, forErr := i.executeFor(step, stepNum)
			if forErr != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, forErr)
			}
			if forReturned {
				return forResult, true, nil
			}
			result = forResult
		case "FAIL":
			failErr := i.executeFail(step, stepNum)
			return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, failErr)
		case "EMIT":
			err = i.executeEmit(step, stepNum)
			if err != nil {
				return nil, false, fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, err)
			}
		default:
			err = fmt.Errorf("step %d: unknown step type '%s'", stepNum+1, step.Type)
			if i.logger != nil {
				i.logger.Info("[ERROR] %v", err)
			}
			return nil, false, err
		}
	}
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] Finished executing steps block.")
	}
	return result, false, nil
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		err = fmt.Errorf("step %d (%s): invalid block format - expected []Step, got %T", parentStepNum+1, blockType, blockValue)
		if i.logger != nil {
			i.logger.Info("[ERROR] %v", err)
		}
		return nil, false, err
	}
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] >> Entering block execution for %s (parent step %d)", blockType, parentStepNum+1)
	}
	result, wasReturn, err = i.executeSteps(steps)
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP] << Exiting block execution for %s (parent step %d), wasReturn: %v, err: %v", blockType, parentStepNum+1, wasReturn, err)
	}
	return result, wasReturn, err
}

// executeFail handles the FAIL statement.
func (i *Interpreter) executeFail(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP]      Executing FAIL")
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

	// Return a specific error type or just a general error
	return errors.New(failMessage) // Using standard error for now
}

// --- Assume executeSet, executeCall, etc. are defined in other files ---
// --- Assume executeFail is defined (added placeholder previously) ---
// --- Assume evaluateExpression is defined in evaluation_main.go ---
// --- Assume evaluateCondition, isTruthy are defined in evaluation_helpers.go ---

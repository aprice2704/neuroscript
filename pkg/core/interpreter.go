// pkg/core/interpreter.go
package core

import (
	"context" // Added for GenAI client init
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os" // Added for Getenv

	// Import core
	"github.com/aprice2704/neuroscript/pkg/core/prompts" // Import core
	// *** ADDED: Import for UUIDs for handles ***
	"github.com/google/uuid"
	// *** ADDED: Imports for GenAI Client ***
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure // Keep unexported
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	sandboxDir      string
	toolRegistry    *ToolRegistry // Use concrete type internally
	logger          *log.Logger
	// *** ADDED: Caches for opaque handles ***
	objectCache map[string]interface{} // Stores actual objects (e.g., *ast.File) keyed by handle ID
	handleTypes map[string]string      // Stores type tag (e.g., "GolangAST") for each handle ID
	// *** ADDED: GenAI Client ***
	genaiClient *genai.Client
}

// getCachedObjectAndType helper for testing cache state.
func (i *Interpreter) getCachedObjectAndType(handleID string) (object interface{}, typeTag string, found bool) {
	if i.objectCache == nil || i.handleTypes == nil {
		return nil, "", false
	}
	typeTag, typeFound := i.handleTypes[handleID]
	object, objFound := i.objectCache[handleID]
	found = typeFound && objFound
	return object, typeTag, found
}

// --- Getter for ToolRegistry ---
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		if i.logger != nil { // Check logger before using
			i.logger.Println("[WARN] ToolRegistry accessed before initialization, creating new one.")
		}
		i.toolRegistry = NewToolRegistry()
		// Core tools are registered by main now, so no need to register here.
	}
	return i.toolRegistry
}

// --- ADDED: Getter for GenAI Client ---
func (i *Interpreter) GenAIClient() *genai.Client {
	// Note: We rely on NewInterpreter having initialized it.
	// If it could be nil, add a check here.
	return i.genaiClient
}

// --- Exported method to add procedures ---
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
	}
	return nil
}

// --- ADDED: Exported method to get known procedures map ---
func (i *Interpreter) KnownProcedures() map[string]Procedure {
	if i.knownProcedures == nil {
		return make(map[string]Procedure) // Return empty map if not initialized
	}
	return i.knownProcedures
}

// --- Methods to satisfy tools.InterpreterContext ---
func (i *Interpreter) Logger() *log.Logger {
	if i.logger == nil {
		return log.New(io.Discard, "", 0)
	}
	return i.logger
}
func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) {
	i.vectorIndex = vi
}

// --- ADDED: Helper methods for handle cache (could be unexported if only used internally) ---

// storeObjectInCache stores an object, assigns it a type tag, and returns a handle ID.
func (i *Interpreter) storeObjectInCache(obj interface{}, typeTag string) string {
	if i.objectCache == nil || i.handleTypes == nil {
		i.logger.Printf("[ERROR] Interpreter object/handle cache not initialized!")
		// Initialize them defensively, though they should be in NewInterpreter
		i.objectCache = make(map[string]interface{})
		i.handleTypes = make(map[string]string)
	}
	handleID := uuid.NewString()
	i.objectCache[handleID] = obj
	i.handleTypes[handleID] = typeTag
	i.logger.Printf("[DEBUG-INTERP] Stored object with handle '%s' and type tag '%s'. Cache size: %d", handleID, typeTag, len(i.objectCache))
	return handleID
}

// retrieveObjectFromCache retrieves an object using its handle ID and validates its type tag.
func (i *Interpreter) retrieveObjectFromCache(handleID string, expectedTypeTag string) (interface{}, error) {
	if i.objectCache == nil || i.handleTypes == nil {
		i.logger.Printf("[ERROR] Interpreter object/handle cache not initialized!")
		return nil, fmt.Errorf("internal error: object cache not initialized")
	}

	storedTypeTag, typeFound := i.handleTypes[handleID]
	if !typeFound {
		return nil, fmt.Errorf("handle '%s' not found in cache", handleID)
	}
	if storedTypeTag != expectedTypeTag {
		return nil, fmt.Errorf("handle '%s' has incorrect type tag: expected '%s', got '%s'", handleID, expectedTypeTag, storedTypeTag)
	}

	obj, objFound := i.objectCache[handleID]
	if !objFound {
		// This indicates an internal inconsistency if the type tag was found
		i.logger.Printf("[ERROR] Internal cache inconsistency: handle type '%s' found for ID '%s', but object is missing!", storedTypeTag, handleID)
		return nil, fmt.Errorf("internal cache error: object for handle '%s' missing", handleID)
	}

	i.logger.Printf("[DEBUG-INTERP] Retrieved object with handle '%s'. Expected type '%s', found '%s'.", handleID, expectedTypeTag, storedTypeTag)
	return obj, nil
}

// --- END ADDED Handle Cache Methods ---

// NewInterpreter creates a new interpreter instance.
// *** MODIFIED: Initialize new maps and GenAI Client ***
func NewInterpreter(logger *log.Logger) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = log.New(io.Discard, "", 0)
	}

	// --- Initialize GenAI Client ---
	// Using GEMINI_API_KEY based on llm.go pattern
	apiKey := os.Getenv("GEMINI_API_KEY")
	var genaiClient *genai.Client
	var clientErr error
	if apiKey == "" {
		effectiveLogger.Println("[WARN] GEMINI_API_KEY environment variable not set. File API tools will likely fail.")
		// Allow creation but client will be nil
	} else {
		ctx := context.Background() // Use background context for initialization
		genaiClient, clientErr = genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if clientErr != nil {
			effectiveLogger.Printf("[ERROR] Failed to create GenAI client: %v. File API tools will likely fail.", clientErr)
			// Continue with nil client
			genaiClient = nil
		} else {
			effectiveLogger.Println("[INFO] GenAI client created successfully.")
		}
	}
	// --- End GenAI Client Init ---

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure), // Initialize map
		vectorIndex:     make(map[string][]float32), // Initialize map
		embeddingDim:    16,                         // Default mock dim
		toolRegistry:    NewToolRegistry(),          // Create registry here
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}), // Initialize object cache
		handleTypes:     make(map[string]string),      // Initialize handle type map
		genaiClient:     genaiClient,                  // Store initialized client (or nil)
	}

	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	// TODO(?): Add a Close() method to Interpreter to close the genaiClient?
	// For now, assume the application manages the client lifetime.

	return interp
}

// RunProcedure executes a named procedure with given arguments.
// (Rest of the function remains the same)
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	proc, exists := i.KnownProcedures()[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined or not loaded", procName)
	}

	localVars := make(map[string]interface{})
	for k, v := range i.variables {
		localVars[k] = v
	}
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), received %d (%v)", procName, len(proc.Params), proc.Params, len(args), args)
	}
	for idx, paramName := range proc.Params {
		if _, isBuiltIn := i.variables[paramName]; isBuiltIn && i.logger != nil {
			i.logger.Printf("[WARN] Procedure '%s' parameter '%s' shadows built-in variable.", procName, paramName)
		}
		localVars[paramName] = args[idx]
	}

	originalVars := i.variables
	originalProcName := i.currentProcName
	i.variables = localVars
	i.currentProcName = procName

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >> Starting Procedure: %s with args: %v", procName, args)
		i.logger.Printf("[DEBUG-INTERP]    Initial local scope size: %d", len(i.variables))
	}

	defer func() {
		i.currentProcName = originalProcName
		i.variables = originalVars
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP] << Finished Procedure: %s (Result: %v (%T), Error: %v)", procName, result, result, err)
			i.logger.Printf("[DEBUG-INTERP]    Restored original scope reference. Current Proc: %q", i.currentProcName)
		}
	}()

	var wasReturn bool
	result, wasReturn, err = i.executeSteps(proc.Steps)
	if err == nil && !wasReturn {
		result = nil
	}

	return result, err
}

// executeSteps iterates through procedure steps and executes them.
// (Function remains the same)
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >>> executeSteps called with %d steps for proc '%s'", len(steps), i.currentProcName)
	}

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool
		var stepErr error

		if i.logger != nil {
			// Logging details remain the same
			valueStr := "<nil>"
			valueType := "<nil>"
			if step.Value != nil {
				valueBytes, _ := json.Marshal(step.Value)
				valueStr = string(valueBytes)
				valueType = fmt.Sprintf("%T", step.Value)
			}
			condStr := "<nil>"
			condType := "<nil>"
			if step.Cond != nil {
				condBytes, _ := json.Marshal(step.Cond)
				condStr = string(condBytes)
				condType = fmt.Sprintf("%T", step.Cond)
			}
			argsStr := "<nil>"
			if len(step.Args) > 0 {
				argsBytes, _ := json.Marshal(step.Args)
				argsStr = string(argsBytes)
			}
			elseValueStr := "<nil>"
			elseValueType := "<nil>"
			if step.ElseValue != nil {
				elseValueBytes, _ := json.Marshal(step.ElseValue)
				elseValueStr = string(elseValueBytes)
				elseValueType = fmt.Sprintf("%T", step.ElseValue)
			}

			i.logger.Printf("[DEBUG-INTERP]    Executing Step %d: Type=%s, Target=%q, Cond=(%s %s), Value=(%s %s), Else=(%s %s), Args=%s",
				stepNum+1, step.Type, step.Target, condType, condStr, valueType, valueStr, elseValueType, elseValueStr, argsStr)
		}

		switch step.Type {
		case "SET":
			stepErr = i.executeSet(step, stepNum)
		case "CALL":
			stepResult, stepErr = i.executeCall(step, stepNum)
		case "IF":
			stepResult, stepReturned, stepErr = i.executeIf(step, stepNum)
		case "RETURN":
			stepResult, stepReturned, stepErr = i.executeReturn(step, stepNum)
		case "EMIT":
			stepErr = i.executeEmit(step, stepNum)
		case "WHILE":
			stepResult, stepReturned, stepErr = i.executeWhile(step, stepNum)
		case "FOR":
			stepResult, stepReturned, stepErr = i.executeFor(step, stepNum)
		default:
			stepErr = fmt.Errorf("unknown step type encountered: %s", step.Type)
		}

		if stepErr != nil {
			return nil, false, fmt.Errorf("error in procedure '%s', step %d (%s): %w", i.currentProcName, stepNum+1, step.Type, stepErr)
		}
		if stepReturned {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]    RETURN encountered in step %d. Returning value: %v (%T)", stepNum+1, stepResult, stepResult)
			}
			return stepResult, true, nil
		}
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] <<< executeSteps finished normally for proc '%s'", i.currentProcName)
	}
	return nil, false, nil
}

// executeBlock helper
// (Function remains the same)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]     >>> Entering executeBlock for %s (from parent step %d)", blockType, parentStepNum+1)
	}
	defer func() {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]     <<< Exiting executeBlock for %s (Returned: %t, Err: %v)", blockType, wasReturn, err)
		}
	}()

	blockBody, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]       Block body is nil for %s, executing 0 steps.", blockType)
			}
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("step %d: internal error - %s block body has unexpected type %T", parentStepNum+1, blockType, blockValue)
	}
	if len(blockBody) == 0 {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]       Block body is empty for %s, executing 0 steps.", blockType)
		}
		return nil, false, nil
	}

	result, wasReturn, err = i.executeSteps(blockBody)
	return result, wasReturn, err
}

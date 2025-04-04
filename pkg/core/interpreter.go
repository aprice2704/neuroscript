// pkg/core/interpreter.go
package core

import (
	"fmt"
	"io"
	"log"
	// No changes needed to imports
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	toolRegistry    *ToolRegistry // Use concrete type internally
	logger          *log.Logger
}

// --- ADDED: Getter for ToolRegistry ---
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

// --- END ADDED Getter ---

// --- ADDED: Exported method to add procedures ---
// This method allows main.go (or other packages) to load parsed procedures
// into the interpreter's internal map, handling potential duplicates.
func (i *Interpreter) AddProcedure(proc Procedure) error {
	// Ensure the map is initialized
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		// Provide more context in the error if possible (e.g., which file defined it)
		// For now, just the name.
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
	}
	return nil
}

// --- END ADDED AddProcedure ---

// --- Methods to satisfy tools.InterpreterContext (or similar interface if defined) ---
func (i *Interpreter) Logger() *log.Logger {
	if i.logger == nil {
		// Return a logger that discards output if none is configured
		return log.New(io.Discard, "", 0)
	}
	return i.logger
}
func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32) // Initialize if nil
	}
	return i.vectorIndex
}
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) {
	i.vectorIndex = vi
}

// GenerateEmbedding is in pkg/core/embeddings.go

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger *log.Logger) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = log.New(io.Discard, "", 0)
	}

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure), // Initialize map
		vectorIndex:     make(map[string][]float32), // Initialize map
		embeddingDim:    16,                         // Default mock dim
		toolRegistry:    NewToolRegistry(),          // Create registry here
		logger:          effectiveLogger,
	}

	// Tool registration is now handled externally (e.g., in main.go)

	// Pre-load standard prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = PromptExecute

	return interp
}

// RunProcedure executes a named procedure with given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined or not loaded", procName)
	}

	// --- Scope handling ---
	localVars := make(map[string]interface{})
	for k, v := range i.variables { // Copy initial global/built-in vars
		localVars[k] = v
	}
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), received %d (%v)", procName, len(proc.Params), proc.Params, len(args), args)
	}
	for idx, paramName := range proc.Params {
		// Check against initial variables map, not localVars, for shadowing built-ins
		if _, isBuiltIn := i.variables[paramName]; isBuiltIn && i.logger != nil {
			i.logger.Printf("[WARN] Procedure '%s' parameter '%s' shadows built-in variable.", procName, paramName)
		}
		localVars[paramName] = args[idx] // Args are passed as strings
	}

	originalVars := i.variables
	originalProcName := i.currentProcName
	i.variables = localVars // Switch to local scope for execution
	i.currentProcName = procName

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >> Starting Procedure: %s with args: %v", procName, args)
		i.logger.Printf("[DEBUG-INTERP]    Initial local scope size: %d", len(i.variables))
	}

	// Defer scope restoration
	defer func() {
		i.currentProcName = originalProcName
		i.variables = originalVars // Restore reference to original map
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP] << Finished Procedure: %s (Result: %v (%T), Error: %v)", procName, result, result, err)
			i.logger.Printf("[DEBUG-INTERP]    Restored original scope reference. Current Proc: %q", i.currentProcName)
		}
	}()

	// Execute the steps
	var wasReturn bool
	result, wasReturn, err = i.executeSteps(proc.Steps) // Call the main step executor
	if err == nil && !wasReturn {
		result = nil // Implicit nil return if no RETURN statement was hit
	}

	return result, err
}

// executeSteps iterates through procedure steps and executes them.
// This is the central dispatcher for step types.
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >>> executeSteps called with %d steps for proc '%s'", len(steps), i.currentProcName)
	}

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool
		var stepErr error

		// Log step details BEFORE execution (as before)
		if i.logger != nil {
			// ... (logging code as before) ...
			valueStr := "<nil>"
			valueType := "<nil>"
			if step.Value != nil {
				valueType = fmt.Sprintf("%T", step.Value)
				if _, ok := step.Value.([]Step); ok {
					valueStr = fmt.Sprintf("[]Step (len %d)", len(step.Value.([]Step)))
				} else {
					valueStr = fmt.Sprintf("%+v", step.Value)
				}
			}
			condStr := "<nil>"
			condType := "<nil>"
			if step.Cond != nil {
				condType = fmt.Sprintf("%T", step.Cond)
				condStr = fmt.Sprintf("%+v", step.Cond)
			}
			argsStr := "<nil>"
			if len(step.Args) > 0 {
				argsStr = fmt.Sprintf("%+v", step.Args)
			}
			elseValueStr := "<nil>"
			elseValueType := "<nil>"
			if step.ElseValue != nil {
				elseValueType = fmt.Sprintf("%T", step.ElseValue)
				if _, ok := step.ElseValue.([]Step); ok {
					elseValueStr = fmt.Sprintf("[]Step (len %d)", len(step.ElseValue.([]Step)))
				} else {
					elseValueStr = fmt.Sprintf("%+v", step.ElseValue)
				}
			}

			i.logger.Printf("[DEBUG-INTERP]    Executing Step %d: Type=%s, Target=%q, Cond=(%s %s), Value=(%s %s), Else=(%s %s), Args=%s",
				stepNum+1, step.Type, step.Target, condType, condStr, valueType, valueStr, elseValueType, elseValueStr, argsStr)
		}

		// Dispatch to specific execution functions (which live in other files)
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

		// Handle errors and returns
		if stepErr != nil {
			return nil, false, fmt.Errorf("error in procedure '%s', step %d (%s): %w", i.currentProcName, stepNum+1, step.Type, stepErr)
		}
		if stepReturned {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]    RETURN encountered in step %d. Returning value: %v (%T)", stepNum+1, stepResult, stepResult)
			}
			return stepResult, true, nil // Propagate return immediately
		}
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] <<< executeSteps finished normally for proc '%s'", i.currentProcName)
	}
	return nil, false, nil // Finished all steps normally
}

// executeBlock helper - Executes a slice of steps (used by IF, ELSE, WHILE, FOR)
// This remains here as it's a general helper for executing nested steps.
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
		// Handle nil block (e.g., empty ELSE) gracefully
		if blockValue == nil {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]       Block body is nil for %s, executing 0 steps.", blockType)
			}
			return nil, false, nil // Not an error, just no steps
		}
		// If it's not nil and not []Step, it's an internal error
		return nil, false, fmt.Errorf("step %d: internal error - %s block body has unexpected type %T", parentStepNum+1, blockType, blockValue)
	}
	if len(blockBody) == 0 {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]       Block body is empty for %s, executing 0 steps.", blockType)
		}
		return nil, false, nil
	}

	// Execute steps recursively using the main dispatcher
	result, wasReturn, err = i.executeSteps(blockBody)
	return result, wasReturn, err
}

// --- REMOVED Redundant Placeholder/Stub Methods ---
// func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) { ... }
// func (i *Interpreter) evaluateCondition(condNode interface{}) (bool, error) { ... }
// func (i *Interpreter) executeSet(step Step, stepNum int) error { ... }
// func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) { ... }
// func (i *Interpreter) executeReturn(step Step, stepNum int) (interface{}, bool, error) { ... }
// func (i *Interpreter) executeEmit(step Step, stepNum int) error { ... }
// func (i *Interpreter) executeIf(step Step, stepNum int) (interface{}, bool, error) { ... }
// func (i *Interpreter) executeWhile(step Step, stepNum int) (interface{}, bool, error) { ... }
// func (i *Interpreter) executeFor(step Step, stepNum int) (interface{}, bool, error) { ... }

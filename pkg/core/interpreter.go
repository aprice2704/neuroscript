package core

import (
	"fmt"
	"io"
	"log"
	// "strings" // Not needed directly here
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32 // Mock vector index
	embeddingDim    int                  // Mock embedding dimension
	currentProcName string               // For error context
	toolRegistry    *ToolRegistry
	logger          *log.Logger
}

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger *log.Logger) *Interpreter {
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Default to discard
	}
	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default mock dim
		toolRegistry:    NewToolRegistry(),
		logger:          logger,
	}
	registerCoreTools(interp.toolRegistry) // Register built-in tools
	return interp
}

// LoadProcedures adds parsed procedures to the interpreter's known set.
func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			i.logger.Printf("[INFO] Reloading procedure: %s", p.Name) // Use logger consistently
		}
		i.knownProcedures[p.Name] = p
	}
	return nil
}

// RunProcedure executes a named procedure with given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined", procName)
	}

	// Create local scope for arguments
	localVars := make(map[string]interface{})
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), but received %d (%v)", procName, len(proc.Params), proc.Params, len(args), args)
	}
	for idx, paramName := range proc.Params {
		localVars[paramName] = args[idx]
	}

	// Save outer scope and set up local scope for execution
	originalVars := i.variables
	originalProcName := i.currentProcName
	i.variables = localVars
	i.currentProcName = procName

	// Use consistent logging prefix for interpreter actions
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >> Starting Procedure: %s with args: %v", procName, args)
		i.logger.Printf("[DEBUG-INTERP]    Initial local scope: %+v", i.variables)
	}

	// Defer restoring the scope and logging completion
	defer func() {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP] << Finished Procedure: %s (Result: %v (%T), Error: %v)", procName, result, result, err)
			i.logger.Printf("[DEBUG-INTERP]    Restoring scope. Previous proc: %q", originalProcName)
		}
		i.currentProcName = originalProcName
		i.variables = originalVars // Restore outer scope
	}()

	// Execute the steps
	var wasReturn bool
	result, wasReturn, err = i.executeSteps(proc.Steps)

	// If execution finished without an explicit RETURN, result is implicitly nil
	if err == nil && !wasReturn {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]    Procedure %s finished without explicit RETURN.", procName)
		}
		result = nil
	}

	return result, err // Return final result and error
}

// --- Step Execution (Main Loop) ---
// executeSteps iterates through procedure steps and executes them.
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >>> executeSteps called with %d steps for proc '%s'", len(steps), i.currentProcName)
	}

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool // Did this specific step cause a RETURN?
		var stepErr error

		// Log the Step struct details BEFORE execution
		if i.logger != nil {
			valueStr := "<nil>"
			valueType := "<nil>"
			if step.Value != nil {
				valueType = fmt.Sprintf("%T", step.Value)
				if _, ok := step.Value.([]Step); ok {
					valueStr = fmt.Sprintf("[]Step (len %d)", len(step.Value.([]Step)))
				} else {
					valueStr = fmt.Sprintf("%+v", step.Value) // Log details for non-slice values
				}
			}

			condStr := "<nil>"
			condType := "<nil>"
			if step.Cond != nil {
				condType = fmt.Sprintf("%T", step.Cond)
				condStr = fmt.Sprintf("%+v", step.Cond) // Log details
			}

			argsStr := "<nil>"
			if len(step.Args) > 0 {
				argsStr = fmt.Sprintf("%+v", step.Args) // Log details
			}

			i.logger.Printf("[DEBUG-INTERP]    Executing Step %d: Type=%s, Target=%q, Cond=(%s %s), Value=(%s %s), Args=%s",
				stepNum+1, step.Type, step.Target, condType, condStr, valueType, valueStr, argsStr)
		}

		switch step.Type {
		case "SET":
			stepErr = i.executeSet(step, stepNum)
		case "CALL":
			i.lastCallResult = nil // Reset before CALL
			stepResult, stepErr = i.executeCall(step, stepNum)
			if stepErr == nil {
				i.lastCallResult = stepResult // Store successful result
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]      CALL %q successful. __last_call_result set to: %v (%T)", step.Target, i.lastCallResult, i.lastCallResult)
				}
			}
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

		// Centralized error handling after each step
		if stepErr != nil {
			// Add context (procedure name, step number/type) to the error
			return nil, false, fmt.Errorf("error in procedure '%s', step %d (%s): %w", i.currentProcName, stepNum+1, step.Type, stepErr)
		}

		// Check if the step caused a RETURN
		if stepReturned {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]    RETURN encountered in step %d. Returning value: %v (%T)", stepNum+1, stepResult, stepResult)
			}
			// Propagate the return value and signal that a return occurred
			return stepResult, true, nil
		}
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] <<< executeSteps finished normally for proc '%s'", i.currentProcName)
	}
	// Finished all steps normally without RETURN or error
	return nil, false, nil
}

// executeBlock helper - Executes a slice of steps (used by IF, ELSE, WHILE, FOR)
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
				i.logger.Printf("[DEBUG-INTERP]       Block body is nil, executing 0 steps.")
			}
			return nil, false, nil // Empty block is valid
		}
		// Should not happen if listener works correctly, but check anyway
		return nil, false, fmt.Errorf("step %d: internal error - %s block body has unexpected type %T", parentStepNum+1, blockType, blockValue)
	}

	if len(blockBody) == 0 {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]       Block body is empty, executing 0 steps.")
		}
		return nil, false, nil // Empty block executes successfully with no result/return
	}

	// Execute the steps within the block recursively
	// The recursive call will handle context wrapping for errors within the block
	result, wasReturn, err = i.executeSteps(blockBody)

	// Return results directly; error context is added by the recursive call
	return result, wasReturn, err
}

// pkg/core/interpreter.go
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
	lastCallResult  interface{} // Renamed internal field, accessed by LAST node
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	toolRegistry    *ToolRegistry
	logger          *log.Logger
}

// NewInterpreter creates a new interpreter instance.
func NewInterpreter(logger *log.Logger) *Interpreter {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	interp := &Interpreter{
		variables:       make(map[string]interface{}), // Initialize empty
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default mock dim
		toolRegistry:    NewToolRegistry(),
		logger:          logger,
	}
	registerCoreTools(interp.toolRegistry) // Register built-in tools

	// Pre-load standard prompts into variables
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = PromptExecute

	return interp
}

// LoadProcedures adds parsed procedures to the interpreter's known set.
func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			i.logger.Printf("[INFO] Reloading procedure: %s", p.Name)
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

	// Create local scope for arguments, copying initial global/built-in vars first.
	localVars := make(map[string]interface{})
	for k, v := range i.variables { // Copy initial global vars (like prompts)
		localVars[k] = v
	}

	// Add arguments, potentially overwriting globals if names clash
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), received %d (%v)", procName, len(proc.Params), proc.Params, len(args), args)
	}
	for idx, paramName := range proc.Params {
		// Warn if overwriting a built-in?
		if _, isBuiltIn := i.variables[paramName]; isBuiltIn && i.logger != nil {
			// Check against initial variables map, not localVars
			i.logger.Printf("[WARN] Procedure '%s' parameter '%s' shadows built-in variable.", procName, paramName)
		}
		localVars[paramName] = args[idx] // Args are passed as strings
	}

	originalVars := i.variables // Save global scope reference
	originalProcName := i.currentProcName
	i.variables = localVars // Execute with the local scope
	i.currentProcName = procName

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >> Starting Procedure: %s with args: %v", procName, args)
		i.logger.Printf("[DEBUG-INTERP]    Initial local scope size: %d", len(i.variables))
	}

	defer func() {
		// Restore only the global scope reference, localVars are discarded
		i.currentProcName = originalProcName
		i.variables = originalVars // Restore reference to original map
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP] << Finished Procedure: %s (Result: %v (%T), Error: %v)", procName, result, result, err)
			i.logger.Printf("[DEBUG-INTERP]    Restored global scope reference. Proc: %q", originalProcName)
		}
	}()

	// Execute the steps
	var wasReturn bool
	result, wasReturn, err = i.executeSteps(proc.Steps)
	if err == nil && !wasReturn {
		result = nil
	} // Implicit nil return

	return result, err
}

// executeSteps iterates through procedure steps and executes them.
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >>> executeSteps called with %d steps for proc '%s'", len(steps), i.currentProcName)
	}

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool
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

		switch step.Type {
		case "SET":
			stepErr = i.executeSet(step, stepNum)
		case "CALL":
			stepResult, stepErr = i.executeCall(step, stepNum) // Sets i.lastCallResult on success
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
		} // Propagate return
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] <<< executeSteps finished normally for proc '%s'", i.currentProcName)
	}
	return nil, false, nil // Finished all steps normally without RETURN or error
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
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("step %d: internal error - %s block body has unexpected type %T", parentStepNum+1, blockType, blockValue)
	}
	if len(blockBody) == 0 {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]       Block body is empty, executing 0 steps.")
		}
		return nil, false, nil
	}

	// Execute steps recursively
	result, wasReturn, err = i.executeSteps(blockBody)
	return result, wasReturn, err
}

package core

import (
	"fmt"
	// "strings" // No longer needed directly
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	toolRegistry    *ToolRegistry
}

func NewInterpreter() *Interpreter {
	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default mock dim
		toolRegistry:    NewToolRegistry(),
	}
	registerCoreTools(interp.toolRegistry) // Assumes this exists in tools_register.go
	return interp
}

func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			// fmt.Printf("[Warning] Reloading procedure: %s\n", p.Name) // DEBUG Commented out
		}
		i.knownProcedures[p.Name] = p
	}
	return nil
}

func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined", procName)
	}

	localVars := make(map[string]interface{})
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), but received %d", procName, len(proc.Params), proc.Params, len(args))
	}
	for idx, paramName := range proc.Params {
		localVars[paramName] = args[idx]
	}

	originalVars := i.variables
	originalProcName := i.currentProcName
	i.variables = localVars
	i.currentProcName = procName
	var wasReturn bool
	result, wasReturn, err = i.executeSteps(proc.Steps)
	i.currentProcName = originalProcName
	i.variables = originalVars

	if err != nil {
		// Error logging handled by executeSteps wrapper
	} else if wasReturn {
		// Return logging handled by executeSteps wrapper
	} else {
		// Normal finish logging handled by executeSteps wrapper
	}
	return result, err
}

// --- Step Execution (Main Loop) ---
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	//skipElse := false // Currently unused

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool
		var stepErr error

		switch step.Type {
		case "SET":
			stepErr = i.executeSet(step, stepNum)
		case "CALL":
			i.lastCallResult = nil
			stepResult, stepErr = i.executeCall(step, stepNum)
			if stepErr == nil {
				i.lastCallResult = stepResult
			}
		case "IF":
			stepResult, stepReturned, stepErr = i.executeIf(step, stepNum)
			if stepErr == nil && stepReturned {
				return stepResult, true, nil
			}
		case "RETURN":
			stepResult, stepReturned, stepErr = i.executeReturn(step, stepNum)
			if stepErr == nil && stepReturned {
				return stepResult, true, nil
			}
		case "WHILE":
			stepResult, stepReturned, stepErr = i.executeWhile(step, stepNum)
			if stepErr == nil && stepReturned {
				return stepResult, true, nil
			}
		case "FOR":
			stepResult, stepReturned, stepErr = i.executeFor(step, stepNum)
			if stepErr == nil && stepReturned {
				return stepResult, true, nil
			}
		default:
			stepErr = fmt.Errorf("unknown step type encountered in step %d: %s", stepNum+1, step.Type)
		}

		if stepErr != nil {
			return nil, false, fmt.Errorf("error in procedure '%s', step %d (%s): %w", i.currentProcName, stepNum+1, step.Type, stepErr)
		}
	}
	return nil, false, nil // Finished all steps without RETURN or error
}

// executeBlock helper - Executes a slice of steps (used by IF, ELSE, WHILE, FOR)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	blockBody, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("step %d: %s block body unexpected type %T", parentStepNum+1, blockType, blockValue)
	}
	if len(blockBody) == 0 {
		return nil, false, nil
	}
	result, wasReturn, err = i.executeSteps(blockBody) // Recursive call
	return result, wasReturn, err
}

// --- REMOVED executeLoopBody ---

// Helper to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

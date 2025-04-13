// pkg/core/interpreter.go
package core

import (
	"fmt"
	"io"
	"log"

	// Import core
	"github.com/aprice2704/neuroscript/pkg/core/prompts" // Import core
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
// This allows main.go to check if a procedure exists after loading.
func (i *Interpreter) KnownProcedures() map[string]Procedure {
	// Return a copy to prevent external modification?
	// For now, return the internal map directly for simplicity.
	// Consider implications if external code modifies this map.
	if i.knownProcedures == nil {
		return make(map[string]Procedure) // Return empty map if not initialized
	}
	return i.knownProcedures
}

// --- END ADDED KnownProcedures ---

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

	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	return interp
}

// RunProcedure executes a named procedure with given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	// Use the exported KnownProcedures() method here
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
// (Rest of the function remains the same)
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] >>> executeSteps called with %d steps for proc '%s'", len(steps), i.currentProcName)
	}

	for stepNum, step := range steps {
		var stepResult interface{}
		var stepReturned bool
		var stepErr error

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

// filename: pkg/core/interpreter_simple_steps.go
package core

import (
	"fmt"
	"strings"
)

// --- Simple Statement Execution Helpers ---

// executeSet (Unchanged)
func (i *Interpreter) executeSet(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing SET for variable '%s'", step.Target)
	}
	targetVar := step.Target
	valueNode := step.Value
	if !isValidIdentifier(targetVar) {
		return fmt.Errorf("SET target '%s' is not valid", targetVar)
	}
	finalValue, err := i.evaluateExpression(valueNode)
	if err != nil {
		return fmt.Errorf("evaluating value for SET '%s': %w", targetVar, err)
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]        SET evaluated value: %v (%T)", finalValue, finalValue)
	}
	// Special handling for 'generated_code'
	if targetVar == "generated_code" {
		if finalStr, isStr := finalValue.(string); isStr {
			trimmedVal := trimCodeFences(finalStr)
			if trimmedVal != finalStr {
				if i.logger != nil {
					i.logger.Debug("-INTERP]        Trimmed code fences for 'generated_code'")
				}
				finalValue = trimmedVal
			}
		}
	}
	i.variables[targetVar] = finalValue
	if i.logger != nil {
		i.logger.Debug("-INTERP]        Stored var '%s' = %v (%T)", targetVar, finalValue, finalValue)
	}
	return nil
}

// executeCall evaluates arguments and performs procedure, TOOL, or built-in calls.
// REMOVED specific KW_LLM handling. Assumes askAI etc. are TOOLs or user procedures.
func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing CALL %q", step.Target)
	}
	target := step.Target
	argNodes := step.Args

	evaluatedArgs := make([]interface{}, len(argNodes))
	var err error
	for idx, argNode := range argNodes {
		evaluatedArgs[idx], err = i.evaluateExpression(argNode)
		if err != nil {
			return nil, fmt.Errorf("evaluating argument %d for CALL %q: %w", idx, target, err)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        CALL Arg %d evaluated: %v (%T)", idx, evaluatedArgs[idx], evaluatedArgs[idx])
		}
	}

	var callResultValue interface{}
	var callErr error

	if strings.HasPrefix(target, "tool.") { // Updated prefix check
		toolName := strings.TrimPrefix(target, "tool.")
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Calling tool.%s", toolName)
		}
		toolImpl, found := i.toolRegistry.GetTool(toolName) // Use base name for lookup
		if !found {
			callErr = fmt.Errorf("unknown tool '%s'", toolName)
		} else {
			preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				callErr = fmt.Errorf("tool %s argument error: %w", toolName, validationErr)
			} else {
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Prepared TOOL args: %+v", preparedArgs)
				}
				callResultValue, callErr = toolImpl.Func(i, preparedArgs) // Captures potential tool execution error
				if callErr == nil {                                       // Only log/set last result on SUCCESS
					if i.logger != nil {
						i.logger.Debug("-INTERP]          tool.%s Result: %v (%T)", toolName, callResultValue, callResultValue)
					}
					i.lastCallResult = callResultValue
				} else {
					callErr = fmt.Errorf("tool %s execution failed: %w", toolName, callErr)
				}
			}
		}
		// --- REMOVED 'LLM' keyword check ---
		// } else if target == "LLM" { ... }
		// --- END REMOVED ---
	} else { // Assume Procedure Call or potential Built-in like askAI (if not TOOLs)
		// Check for specific built-in actor interactions first?
		// For now, assume they are dispatched as procedure calls or TOOL calls.
		// If they were distinct keywords, they'd need specific step types.
		procToCall := target
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Calling Procedure/Function %q", procToCall)
		}
		stringArgs := make([]string, len(evaluatedArgs)) // Prepare args for RunProcedure if needed
		for idx, val := range evaluatedArgs {
			stringArgs[idx] = fmt.Sprintf("%v", val)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]          Procedure args (as strings): %v", stringArgs)
		}

		// Try RunProcedure
		procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...)
		if procCallErr != nil {
			// If not found, maybe it's a built-in? (This logic needs refinement)
			// For now, just propagate the error.
			callErr = procCallErr // Error context added within RunProcedure
		} else {
			callResultValue = procResultValue
			callErr = nil
			if i.logger != nil {
				i.logger.Debug("-INTERP]          Procedure %q Result: %v (%T)", procToCall, callResultValue, callResultValue)
			}
			i.lastCallResult = callResultValue
		}
	}

	return callResultValue, callErr
}

// executeReturn (Unchanged)
func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing RETURN")
	}
	valueNode := step.Value
	var returnValue interface{}
	var evalErr error
	if valueNode != nil {
		returnValue, evalErr = i.evaluateExpression(valueNode)
		if evalErr != nil {
			return nil, true, fmt.Errorf("evaluating RETURN value: %w", evalErr)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        RETURN evaluated value: %v (%T)", returnValue, returnValue)
		}
	} else {
		returnValue = nil
		if i.logger != nil {
			i.logger.Debug("-INTERP]        RETURN with no value (implicit nil)")
		}
	}
	return returnValue, true, nil
}

// executeEmit (Unchanged)
func (i *Interpreter) executeEmit(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing EMIT")
	}
	valueNode := step.Value
	var evaluatedValue interface{}
	var evalErr error
	if valueNode != nil {
		evaluatedValue, evalErr = i.evaluateExpression(valueNode)
		if evalErr != nil {
			return fmt.Errorf("evaluating EMIT value: %w", evalErr)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        EMIT evaluated value: %v (%T)", evaluatedValue, evaluatedValue)
		}
	} else {
		evaluatedValue = ""
		if i.logger != nil {
			i.logger.Debug("-INTERP]        EMIT with no value (implicit empty string)")
		}
	}
	fmt.Printf("[EMIT] %v\n", evaluatedValue)
	return nil
}

// --- NEW: executeMust ---
func (i *Interpreter) executeMust(step Step, stepNum int) error {
	stepType := step.Type // "must" or "mustbe"
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing %s", strings.ToUpper(stepType))
	}
	valueNode := step.Value // This holds the expression node or FunctionCallNode

	// Evaluate the expression/function call
	evaluatedValue, evalErr := i.evaluateExpression(valueNode)
	if evalErr != nil {
		// An error during evaluation itself is considered a MUST failure
		err := fmt.Errorf("%s check failed during evaluation: %w", stepType, evalErr)
		if i.logger != nil {
			i.logger.Debug("-INTERP]        %s", err)
		}
		// Wrap with sentinel error ErrMustConditionFailed
		return fmt.Errorf("%w: %w", ErrMustConditionFailed, err)
	}

	// Check the truthiness of the result
	isOk := isTruthy(evaluatedValue)
	if i.logger != nil {
		i.logger.Debug("-INTERP]        %s evaluated value: %v (%T), Truthiness: %t", stepType, evaluatedValue, evaluatedValue, isOk)
	}

	if !isOk {
		// Condition failed, return specific error
		err := fmt.Errorf("%s condition evaluated to false (value: %v)", stepType, evaluatedValue)
		if i.logger != nil {
			i.logger.Debug("-INTERP]        %s", err)
		}
		// Wrap with sentinel error ErrMustConditionFailed
		return fmt.Errorf("%w: %w", ErrMustConditionFailed, err)
	}

	// Condition passed
	if i.logger != nil {
		i.logger.Debug("-INTERP]        %s check PASSED.", stepType)
	}
	return nil // Success
}

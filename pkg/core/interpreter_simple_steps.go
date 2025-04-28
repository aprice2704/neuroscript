// filename: pkg/core/interpreter_simple_steps.go
package core

import (
	"fmt"
	"strings"
)

// --- Simple Statement Execution Helpers ---

// executeSet (Unchanged)
// ...
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

// executeCall (Unchanged)
// ...
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
	isToolCall := false
	var baseToolName string
	if len(target) > 5 && strings.ToLower(target[:5]) == "tool." {
		isToolCall = true
		baseToolName = target[5:]
	}

	if isToolCall {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Calling tool with base name: %s (Original target: %s)", baseToolName, target)
		}
		toolImpl, found := i.toolRegistry.GetTool(baseToolName)
		if !found {
			callErr = fmt.Errorf("%w: TOOL '%s' (base name: '%s')", ErrProcedureNotFound, target, baseToolName)
		} else {
			preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				callErr = fmt.Errorf("tool '%s' argument error: %w", baseToolName, validationErr)
			} else {
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Prepared TOOL '%s' args: %+v", baseToolName, preparedArgs)
				}
				callResultValue, callErr = toolImpl.Func(i, preparedArgs)
				if callErr == nil {
					if i.logger != nil {
						i.logger.Debug("-INTERP]          Tool '%s' Result: %v (%T)", baseToolName, callResultValue, callResultValue)
					}
					i.lastCallResult = callResultValue
				} else {
					callErr = fmt.Errorf("tool '%s' execution failed: %w", baseToolName, callErr)
				}
			}
		}
	} else { // Assume Procedure Call
		procToCall := target
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Calling Procedure/Function %q", procToCall)
		}
		procResultValue, procCallErr := i.RunProcedure(procToCall, evaluatedArgs...)
		if procCallErr != nil {
			callErr = procCallErr
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

// FIX: Simplify executeReturn logic - always return nil or []interface{}
func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing RETURN")
	}

	// Step.Value should hold []interface{} (slice of expression nodes) or nil
	valueNodesRaw := step.Value

	if valueNodesRaw == nil {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        RETURN with no value (result is nil)")
		}
		return nil, true, nil // Return nil result for `return;`
	}

	// Attempt to cast to a slice of interfaces
	valueNodes, ok := valueNodesRaw.([]interface{})
	if !ok {
		// This should not happen if AST builder is correct
		return nil, true, fmt.Errorf("internal error: RETURN step value is not nil or []interface{} (%T)", valueNodesRaw)
	}

	// If the slice of nodes is empty (shouldn't happen with current AST builder logic, but defensive)
	if len(valueNodes) == 0 {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        RETURN with empty expression list (result is nil)")
		}
		return nil, true, nil // Return nil result for `return` followed by nothing? Or empty slice? Nil seems better.
	}

	// Evaluate each expression node in the list
	returnValues := make([]interface{}, len(valueNodes))
	for idx, node := range valueNodes {
		evaluatedValue, evalErr := i.evaluateExpression(node)
		if evalErr != nil {
			return nil, true, fmt.Errorf("evaluating RETURN value #%d: %w", idx+1, evalErr)
		}
		returnValues[idx] = evaluatedValue
		if i.logger != nil {
			i.logger.Debug("-INTERP]        RETURN evaluated value #%d: %v (%T)", idx+1, evaluatedValue, evaluatedValue)
		}
	}

	// Return the slice of evaluated results.
	// The RunProcedure function will handle unpacking if only one value is expected/returned.
	if i.logger != nil {
		i.logger.Debug("-INTERP]        Returning %d value(s) as slice.", len(returnValues))
	}
	return returnValues, true, nil
}

// executeEmit (Unchanged)
// ...
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

// executeMust (Unchanged)
// ...
func (i *Interpreter) executeMust(step Step, stepNum int) error {
	stepType := step.Type // "must" or "mustbe"
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing %s", strings.ToUpper(stepType))
	}
	valueNode := step.Value

	evaluatedValue, evalErr := i.evaluateExpression(valueNode)
	if evalErr != nil {
		err := fmt.Errorf("%s check failed during evaluation: %w", stepType, evalErr)
		if i.logger != nil {
			i.logger.Debug("-INTERP]        %s", err)
		}
		return fmt.Errorf("%w: %w", ErrMustConditionFailed, err)
	}
	isOk := isTruthy(evaluatedValue)
	if i.logger != nil {
		i.logger.Debug("-INTERP]        %s evaluated value: %v (%T), Truthiness: %t", stepType, evaluatedValue, evaluatedValue, isOk)
	}
	if !isOk {
		err := fmt.Errorf("%s condition evaluated to false (value: %v)", stepType, evaluatedValue)
		if i.logger != nil {
			i.logger.Debug("-INTERP]        %s", err)
		}
		return fmt.Errorf("%w: %w", ErrMustConditionFailed, err)
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]        %s check PASSED.", stepType)
	}
	return nil
}

// executeFail (Unchanged)
// ...
func (i *Interpreter) executeFail(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Info("[DEBUG-INTERP]      Executing FAIL (Step %d)", stepNum+1)
	}
	valueNode := step.Value
	var failMessage string
	if valueNode != nil {
		evaluatedValue, evalErr := i.evaluateExpression(valueNode)
		if evalErr != nil {
			return fmt.Errorf("evaluating FAIL message: %w", evalErr)
		}
		failMessage = fmt.Sprintf("%v", evaluatedValue)
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]        FAIL evaluated message: %q", failMessage)
		}
	} else {
		failMessage = "FAIL statement encountered"
		if i.logger != nil {
			i.logger.Info("[DEBUG-INTERP]        FAIL with no message (using default)")
		}
	}
	return fmt.Errorf("%w: %s", ErrFailStatement, failMessage)
}

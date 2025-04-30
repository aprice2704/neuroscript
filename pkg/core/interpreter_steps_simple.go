// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	"fmt"
	"strings"
	// Assuming NsError, RuntimeError, error codes, etc. defined in errors.go
)

// executeSet handles the "set" step.
// It evaluates the RHS expression (which might now involve a function call)
// and sets the variable. Returns the set value for 'LAST'.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing SET: Target=%s", step.Target)
	// Evaluate the RHS expression node. This call might trigger function/tool execution.
	value, err := i.evaluateExpression(step.Value)
	if err != nil {
		// Wrap underlying evaluation error (could be from call execution now)
		return nil, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating value for SET %s", step.Target), fmt.Errorf("evaluating value for set %s: %w", step.Target, err))
	}

	// Check for read-only assignment in error handler
	if isInHandler && (step.Target == "err_code" || step.Target == "err_msg") {
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler", step.Target)
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, fmt.Errorf("%s: %w", errMsg, ErrReadOnlyViolation))
	}

	// Set the variable in the current scope
	setErr := i.SetVariable(step.Target, value)
	if setErr != nil {
		// Wrap internal error from variable scope management
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("setting variable '%s'", step.Target), fmt.Errorf("setting variable %s: %w", step.Target, setErr))
	}

	// Return the successfully evaluated and set value.
	return value, nil
}

// --- REMOVED executeCall ---

// executeReturn handles the "return" step.
// Evaluates return expressions and signals return.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing RETURN")
	rawValue := step.Value // This is usually []interface{} (AST nodes) or nil

	if rawValue == nil {
		i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)")
		return nil, true, nil // Return nil value, signal return=true, no error
	}

	// Expect rawValue to be a slice of expression nodes from the AST builder
	if exprSlice, ok := rawValue.([]interface{}); ok {
		i.Logger().Debug("[DEBUG-INTERP]     Return has %d expression(s)", len(exprSlice))
		if len(exprSlice) == 0 {
			i.Logger().Debug("[DEBUG-INTERP]     Return has empty expression list (equivalent to nil)")
			return nil, true, nil // Treat 'return ()' like 'return'
		}

		results := make([]interface{}, len(exprSlice))
		for idx, exprNode := range exprSlice {
			evaluatedValue, err := i.evaluateExpression(exprNode) // Evaluate each expression
			if err != nil {
				// Wrap evaluation error
				return nil, true, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating return expression %d", idx+1), fmt.Errorf("evaluating return expression %d: %w", idx+1, err))
			}
			results[idx] = evaluatedValue
		}
		// Return evaluated results slice, signal return=true, no error
		return results, true, nil
	} else {
		// This case should ideally not happen
		i.Logger().Error("[ERROR INTERP] RETURN step value was not []interface{}: %T", rawValue)
		errMsg := fmt.Sprintf("internal error: RETURN step value was not []interface{}, but %T", rawValue)
		return nil, true, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInternal))
	}
}

// executeEmit handles the "emit" step.
// Evaluates expression and prints. Returns emitted value for 'LAST'.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT")
	value, err := i.evaluateExpression(step.Value) // Evaluate the expression
	if err != nil {
		return nil, NewRuntimeError(ErrorCodeGeneric, "evaluating value for EMIT", fmt.Errorf("evaluating emit value: %w", err))
	}
	fmt.Printf("EMIT: %v\n", value) // Default print mechanism
	return value, nil               // Return emitted value for potential 'LAST' use
}

// executeMust handles "must" and "mustbe" steps.
// Uses NodeToString for error message source.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type) // must or mustbe
	i.Logger().Debug("[DEBUG-INTERP]   Executing %s", strings.ToUpper(stepType))

	// Evaluate the condition expression (which might be a CallableExprNode for mustbe)
	value, err := i.evaluateExpression(step.Value)

	// Handle evaluation errors (could be from the call in 'mustbe')
	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s", stepType)
		// If it was 'mustbe', the step.Target contains the function name.
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("error executing check function '%s' for mustbe", step.Target)
		}
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	// Check truthiness of the successfully evaluated result
	if !isTruthy(value) {
		errMsg := ""
		// Use step.Target for 'mustbe' error messages if available
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("'%s %s' evaluated to false", stepType, step.Target)
		} else {
			// Fallback for regular 'must' or if target wasn't captured
			nodeStr := NodeToString(step.Value) // Try to stringify the original AST node
			errMsg = fmt.Sprintf("'%s %s' condition evaluated to false", stepType, nodeStr)
		}
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	i.Logger().Debug("[DEBUG-INTERP]     %s condition TRUE.", strings.ToUpper(stepType))
	// Return the successfully evaluated condition value
	return value, nil
}

// executeFail handles the "fail" step. (Unchanged)
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	i.Logger().Debug("[DEBUG-INTERP]   Executing FAIL")
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var evalErr error = nil
	var wrappedErr error = ErrFailStatement

	if step.Value != nil {
		failValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			evalErr = err
			errMsg = fmt.Sprintf("fail statement executed (error evaluating message/code: %v)", err)
		} else {
			switch v := failValue.(type) {
			case string:
				errMsg = v
			case int:
				errCode = v
			case int64:
				errCode = int(v)
			case float64:
				errCode = int(v)
				errMsg = fmt.Sprintf("fail statement executed with code %d (from float %v)", errCode, v)
			default:
				errMsg = fmt.Sprintf("fail statement executed with value: %v", failValue)
			}
		}
	}
	finalErrMsg := errMsg
	if evalErr != nil {
		finalErrMsg = fmt.Sprintf("%s [evaluation error: %v]", errMsg, evalErr)
	}
	return NewRuntimeError(errCode, finalErrMsg, wrappedErr)
}

// executeOnError handles the "on_error" step setup. (Unchanged)
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing ON_ERROR - Handler now active for subsequent steps in this scope.")
	handlerStep := step // The step contains the body ([]Step) in its Value field
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step. (Unchanged)
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing CLEAR_ERROR")
	if !isInHandler {
		errMsg := fmt.Sprintf("step %d: 'clear_error' can only be used inside an on_error block", stepNum+1)
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, fmt.Errorf("%s: %w", errMsg, ErrClearViolation))
	}
	return true, nil // Signal clear was called
}

// --- ADDED NodeToString Helper ---
// NodeToString converts an AST node to a string representation for error messages.
func NodeToString(node interface{}) string {
	if node == nil {
		return "<nil>"
	}
	// Basic fallback using fmt.Sprintf
	str := fmt.Sprintf("%#v", node) // Use %#v for potentially more detail

	// Attempt to use String() method if available (common pattern)
	if stringer, ok := node.(fmt.Stringer); ok {
		str = stringer.String()
	}

	// Truncate long representations for brevity in error messages
	maxLen := 50
	if len(str) > maxLen {
		str = str[:maxLen-3] + "..."
	}
	return str
}

// isTruthy needs to be defined (e.g., in evaluation_helpers.go)
// func isTruthy(value interface{}) bool { ... }

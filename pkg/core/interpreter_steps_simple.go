// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	// Needed for errors.Is if used, keep for now
	"fmt"
	"strings"
	// Assuming NsError, RuntimeError, error codes, etc. defined in errors.go
)

// executeSet handles the "set" step.
// Evaluates the RHS expression and sets the variable. Returns the set value for 'LAST'.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String() // Get position string once
	i.Logger().Debug("[DEBUG-INTERP]   Executing SET", "Target", step.Target, "pos", posStr)
	// Evaluate the RHS expression node. This call might trigger function/tool execution.
	value, err := i.evaluateExpression(step.Value) // Value field holds the RHS expression node
	if err != nil {
		// Wrap underlying evaluation error
		errMsg := fmt.Sprintf("evaluating value for SET %s at %s", step.Target, posStr)
		// Use ErrorCodeEvaluation for evaluation errors
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}

	// Check for read-only assignment in error handler
	if isInHandler && (step.Target == "err_code" || step.Target == "err_msg") {
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler at %s", step.Target, posStr)
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, fmt.Errorf("%s: %w", errMsg, ErrReadOnlyViolation))
	}

	// Set the variable in the current scope
	setErr := i.SetVariable(step.Target, value)
	if setErr != nil {
		// Wrap internal error from variable scope management
		errMsg := fmt.Sprintf("setting variable '%s' at %s", step.Target, posStr)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
	}

	// Return the successfully evaluated and set value.
	return value, nil
}

// executeReturn handles the "return" step.
// Evaluates return expressions and signals return.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing RETURN", "pos", posStr)
	rawValue := step.Value // This is usually []Expression or nil

	if rawValue == nil {
		i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)", "pos", posStr)
		return nil, true, nil // Return nil value, signal return=true, no error
	}

	// Expect rawValue to be a slice of expression nodes from the AST builder
	// Now expect []Expression directly from AST
	if exprSlice, ok := rawValue.([]Expression); ok {
		i.Logger().Debug("[DEBUG-INTERP]     Return has %d expression(s)", "count", len(exprSlice), "pos", posStr)
		if len(exprSlice) == 0 {
			i.Logger().Debug("[DEBUG-INTERP]     Return has empty expression list (equivalent to nil)", "pos", posStr)
			return nil, true, nil // Treat 'return ()' like 'return'
		}

		results := make([]interface{}, len(exprSlice))
		for idx, exprNode := range exprSlice {
			evaluatedValue, err := i.evaluateExpression(exprNode) // Evaluate each expression
			if err != nil {
				// Wrap evaluation error, include position if possible from exprNode
				exprPosStr := "<unknown>"
				if exprNode != nil && exprNode.GetPos() != nil {
					exprPosStr = exprNode.GetPos().String()
				}
				errMsg := fmt.Sprintf("evaluating return expression %d at %s", idx+1, exprPosStr)
				// Use ErrorCodeEvaluation for evaluation errors
				return nil, true, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
			}
			results[idx] = evaluatedValue
		}
		// Return evaluated results slice, signal return=true, no error
		// If only one result, return it directly instead of a slice
		if len(results) == 1 {
			return results[0], true, nil
		}
		return results, true, nil
	} else {
		// This case should ideally not happen if AST builder is correct
		errMsg := fmt.Sprintf("internal error at %s: RETURN step value was not []Expression, but %T", posStr, rawValue)
		i.Logger().Error("[ERROR INTERP] %s", errMsg)
		return nil, true, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInternal))
	}
}

// executeEmit handles the "emit" step.
// Evaluates expression and prints. Returns emitted value for 'LAST'.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT", "pos", posStr)
	value, err := i.evaluateExpression(step.Value) // Evaluate the expression node
	if err != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}
	fmt.Printf("EMIT: %v\n", value) // Default print mechanism
	return value, nil               // Return emitted value for potential 'LAST' use
}

// executeMust handles "must" and "mustbe" steps.
// Uses NodeToString for error message source.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type) // must or mustbe
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing %s", "type", strings.ToUpper(stepType), "pos", posStr)

	// Evaluate the condition expression (Value field holds the condition for 'must' or the CallableExprNode for 'mustbe')
	value, err := i.evaluateExpression(step.Value)

	// Handle evaluation errors (could be from the call in 'mustbe')
	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s at %s", stepType, posStr)
		// If it was 'mustbe', the step.Target contains the function name.
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("error executing check function '%s' for mustbe at %s", step.Target, posStr)
		}
		// Wrap error using ErrMustConditionFailed sentinel
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	// Check truthiness of the successfully evaluated result
	if !isTruthy(value) {
		errMsg := ""
		// Use step.Target for 'mustbe' error messages if available
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("'%s %s' check failed (returned falsy) at %s", stepType, step.Target, posStr)
		} else {
			// Fallback for regular 'must' or if target wasn't captured
			nodeStr := NodeToString(step.Value) // Try to stringify the original AST node
			errMsg = fmt.Sprintf("'%s %s' condition evaluated to false at %s", stepType, nodeStr, posStr)
		}
		// Use ErrMustConditionFailed sentinel
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	i.Logger().Debug("[DEBUG-INTERP]     %s condition TRUE.", "type", strings.ToUpper(stepType), "pos", posStr)
	// Return the successfully evaluated condition value
	return value, nil
}

// executeFail handles the "fail" step.
// *** MODIFIED: Cast integers to ErrorCode ***
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement // Default error code
	errMsg := "fail statement executed"
	var evalErr error = nil
	wrappedErr := ErrFailStatement // Use specific sentinel error

	if step.Value != nil { // Value field holds the optional expression node
		failValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			evalErr = err // Store evaluation error
			errMsg = fmt.Sprintf("fail statement executed at %s (error evaluating message/code: %v)", posStr, err)
		} else {
			errMsg = fmt.Sprintf("fail statement executed at %s with value: %v", posStr, failValue) // Include pos
			// Check type of evaluated value to determine code/message
			switch v := failValue.(type) {
			case string:
				errMsg = v // Use string directly as message
			case int:
				errCode = ErrorCode(v) // *** CAST int to ErrorCode ***
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case int64:
				errCode = ErrorCode(int(v)) // *** CAST int64 to ErrorCode ***
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case float64:
				// Potential precision loss, but try casting
				errCode = ErrorCode(int(v)) // *** CAST float64 to ErrorCode ***
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d (from float %v)", posStr, errCode, v)
			}
		}
	} else {
		errMsg = fmt.Sprintf("fail statement executed at %s", posStr) // Include pos
	}

	finalErrMsg := errMsg
	if evalErr != nil {
		finalErrMsg = fmt.Sprintf("%s [evaluation error: %v]", errMsg, evalErr)
		wrappedErr = evalErr // Wrap the evaluation error instead of ErrFailStatement
	}

	return NewRuntimeError(errCode, finalErrMsg, wrappedErr)
}

// executeOnError handles the "on_error" step setup.
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing ON_ERROR - Handler now active.", "pos", posStr)
	// Return the step itself (which contains the handler body in Value) to be used as the active handler
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := fmt.Sprintf("step %d at %s: 'clear_error' can only be used inside an on_error block", stepNum+1, posStr)
		// Use ErrorCodeClearViolation
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, fmt.Errorf("%s: %w", errMsg, ErrClearViolation))
	}
	return true, nil // Signal clear was called
}

// --- Helper: NodeToString --- (Moved from interpreter_exec.go to avoid dependency)
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

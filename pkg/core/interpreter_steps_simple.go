// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 13:08:08 PDT
// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	"fmt"
	"strings"
	// Keep errors import if needed by other functions in the file
	// Keep strconv import if needed by other functions in the file
)

// executeSet handles the "set" step.
// Note: Assumes Step struct has Target and Value fields accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing SET", "Target", step.Target, "pos", posStr)
	value, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating value for SET %s at %s", step.Target, posStr)
		// Wrap the underlying error for context
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}
	if isInHandler && (step.Target == "err_code" || step.Target == "err_msg") {
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler at %s", step.Target, posStr)
		// Wrap the sentinel error
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, fmt.Errorf("%s: %w", errMsg, ErrReadOnlyViolation))
	}
	setErr := i.SetVariable(step.Target, value)
	if setErr != nil {
		errMsg := fmt.Sprintf("setting variable '%s' at %s", step.Target, posStr)
		// Wrap the underlying error
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
	}
	return value, nil
}

// executeReturn handles the "return" step.
// Note: Assumes Step struct has Value field accessible.
// Note: Assumes step.Pos.String() is valid.
// Note: Assumes Expression interface has GetPos() method.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing RETURN", "pos", posStr)
	rawValue := step.Value // Value from the Step struct

	if rawValue == nil {
		// Use structured logging
		i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)", "pos", posStr)
		return nil, true, nil
	}

	// Check if the Value is a slice of Expression
	if exprSlice, ok := rawValue.([]Expression); ok {
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-INTERP]     Return has expression(s)", "count", len(exprSlice), "pos", posStr)
		if len(exprSlice) == 0 {
			// Use structured logging
			i.Logger().Debug("[DEBUG-INTERP]     Return has empty expression list (equivalent to nil)", "pos", posStr)
			return nil, true, nil
		}

		results := make([]interface{}, len(exprSlice))
		for idx, exprNode := range exprSlice {
			evaluatedValue, err := i.evaluateExpression(exprNode)
			if err != nil {
				exprPosStr := "<unknown>"
				// Check if exprNode and its position are valid before accessing
				if exprNode != nil {
					nodePos := exprNode.GetPos() // Assuming GetPos returns *Position
					if nodePos != nil {
						exprPosStr = nodePos.String()
					}
				}
				errMsg := fmt.Sprintf("evaluating return expression %d at %s", idx+1, exprPosStr)
				// Ensure the error returned here is a RuntimeError
				if _, ok := err.(*RuntimeError); !ok {
					err = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
				}
				return nil, true, err // Return the evaluation error directly
			}
			results[idx] = evaluatedValue
		}
		// Return single value if only one expression, otherwise the slice
		if len(results) == 1 {
			return results[0], true, nil
		}
		return results, true, nil
	} else {
		// Handle the case where step.Value is not []Expression
		errMsg := fmt.Sprintf("internal error at %s: RETURN step value was not []Expression, but %T", posStr, rawValue)
		// Use structured logging - CORRECTED
		i.Logger().Error("Internal error: RETURN step value type mismatch", "error", errMsg, "pos", posStr, "actual_type", fmt.Sprintf("%T", rawValue))
		// Return an internal error
		return nil, true, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInternal))
	}
}

// executeEmit handles the "emit" step.
// Note: Assumes Step struct has Value field accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT", "pos", posStr)
	value, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		// Wrap the underlying error
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}
	// EMIT's purpose is console output
	fmt.Printf("EMIT: %v\n", value)
	return value, nil
}

// executeMust handles "must" and "mustbe" steps.
// Note: Assumes Step struct has Type, Pos, Value, Target fields accessible.
// Note: Assumes step.Pos.String() is valid.
// Note: Assumes NodeToString helper exists.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type) // e.g., "must", "mustbe"
	posStr := step.Pos.String()
	// Use structured logging - CORRECTED
	i.Logger().Debug("[DEBUG-INTERP]   Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	value, err := i.evaluateExpression(step.Value) // Value holds the condition or check function call result

	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s at %s", stepType, posStr)
		// Special message for mustbe if target (function name) exists
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("error executing check function '%s' for mustbe at %s", step.Target, posStr)
		}
		// Wrap the underlying evaluation error
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	// Check if the result is truthy
	if !isTruthy(value) {
		errMsg := ""
		if stepType == "mustbe" && step.Target != "" {
			// MustBe failed - provide specific message
			errMsg = fmt.Sprintf("'%s %s' check failed (returned falsy) at %s", stepType, step.Target, posStr)
		} else {
			// Must failed - provide specific message
			nodeStr := NodeToString(step.Value) // Get string representation of the condition node
			errMsg = fmt.Sprintf("'%s %s' condition evaluated to false at %s", stepType, nodeStr, posStr)
		}
		// Return the failure error
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	// Use structured logging - CORRECTED
	i.Logger().Debug("[DEBUG-INTERP]     MUST/MUSTBE condition TRUE", "type", strings.ToUpper(stepType), "pos", posStr)
	return value, nil // Return the truthy value (might be useful for LAST)
}

// executeFail handles the "fail" step.
// Note: Assumes Step struct has Value field accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement   // Default code
	errMsg := "fail statement executed" // Default message
	var evalErr error = nil
	wrappedErr := ErrFailStatement // Default underlying error

	if step.Value != nil {
		failValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			// Error evaluating the message/code itself
			evalErr = err // Store the evaluation error
			errMsg = fmt.Sprintf("fail statement executed at %s (error evaluating message/code: %v)", posStr, err)
		} else {
			// Successfully evaluated the value, use it for message/code
			errMsg = fmt.Sprintf("fail statement executed at %s with value: %v", posStr, failValue)
			switch v := failValue.(type) {
			case string:
				errMsg = v // Use string directly as message
			case int:
				errCode = ErrorCode(v) // Use int as error code
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case int64:
				errCode = ErrorCode(int(v)) // Convert int64 to ErrorCode
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case float64:
				// Convert float to int for ErrorCode (potential precision loss)
				errCode = ErrorCode(int(v))
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d (from float %v)", posStr, errCode, v)
				// Maybe add warning about float conversion?
			}
		}
	} else {
		// FAIL statement without a value
		errMsg = fmt.Sprintf("fail statement executed at %s", posStr)
	}

	// Construct the final error message, including evaluation error if it occurred
	finalErrMsg := errMsg
	if evalErr != nil {
		finalErrMsg = fmt.Sprintf("%s [evaluation error: %v]", errMsg, evalErr)
		wrappedErr = evalErr // Use the evaluation error as the wrapped error
	}

	// Return the final RuntimeError
	return NewRuntimeError(errCode, finalErrMsg, wrappedErr)
}

// executeOnError handles the "on_error" step setup.
// Note: Assumes Step struct has Pos field accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing ON_ERROR - Handler now active.", "pos", posStr)
	// Return the step itself to signal activation of the handler for subsequent steps
	// Need to return a pointer to the Step struct
	handlerStep := step // Create a copy? Or assume Step is already a pointer?
	// If Step is a struct, we need to return its address. If it's already a pointer, return it directly.
	// Let's assume Step is a struct based on the original code's usage.
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
// Note: Assumes Step struct has Pos field accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	posStr := step.Pos.String()
	// Use structured logging
	i.Logger().Debug("[DEBUG-INTERP]   Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := fmt.Sprintf("step %d at %s: 'clear_error' can only be used inside an on_error block", stepNum+1, posStr)
		// Wrap the sentinel error
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, fmt.Errorf("%s: %w", errMsg, ErrClearViolation))
	}
	// Return true to signal that the error should be cleared
	return true, nil
}

// --- ADDED executeAsk stub ---
// executeAsk handles the "ask" step (placeholder).
// Note: Assumes Step struct has Target, Value, Pos fields accessible.
// Note: Assumes step.Pos.String() is valid.
func (i *Interpreter) executeAsk(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	targetVar := step.Target
	// Use structured logging
	i.Logger().Info("[INFO-INTERP] Executing ASK (Placeholder)", "pos", posStr, "target_var", targetVar)

	// 1. Evaluate the prompt expression
	promptValue, err := i.evaluateExpression(step.Value) // Value holds the prompt Expression
	if err != nil {
		errMsg := fmt.Sprintf("evaluating prompt for ASK at %s", posStr)
		// Wrap the underlying error
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}

	promptStr, ok := promptValue.(string)
	if !ok {
		errMsg := fmt.Sprintf("prompt for ASK evaluated to non-string type (%T) at %s", promptValue, posStr)
		// Wrap the sentinel error
		return nil, NewRuntimeError(ErrorCodeType, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInvalidOperandTypeString))
	}

	// 2. TODO: Implement LLM call using i.llmClient.GenerateContent(promptStr)
	// Use structured logging
	i.Logger().Warn("[WARN INTERP] ASK step execution not fully implemented - LLM call skipped.", "prompt", promptStr)
	llmResult := fmt.Sprintf("LLM Response placeholder for: %s", promptStr) // Placeholder result

	// 3. Store result if target variable exists
	if targetVar != "" {
		if setErr := i.SetVariable(targetVar, llmResult); setErr != nil {
			errMsg := fmt.Sprintf("setting variable '%s' for ASK result at %s", targetVar, posStr)
			// Wrap the underlying error
			return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
		}
		// Use structured logging
		i.Logger().Debug("[DEBUG-INTERP] Stored ASK result in variable", "variable", targetVar)
	}

	// ASK step result itself (for LAST) is the LLM's response
	return llmResult, nil
}

// --- ADDED Handlers for Break/Continue ---

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing BREAK", "pos", posStr)
	// Break simply returns the sentinel error to signal the loop execution to stop.
	// No value is returned.
	return nil, ErrBreak
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing CONTINUE", "pos", posStr)
	// Continue simply returns the sentinel error to signal the loop execution to skip to the next iteration.
	// No value is returned.
	return nil, ErrContinue
}

// --- Helpers --- (NodeToString, isTruthy etc. assumed to exist elsewhere)
// Note: Need definitions for Step struct, Expression interface, Position struct,
//       NewRuntimeError, ErrorCode*, Err*, evaluateExpression, SetVariable,
//       isTruthy, NodeToString, evaluateBuiltInFunction, isBuiltInFunction, etc.
//       These are assumed to be defined correctly in other files.

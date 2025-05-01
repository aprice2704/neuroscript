// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	// Keep errors import
	"fmt"
	"strings"
	// Keep strconv import
)

// ... (executeSet, executeReturn, executeEmit, executeMust, executeFail, executeOnError, executeClearError remain the same as previous correct version) ...

// executeSet handles the "set" step.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing SET", "Target", step.Target, "pos", posStr)
	value, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating value for SET %s at %s", step.Target, posStr)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}
	if isInHandler && (step.Target == "err_code" || step.Target == "err_msg") {
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler at %s", step.Target, posStr)
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, fmt.Errorf("%s: %w", errMsg, ErrReadOnlyViolation))
	}
	setErr := i.SetVariable(step.Target, value)
	if setErr != nil {
		errMsg := fmt.Sprintf("setting variable '%s' at %s", step.Target, posStr)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
	}
	return value, nil
}

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing RETURN", "pos", posStr)
	rawValue := step.Value

	if rawValue == nil {
		i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)", "pos", posStr)
		return nil, true, nil
	}

	if exprSlice, ok := rawValue.([]Expression); ok {
		i.Logger().Debug("[DEBUG-INTERP]     Return has %d expression(s)", "count", len(exprSlice), "pos", posStr)
		if len(exprSlice) == 0 {
			i.Logger().Debug("[DEBUG-INTERP]     Return has empty expression list (equivalent to nil)", "pos", posStr)
			return nil, true, nil
		}

		results := make([]interface{}, len(exprSlice))
		for idx, exprNode := range exprSlice {
			evaluatedValue, err := i.evaluateExpression(exprNode)
			if err != nil {
				exprPosStr := "<unknown>"
				if exprNode != nil && exprNode.GetPos() != nil {
					exprPosStr = exprNode.GetPos().String()
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
		if len(results) == 1 {
			return results[0], true, nil
		}
		return results, true, nil
	} else {
		errMsg := fmt.Sprintf("internal error at %s: RETURN step value was not []Expression, but %T", posStr, rawValue)
		i.Logger().Error("[ERROR INTERP] %s", errMsg)
		// Return an internal error, signaling return=true might be misleading, maybe false?
		// Let's stick to true for now as the intent was to return.
		return nil, true, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInternal))
	}
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT", "pos", posStr)
	value, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}
	fmt.Printf("EMIT: %v\n", value)
	return value, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type)
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing %s", "type", strings.ToUpper(stepType), "pos", posStr)

	value, err := i.evaluateExpression(step.Value)

	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s at %s", stepType, posStr)
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("error executing check function '%s' for mustbe at %s", step.Target, posStr)
		}
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		// Ensure the final error is a RuntimeError
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	if !isTruthy(value) {
		errMsg := ""
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("'%s %s' check failed (returned falsy) at %s", stepType, step.Target, posStr)
		} else {
			nodeStr := NodeToString(step.Value)
			errMsg = fmt.Sprintf("'%s %s' condition evaluated to false at %s", stepType, nodeStr, posStr)
		}
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	i.Logger().Debug("[DEBUG-INTERP]     %s condition TRUE.", "type", strings.ToUpper(stepType), "pos", posStr)
	return value, nil
}

// executeFail handles the "fail" step.
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var evalErr error = nil
	wrappedErr := ErrFailStatement

	if step.Value != nil {
		failValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			evalErr = err
			errMsg = fmt.Sprintf("fail statement executed at %s (error evaluating message/code: %v)", posStr, err)
		} else {
			errMsg = fmt.Sprintf("fail statement executed at %s with value: %v", posStr, failValue)
			switch v := failValue.(type) {
			case string:
				errMsg = v
			case int:
				errCode = ErrorCode(v)
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case int64:
				errCode = ErrorCode(int(v))
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d", posStr, errCode)
			case float64:
				errCode = ErrorCode(int(v))
				errMsg = fmt.Sprintf("fail statement executed at %s with code %d (from float %v)", posStr, errCode, v)
			}
		}
	} else {
		errMsg = fmt.Sprintf("fail statement executed at %s", posStr)
	}

	finalErrMsg := errMsg
	if evalErr != nil {
		finalErrMsg = fmt.Sprintf("%s [evaluation error: %v]", errMsg, evalErr)
		wrappedErr = evalErr
	}

	return NewRuntimeError(errCode, finalErrMsg, wrappedErr)
}

// executeOnError handles the "on_error" step setup.
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing ON_ERROR - Handler now active.", "pos", posStr)
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := fmt.Sprintf("step %d at %s: 'clear_error' can only be used inside an on_error block", stepNum+1, posStr)
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, fmt.Errorf("%s: %w", errMsg, ErrClearViolation))
	}
	return true, nil
}

// --- ADDED executeAsk stub ---
// executeAsk handles the "ask" step (placeholder).
func (i *Interpreter) executeAsk(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	targetVar := step.Target
	i.Logger().Info("[INFO-INTERP] Executing ASK (Placeholder)", "pos", posStr, "target_var", targetVar)

	// 1. Evaluate the prompt expression
	promptValue, err := i.evaluateExpression(step.Value) // Value holds the prompt Expression
	if err != nil {
		errMsg := fmt.Sprintf("evaluating prompt for ASK at %s", posStr)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}

	promptStr, ok := promptValue.(string)
	if !ok {
		errMsg := fmt.Sprintf("prompt for ASK evaluated to non-string type (%T) at %s", promptValue, posStr)
		return nil, NewRuntimeError(ErrorCodeType, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInvalidOperandTypeString))
	}

	// 2. TODO: Implement LLM call using i.llmClient.GenerateContent(promptStr)
	i.Logger().Warn("[WARN INTERP] ASK step execution not fully implemented - LLM call skipped.", "prompt", promptStr)
	llmResult := fmt.Sprintf("LLM Response placeholder for: %s", promptStr) // Placeholder result

	// 3. Store result if target variable exists
	if targetVar != "" {
		if setErr := i.SetVariable(targetVar, llmResult); setErr != nil {
			errMsg := fmt.Sprintf("setting variable '%s' for ASK result at %s", targetVar, posStr)
			return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
		}
		i.Logger().Debug("[DEBUG-INTERP] Stored ASK result in variable", "variable", targetVar)
	}

	// ASK step result itself (for LAST) is the LLM's response
	return llmResult, nil
}

// --- Helpers --- (NodeToString, isTruthy etc. assumed to exist elsewhere)

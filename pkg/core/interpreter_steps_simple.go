// NeuroScript Version: 0.3.0
// File version: 0.0.3 // Correct executeEmit to use step.Value, i.stdout, and fmt.Sprintf for string conversion.
// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	"fmt"
	"strings"
	// Keep errors import if needed by other functions in the file
	// Keep strconv import if needed by other functions in the file
)

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

	if len(step.Values) > 0 {
		i.Logger().Debug("[DEBUG-INTERP]     Return has multiple expressions", "count", len(step.Values), "pos", posStr)
		results := make([]interface{}, len(step.Values))
		for idx, exprNode := range step.Values {
			evaluatedValue, err := i.evaluateExpression(exprNode)
			if err != nil {
				exprPosStr := "<unknown>"
				if exprNode != nil {
					nodePos := exprNode.GetPos()
					if nodePos != nil {
						exprPosStr = nodePos.String()
					}
				}
				errMsg := fmt.Sprintf("evaluating return expression %d at %s", idx+1, exprPosStr)
				if _, ok := err.(*RuntimeError); !ok {
					err = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
				}
				return nil, true, err
			}
			results[idx] = evaluatedValue
		}
		return results, true, nil
	}

	if step.Value != nil {
		i.Logger().Debug("[DEBUG-INTERP]     Return has a single expression", "pos", posStr)
		evaluatedValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			exprPosStr := "<unknown>"
			nodePos := step.Value.GetPos()
			if nodePos != nil {
				exprPosStr = nodePos.String()
			}
			errMsg := fmt.Sprintf("evaluating return expression at %s", exprPosStr)
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
			}
			return nil, true, err
		}
		return evaluatedValue, true, nil
	}

	i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)", "pos", posStr)
	return nil, true, nil
}

// executeEmit handles the "emit" step.
// It evaluates the expression provided with the EMIT statement and prints its string representation
// to the interpreter's configured stdout writer.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT", "pos", posStr)

	var valueToEmit interface{}
	var evalErr error

	// EMIT expects a single expression in step.Value.
	if step.Value != nil {
		valueToEmit, evalErr = i.evaluateExpression(step.Value)
	} else {
		// EMIT with no arguments results in emitting an empty string (effectively a newline via Fprintln).
		valueToEmit = ""
	}

	if evalErr != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		// Ensure the error is a RuntimeError
		if _, ok := evalErr.(*RuntimeError); !ok {
			evalErr = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, evalErr))
		}
		return nil, evalErr
	}

	// Convert the evaluated value to a string for printing.
	// Using fmt.Sprintf("%v", ...) for general purpose string conversion.
	formattedOutput := fmt.Sprintf("%v", valueToEmit)

	if i.stdout == nil {
		i.Logger().Error("executeEmit: Interpreter stdout is nil! This is a critical setup error. Falling back to os.Stdout.")
		fmt.Println(formattedOutput) // Emergency fallback
	} else {
		if _, err := fmt.Fprintln(i.stdout, formattedOutput); err != nil {
			i.Logger().Error("Failed to write EMIT output via i.stdout", "error", err)
			return nil, NewRuntimeError(ErrorCodeIOFailed, "failed to emit output", err)
		}
	}

	// The EMIT statement's "result" for the purpose of LAST is the raw value that was evaluated.
	return valueToEmit, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type)
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	value, err := i.evaluateExpression(step.Value)

	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s at %s", stepType, posStr)
		if stepType == "mustbe" && step.Target != "" {
			errMsg = fmt.Sprintf("error executing check function '%s' for mustbe at %s", step.Target, posStr)
		}
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	if !isTruthy(value) {
		errMsg := ""
		if stepType == "mustbe" && step.Target != "" {
			nodeStr := NodeToString(step.Value)
			errMsg = fmt.Sprintf("'%s %s(%s)' check failed (returned falsy) at %s", stepType, step.Target, nodeStr, posStr)
		} else {
			nodeStr := NodeToString(step.Value)
			errMsg = fmt.Sprintf("'%s %s' condition evaluated to false at %s", stepType, nodeStr, posStr)
		}
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	i.Logger().Debug("[DEBUG-INTERP]     MUST/MUSTBE condition TRUE", "type", strings.ToUpper(stepType), "pos", posStr)
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

// executeAsk handles the "ask" step (placeholder).
func (i *Interpreter) executeAsk(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	targetVar := step.Target
	i.Logger().Debug("[DEBUG-INTERP] Executing ASK (Placeholder)", "pos", posStr, "target_var", targetVar)

	promptValue, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating prompt for ASK at %s", posStr)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, err))
	}

	promptStr, ok := promptValue.(string)
	if !ok {
		errMsg := fmt.Sprintf("prompt for ASK evaluated to non-string type (%T) at %s", promptValue, posStr)
		return nil, NewRuntimeError(ErrorCodeType, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInvalidOperandTypeString))
	}

	i.Logger().Warn("[WARN INTERP] ASK step execution not fully implemented - LLM call skipped.", "prompt", promptStr)
	llmResult := fmt.Sprintf("LLM Response placeholder for: %s", promptStr)

	if targetVar != "" {
		if setErr := i.SetVariable(targetVar, llmResult); setErr != nil {
			errMsg := fmt.Sprintf("setting variable '%s' for ASK result at %s", targetVar, posStr)
			return nil, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, setErr))
		}
		i.Logger().Debug("[DEBUG-INTERP] Stored ASK result in variable", "variable", targetVar)
	}

	return llmResult, nil
}

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing BREAK", "pos", posStr)
	return nil, ErrBreak
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP]   Executing CONTINUE", "pos", posStr)
	return nil, ErrContinue
}

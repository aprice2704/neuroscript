// NeuroScript Version: 0.4.1
// File version: 10
// Purpose: Corrected executeMust to use lastCallResult directly as a core.Value, removing the now-incorrect type assertion.
// filename: pkg/core/interpreter_steps_simple.go

package core

import (
	"fmt"
	"strings"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step Step) (Value, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing RETURN", "pos", posStr)

	if len(step.Values) == 0 {
		return NilValue{}, true, nil
	}

	if len(step.Values) == 1 {
		evaluatedValue, err := i.evaluateExpression(step.Values[0])
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, WrapErrorWithPosition(err, step.Values[0].GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

	i.Logger().Debug("[DEBUG-INTERP] Return has multiple expressions", "count", len(step.Values), "pos", posStr)
	results := make([]Value, len(step.Values))
	for idx, exprNode := range step.Values {
		evaluatedValue, err := i.evaluateExpression(exprNode)
		if err != nil {
			errMsg := fmt.Sprintf("evaluating return expression %d", idx+1)
			return nil, true, WrapErrorWithPosition(err, exprNode.GetPos(), errMsg)
		}
		results[idx] = evaluatedValue
	}
	return NewListValue(results), true, nil
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step Step) (Value, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing EMIT", "pos", posStr)

	if len(step.Values) == 0 {
		fmt.Fprintln(i.stdout)
		return NilValue{}, nil
	}

	var lastVal Value = NilValue{}
	var outputParts []string

	for _, expr := range step.Values {
		valToEmit, evalErr := i.evaluateExpression(expr)
		if evalErr != nil {
			errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
			return nil, WrapErrorWithPosition(evalErr, expr.GetPos(), errMsg)
		}
		lastVal = valToEmit
		formattedOutput, _ := toString(valToEmit)
		outputParts = append(outputParts, formattedOutput)
	}

	if i.stdout == nil {
		i.Logger().Error("executeEmit: Interpreter stdout is nil! This is a critical setup error.")
		fmt.Println(strings.Join(outputParts, " "))
	} else {
		if _, err := fmt.Fprintln(i.stdout, strings.Join(outputParts, " ")); err != nil {
			i.Logger().Error("Failed to write EMIT output via i.stdout", "error", err)
			return nil, NewRuntimeError(ErrorCodeIOFailed, "failed to emit output", err).WithPosition(step.Pos)
		}
	}

	return lastVal, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step Step) (Value, error) {
	posStr := step.Pos.String()
	stepType := strings.ToLower(step.Type)
	i.Logger().Debug("[DEBUG-INTERP] Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	var val Value
	var err error

	var exprToEval Expression
	if step.Cond != nil {
		exprToEval = step.Cond
	} else if step.Call != nil { // For mustbe
		exprToEval = step.Call
	}

	if exprToEval == nil {
		// This handles the 'must last' case. lastCallResult is already a Value.
		if i.lastCallResult == nil {
			return nil, ErrMustConditionFailed
		}
		val = i.lastCallResult
	} else {
		val, err = i.evaluateExpression(exprToEval)
		if err != nil {
			return nil, WrapErrorWithPosition(err, exprToEval.GetPos(), "evaluating MUST condition")
		}
	}

	if ev, ok := val.(ErrorValue); ok {
		return nil, ev
	}

	if !IsTruthy(val) {
		return nil, ErrMustConditionFailed
	}

	return val, nil
}

// executeFail handles the "fail" step.
func (i *Interpreter) executeFail(step Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var wrappedErr error = ErrFailStatement
	var finalPos = step.Pos
	var exprToEval Expression

	if len(step.Values) > 0 {
		exprToEval = step.Values[0]
	}

	if exprToEval != nil {
		finalPos = exprToEval.GetPos()
		failValue, err := i.evaluateExpression(exprToEval)
		if err != nil {
			evalFailMsg := fmt.Sprintf("error evaluating message/code for FAIL statement: %s", err.Error())
			return NewRuntimeError(errCode, evalFailMsg, err).WithPosition(finalPos)
		}
		errMsg, _ = toString(failValue)
	}
	return NewRuntimeError(errCode, errMsg, wrappedErr).WithPosition(finalPos)
}

// executeOnError handles the "on_error" step setup.
func (i *Interpreter) executeOnError(step Step) (*Step, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing ON_ERROR - Handler now active.", "pos", posStr)
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step Step, isInHandler bool) (bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := "'clear_error' can only be used inside an on_error block"
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, ErrClearViolation).WithPosition(step.Pos)
	}
	return true, nil
}

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing BREAK", "pos", posStr)
	return ErrBreak
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CONTINUE", "pos", posStr)
	return ErrContinue
}

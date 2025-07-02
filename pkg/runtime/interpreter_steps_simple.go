// NeuroScript Version: 0.4.1
// File version: 11
// Purpose: Corrected executeMust to immediately return evaluation errors, ensuring error propagation.
// filename: pkg/runtime/interpreter_steps_simple.go
// nlines: 200 // Approximate
// risk_rating: MEDIUM

package runtime

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step ast.Step) (Value, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing RETURN", "pos", posStr)

	if len(step.Values) == 0 {
		return NilValue{}, true, nil
	}

	if len(step.Values) == 1 {
		evaluatedValue, err := i.evaluate.Expression(step.Values[0])
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, WrapErrorWithPosition(err, step.Values[0].GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

	i.Logger().Debug("[DEBUG-INTERP] Return has multiple expressions", "count", len(step.Values), "pos", posStr)
	results := make([]Value, len(step.Values))
	for idx, exprNode := range step.Values {
		evaluatedValue, err := i.evaluate.Expression(exprNode)
		if err != nil {
			errMsg := fmt.Sprintf("evaluating return expression %d", idx+1)
			return nil, true, WrapErrorWithPosition(err, exprNode.GetPos(), errMsg)
		}
		results[idx] = evaluatedValue
	}
	return NewListValue(results), true, nil
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step ast.Step) (Value, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing EMIT", "pos", posStr)

	if len(step.Values) == 0 {
		fmt.Fprintln(i.stdout)
		return NilValue{}, nil
	}

	var lastVal Value = NilValue{}
	var outputParts []string

	for _, expr := range step.Values {
		valToEmit, evalErr := i.evaluate.Expression(expr)
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
			return nil, lang.NewRuntimeError(ErrorCodeIOFailed, "failed to emit output", err).WithPosition(step.Pos)
		}
	}

	return lastVal, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step ast.Step) (Value, error) {
	posStr := step.Pos.String()
	stepType := strings.ToLower(step.Type)
	i.Logger().Debug("[DEBUG-INTERP] Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	var val Value
	var err error

	var exprToEval ast.Expression
	if step.Cond != nil {
		exprToEval = step.Cond
	} else if step.Call != nil { // For mustbe
		exprToEval = step.Call
	}

	if exprToEval == nil {
		if i.lastCallResult == nil {
			// This handles 'must last' when lastCallResult is nil.
			return nil, ErrMustConditionFailed
		}
		val = i.lastCallResult
	} else {
		val, err = i.evaluate.Expression(exprToEval)
		// CRITICAL FIX: Prioritize the evaluation error. If the expression itself
		// fails to evaluate, that is the error we must return.
		if err != nil {
			return nil, err
		}
	}

	// If the evaluation results in an ErrorValue (e.g. from a tool call), fail with that error.
	if ev, ok := val.(ErrorValue); ok {
		return nil, ev
	}

	// If the evaluation results in any other non-truthy value, fail with the generic condition error.
	if !IsTruthy(val) {
		return nil, ErrMustConditionFailed
	}

	return val, nil
}

// executeFail handles the "fail" step.
func (i *Interpreter) executeFail(step ast.Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var wrappedErr error = ErrFailStatement
	var finalPos = step.Pos
	var exprToEval ast.Expression

	if len(step.Values) > 0 {
		exprToEval = step.Values[0]
	}

	if exprToEval != nil {
		finalPos = exprToEval.GetPos()
		failValue, err := i.evaluate.Expression(exprToEval)
		if err != nil {
			evalFailMsg := fmt.Sprintf("error evaluating message/code for FAIL statement: %s", err.Error())
			return lang.NewRuntimeError(errCode, evalFailMsg, err).WithPosition(finalPos)
		}
		errMsg, _ = toString(failValue)
	}
	return lang.NewRuntimeError(errCode, errMsg, wrappedErr).WithPosition(finalPos)
}

// executeOnError handles the "on error" step setup.
func (i *Interpreter) executeOnError(step ast.Step) (*Step, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing ON_ERROR - Handler now active.", "pos", posStr)
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step ast.Step, isInHandler bool) (bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := "'clear_error' can only be used inside an on_error block"
		return false, lang.NewRuntimeError(ErrorCodeClearViolation, errMsg, ErrClearViolation).WithPosition(step.Pos)
	}
	return true, nil
}

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step ast.Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing BREAK", "pos", posStr)
	return ErrBreak
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step ast.Step) error {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CONTINUE", "pos", posStr)
	return ErrContinue
}

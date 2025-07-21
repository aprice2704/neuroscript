// NeuroScript Version: 0.5.2
// File version: 35
// Purpose: Replaced all direct access to the removed 'Position' field with calls to the GetPos() method.
// filename: pkg/interpreter/interpreter_steps_simple.go

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step ast.Step) (lang.Value, bool, error) {

	fmt.Printf("in executeReturn -------------> %v step", step)

	if len(step.Values) == 0 {
		return &lang.NilValue{}, true, nil
	}

	if len(step.Values) == 1 {
		// Evaluate the single return expression.
		evaluatedValue, err := i.evaluate.Expression(step.Values[0])
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, lang.WrapErrorWithPosition(err, step.Values[0].GetPos(), errMsg)
		}
		// =========================================================================
		fmt.Printf(">>>> [DEBUG] executeReturn: Evaluated return value is: %#v\n", evaluatedValue)
		// =========================================================================
		return evaluatedValue, true, nil
	}

	// Handle multiple return values by creating a list.
	results := make([]lang.Value, len(step.Values))
	for idx, exprNode := range step.Values {
		evaluatedValue, err := i.evaluate.Expression(exprNode)
		if err != nil {
			errMsg := fmt.Sprintf("evaluating return expression %d", idx+1)
			return nil, true, lang.WrapErrorWithPosition(err, exprNode.GetPos(), errMsg)
		}
		results[idx] = evaluatedValue
	}
	return lang.ListValue{Value: results}, true, nil
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step ast.Step) (lang.Value, error) {
	if len(step.Values) == 0 {
		fmt.Fprintln(i.stdout)
		return &lang.NilValue{}, nil
	}

	var lastVal lang.Value = &lang.NilValue{}
	var outputParts []string

	for _, expr := range step.Values {
		valToEmit, evalErr := i.evaluate.Expression(expr)
		if evalErr != nil {
			errMsg := fmt.Sprintf("evaluating value for EMIT at %s", step.GetPos().String())
			return nil, lang.WrapErrorWithPosition(evalErr, expr.GetPos(), errMsg)
		}
		lastVal = valToEmit
		formattedOutput, _ := lang.ToString(valToEmit)
		outputParts = append(outputParts, formattedOutput)
	}

	if i.stdout == nil {
		i.Logger().Error("executeEmit: Interpreter stdout is nil! This is a critical setup error.")
		fmt.Println(strings.Join(outputParts, " "))
	} else {
		if _, err := fmt.Fprintln(i.stdout, strings.Join(outputParts, " ")); err != nil {
			i.Logger().Error("Failed to write EMIT output via i.stdout", "error", err)
			return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, "failed to emit output", err).WithPosition(step.GetPos())
		}
	}

	return lastVal, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step ast.Step) (lang.Value, error) {
	var val lang.Value
	var err error

	var exprToEval ast.Expression
	if step.Cond != nil {
		exprToEval = step.Cond
	} else if step.Call != nil { // For mustbe
		exprToEval = step.Call
	}

	if exprToEval == nil {
		val = i.lastCallResult
	} else {
		val, err = i.evaluate.Expression(exprToEval)
		if err != nil {
			return nil, lang.WrapErrorWithPosition(err, exprToEval.GetPos(), "evaluating expression for 'must'")
		}
	}

	if !lang.IsTruthy(val) {
		return nil, lang.ErrMustConditionFailed
	}

	return val, nil
}

// executeFail handles the "fail" step.
func (i *Interpreter) executeFail(step ast.Step) error {
	errCode := lang.ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var wrappedErr error = lang.ErrFailStatement
	var finalPos = step.GetPos()
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
		errMsg, _ = lang.ToString(failValue)
	}
	return lang.NewRuntimeError(errCode, errMsg, wrappedErr).WithPosition(finalPos)
}

// executeOnError handles the "on error" step setup.
func (i *Interpreter) executeOnError(step ast.Step) (*ast.Step, error) {
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step ast.Step, isInHandler bool) (bool, error) {
	if !isInHandler {
		errMsg := "'clear_error' can only be used inside an on_error block"
		return false, lang.NewRuntimeError(lang.ErrorCodeClearViolation, errMsg, lang.ErrClearViolation).WithPosition(step.GetPos())
	}
	return true, nil
}

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step ast.Step) error {
	return lang.NewRuntimeError(lang.ErrorCodeControlFlow, "'break' used outside of a loop", lang.ErrBreak).WithPosition(step.GetPos())
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step ast.Step) error {
	return lang.NewRuntimeError(lang.ErrorCodeControlFlow, "'continue' used outside of a loop", lang.ErrContinue).WithPosition(step.GetPos())
}

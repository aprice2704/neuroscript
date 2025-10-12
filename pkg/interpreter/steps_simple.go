// NeuroScript Version: 0.8.0
// File version: 42
// Purpose: Re-plumbed all expression evaluation to use the external 'eval' package.
// filename: pkg/interpreter/steps_simple.go
// nlines: 180
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step ast.Step) (lang.Value, bool, error) {
	if len(step.Values) == 0 {
		return &lang.NilValue{}, true, nil
	}

	if len(step.Values) == 1 {
		evaluatedValue, err := eval.Expression(i, step.Values[0])
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, lang.WrapErrorWithPosition(err, step.Values[0].GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

	results := make([]lang.Value, len(step.Values))
	for idx, exprNode := range step.Values {
		evaluatedValue, err := eval.Expression(i, exprNode)
		if err != nil {
			errMsg := fmt.Sprintf("evaluating return expression %d", idx+1)
			return nil, true, lang.WrapErrorWithPosition(err, exprNode.GetPos(), errMsg)
		}
		results[idx] = evaluatedValue
	}
	return lang.ListValue{Value: results}, true, nil
}

// executeEmit handles the "emit" statement.
func (i *Interpreter) executeEmit(step ast.Step) (lang.Value, error) {
	if len(step.Values) == 0 {
		return &lang.NilValue{}, nil
	}
	val, err := eval.Expression(i, step.Values[0])
	if err != nil {
		return nil, err
	}

	if i.hostContext != nil && i.hostContext.EmitFunc != nil {
		i.hostContext.EmitFunc(val)
	} else {
		if _, err := fmt.Fprintln(i.Stdout(), val.String()); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, "failed to write to stdout", err).WithPosition(step.GetPos())
		}
	}
	return val, nil
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
		val, err = eval.Expression(i, exprToEval)
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
		failValue, err := eval.Expression(i, exprToEval)
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

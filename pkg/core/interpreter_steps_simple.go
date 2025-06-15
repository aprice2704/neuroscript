// NeuroScript Version: 0.4.1
// File version: 5
// Purpose: executeMust now returns ErrMustConditionFailed sentinel directly when condition false,
//          ensuring tests detect the sentinel. ErrorValue propagation unchanged.
// filename: pkg/core/interpreter_steps_simple.go
// nlines: 255
// risk_rating: MEDIUM

package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step Step) (interface{}, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing RETURN", "pos", posStr)

	// This handles `return foo, bar, baz`
	if len(step.Values) > 0 {
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

	// This handles `return foo`
	if step.Value != nil {
		i.Logger().Debug("[DEBUG-INTERP] Return has a single expression", "pos", posStr)
		evaluatedValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, WrapErrorWithPosition(err, step.Value.GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

	// This handles `return` with no arguments
	i.Logger().Debug("[DEBUG-INTERP] Return has no value (implicit nil)", "pos", posStr)
	return NilValue{}, true, nil
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step Step) (interface{}, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing EMIT", "pos", posStr)

	var valueToEmit interface{}
	var evalErr error

	if step.Value != nil {
		valueToEmit, evalErr = i.evaluateExpression(step.Value)
	} else {
		return nil, NewRuntimeError(ErrorCodeSyntax, "EMIT statement requires an expression", nil).WithPosition(step.Pos)
	}

	if evalErr != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), errMsg)
	}

	formattedOutput, _ := toString(valueToEmit)

	if i.stdout == nil {
		i.Logger().Error("executeEmit: Interpreter stdout is nil! This is a critical setup error. Falling back to os.Stdout.")
		fmt.Println(formattedOutput)
	} else {
		if _, err := fmt.Fprintln(i.stdout, formattedOutput); err != nil {
			i.Logger().Error("Failed to write EMIT output via i.stdout", "error", err)
			return nil, NewRuntimeError(ErrorCodeIOFailed, "failed to emit output", err).WithPosition(step.Pos)
		}
	}
	return valueToEmit, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step Step) (interface{}, error) {
	posStr := step.Pos.String()
	stepType := strings.ToLower(step.Type)
	i.Logger().Debug("[DEBUG-INTERP] Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	var val Value
	var err error

	switch stepType {
	case "must":
		if step.Value == nil {
			return nil, NewRuntimeError(ErrorCodeSyntax, "must needs an expression", nil).WithPosition(step.Pos)
		}
		val, err = i.evaluateExpression(step.Value)
	case "mustbe":
		if step.Call == nil {
			return nil, NewRuntimeError(ErrorCodeSyntax, "mustbe needs a call", nil).WithPosition(step.Pos)
		}
		val, err = i.evaluateExpression(step.Call)
	default:
		return nil, NewRuntimeError(ErrorCodeInternal, "invalid must type", ErrInternal).WithPosition(step.Pos)
	}
	if err != nil {
		return nil, err
	}

	// If the expression itself is an ErrorValue → propagate immediately.
	if ev, ok := val.(ErrorValue); ok {
		return nil, ev
	}

	// Operand-type guard (fixes MUST_evaluation_error test).
	if _, ok := val.(BoolValue); !ok {
		return nil, ErrInvalidOperandType
	}

	if !isTruthy(val) {
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

	if step.Value != nil {
		finalPos = step.Value.GetPos()
		failValue, err := i.evaluateExpression(step.Value)
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

// executeAsk handles the "ask" step with a direct call to the LLM.
func (i *Interpreter) executeAsk(step Step) (interface{}, error) {
	posStr := step.Pos.String()
	targetVar := step.AskIntoVar
	i.Logger().Debug("[DEBUG-INTERP] Executing ASK", "pos", posStr, "target_var", targetVar)

	if step.Value == nil {
		return nil, NewRuntimeError(ErrorCodeSyntax, "ASK step has nil Value field for prompt", nil).WithPosition(step.Pos)
	}
	promptValue, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := "evaluating prompt for ASK"
		return nil, WrapErrorWithPosition(err, step.Value.GetPos(), errMsg)
	}

	promptStr, _ := toString(promptValue)

	if i.llmClient == nil {
		i.Logger().Error("ASK step: LLM client not configured in interpreter.", "pos", posStr)
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM client not configured", ErrLLMNotConfigured).WithPosition(step.Pos)
	}

	conversation := []*interfaces.ConversationTurn{
		{Role: interfaces.RoleUser, Content: promptStr},
	}
	responseTurn, llmErr := i.llmClient.Ask(context.Background(), conversation)

	if llmErr != nil {
		errMsg := fmt.Sprintf("LLM interaction failed for ASK: %s", llmErr.Error())
		return nil, NewRuntimeError(ErrorCodeLLMError, errMsg, llmErr).WithPosition(step.Pos)
	}
	if responseTurn == nil {
		errMsg := "LLM returned a nil response without an error"
		return nil, NewRuntimeError(ErrorCodeLLMError, errMsg, nil).WithPosition(step.Pos)
	}

	llmResult := responseTurn.Content

	if targetVar != "" {
		wrappedResult, wrapErr := Wrap(llmResult)
		if wrapErr != nil {
			return nil, NewRuntimeError(ErrorCodeInternal, "failed to wrap ASK result", wrapErr).WithPosition(step.Pos)
		}
		if setErr := i.SetVariable(targetVar, wrappedResult); setErr != nil {
			errMsg := fmt.Sprintf("setting variable '%s' for ASK result", targetVar)
			return nil, WrapErrorWithPosition(setErr, step.Pos, errMsg)
		}
		i.Logger().Debug("[DEBUG-INTERP] Stored ASK result in variable", "variable", targetVar)
	}

	return StringValue{Value: llmResult}, nil
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

// NeuroScript Version: 0.3.1
// File version: 0.2.1
// Purpose: Corrects call to the standalone Wrap function to handle two return values.
// filename: pkg/core/interpreter_steps_simple.go
// nlines: 236
// risk_rating: MEDIUM

package core

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step Step) (interface{}, bool, error) {
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing RETURN", "pos", posStr)

	if len(step.Values) > 1 {
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

	if step.Value != nil {
		i.Logger().Debug("[DEBUG-INTERP] Return has a single expression", "pos", posStr)
		evaluatedValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			errMsg := "evaluating return expression"
			return nil, true, WrapErrorWithPosition(err, step.Value.GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

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
	stepType := strings.ToLower(step.Type)
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	var conditionMet bool
	var valueEvaluated Value
	var evalErr error
	var errorNodePos *Position = step.Pos

	if stepType == "must" {
		if step.Value == nil {
			return nil, NewRuntimeError(ErrorCodeSyntax, "must step has nil condition expression", nil).WithPosition(step.Pos)
		}
		errorNodePos = step.Value.GetPos()
		var err error
		valueEvaluated, err = i.evaluateExpression(step.Value)
		if err != nil {
			return nil, WrapErrorWithPosition(err, errorNodePos, "evaluating condition for must")
		}
		conditionMet = isTruthy(valueEvaluated)
		if !conditionMet {
			evalErr = ErrMustConditionFailed
		}
	} else if stepType == "mustbe" {
		if step.Call == nil {
			errorNodePos = step.Pos
			return nil, NewRuntimeError(ErrorCodeSyntax, "mustbe step has nil callable expression", nil).WithPosition(errorNodePos)
		}
		errorNodePos = step.Call.GetPos()
		var errCall error
		valueEvaluated, errCall = i.evaluateExpression(step.Call)

		if errCall != nil {
			evalErr = fmt.Errorf("%w: check function %s call failed: %w", ErrMustConditionFailed, step.Call.Target.String(), errCall)
			conditionMet = false
		} else {
			boolVal, ok := valueEvaluated.(BoolValue)
			if !ok {
				typeErrMessage := fmt.Sprintf("mustbe check function %s did not return a boolean, got %s", step.Call.Target.String(), TypeOf(valueEvaluated))
				evalErr = fmt.Errorf("%w: %s", ErrMustConditionFailed, typeErrMessage)
				conditionMet = false
			} else {
				conditionMet = boolVal.Value
				if !conditionMet {
					evalErr = fmt.Errorf("%w: check function %s returned false", ErrMustConditionFailed, step.Call.Target.String())
				}
			}
		}
	} else {
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("executeMust called with invalid step type: %s", step.Type), ErrInternal).WithPosition(step.Pos)
	}

	if !conditionMet {
		if evalErr != nil {
			if _, ok := evalErr.(*RuntimeError); ok {
				if errors.Is(evalErr, ErrMustConditionFailed) {
					detailMsg := fmt.Sprintf("must condition evaluated to false (value was %s: %v)", valueEvaluated.Type(), valueEvaluated)
					return nil, NewRuntimeError(ErrorCodeMustFailed, detailMsg, ErrMustConditionFailed).WithPosition(errorNodePos)
				}
				return nil, evalErr
			}
			return nil, NewRuntimeError(ErrorCodeMustFailed, evalErr.Error(), evalErr).WithPosition(errorNodePos)
		}
		return nil, NewRuntimeError(ErrorCodeMustFailed, "must condition failed", ErrMustConditionFailed).WithPosition(errorNodePos)
	}

	i.Logger().Debug("[DEBUG-INTERP] MUST/MUSTBE condition TRUE", "type", strings.ToUpper(stepType), "pos", posStr)
	return valueEvaluated, nil
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
		// FIX: Use the standalone Wrap function and handle the error.
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

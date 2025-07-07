// NeuroScript Version: 0.5.2
// File version: 58
// Purpose: Corrected the on_error handler invocation logic to properly check the current interpreter's own error handler stack, fixing the root cause of all fixture test failures.
// filename: pkg/interpreter/interpreter_exec.go
// nlines: 300
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func (i *Interpreter) executeSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError) (lang.Value, bool, bool, error) {
	return i.recExecuteSteps(steps, isInHandler, activeError, 0)
}

func getStepSubjectForLogging(step ast.Step) string {
	// ... (implementation is unchanged)
	switch strings.ToLower(step.Type) {
	case "set", "assign":
		if len(step.LValues) > 0 {
			var parts []string
			for _, lval := range step.LValues {
				if lval != nil {
					parts = append(parts, lval.String())
				}
			}
			return strings.Join(parts, ", ")
		}
	case "call":
		if step.Call != nil {
			return step.Call.Target.String()
		}
	case "ask":
		var promptStr string
		if len(step.Values) > 0 && step.Values[0] != nil {
			promptStr = step.Values[0].String()
		}
		if step.AskIntoVar != "" {
			return fmt.Sprintf("into %s (prompt: %s)", step.AskIntoVar, promptStr)
		}
		return fmt.Sprintf("prompt: %s", promptStr)

	case "for_each", "for":
		return fmt.Sprintf("loopVar: %s, collection: %s", step.LoopVarName, step.Collection.String())
	case "must", "mustbe":
		if step.Cond != nil {
			return fmt.Sprintf("must %s", step.Cond.String())
		}
		if step.Call != nil {
			return fmt.Sprintf("mustbe %s", step.Call.String())
		}
		return "must last"
	case "return":
		if len(step.Values) > 0 {
			var parts []string
			for _, v := range step.Values {
				if v != nil {
					parts = append(parts, v.String())
				}
			}
			return fmt.Sprintf("returning %d values: %s", len(step.Values), strings.Join(parts, ", "))
		}
		return "return (nil)"
	case "emit":
		if len(step.Values) > 0 {
			var parts []string
			for _, v := range step.Values {
				if v != nil {
					parts = append(parts, v.String())
				}
			}
			return strings.Join(parts, ", ")
		}
		return "emit (empty)"
	}
	return "<no specific subject>"
}

func (i *Interpreter) recExecuteSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError, depth int) (finalResult lang.Value, wasReturn bool, wasCleared bool, finalError error) {
	finalResult = &lang.NilValue{}

	for stepIdx, step := range steps {
		var stepResult lang.Value
		var stepErr error

		stepTypeLower := strings.ToLower(step.Type)
		subject := getStepSubjectForLogging(step)
		fmt.Printf("[DEBUG-EXEC] Step %d: %s %s (Pos: %s)\n", stepIdx, strings.ToUpper(stepTypeLower), subject, step.GetPos())

		switch stepTypeLower {
		case "set", "assign":
			stepResult, stepErr = i.executeSet(step)
		case "call":
			stepResult, stepErr = i.executeCall(step)
		case "return":
			if isInHandler {
				stepErr = lang.NewRuntimeError(lang.ErrorCodeReturnViolation, "'return' is not permitted inside an on_error block", lang.ErrReturnViolation).WithPosition(&step.Position)
			} else {
				var returnValue lang.Value
				returnValue, wasReturn, stepErr = i.executeReturn(step)
				if stepErr == nil && wasReturn {
					i.lastCallResult = returnValue
					return returnValue, true, false, nil
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step)
		case "if":
			var ifReturned, ifCleared bool
			var ifBlockResult lang.Value
			ifBlockResult, ifReturned, ifCleared, stepErr = i.executeIf(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = ifBlockResult
				if ifReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if ifCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "while":
			var whileReturned, whileCleared bool
			var whileBlockResult lang.Value
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, wasCleared, nil
				}
				if whileCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "for", "for_each":
			var forReturned, forCleared bool
			var forResult lang.Value
			forResult, forReturned, forCleared, stepErr = i.executeFor(step, isInHandler, activeError)
			if stepErr == nil {
				stepResult = forResult
				if forReturned {
					return forResult, true, false, nil
				}
				if forCleared {
					wasCleared = true
					activeError = nil
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step)
		case "fail":
			stepErr = i.executeFail(step)
		case "on_error":
			continue
		case "clear_error":
			var clearedNow bool
			clearedNow, stepErr = i.executeClearError(step, isInHandler)
			if stepErr == nil && clearedNow {
				wasCleared = true
				activeError = nil
			}
		case "ask":
			stepResult, stepErr = i.executeAsk(step)
		case "break":
			stepErr = i.executeBreak(step)
		case "continue":
			stepErr = i.executeContinue(step)
		case "expression_statement":
			if len(step.Values) > 0 {
				stepResult, stepErr = i.evaluate.Expression(step.Values[0])
			} else {
				stepResult = &lang.NilValue{}
			}
		default:
			errMsg := fmt.Sprintf("unknown step type '%s'", step.Type)
			stepErr = lang.NewRuntimeError(lang.ErrorCodeUnknownKeyword, errMsg, lang.ErrUnknownKeyword).WithPosition(&step.Position)
		}

		if stepErr != nil {
			fmt.Printf("[DEBUG-EXEC] Step %d FAILED. Error: %v\n", stepIdx, stepErr)
		} else {
			fmt.Printf("[DEBUG-EXEC] Step %d SUCCEEDED. Result: %v (%T)\n", stepIdx, stepResult, stepResult)
		}

		if stepErr != nil {
			rtErr := ensureRuntimeError(stepErr, &step.Position, stepTypeLower)

			if errors.Is(rtErr.Unwrap(), lang.ErrBreak) || errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
				return nil, false, wasCleared, rtErr
			}

			// FIX: Check the *current* interpreter's stack for a handler.
			if !isInHandler && len(i.state.errorHandlerStack) > 0 {
				handlerBlock := i.state.errorHandlerStack[len(i.state.errorHandlerStack)-1]
				handlerToExecute := handlerBlock[0] // Assuming one handler per level for now

				// Execute handler in the *current* interpreter context.
				_, _, handlerCleared, handlerErr := i.executeSteps(handlerToExecute.Body, true, rtErr)

				if handlerErr != nil {
					return nil, false, false, ensureRuntimeError(handlerErr, &handlerToExecute.Position, "ON_ERROR_HANDLER")
				}

				if handlerCleared {
					// The error was handled and we can continue execution.
					stepErr = nil
					continue
				} else {
					// The handler ran but didn't clear the error, so it propagates.
					return nil, false, false, rtErr
				}
			} else {
				// No handler available or we're already inside one.
				return nil, false, wasCleared, rtErr
			}
		}

		if stepErr == nil {
			if shouldUpdateLastResult(stepTypeLower) {
				finalResult = stepResult
				i.lastCallResult = stepResult
			}
		}
	}

	return finalResult, false, wasCleared, nil
}

func (i *Interpreter) executeCall(step ast.Step) (lang.Value, error) {
	if step.Call == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(&step.Position)
	}
	return i.evaluate.Expression(step.Call)
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentPos *lang.Position, blockType string, isInHandler bool, activeError *lang.RuntimeError, depth int) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]ast.Step)
	if !ok {
		if blockValue == nil {
			return &lang.NilValue{}, false, false, nil
		}
		errMsg := fmt.Sprintf("internal error: invalid block format for %s - expected []Step, got %T", blockType, blockValue)
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, lang.ErrInternal).WithPosition(parentPos)
	}
	return i.recExecuteSteps(steps, isInHandler, activeError, depth)
}

func shouldUpdateLastResult(stepTypeLower string) bool {
	switch stepTypeLower {
	case "set", "assign", "emit", "ask", "call", "expression_statement":
		return true
	default:
		return false
	}
}

func ensureRuntimeError(err error, pos *lang.Position, context string) *lang.RuntimeError {
	if rtErr, ok := err.(*lang.RuntimeError); ok {
		if rtErr.Position == nil && pos != nil {
			return rtErr.WithPosition(pos)
		}
		return rtErr
	}
	return lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error during %s", context), err).WithPosition(pos)
}

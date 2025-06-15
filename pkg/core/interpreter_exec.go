// NeuroScript Version: 0.4.1
// File version: 11
// Purpose: Added detailed debugging to executeSteps to trace return value issues.
// filename: pkg/core/interpreter_exec.go
// nlines: 300
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	"strings"
)

// getStepSubjectForLogging creates a descriptive string for a step's main subject for logging.
func getStepSubjectForLogging(step Step) string {
	switch strings.ToLower(step.Type) {
	case "set", "assign":
		if step.LValue != nil {
			return step.LValue.String()
		}
	case "call":
		if step.Call != nil {
			return step.Call.Target.String()
		}
	case "ask":
		if step.AskIntoVar != "" {
			return fmt.Sprintf("into %s (prompt: %s)", step.AskIntoVar, step.Value.String())
		}
		if step.Value != nil {
			return fmt.Sprintf("prompt: %s", step.Value.String())
		}
		return "ask"
	case "for_each", "for":
		return fmt.Sprintf("loopVar: %s, collection: %s", step.LoopVarName, step.Collection.String())
	case "must", "mustbe":
		if strings.ToLower(step.Type) == "mustbe" && step.Call != nil {
			return fmt.Sprintf("mustbe %s", step.Call.Target.String())
		}
		if step.Value != nil {
			return fmt.Sprintf("must %s", step.Value.String())
		}
		return "must condition"
	case "return":
		if len(step.Values) > 0 {
			return fmt.Sprintf("returning %d values", len(step.Values))
		}
		if step.Value != nil {
			return fmt.Sprintf("returning %s", step.Value.String())
		}
		return "return (nil)"
	case "emit":
		if step.Value != nil {
			return step.Value.String()
		}
		return "emit (empty)"
	}
	return "<no specific subject>"
}

// executeSteps iterates through and executes steps, handling control flow and errors.
func (i *Interpreter) executeSteps(steps []Step, isInHandler bool, activeError *RuntimeError) (finalResult interface{}, wasReturn bool, wasCleared bool, finalError error) {
	modeStr := "normal"
	activeErrorStr := "nil"
	if isInHandler {
		modeStr = "handler"
		if activeError != nil {
			activeErrorStr = fmt.Sprintf("%d: %s", activeError.Code, activeError.Message)
		}
	}
	i.Logger().Debug("[DEBUG-INTERP] Executing steps", "count", len(steps), "mode", modeStr, "activeError", activeErrorStr)

	var currentErrorHandler *Step = nil

	for stepNum, step := range steps {
		stepResult := interface{}(nil)
		stepErr := error(nil)

		stepTypeLower := strings.ToLower(step.Type)
		stepTypeStr := strings.ToUpper(stepTypeLower)
		stepSubjectStr := getStepSubjectForLogging(step)
		logPos := "<unknown_pos>"
		if step.Pos != nil {
			logPos = step.Pos.String()
		}
		i.Logger().Debug("[DEBUG-INTERP]   Executing Step", "step_num", stepNum+1, "type", stepTypeStr, "subject", stepSubjectStr, "pos", logPos)

		switch stepTypeLower {
		case "set", "assign":
			stepResult, stepErr = i.executeSet(step)
		case "call":
			if step.Call != nil {
				var callRes interface{}
				callRes, stepErr = i.evaluateExpression(step.Call)
				if stepErr == nil {
					stepResult = callRes
				}
			} else {
				errMsg := fmt.Sprintf("step %d: 'call' step type without Call details", stepNum+1)
				stepErr = NewRuntimeError(ErrorCodeInternal, errMsg, errors.New(errMsg)).WithPosition(step.Pos)
			}
		case "return":
			if isInHandler {
				errMsg := fmt.Sprintf("step %d: 'return' statement is not permitted inside an on_error block", stepNum+1)
				stepErr = NewRuntimeError(ErrorCodeReturnViolation, errMsg, ErrReturnViolation).WithPosition(step.Pos)
			} else {
				var returnValue interface{}
				returnValue, wasReturn, stepErr = i.executeReturn(step)
				// DEBUGGING: Log what executeReturn gives us.
				i.Logger().Debug("[DEBUG-INTERP] RETURN statement executed", "returnValue", returnValue, "returnValueType", fmt.Sprintf("%T", returnValue), "wasReturn", wasReturn, "err", stepErr)
				if stepErr == nil && wasReturn {
					i.lastCallResult = returnValue
					return returnValue, true, false, nil
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step)
		case "if":
			var ifReturned, ifCleared bool
			var ifBlockResult interface{}
			ifBlockResult, ifReturned, ifCleared, stepErr = i.executeIf(step, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			if stepErr == nil {
				stepResult = ifBlockResult
				if ifReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if ifCleared {
					wasCleared = true
				}
			}
		case "while":
			var whileReturned, whileCleared bool
			var whileBlockResult interface{}
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) {
				stepErr = nil
			} else if errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] CONTINUE propagated out of WHILE loop unexpectedly", "step_num", stepNum+1)
			} else if stepErr == nil {
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if whileCleared {
					wasCleared = true
				}
			}
		case "for", "for_each":
			var forReturned, forCleared bool
			var forBlockResult interface{}
			forBlockResult, forReturned, forCleared, stepErr = i.executeFor(step, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) {
				stepErr = nil
			} else if errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] CONTINUE propagated out of FOR loop unexpectedly", "step_num", stepNum+1)
			} else if stepErr == nil {
				stepResult = forBlockResult
				if forReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if forCleared {
					wasCleared = true
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step)
		case "fail":
			stepErr = i.executeFail(step)
		case "on_error":
			currentErrorHandler, stepErr = i.executeOnError(step)
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
		default:
			errMsg := fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type)
			stepErr = NewRuntimeError(ErrorCodeUnknownKeyword, errMsg, ErrUnknownKeyword).WithPosition(step.Pos)
		}

		if stepErr != nil {
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			rtErr := ensureRuntimeError(stepErr, step.Pos, stepTypeStr)
			if currentErrorHandler != nil {
				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(currentErrorHandler.Body, true, rtErr)
				if handlerErr != nil {
					return nil, false, false, ensureRuntimeError(handlerErr, currentErrorHandler.Pos, "ON_ERROR_HANDLER")
				}
				if handlerReturned {
					return nil, true, false, NewRuntimeError(ErrorCodeReturnViolation, "return from on_error handler is not permitted", ErrReturnViolation).WithPosition(currentErrorHandler.Pos)
				}
				if handlerCleared {
					wasCleared = true
					activeError = nil
					stepErr = nil // The error was handled and cleared.
				} else {
					return nil, false, false, rtErr // Propagate original error.
				}
			} else {
				return nil, false, false, rtErr // No handler, propagate error.
			}
		}

		if stepErr == nil {
			if shouldUpdateLastResult(stepTypeLower) {
				i.lastCallResult = stepResult
			}
		}
	}

	// DEBUGGING: Log what we are about to return at the end of the block.
	i.Logger().Debug("[DEBUG-INTERP] Finished executing steps block normally", "final_lastCallResult", i.lastCallResult, "final_lastCallResultType", fmt.Sprintf("%T", i.lastCallResult), "final_wasCleared", wasCleared)
	return i.lastCallResult, false, wasCleared, nil
}

// executeBlock is a wrapper around executeSteps that handles casting and logging for block-based statements like if/for.
func (i *Interpreter) executeBlock(blockValue interface{}, parentPos *Position, blockType string, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		// This handles empty blocks (e.g., `if true then end`), which are valid.
		if blockValue == nil {
			return nil, false, false, nil
		}
		errMsg := fmt.Sprintf("internal error: invalid block format for %s - expected []Step, got %T", blockType, blockValue)
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal).WithPosition(parentPos)
	}

	return i.executeSteps(steps, isInHandler, activeError)
}

// shouldUpdateLastResult determines if a step type should modify the interpreter's `lastCallResult`.
func shouldUpdateLastResult(stepTypeLower string) bool {
	switch stepTypeLower {
	case "set", "assign", "emit", "must", "mustbe", "ask", "call":
		return true
	default:
		return false
	}
}

// ensureRuntimeError wraps a generic error in a RuntimeError if it isn't one already.
func ensureRuntimeError(err error, pos *Position, context string) *RuntimeError {
	if rtErr, ok := err.(*RuntimeError); ok {
		if rtErr.Position == nil && pos != nil {
			return rtErr.WithPosition(pos)
		}
		return rtErr
	}
	return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error during %s", context), err).WithPosition(pos)
}

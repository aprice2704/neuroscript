// NeuroScript Version: 0.5.2
// File version: 27
// Purpose: Unified error handler registration and lookup to fix bug where handlers in nested scopes were not found.
// filename: pkg/runtime/interpreter_exec.go
// nlines: 278
// risk_rating: HIGH

package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// getStepSubjectForLogging creates a descriptive string for a step's main subject for logging.
func getStepSubjectForLogging(step ast.Step) string {
	// NOTE: This function is unchanged from your version.
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

// executeSteps iterates through and executes steps, handling control flow and errors.
func (i *Interpreter) executeSteps(steps []Step, isInHandler bool, activeError *RuntimeError) (finalResult Value, wasReturn bool, wasCleared bool, finalError error) {
	modeStr := "normal"
	if isInHandler {
		modeStr = "handler"
	}
	i.Logger().Debug("[DEBUG-INTERP] Executing steps", "count", len(steps), "mode", modeStr)

	// FIX: Removed `currentErrorHandler` local variable to rely solely on the interpreter's handler stack.
	finalResult = NilValue{}

	for stepNum, step := range steps {
		var stepResult Value
		var stepErr error

		stepTypeLower := strings.ToLower(step.Type)
		logPos := "<unknown_pos>"
		if step.Pos != nil {
			logPos = step.Pos.String()
		}
		i.Logger().Debug("[DEBUG-INTERP]   Executing ast.Step", "step_num", stepNum+1, "type", strings.ToUpper(stepTypeLower), "subject", getStepSubjectForLogging(step), "pos", logPos)

		switch stepTypeLower {
		// ... (cases "set" through "fail" are unchanged)
		case "set", "assign":
			stepResult, stepErr = i.executeSet(step)
		case "call":
			if step.Call != nil {
				stepResult, stepErr = i.evaluate.Expression(step.Call)
			} else {
				stepErr = lang.NewRuntimeError(ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(step.Pos)
			}
		case "return":
			if isInHandler {
				stepErr = lang.NewRuntimeError(ErrorCodeReturnViolation, "'return' is not permitted inside an on_error block", ErrReturnViolation).WithPosition(step.Pos)
			} else {
				var returnValue Value
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
			var ifBlockResult Value
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
			var whileBlockResult Value
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) {
				stepErr = nil
			} else if errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] CONTINUE propagated out of WHILE loop unexpectedly", "step_num", stepNum+1)
				stepErr = nil
			} else if stepErr == nil {
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, wasCleared, nil
				}
				if whileCleared {
					wasCleared = true
				}
			}
		case "for", "for_each":
			var forReturned, forCleared bool
			var forResult Value
			forResult, forReturned, forCleared, stepErr = i.executeFor(step, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) {
				stepErr = nil
			} else if errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] CONTINUE propagated out of FOR loop unexpectedly", "step_num", stepNum+1)
			} else if stepErr == nil {
				stepResult = forResult
				if forReturned {
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
			// FIX: This now only calls the registration function. We no longer use a local variable.
			// This assumes i.executeOnError correctly pushes the handler to i.errorHandlerStack.
			_, stepErr = i.executeOnError(step)
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
			errMsg := fmt.Sprintf("unknown step type '%s'", step.Type)
			stepErr = lang.NewRuntimeError(ErrorCodeUnknownKeyword, errMsg, ErrUnknownKeyword).WithPosition(step.Pos)
		}

		if stepErr != nil {
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			rtErr := ensureRuntimeError(stepErr, step.Pos, stepTypeLower)

			// FIX: Simplified handler lookup to ONLY use the central interpreter stack.
			var handlerToExecute *Step = nil
			if !isInHandler {
				if len(i.errorHandlerStack) > 0 {
					procHandlers := i.errorHandlerStack[len(i.errorHandlerStack)-1]
					if len(procHandlers) > 0 {
						// This still takes the first generic handler. Assumed correct for now.
						handlerToExecute = procHandlers[0]
					}
				}
			}

			if handlerToExecute != nil {
				var handlerCleared bool
				var handlerErr error
				_, _, handlerCleared, handlerErr = i.executeSteps(handlerToExecute.Body, true, rtErr)

				if handlerErr != nil {
					return nil, false, false, ensureRuntimeError(handlerErr, handlerToExecute.Pos, "ON_ERROR_HANDLER")
				}

				if handlerCleared {
					wasCleared = true
					activeError = nil
					stepErr = nil
					continue // This ensures execution resumes after the failed statement.
				} else {
					// The re-raise behavior is correct per your instruction.
					return nil, false, false, rtErr
				}
			} else {
				return nil, false, false, rtErr
			}
		}

		if stepErr == nil {
			if shouldUpdateLastResult(stepTypeLower) {
				finalResult = stepResult
				i.lastCallResult = stepResult
			}
		}
	}

	i.Logger().Debug("[DEBUG-INTERP] Finished executing steps block normally")
	return finalResult, false, wasCleared, nil
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentPos *lang.Position, blockType string, isInHandler bool, activeError *RuntimeError) (result Value, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil {
			return NilValue{}, false, false, nil
		}
		errMsg := fmt.Sprintf("internal error: invalid block format for %s - expected []Step, got %T", blockType, blockValue)
		return nil, false, false, lang.NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal).WithPosition(parentPos)
	}
	return i.executeSteps(steps, isInHandler, activeError)
}

func shouldUpdateLastResult(stepTypeLower string) bool {
	switch stepTypeLower {
	case "set", "assign", "emit", "must", "mustbe", "ask", "call":
		return true
	default:
		return false
	}
}

func ensureRuntimeError(err error, pos *lang.Position, context string) *RuntimeError {
	if rtErr, ok := err.(*RuntimeError); ok {
		if rtErr.Position == nil && pos != nil {
			return rtErr.WithPosition(pos)
		}
		return rtErr
	}
	return lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error during %s", context), err).WithPosition(pos)
}

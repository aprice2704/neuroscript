// NeuroScript Version: 0.5.2
// File version: 51
// Purpose: Added comprehensive debug logging for all loop counter variables (i, j, inner_count, outer_count).
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

// getStepSubjectForLogging creates a descriptive string for a step's main subject for logging.
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

// recExecuteSteps is the recursive core of the execution loop, with added depth for debugging.
func (i *Interpreter) recExecuteSteps(steps []ast.Step, isInHandler bool, activeError *lang.RuntimeError, depth int) (finalResult lang.Value, wasReturn bool, wasCleared bool, finalError error) {

	finalResult = &lang.NilValue{}

	for _, step := range steps {
		var stepResult lang.Value
		var stepErr error

		// **NEW DEBUGGING:** Print all relevant counter states at the start of every step.
		var stateParts []string
		if val, exists := i.GetVariable("i"); exists {
			stateParts = append(stateParts, fmt.Sprintf("i=%v", val))
		}
		if val, exists := i.GetVariable("j"); exists {
			stateParts = append(stateParts, fmt.Sprintf("j=%v", val))
		}
		if val, exists := i.GetVariable("outer_count"); exists {
			stateParts = append(stateParts, fmt.Sprintf("outer=%v", val))
		}
		if val, exists := i.GetVariable("inner_count"); exists {
			stateParts = append(stateParts, fmt.Sprintf("inner=%v", val))
		}
		stateStr := strings.Join(stateParts, ", ")
		if stateStr == "" {
			stateStr = "counters not set"
		}

		stepTypeLower := strings.ToLower(step.Type)

		switch stepTypeLower {
		case "set", "assign":
			stepResult, stepErr = i.executeSet(step)
		case "call":
			if step.Call != nil {
				stepResult, stepErr = i.evaluate.Expression(step.Call)
			} else {
				stepErr = lang.NewRuntimeError(lang.ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(&step.Position)
			}
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
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step)
		case "fail":
			stepErr = i.executeFail(step)
		case "on_error":
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
			stepErr = lang.NewRuntimeError(lang.ErrorCodeUnknownKeyword, errMsg, lang.ErrUnknownKeyword).WithPosition(&step.Position)
		}

		if stepErr != nil {

			rtErr := ensureRuntimeError(stepErr, &step.Position, stepTypeLower)

			if errors.Is(rtErr.Unwrap(), lang.ErrBreak) || errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
				return nil, false, wasCleared, rtErr
			}

			var handlerToExecute *ast.Step = nil
			if !isInHandler {
				if len(i.state.errorHandlerStack) > 0 {
					procHandlers := i.state.errorHandlerStack[len(i.state.errorHandlerStack)-1]
					if len(procHandlers) > 0 {
						handlerToExecute = procHandlers[0]
					}
				}
			}

			if handlerToExecute != nil {
				var handlerCleared bool
				var handlerErr error
				_, _, handlerCleared, handlerErr = i.recExecuteSteps(handlerToExecute.Body, true, rtErr, depth+1)

				if handlerErr != nil {
					//				fmt.Printf("%s    (ON_ERROR handler ITSELF failed. Propagating handler's error.)\n", indent)
					return nil, false, false, ensureRuntimeError(handlerErr, &handlerToExecute.Position, "ON_ERROR_HANDLER")
				}

				if handlerCleared {
					wasCleared = true
					activeError = nil
					stepErr = nil
					continue
				} else {
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

	return finalResult, false, wasCleared, nil
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
	case "set", "assign", "emit", "ask", "call":
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

// NeuroScript Version: 0.3.0
// File version: 0.0.4 // Defensive shouldUpdateLast logic for IF/control flow.
// filename: pkg/core/interpreter_exec.go
package core

import (
	"errors"
	"fmt"
	"strings"
)

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
	i.Logger().Info("[DEBUG-INTERP] Executing steps", "count", len(steps), "mode", modeStr, "activeError", activeErrorStr)

	var currentErrorHandler *Step = nil

	for stepNum, step := range steps {
		stepResult := interface{}(nil)
		stepErr := error(nil)

		stepTypeLower := strings.ToLower(step.Type) // Lowercase once
		stepTypeStr := strings.ToUpper(stepTypeLower)
		stepTargetStr := step.Target
		i.Logger().Info("[DEBUG-INTERP]   Executing Step", "step_num", stepNum+1, "type", stepTypeStr, "target", stepTargetStr)

		switch stepTypeLower {
		case "set":
			stepResult, stepErr = i.executeSet(step, stepNum, isInHandler, activeError)
		case "call":
			if step.Call != nil {
				var callRes interface{}
				callRes, stepErr = i.evaluateExpression(step.Call)
				if stepErr == nil {
					stepResult = callRes
				}
			} else {
				stepErr = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("step %d: 'call' step type without Call details", stepNum+1), ErrUnknownKeyword)
			}
		case "return":
			if isInHandler {
				stepErr = NewRuntimeError(ErrorCodeReturnViolation, fmt.Sprintf("step %d: 'return' statement is not permitted inside an on_error block", stepNum+1), ErrReturnViolation)
			} else {
				var returnValue interface{}
				returnValue, wasReturn, stepErr = i.executeReturn(step, stepNum, isInHandler, activeError)
				if stepErr == nil && wasReturn {
					i.lastCallResult = returnValue
					return returnValue, true, false, nil
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step, stepNum, isInHandler, activeError)
		case "if":
			var ifReturned, ifCleared bool
			var ifBlockResult interface{}
			ifBlockResult, ifReturned, ifCleared, stepErr = i.executeIf(step, stepNum, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			if stepErr == nil {
				stepResult = ifBlockResult // This is the result of an explicit return from the block
				if ifReturned {
					i.lastCallResult = stepResult // If the IF block returned, update LAST
					return stepResult, true, false, nil
				}
				if ifCleared {
					wasCleared = true
				}
			}
		case "while":
			var whileReturned, whileCleared bool
			var whileBlockResult interface{}
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			if stepErr == nil {
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult // If WHILE block returned, update LAST
					return stepResult, true, false, nil
				}
				if whileCleared {
					wasCleared = true
				}
			}
		case "for":
			var forReturned, forCleared bool
			var forBlockResult interface{}
			forBlockResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				return nil, false, wasCleared, stepErr
			}
			if stepErr == nil {
				stepResult = forBlockResult
				if forReturned {
					i.lastCallResult = stepResult // If FOR block returned, update LAST
					return stepResult, true, false, nil
				}
				if forCleared {
					wasCleared = true
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step, stepNum, isInHandler, activeError)
		case "fail":
			stepErr = i.executeFail(step, stepNum, isInHandler, activeError)
			stepResult = nil
		case "on_error":
			var handlerStep *Step
			handlerStep, stepErr = i.executeOnError(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				currentErrorHandler = handlerStep
			}
			stepResult = nil
		case "clear_error":
			var clearedNow bool
			clearedNow, stepErr = i.executeClearError(step, stepNum, isInHandler, activeError)
			if stepErr == nil && clearedNow {
				wasCleared = true
			}
			stepResult = nil
		case "ask":
			stepResult, stepErr = i.executeAsk(step, stepNum, isInHandler, activeError)
		case "break":
			stepResult, stepErr = i.executeBreak(step, stepNum, isInHandler, activeError)
		case "continue":
			stepResult, stepErr = i.executeContinue(step, stepNum, isInHandler, activeError)
		default:
			stepErr = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type), ErrUnknownKeyword)
			stepResult = nil
		}

		if stepErr != nil {
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP] Propagating control flow signal", "signal", stepErr, "step_num", stepNum+1)
				return nil, false, wasCleared, stepErr
			}
			rtErr, isRuntimeErr := stepErr.(*RuntimeError)
			if !isRuntimeErr {
				i.Logger().Warn("Wrapping non-RuntimeError from step execution", "original_error", stepErr, "step_num", stepNum+1)
				rtErr = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error during step %d (%s)", stepNum+1, stepTypeStr), stepErr)
			}

			if currentErrorHandler != nil {
				i.Logger().Info("Error occurred, executing active ON_ERROR handler", "original_error", rtErr, "step_num", stepNum+1)
				handlerSteps := currentErrorHandler.Body
				outerHandler := currentErrorHandler
				currentErrorHandler = nil
				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(handlerSteps, true, rtErr)
				if handlerErr == nil && !handlerReturned {
					currentErrorHandler = outerHandler
				} else {
					i.Logger().Debug("Not restoring error handler due to handler error or return", "handler_error", handlerErr, "handler_returned", handlerReturned)
				}

				if handlerErr != nil {
					i.Logger().Warn("Error occurred inside ON_ERROR handler, propagating handler error", "handler_error", handlerErr, "original_error", rtErr)
					if _, ok := handlerErr.(*RuntimeError); !ok {
						errMsg := fmt.Sprintf("internal error processing on_error handler at %s", outerHandler.Pos.String())
						handlerErr = NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, handlerErr))
					}
					return nil, false, false, handlerErr
				}
				if handlerReturned {
					errMsg := "internal error: 'return' propagated incorrectly from on_error handler"
					finalError = NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
					i.Logger().Error("Internal error: Return from handler", "error", finalError)
					return nil, false, false, finalError
				}
				if handlerCleared {
					i.Logger().Info("ON_ERROR handler executed and cleared the error", "cleared_error", rtErr)
					stepErr = nil
					wasCleared = true
				} else {
					i.Logger().Info("ON_ERROR handler executed but did not clear error, propagating original error", "original_error", rtErr)
					finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, stepTypeStr, rtErr)
					return nil, false, false, finalError
				}
			} else {
				i.Logger().Debug("Error occurred, no active handler, propagating", "error", rtErr, "step_num", stepNum+1)
				finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, stepTypeStr, rtErr)
				return nil, false, false, finalError
			}
		}

		if stepErr == nil {
			// Determine if this step type should update the interpreter's lastCallResult.
			// Control flow structures (if, while, for) do not update it themselves;
			// their internal value-producing steps do.
			// 'return' updates it directly when it exits.
			shouldUpdateLast := false
			if stepTypeLower == "set" || stepTypeLower == "emit" ||
				stepTypeLower == "must" || stepTypeLower == "mustbe" ||
				stepTypeLower == "ask" || stepTypeLower == "call" {
				shouldUpdateLast = true
			}

			if shouldUpdateLast {
				i.lastCallResult = stepResult
				i.Logger().Debug("[DEBUG-INTERP]     Step successful. Last result updated", "step_num", stepNum+1, "last_result", i.lastCallResult, "last_result_type", fmt.Sprintf("%T", i.lastCallResult))
			} else {
				// For steps that don't update LAST (like if, while, for, on_error, clear_error, break, continue)
				// i.lastCallResult remains unchanged by this specific step's completion.
				// It will reflect the result of the last *value-producing* step.
				i.Logger().Debug("[DEBUG-INTERP]     Step successful", "step_num", stepNum+1, "type", stepTypeStr, "info", fmt.Sprintf("does not update LAST directly (current LAST: %v (%T))", i.lastCallResult, i.lastCallResult))
			}
		}
	}

	i.Logger().Info("[DEBUG-INTERP] Finished executing steps block normally", "mode", modeStr)
	return nil, false, wasCleared, nil
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		errMsg := fmt.Sprintf("step %d (%s): invalid block format - expected []Step", parentStepNum+1, blockType)
		if blockValue != nil {
			errMsg = fmt.Sprintf("%s, got %T", errMsg, blockValue)
		} else {
			i.Logger().Debug("Entering block execution for nil/empty block", "block_type", blockType, "parent_step", parentStepNum+1)
			return nil, false, false, nil
		}
		err = NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
		i.Logger().Error("Invalid block format", "error", err)
		return nil, false, false, err
	}

	if len(steps) == 0 {
		i.Logger().Debug("Entering empty block execution", "block_type", blockType, "parent_step", parentStepNum+1)
		return nil, false, false, nil
	}

	activeErrorStr := "nil"
	if isInHandler {
		if activeError != nil {
			activeErrorStr = fmt.Sprintf("%d", activeError.Code)
		}
	}
	i.Logger().Info(">> Entering block execution",
		"block_type", blockType,
		"handler_mode", isInHandler,
		"parent_step", parentStepNum+1,
		"step_count", len(steps),
		"active_error_code", activeErrorStr)

	// Store lastCallResult before executing the block
	// lastResultBeforeBlock := i.lastCallResult

	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	// If the block did not have an explicit return, the i.lastCallResult
	// should reflect the last value-producing step *within* the block.
	// The 'result' returned here from executeSteps for the block itself is nil if no return.
	// The 'shouldUpdateLast' logic within executeSteps handles i.lastCallResult correctly.

	i.Logger().Info("<< Exiting block execution",
		"block_type", blockType,
		"parent_step", parentStepNum+1,
		"result_from_block_return", result, // This is nil if no explicit return in block
		"was_return", wasReturn,
		"was_cleared", wasCleared,
		"error", err,
		"lastCallResult_after_block", fmt.Sprintf("%v (%T)", i.lastCallResult, i.lastCallResult))

	return result, wasReturn, wasCleared, err
}

// nlines: 252
// risk_rating: HIGH

// filename: pkg/core/interpreter_exec.go
package core

import (
	// Keep for Is/As usage
	"fmt"
	"strings"
)

// executeSteps iterates through and executes steps, handling control flow and errors.
// Takes context: isInHandler bool, activeError *RuntimeError (the error being handled).
// Returns: finalResult (nil on implicit return), wasReturn, wasCleared, finalError (error should be *RuntimeError if script-related).
func (i *Interpreter) executeSteps(steps []Step, isInHandler bool, activeError *RuntimeError) (finalResult interface{}, wasReturn bool, wasCleared bool, finalError error) {
	mode := "normal"
	errorStr := "nil"
	if isInHandler {
		mode = "handler"
		if activeError != nil {
			errorStr = fmt.Sprintf("%d: %s", activeError.Code, activeError.Message)
		}
	}
	i.Logger().Info("[DEBUG-INTERP] Executing %d steps (%s mode, activeError: %s)...", len(steps), mode, errorStr)

	var currentErrorHandler *Step = nil
	// REMOVED: var lastStepResult interface{} = nil // No longer needed

	for stepNum, step := range steps {
		stepResult := interface{}(nil)
		stepErr := error(nil)

		i.Logger().Info("[DEBUG-INTERP]   Step %d: Type=%s, Target=%s", stepNum+1, strings.ToUpper(step.Type), step.Target)

		// --- Execute Step ---
		switch strings.ToLower(step.Type) {
		case "set":
			stepResult, stepErr = i.executeSet(step, stepNum, isInHandler, activeError)
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
			var ifResult interface{}
			ifResult, ifReturned, ifCleared, stepErr = i.executeIf(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				stepResult = ifResult
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
			var whileResult interface{}
			whileResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				stepResult = whileResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if whileCleared {
					wasCleared = true
				}
			}
		case "for":
			var forReturned, forCleared bool
			var forResult interface{}
			forResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				stepResult = forResult
				if forReturned {
					i.lastCallResult = stepResult
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
			stepResult, stepErr = i.executeAsk(step, stepNum, isInHandler, activeError) // Call the (currently missing) function
		default:
			stepErr = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type), ErrUnknownKeyword)
			stepResult = nil
		}
		// --- End Execute Step ---

		// --- Error Handling Check ---
		if stepErr != nil {
			rtErr, isRuntimeErr := stepErr.(*RuntimeError)
			if !isRuntimeErr {
				i.Logger().Warn("[WARN INTERP] Wrapping non-RuntimeError: %v", stepErr)
				rtErr = NewRuntimeError(ErrorCodeInternal, stepErr.Error(), stepErr)
			}

			if currentErrorHandler != nil {
				i.Logger().Info("[DEBUG-INTERP]   Error occurred: %v. Handler is active. Executing handler.", rtErr)
				handlerSteps, ok := currentErrorHandler.Value.([]Step)
				if !ok {
					finalError = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error: on_error step value is not []Step (%T)", currentErrorHandler.Value), ErrInternal)
					return nil, false, false, finalError
				}
				outerHandler := currentErrorHandler
				currentErrorHandler = nil
				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(handlerSteps, true, rtErr)
				currentErrorHandler = outerHandler

				if handlerErr != nil {
					i.Logger().Warn("[WARN INTERP] Error occurred inside on_error handler: %v. Propagating this new error.", handlerErr)
					return nil, false, false, handlerErr
				}
				if handlerReturned {
					finalError = NewRuntimeError(ErrorCodeInternal, "internal error: 'return' propagated incorrectly from handler", ErrInternal)
					i.Logger().Error("[ERROR INTERP] %v", finalError)
					return nil, false, false, finalError
				}
				if handlerCleared {
					i.Logger().Info("[DEBUG-INTERP]   Handler executed and cleared the error: %v", rtErr)
					stepErr = nil
					wasCleared = true
				} else {
					i.Logger().Info("[DEBUG-INTERP]   Handler executed but did not clear error. Propagating original error: %v", rtErr)
					finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, rtErr)
					return nil, false, false, finalError
				}
			} else {
				i.Logger().Debug("[DEBUG-INTERP]   Error occurred: %v. No active handler in scope. Propagating.", rtErr)
				finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, rtErr)
				return nil, false, false, finalError
			}
		} // --- End Error Handling Check ---

		// If the step executed successfully (stepErr is nil now)
		if stepErr == nil {
			// Update Interpreter's lastCallResult field based on step type
			switch strings.ToLower(step.Type) {
			case "set", "emit", "must", "mustbe", "if", "while", "for", "ask": // Added "ask" here
				i.lastCallResult = stepResult
				i.Logger().Debug("[DEBUG-INTERP]     Step %d successful. Last result updated: %v (%T)", stepNum+1, i.lastCallResult, i.lastCallResult)
			default:
				i.Logger().Debug("[DEBUG-INTERP]     Step %d successful. (Type %s does not update LAST)", stepNum+1, strings.ToUpper(step.Type))
			}
			// NOTE: We no longer update a separate 'finalResult' here
		}

		if wasCleared {
			i.Logger().Info("[DEBUG-INTERP]     Error was cleared in this step or sub-block. Continuing execution.")
		}

	} // End of steps loop

	i.Logger().Info("[DEBUG-INTERP] Finished executing steps block (%s mode) normally.", mode)
	// Corrected: Return nil result for implicit return.
	return nil, false, wasCleared, nil
}

// executeBlock executes a block of steps, passing context flags.
// (Unchanged)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil {
			if i.Logger() != nil {
				i.Logger().Debug("[DEBUG-INTERP] >> Entering empty block execution for %s (parent step %d)", blockType, parentStepNum+1)
			}
			return nil, false, false, nil
		}
		err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("step %d (%s): invalid block format - expected []Step, got %T", parentStepNum+1, blockType, blockValue), ErrInternal)
		if i.Logger() != nil {
			i.Logger().Error("[ERROR] %v", err)
		}
		return nil, false, false, err
	}

	handlerModeStr := ""
	errorStr := "nil"
	if isInHandler {
		handlerModeStr = " (handler mode)"
		if activeError != nil {
			errorStr = fmt.Sprintf("%d", activeError.Code)
		}
	}
	if i.Logger() != nil {
		i.Logger().Info("[DEBUG-INTERP] >> Entering block execution for %s%s (parent step %d, %d steps, activeError: %s)", blockType, handlerModeStr, parentStepNum+1, len(steps), errorStr)
	}

	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	if i.Logger() != nil {
		i.Logger().Info("[DEBUG-INTERP] << Exiting block execution for %s (parent step %d), result: %v, wasReturn: %v, wasCleared: %v, err: %v", blockType, parentStepNum+1, result, wasReturn, wasCleared, err)
	}
	return result, wasReturn, wasCleared, err
}

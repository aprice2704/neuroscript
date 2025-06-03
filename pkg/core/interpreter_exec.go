// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Modified logging in executeSteps to use getStepSubjectForLogging instead of step.Target.
// filename: pkg/core/interpreter_exec.go
// nlines: 265 // Approximate
// risk_rating: MEDIUM
package core

import (
	"errors"
	"fmt"
	"strings"
)

// getStepSubjectForLogging creates a descriptive string for a step's main subject for logging.
func getStepSubjectForLogging(step Step) string {
	switch strings.ToLower(step.Type) {
	case "set":
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
	case "for_each", "for": // Assuming "for" might be an alias
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
		// Add other cases for if, while, etc. if a "subject" is relevant for their top-level log
		// For now, they mostly log their conditions internally or are block structures.
	}
	return "<no specific subject>" // Default for steps like on_error, clear_error, break, continue
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
		// MODIFIED: Use helper function for step subject logging
		stepSubjectStr := getStepSubjectForLogging(step)
		logPos := "<unknown_pos>"
		if step.Pos != nil {
			logPos = step.Pos.String()
		}
		i.Logger().Debug("[DEBUG-INTERP]   Executing Step", "step_num", stepNum+1, "type", stepTypeStr, "subject", stepSubjectStr, "pos", logPos)

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
				errMsg := fmt.Sprintf("step %d: 'call' step type without Call details", stepNum+1)
				stepErr = NewRuntimeError(ErrorCodeInternal, errMsg, errors.New(errMsg)).WithPosition(step.Pos) // Use a base error for Wrapped
			}
		case "return":
			if isInHandler {
				errMsg := fmt.Sprintf("step %d: 'return' statement is not permitted inside an on_error block", stepNum+1)
				stepErr = NewRuntimeError(ErrorCodeReturnViolation, errMsg, ErrReturnViolation).WithPosition(step.Pos)
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
				return nil, false, wasCleared, stepErr // Propagate wasCleared if break/continue happened after a clear
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
			whileBlockResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			// Do NOT propagate ErrBreak from executeWhile if it completed normally (break was handled internally)
			// Only propagate if it's an actual error OR if the loop was exited by break and executeWhile returns ErrBreak
			if errors.Is(stepErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP] WHILE loop broken", "step_num", stepNum+1)
				stepErr = nil // Consume ErrBreak if loop structure handles it
			} else if errors.Is(stepErr, ErrContinue) {
				// Should not happen, continue is handled within the loop. If it propagates, it's an issue.
				i.Logger().Warn("[DEBUG-INTERP] CONTINUE propagated out of WHILE loop unexpectedly", "step_num", stepNum+1)
			} else if stepErr == nil { // Normal completion or successful break
				stepResult = whileBlockResult
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if whileCleared {
					wasCleared = true
				}
			}
		case "for", "for_each": // Allow "for" as an alias
			var forReturned, forCleared bool
			var forBlockResult interface{}
			forBlockResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			if errors.Is(stepErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP] FOR loop broken", "step_num", stepNum+1)
				stepErr = nil // Consume ErrBreak
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
				activeError = nil // Explicitly clear the active error for subsequent steps in this block
				i.Logger().Debug("Active error cleared by CLEAR_ERROR step", "step_num", stepNum+1)
			}
			stepResult = nil
		case "ask":
			stepResult, stepErr = i.executeAsk(step, stepNum, isInHandler, activeError)
		case "break":
			stepResult, stepErr = i.executeBreak(step, stepNum, isInHandler, activeError)
		case "continue":
			stepResult, stepErr = i.executeContinue(step, stepNum, isInHandler, activeError)
		default:
			errMsg := fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type)
			stepErr = NewRuntimeError(ErrorCodeUnknownKeyword, errMsg, ErrUnknownKeyword).WithPosition(step.Pos) // Use specific error
			stepResult = nil
		}

		if stepErr != nil {
			// Handle control flow signals (break/continue)
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP] Propagating control flow signal", "signal", stepErr.Error(), "step_num", stepNum+1)
				return nil, false, wasCleared, stepErr // Propagate to be handled by loop structures
			}

			// Convert to RuntimeError if it isn't already
			rtErr, isRuntimeErr := stepErr.(*RuntimeError)
			if !isRuntimeErr {
				i.Logger().Warn("Wrapping non-RuntimeError from step execution", "original_error", stepErr, "step_num", stepNum+1)
				rtErr = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error during step %d (%s)", stepNum+1, stepTypeStr), stepErr).WithPosition(step.Pos)
			} else if rtErr.Position == nil && step.Pos != nil { // Ensure position is set
				rtErr = rtErr.WithPosition(step.Pos)
			}

			if currentErrorHandler != nil {
				i.Logger().Debug("Error occurred, executing active ON_ERROR handler", "original_error", rtErr.Error(), "step_num", stepNum+1)
				handlerSteps := currentErrorHandler.Body

				// Temporarily disable this handler to prevent recursion on error within handler
				tempHandlerForRecursion := currentErrorHandler
				currentErrorHandler = nil

				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(handlerSteps, true, rtErr)

				// Restore handler only if the handler itself didn't error out or return (or clear).
				if handlerErr == nil && !handlerReturned && !handlerCleared {
					currentErrorHandler = tempHandlerForRecursion
				} else {
					i.Logger().Debug("Not restoring error handler", "handler_error", handlerErr, "handler_returned", handlerReturned, "handler_cleared", handlerCleared)
				}

				if handlerErr != nil {
					i.Logger().Warn("Error occurred inside ON_ERROR handler, propagating handler error", "handler_error", handlerErr.Error(), "original_error", rtErr.Error())
					// Ensure handlerErr is RuntimeError with position
					if _, ok := handlerErr.(*RuntimeError); !ok {
						errMsg := fmt.Sprintf("internal error processing on_error handler at %s", tempHandlerForRecursion.Pos.String())
						handlerErr = NewRuntimeError(ErrorCodeInternal, errMsg, handlerErr).WithPosition(tempHandlerForRecursion.Pos)
					}
					return nil, false, false, handlerErr // Propagate handler's error
				}
				if handlerReturned { // A 'return' inside on_error should ideally not occur or be handled very carefully.
					errMsg := "execution flow error: 'return' from 'on_error' handler is not standard behavior and implies termination of procedure"
					finalError = NewRuntimeError(ErrorCodeReturnViolation, errMsg, ErrReturnViolation).WithPosition(tempHandlerForRecursion.Pos)
					i.Logger().Error("Return from on_error handler", "error", finalError.Error())
					// This return effectively ends the current procedure if not caught.
					return nil, true, false, finalError // Signal return from the original executeSteps call
				}
				if handlerCleared {
					i.Logger().Debug("ON_ERROR handler executed and cleared the error", "cleared_error", rtErr.Error())
					stepErr = nil     // Original error is cleared
					wasCleared = true // Signal that an error was cleared
					activeError = nil // Clear the active error for this scope
					// Continue to the next step in the current block
				} else {
					i.Logger().Debug("ON_ERROR handler executed but did not clear error, propagating original error", "original_error", rtErr.Error())
					// Original error rtErr should be propagated
					return nil, false, false, rtErr
				}
			} else { // No error handler active
				i.Logger().Debug("Error occurred, no active ON_ERROR handler, propagating", "error", rtErr.Error(), "step_num", stepNum+1)
				return nil, false, false, rtErr // Propagate the original (or wrapped) RuntimeError
			}
		} // end if stepErr != nil

		// If stepErr was cleared by a handler, continue to next step.
		if stepErr == nil {
			shouldUpdateLast := false
			switch stepTypeLower {
			case "set", "emit", "must", "mustbe", "ask", "call":
				shouldUpdateLast = true
			}

			if shouldUpdateLast {
				i.lastCallResult = stepResult
				i.Logger().Debug("[DEBUG-INTERP]     Step successful. Last result updated", "step_num", stepNum+1, "last_result", i.lastCallResult, "last_result_type", fmt.Sprintf("%T", i.lastCallResult))
			} else {
				i.Logger().Debug("[DEBUG-INTERP]     Step successful", "step_num", stepNum+1, "type", stepTypeStr, "info", "does not update LAST directly")
			}
		}
	} // end for loop over steps

	i.Logger().Debug("[DEBUG-INTERP] Finished executing steps block normally", "mode", modeStr, "final_wasCleared", wasCleared)
	return nil, false, wasCleared, nil // Normal completion of the block
}

// executeBlock remains unchanged from your provided file.
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
		// Use existing NewRuntimeError and WithPosition
		newErr := NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
		if len(steps) > 0 && steps[0].Pos != nil { // Try to get a position if possible
			newErr = newErr.WithPosition(steps[0].Pos)
		}
		i.Logger().Error("Invalid block format", "error", newErr.Error())
		return nil, false, false, newErr
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
	i.Logger().Debug(">> Entering block execution",
		"block_type", blockType,
		"handler_mode", isInHandler,
		"parent_step", parentStepNum+1,
		"step_count", len(steps),
		"active_error_code", activeErrorStr)

	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	logFields := []interface{}{
		"block_type", blockType,
		"parent_step", parentStepNum + 1,
		"result_from_block_return", result,
		"was_return", wasReturn,
		"was_cleared", wasCleared,
		"lastCallResult_after_block", fmt.Sprintf("%v (%T)", i.lastCallResult, i.lastCallResult),
	}
	if err != nil {
		logFields = append(logFields, "error", err.Error())
	}
	i.Logger().Debug("<< Exiting block execution", logFields...)

	return result, wasReturn, wasCleared, err
}

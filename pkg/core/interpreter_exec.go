// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 13:15:23 PDT
// filename: pkg/core/interpreter_exec.go
package core

import (
	"errors" // Needed for errors.Is
	"fmt"
	"strings"
)

// executeSteps iterates through and executes steps, handling control flow and errors.
// Takes context: isInHandler bool, activeError *RuntimeError (the error being handled).
// Returns: finalResult (nil on implicit return), wasReturn, wasCleared, finalError (error should be *RuntimeError if script-related, or ErrBreak/ErrContinue).
func (i *Interpreter) executeSteps(steps []Step, isInHandler bool, activeError *RuntimeError) (finalResult interface{}, wasReturn bool, wasCleared bool, finalError error) {
	modeStr := "normal"
	activeErrorStr := "nil"
	if isInHandler {
		modeStr = "handler"
		if activeError != nil {
			// Use fmt.Sprintf here because we are preparing a string *value* for logging, not the log message itself.
			activeErrorStr = fmt.Sprintf("%d: %s", activeError.Code, activeError.Message)
		}
	}
	// Use structured logging - CORRECTED
	i.Logger().Info("[DEBUG-INTERP] Executing steps", "count", len(steps), "mode", modeStr, "activeError", activeErrorStr)

	var currentErrorHandler *Step = nil

	for stepNum, step := range steps { // 'step' here is a copy of the struct from the slice
		stepResult := interface{}(nil)
		stepErr := error(nil)

		// Use structured logging - Assuming Step struct has Type and Target fields
		stepTypeStr := strings.ToUpper(step.Type)
		stepTargetStr := step.Target
		i.Logger().Info("[DEBUG-INTERP]   Executing Step", "step_num", stepNum+1, "type", stepTypeStr, "target", stepTargetStr)

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
					return returnValue, true, false, nil // Early exit on return
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step, stepNum, isInHandler, activeError)
		case "if":
			var ifReturned, ifCleared bool
			var ifResult interface{}
			ifResult, ifReturned, ifCleared, stepErr = i.executeIf(step, stepNum, isInHandler, activeError)
			// Check for break/continue propagating from within the if block
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP] Propagating control flow signal from IF block", "signal", stepErr)
				return nil, false, wasCleared, stepErr // Propagate signal immediately
			}
			if stepErr == nil {
				stepResult = ifResult // Propagate result from block
				if ifReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil // Early exit on return
				}
				if ifCleared {
					wasCleared = true // Propagate clear signal
				}
			}
		case "while":
			var whileReturned, whileCleared bool
			var whileResult interface{}
			whileResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			// Check for break/continue propagating from within the while block
			// NOTE: executeWhile already handles internal break/continue, so this check *shouldn't* be hit
			// unless executeWhile changes its return behavior. Keeping it defensively.
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] Propagating unexpected control flow signal from WHILE block", "signal", stepErr)
				return nil, false, wasCleared, stepErr // Propagate signal immediately
			}
			if stepErr == nil {
				stepResult = whileResult // Propagate result from block
				if whileReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil // Early exit on return
				}
				if whileCleared {
					wasCleared = true // Propagate clear signal
				}
			}
		case "for":
			var forReturned, forCleared bool
			var forResult interface{}
			forResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			// Check for break/continue propagating from within the for block
			// NOTE: executeFor already handles internal break/continue, so this check *shouldn't* be hit
			// unless executeFor changes its return behavior. Keeping it defensively.
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Warn("[DEBUG-INTERP] Propagating unexpected control flow signal from FOR block", "signal", stepErr)
				return nil, false, wasCleared, stepErr // Propagate signal immediately
			}
			if stepErr == nil {
				stepResult = forResult // Propagate result from block
				if forReturned {
					i.lastCallResult = stepResult
					return stepResult, true, false, nil // Early exit on return
				}
				if forCleared {
					wasCleared = true // Propagate clear signal
				}
			}
		case "must", "mustbe":
			stepResult, stepErr = i.executeMust(step, stepNum, isInHandler, activeError)
		case "fail":
			stepErr = i.executeFail(step, stepNum, isInHandler, activeError)
			stepResult = nil // Fail doesn't produce a result
		case "on_error":
			var handlerStep *Step
			handlerStep, stepErr = i.executeOnError(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				currentErrorHandler = handlerStep // Activate the handler
			}
			stepResult = nil // OnError itself doesn't produce a result
		case "clear_error":
			var clearedNow bool
			clearedNow, stepErr = i.executeClearError(step, stepNum, isInHandler, activeError)
			if stepErr == nil && clearedNow {
				wasCleared = true // Signal that error was cleared
			}
			stepResult = nil // ClearError doesn't produce a result
		case "ask":
			stepResult, stepErr = i.executeAsk(step, stepNum, isInHandler, activeError) // Call the ask function

		// --- ADDED Cases for Break/Continue ---
		case "break":
			stepResult, stepErr = i.executeBreak(step, stepNum, isInHandler, activeError)
			// stepErr should be ErrBreak if successful
		case "continue":
			stepResult, stepErr = i.executeContinue(step, stepNum, isInHandler, activeError)
			// stepErr should be ErrContinue if successful
		// --- End ADDED Cases ---

		// Add cases for other step types (call, etc.) if they exist
		default:
			stepErr = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type), ErrUnknownKeyword)
			stepResult = nil
		}
		// --- End Execute Step ---

		// --- Error Handling Check ---
		if stepErr != nil {
			// --- ADDED: Check for Break/Continue first ---
			// These signals should bypass the on_error handler and propagate up to the loop constructs
			if errors.Is(stepErr, ErrBreak) || errors.Is(stepErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP] Propagating control flow signal", "signal", stepErr, "step_num", stepNum+1)
				return nil, false, wasCleared, stepErr // Propagate break/continue immediately
			}
			// --- End ADDED Check ---

			// Existing error handling for other errors (RuntimeError wrapping, on_error handler call)
			rtErr, isRuntimeErr := stepErr.(*RuntimeError)
			if !isRuntimeErr {
				// Wrap non-RuntimeError errors encountered during step execution
				i.Logger().Warn("Wrapping non-RuntimeError from step execution", "original_error", stepErr, "step_num", stepNum+1)
				rtErr = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error during step %d (%s)", stepNum+1, stepTypeStr), stepErr)
			}

			if currentErrorHandler != nil {
				// Use structured logging
				i.Logger().Info("Error occurred, executing active ON_ERROR handler", "original_error", rtErr, "step_num", stepNum+1)

				handlerSteps, ok := currentErrorHandler.Value.([]Step) // Assuming Value holds the handler steps
				if !ok {
					errMsg := fmt.Sprintf("internal error: on_error step value is not []Step (%T)", currentErrorHandler.Value)
					finalError = NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
					return nil, false, false, finalError // Fatal internal error
				}

				// Temporarily disable the current handler while executing it
				outerHandler := currentErrorHandler
				currentErrorHandler = nil
				// Execute the handler steps, passing the error that occurred
				// Note: Pass the Runtime Error 'rtErr' as the activeError context
				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(handlerSteps, true, rtErr) // Pass true for isInHandler

				// Restore the outer handler *unless* the handler itself errored or returned
				if handlerErr == nil && !handlerReturned {
					currentErrorHandler = outerHandler
				} else {
					// If handler errored or returned, we don't restore the handler
					// as execution is stopping or propagating a different error/return.
					i.Logger().Debug("Not restoring error handler due to handler error or return", "handler_error", handlerErr, "handler_returned", handlerReturned)
				}

				if handlerErr != nil {
					// Error inside the handler itself - this supersedes the original error
					// Use structured logging
					i.Logger().Warn("Error occurred inside ON_ERROR handler, propagating handler error", "handler_error", handlerErr, "original_error", rtErr)
					// Ensure handler error is RuntimeError
					if _, ok := handlerErr.(*RuntimeError); !ok {
						errMsg := fmt.Sprintf("internal error processing on_error handler at %s", currentErrorHandler.Pos.String())
						handlerErr = NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, handlerErr))
					}
					return nil, false, false, handlerErr // Propagate the handler's error
				}
				if handlerReturned {
					// This should be prevented by executeReturn check, but handle defensively
					errMsg := "internal error: 'return' propagated incorrectly from on_error handler"
					finalError = NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
					i.Logger().Error("Internal error: Return from handler", "error", finalError)
					return nil, false, false, finalError
				}
				if handlerCleared {
					// Handler executed and cleared the error
					// Use structured logging
					i.Logger().Info("ON_ERROR handler executed and cleared the error", "cleared_error", rtErr)
					stepErr = nil     // Clear the error for the current step loop iteration
					wasCleared = true // Signal that an error was cleared within this block
					// Continue to the next step in the original block
				} else {
					// Handler executed but did NOT clear the error
					// Use structured logging
					i.Logger().Info("ON_ERROR handler executed but did not clear error, propagating original error", "original_error", rtErr)
					// Ensure we propagate the original error as a RuntimeError
					finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, stepTypeStr, rtErr) // Keep original error context
					return nil, false, false, finalError                                       // Stop execution and return the original error
				}
			} else {
				// Error occurred, and no handler is active
				// Use structured logging
				i.Logger().Debug("Error occurred, no active handler, propagating", "error", rtErr, "step_num", stepNum+1)
				finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, stepTypeStr, rtErr) // Keep original error context
				return nil, false, false, finalError                                       // Stop execution and return the error
			}
		} // --- End Error Handling Check ---

		// --- Update Last Result ---
		// Only update if the step executed successfully (stepErr is nil)
		if stepErr == nil {
			// Update Interpreter's lastCallResult field based on step type
			// Determine which steps should update LAST
			shouldUpdateLast := false
			switch strings.ToLower(step.Type) {
			// List the step types that produce a meaningful result for LAST
			case "set", "emit", "must", "mustbe", "if", "while", "for", "ask", "call": // Added 'call' assuming it exists
				shouldUpdateLast = true
				// break and continue do not produce a value for LAST
			}

			if shouldUpdateLast {
				i.lastCallResult = stepResult
				// Use structured logging - CORRECTED
				i.Logger().Debug("[DEBUG-INTERP]     Step successful. Last result updated", "step_num", stepNum+1, "last_result", i.lastCallResult, "last_result_type", fmt.Sprintf("%T", i.lastCallResult))
			} else {
				// Use structured logging - CORRECTED
				i.Logger().Debug("[DEBUG-INTERP]     Step successful", "step_num", stepNum+1, "type", stepTypeStr, "info", "does not update LAST")
			}
		}
		// --- End Update Last Result ---

	} // End of steps loop

	// Use structured logging - CORRECTED
	i.Logger().Info("[DEBUG-INTERP] Finished executing steps block normally", "mode", modeStr)
	// If loop completes without error or early return
	return nil, false, wasCleared, nil // Implicit return (nil result), not a return statement, potentially cleared, no error
}

// executeBlock executes a block of steps, passing context flags.
// (Assumes this function exists and calls executeSteps internally)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		// Handle case where blockValue is nil or not []Step
		errMsg := fmt.Sprintf("step %d (%s): invalid block format - expected []Step", parentStepNum+1, blockType)
		if blockValue != nil {
			errMsg = fmt.Sprintf("%s, got %T", errMsg, blockValue)
		} else {
			errMsg = fmt.Sprintf("%s, got nil", errMsg)
			// If nil is acceptable (e.g., empty block), handle it gracefully
			i.Logger().Debug("Entering empty block execution", "block_type", blockType, "parent_step", parentStepNum+1)
			return nil, false, false, nil
		}
		err = NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternal)
		i.Logger().Error("Invalid block format", "error", err)
		return nil, false, false, err
	}

	// Prepare logging context
	activeErrorStr := "nil"
	if isInHandler {
		if activeError != nil {
			activeErrorStr = fmt.Sprintf("%d", activeError.Code)
		}
	}
	// Use structured logging
	i.Logger().Info(">> Entering block execution",
		"block_type", blockType,
		"handler_mode", isInHandler, // Keep boolean flag
		"parent_step", parentStepNum+1,
		"step_count", len(steps),
		"active_error_code", activeErrorStr)

	// Execute the steps within the block
	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	// Use structured logging
	i.Logger().Info("<< Exiting block execution",
		"block_type", blockType,
		"parent_step", parentStepNum+1,
		"result", result, // Log the actual result
		"was_return", wasReturn,
		"was_cleared", wasCleared,
		"error", err) // Log the error if any

	return result, wasReturn, wasCleared, err
}

// --- Stubs for other execute functions referenced above ---
// These need to be defined elsewhere, matching the signatures used in the switch statement.
// func (i *Interpreter) executeIf(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) { panic("executeIf not implemented") }
// func (i *Interpreter) executeWhile(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) { panic("executeWhile not implemented") }
// func (i *Interpreter) executeFor(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) { panic("executeFor not implemented") }
// Ensure the functions called in the switch statement (executeSet, executeReturn, etc.) are defined correctly, likely in other interpreter_steps_*.go files.

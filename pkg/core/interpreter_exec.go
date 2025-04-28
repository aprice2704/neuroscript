// filename: pkg/core/interpreter_exec.go
package core

import (
	// Import errors for Is/As if needed later
	"fmt"
	"strings"
)

// executeSteps iterates through and executes steps, handling control flow and errors.
// Takes context: isInHandler bool, activeError *RuntimeError (the error being handled).
// Returns: finalResult, wasReturn, wasCleared, finalError (error should be *RuntimeError if script-related).
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

	var currentErrorHandler *Step = nil // Tracks the active on_error for *this execution scope*
	// Use the Interpreter's lastCallResult field (as per interpreter_new.go)
	// Initialize finalResult to nil; it will be updated by the last successful step if no return occurs.
	finalResult = nil

	for stepNum, step := range steps {
		stepResult := interface{}(nil) // Result of the current step, default nil
		stepErr := error(nil)          // Error from the current step

		i.Logger().Info("[DEBUG-INTERP]   Step %d: Type=%s, Target=%s", stepNum+1, strings.ToUpper(step.Type), step.Target)

		// --- Execute Step ---
		switch strings.ToLower(step.Type) {
		case "set":
			// FIX: Capture 2 return values
			stepResult, stepErr = i.executeSet(step, stepNum, isInHandler, activeError)
		case "call":
			stepResult, stepErr = i.executeCall(step, stepNum, isInHandler, activeError)
			// CALL contributes to the final block result if it's the last step, potentially
			// We handle updating i.lastCallResult below if stepErr is nil
		case "return":
			if isInHandler {
				// FIX: Wrap ErrReturnViolation (now defined in errors.go)
				stepErr = NewRuntimeError(ErrorCodeReturnViolation, fmt.Sprintf("step %d: 'return' statement is not permitted inside an on_error block", stepNum+1), ErrReturnViolation)
			} else {
				// executeReturn returns the value(s), true flag, and error
				finalResult, wasReturn, stepErr = i.executeReturn(step, stepNum, isInHandler, activeError)
				if stepErr == nil && wasReturn {
					// FIX: Update lastCallResult with the returned value(s) before exiting
					i.lastCallResult = finalResult
					return finalResult, true, false, nil // Immediate return on success
				}
				// If executeReturn had an error, stepErr will be set, handled below.
			}
		case "emit":
			// FIX: Capture 2 return values
			stepResult, stepErr = i.executeEmit(step, stepNum, isInHandler, activeError)
		case "if":
			var ifReturned, ifCleared bool
			// executeIf handles block execution and returns final result of the block
			stepResult, ifReturned, ifCleared, stepErr = i.executeIf(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				if ifReturned {
					// Propagate return immediately, update lastCallResult first
					// FIX: Use i.lastCallResult
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if ifCleared {
					wasCleared = true // Signal clear upwards
				}
				// Result of IF block becomes the step result if no error/return
			}
		case "while":
			var whileReturned, whileCleared bool
			// executeWhile returns final result of the *last successful iteration*
			stepResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				if whileReturned {
					// FIX: Use i.lastCallResult
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if whileCleared {
					wasCleared = true
				}
				// Result of WHILE becomes step result
			}
		case "for":
			var forReturned, forCleared bool
			// executeFor returns final result of the *last successful iteration*
			stepResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				if forReturned {
					// FIX: Use i.lastCallResult
					i.lastCallResult = stepResult
					return stepResult, true, false, nil
				}
				if forCleared {
					wasCleared = true
				}
				// Result of FOR becomes step result
			}
		case "must", "mustbe":
			// FIX: Capture 2 return values
			stepResult, stepErr = i.executeMust(step, stepNum, isInHandler, activeError)
		case "fail":
			stepErr = i.executeFail(step, stepNum, isInHandler, activeError) // Should return *RuntimeError
			// FAIL intentionally does not produce a stepResult
			stepResult = nil
		case "on_error":
			var handlerStep *Step
			handlerStep, stepErr = i.executeOnError(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
				currentErrorHandler = handlerStep // Activate for subsequent steps
			}
			// ON_ERROR does not produce a stepResult
			stepResult = nil
		case "clear_error":
			var clearedNow bool
			clearedNow, stepErr = i.executeClearError(step, stepNum, isInHandler, activeError)
			if stepErr == nil && clearedNow {
				wasCleared = true // Signal clear upwards
			}
			// CLEAR_ERROR does not produce a stepResult
			stepResult = nil
		default:
			stepErr = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("step %d: unknown step type '%s'", stepNum+1, step.Type), ErrUnknownKeyword)
			stepResult = nil
		}
		// --- End Execute Step ---

		// --- Error Handling Check ---
		if stepErr != nil {
			// Ensure stepErr is a *RuntimeError for consistent handling
			rtErr, isRuntimeErr := stepErr.(*RuntimeError)
			if !isRuntimeErr {
				// Wrap non-RuntimeError (e.g., internal Go errors from tools/evaluation)
				i.Logger().Warn("[WARN INTERP] Wrapping non-RuntimeError: %v", stepErr)
				rtErr = NewRuntimeError(ErrorCodeInternal, stepErr.Error(), stepErr)
			}

			// An errored step doesn't produce a result for LAST. Keep i.lastCallResult as it was.

			if currentErrorHandler != nil {
				// Active handler exists!
				i.Logger().Info("[DEBUG-INTERP]   Error occurred: %v. Handler is active. Executing handler.", rtErr)

				handlerSteps, ok := currentErrorHandler.Value.([]Step)
				if !ok {
					// This is an internal error - the AST wasn't built correctly
					finalError = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error: on_error step value is not []Step (%T)", currentErrorHandler.Value), ErrInternal)
					return nil, false, false, finalError
				}

				// Execute handler steps recursively, passing the *current* error (rtErr)
				outerHandler := currentErrorHandler // Save current handler
				currentErrorHandler = nil           // Deactivate handler for the recursive call initially

				// --- Scope/Variable Handling for Handler ---
				// Need to update Interpreter's variable stack or pass context for err_code/err_msg
				// For now, assume implicit variables are handled by GetVariable if context is right
				_, handlerReturned, handlerCleared, handlerErr := i.executeSteps(handlerSteps, true, rtErr)
				// Restore scope if saved
				// --- End Scope Handling ---

				currentErrorHandler = outerHandler // Restore outer handler

				// --- Post Handler ---
				if handlerErr != nil { // Error *inside* handler replaces original
					i.Logger().Warn("[WARN INTERP] Error occurred inside on_error handler: %v. Propagating this new error.", handlerErr)
					// Wrap with outer step info? Maybe not, handler error is distinct.
					return nil, false, false, handlerErr // Propagate handler's error
				}
				if handlerReturned { // Should not happen per spec
					finalError = NewRuntimeError(ErrorCodeInternal, "internal error: 'return' propagated incorrectly from handler", ErrInternal)
					i.Logger().Error("[ERROR INTERP] %v", finalError)
					return nil, false, false, finalError
				}
				if handlerCleared { // Handler called clear_error
					i.Logger().Info("[DEBUG-INTERP]   Handler executed and cleared the error: %v", rtErr)
					stepErr = nil     // Nullify the error, allowing the loop to continue
					wasCleared = true // Signal that a clear happened within this block
					// Clear the active error context *if* this was the top-level handler?
					// If called recursively, just signal upwards.
					// Continue to the next step in the *current* block.
				} else { // Handler finished, but didn't clear
					i.Logger().Info("[DEBUG-INTERP]   Handler executed but did not clear error. Propagating original error: %v", rtErr)
					// Wrap error with step info before returning
					finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, rtErr)
					return nil, false, false, finalError // Propagate the original error
				}
			} else {
				// No active handler, propagate error immediately
				i.Logger().Debug("[DEBUG-INTERP]   Error occurred: %v. No active handler in scope. Propagating.", rtErr)
				// Wrap error with step info before returning
				finalError = fmt.Errorf("step %d (%s): %w", stepNum+1, step.Type, rtErr)
				return nil, false, false, finalError
			}
		} // --- End Error Handling Check ---

		// If the step executed successfully (stepErr is nil now, possibly cleared by handler)
		if stepErr == nil {
			// FIX: Update Interpreter's lastCallResult field (not lastStepResult)
			i.lastCallResult = stepResult
			i.Logger().Debug("[DEBUG-INTERP]     Step %d successful. Last result set to: %v (%T)", stepNum+1, i.lastCallResult, i.lastCallResult)
			// Update finalResult for the block - the result of the *last* successful step becomes the block result
			finalResult = i.lastCallResult
		}

		// If error was cleared inside the handler, stepErr is now nil, and loop continues.
		// If clear_error was called directly, need to ensure loop continues.
		if wasCleared && strings.ToLower(step.Type) == "clear_error" {
			i.Logger().Info("[DEBUG-INTERP]     Continuing loop after successful clear_error step.")
			// Reset wasCleared for next steps? Yes, it only applies to the step that was cleared.
			// But the flag needs to propagate *up* if the block finishes.
			// Let's keep wasCleared=true if set by handler/clear_error, but only continue here if it was a direct clear_error step.
			continue
		}

	} // End of steps loop

	i.Logger().Info("[DEBUG-INTERP] Finished executing steps block (%s mode) normally.", mode)
	// Return the result of the *last successfully executed step* if no explicit return occurred.
	// finalResult was updated in the loop.
	return finalResult, false, wasCleared, nil // Normal finish
}

// executeBlock executes a block of steps, passing context flags.
// It now returns the final result of the block (often the result of the last step).
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]Step)
	if !ok {
		if blockValue == nil { // Handle nil block (e.g., empty else)
			if i.Logger() != nil {
				i.Logger().Debug("[DEBUG-INTERP] >> Entering empty block execution for %s (parent step %d)", blockType, parentStepNum+1)
			}
			return nil, false, false, nil // Empty block returns nil result
		}
		// Return a RuntimeError for internal errors too
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

	// --- Scope Management (Start) ---
	// Implement scope pushing/popping if needed for block-level variables
	// i.pushScope() // Example placeholder
	// --- Scope Management (Start) ---

	// Recursive call passing down context
	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	// --- Scope Management (End) ---
	// i.popScope() // Example placeholder
	// --- Scope Management (End) ---

	if i.Logger() != nil {
		i.Logger().Info("[DEBUG-INTERP] << Exiting block execution for %s (parent step %d), result: %v, wasReturn: %v, wasCleared: %v, err: %v", blockType, parentStepNum+1, result, wasReturn, wasCleared, err)
	}
	// Propagate all results upwards (result now holds the value of the last step if no return/error)
	return result, wasReturn, wasCleared, err
}

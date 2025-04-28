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
	finalResult = nil // Initialize finalResult

	for stepNum, step := range steps {
		stepResult := interface{}(nil) // Result of the current step, default nil
		stepErr := error(nil)          // Error from the current step

		i.Logger().Info("[DEBUG-INTERP]   Step %d: Type=%s, Target=%s", stepNum+1, strings.ToUpper(step.Type), step.Target)

		// --- Execute Step ---
		switch strings.ToLower(step.Type) {
		case "set":
			stepResult, stepErr = i.executeSet(step, stepNum, isInHandler, activeError)
		case "call":
			// --- FIX: Dispatch ask... calls here ---
			switch strings.ToLower(step.Target) {
			case "askai":
				// Need evaluated args for ask functions
				evaluatedArgs, evalErr := i.evaluateCallArgs(step.Args)
				if evalErr != nil {
					stepErr = evalErr // Assign evaluation error
				} else {
					stepResult, stepErr = i.executeAskAI(step, stepNum, evaluatedArgs)
				}
			case "askhuman":
				evaluatedArgs, evalErr := i.evaluateCallArgs(step.Args)
				if evalErr != nil {
					stepErr = evalErr
				} else {
					stepResult, stepErr = i.executeAskHuman(step, stepNum, evaluatedArgs)
				}
			case "askcomputer":
				evaluatedArgs, evalErr := i.evaluateCallArgs(step.Args)
				if evalErr != nil {
					stepErr = evalErr
				} else {
					stepResult, stepErr = i.executeAskComputer(step, stepNum, evaluatedArgs)
				}
			default:
				// Original procedure/tool call logic
				stepResult, stepErr = i.executeCall(step, stepNum, isInHandler, activeError)
			}
			// --- End FIX ---
		case "return":
			if isInHandler {
				stepErr = NewRuntimeError(ErrorCodeReturnViolation, fmt.Sprintf("step %d: 'return' statement is not permitted inside an on_error block", stepNum+1), ErrReturnViolation)
			} else {
				finalResult, wasReturn, stepErr = i.executeReturn(step, stepNum, isInHandler, activeError)
				if stepErr == nil && wasReturn {
					i.lastCallResult = finalResult
					return finalResult, true, false, nil
				}
			}
		case "emit":
			stepResult, stepErr = i.executeEmit(step, stepNum, isInHandler, activeError)
		case "if":
			var ifReturned, ifCleared bool
			stepResult, ifReturned, ifCleared, stepErr = i.executeIf(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
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
			stepResult, whileReturned, whileCleared, stepErr = i.executeWhile(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
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
			stepResult, forReturned, forCleared, stepErr = i.executeFor(step, stepNum, isInHandler, activeError)
			if stepErr == nil {
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
					stepErr = nil // Nullify the error
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
			// Update Interpreter's lastCallResult field
			i.lastCallResult = stepResult
			i.Logger().Debug("[DEBUG-INTERP]     Step %d successful. Last result set to: %v (%T)", stepNum+1, i.lastCallResult, i.lastCallResult)
			// Update finalResult for the block - result of the *last* successful step
			finalResult = i.lastCallResult
		}

		if wasCleared && strings.ToLower(step.Type) == "clear_error" && stepErr == nil {
			i.Logger().Info("[DEBUG-INTERP]     Continuing loop after successful clear_error step.")
			// Reset wasCleared? No, let it propagate up if block finishes.
			continue
		}

	} // End of steps loop

	i.Logger().Info("[DEBUG-INTERP] Finished executing steps block (%s mode) normally.", mode)
	// Return the result of the *last successfully executed step* if no explicit return occurred.
	return finalResult, false, wasCleared, nil // Normal finish
}

// evaluateCallArgs helper to evaluate arguments for ask... calls
func (i *Interpreter) evaluateCallArgs(args []interface{}) ([]interface{}, error) {
	evaluatedArgs := make([]interface{}, len(args))
	for idx, arg := range args {
		evaluatedArg, err := i.evaluateExpression(arg)
		if err != nil {
			// Wrap error with arg index info
			return nil, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating arg %d for call", idx+1), fmt.Errorf("evaluating arg %d: %w", idx+1, err))
		}
		evaluatedArgs[idx] = evaluatedArg
	}
	return evaluatedArgs, nil
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

	// --- Scope Management (Placeholder) ---
	// Implement scope push/pop if necessary for variable isolation
	// i.pushScope()
	// defer i.popScope()
	// --- End Scope Management ---

	result, wasReturn, wasCleared, err = i.executeSteps(steps, isInHandler, activeError)

	if i.Logger() != nil {
		i.Logger().Info("[DEBUG-INTERP] << Exiting block execution for %s (parent step %d), result: %v, wasReturn: %v, wasCleared: %v, err: %v", blockType, parentStepNum+1, result, wasReturn, wasCleared, err)
	}
	return result, wasReturn, wasCleared, err
}

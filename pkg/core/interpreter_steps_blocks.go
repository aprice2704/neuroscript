// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 13:10:35 PDT
// filename: pkg/core/interpreter_steps_blocks.go
package core

import (
	"errors" // Needed for errors.Is
	"fmt"
	"reflect"
	// Assuming NsError and error constants are defined in "errors.go"
)

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing IF")
	// CORRECTED: Call evaluateExpression
	condResult, evalErr := i.evaluateExpression(step.Cond) // Pass context flags if evaluateExpression needs them
	if evalErr != nil {
		// Ensure error is RuntimeError
		if _, ok := evalErr.(*RuntimeError); !ok {
			errMsg := fmt.Sprintf("evaluating IF condition at %s", step.Pos.String())
			evalErr = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, evalErr))
		}
		return nil, false, false, evalErr // Return evaluation error
	}

	// isTruthy needs to be defined elsewhere
	if isTruthy(condResult) {
		i.Logger().Debug("[DEBUG-INTERP]   IF condition TRUE, executing THEN block")
		result, wasReturn, wasCleared, err = i.executeBlock(step.Value, stepNum, "IF-THEN", isInHandler, activeError)
	} else {
		i.Logger().Debug("[DEBUG-INTERP]   IF condition FALSE, executing ELSE block (if exists)")
		result, wasReturn, wasCleared, err = i.executeBlock(step.ElseValue, stepNum, "IF-ELSE", isInHandler, activeError)
	}
	// Propagate break/continue signals if they occurred within the chosen block
	if errors.Is(err, ErrBreak) || errors.Is(err, ErrContinue) {
		return nil, false, false, err
	}
	// Handle other potential errors from the block execution
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			errMsg := fmt.Sprintf("executing IF block at %s", step.Pos.String())
			err = NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, err))
		}
		return nil, false, false, err
	}

	return result, wasReturn, wasCleared, nil // Return normal result/state
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := step.Pos.String() // Get position once
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE", "pos", posStr)
	iteration := 0
	maxIterations := 10000 // Safety break

	// Initialize result to nil, only update if loop body runs successfully
	result = nil

	for iteration < maxIterations {
		iteration++
		// Evaluate condition
		condResult, evalErr := i.evaluateExpression(step.Cond)
		if evalErr != nil {
			errMsg := fmt.Sprintf("iteration %d: evaluating WHILE condition at %s", iteration, posStr)
			if _, ok := evalErr.(*RuntimeError); !ok {
				evalErr = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, evalErr))
			}
			err = evalErr // Assign to outer err
			break         // Exit loop on condition evaluation error
		}

		// Check condition truthiness
		if !isTruthy(condResult) {
			i.Logger().Debug("[DEBUG-INTERP]   WHILE condition FALSE on iteration %d. Exiting loop.", iteration, "pos", posStr)
			break // Exit the loop normally
		}

		// Execute loop body
		i.Logger().Debug("[DEBUG-INTERP]   WHILE condition TRUE on iteration %d. Executing block.", iteration, "pos", posStr)
		var blockResult interface{}
		var blockReturned, blockCleared bool
		var blockErr error
		blockResult, blockReturned, blockCleared, blockErr = i.executeBlock(step.Value, stepNum, "WHILE-BODY", isInHandler, activeError)

		// --- Check for control flow signals ---
		if errors.Is(blockErr, ErrBreak) {
			i.Logger().Debug("[DEBUG-INTERP]   BREAK received in WHILE loop body on iteration %d. Exiting loop.", iteration, "pos", posStr)
			err = nil // Consume ErrBreak signal
			break     // Exit the Go loop
		}
		if errors.Is(blockErr, ErrContinue) {
			i.Logger().Debug("[DEBUG-INTERP]   CONTINUE received in WHILE loop body on iteration %d. Skipping to next iteration.", iteration, "pos", posStr)
			// Consume ErrContinue signal and continue the Go loop
			result = blockResult // Keep last successful result before continue
			if blockCleared {    // Persist cleared state
				wasCleared = true
			}
			continue // Skip to the next iteration of the Go loop
		}
		// --- End control flow check ---

		// Handle other errors from the block
		if blockErr != nil {
			errMsg := fmt.Sprintf("iteration %d: executing WHILE body at %s", iteration, posStr)
			if _, ok := blockErr.(*RuntimeError); !ok {
				blockErr = NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, blockErr))
			}
			err = blockErr // Assign to outer err
			break          // Exit loop on error
		}

		// Handle return from the block
		if blockReturned {
			return blockResult, true, false, nil // Propagate return immediately
		}

		// Handle clear_error from the block
		if blockCleared {
			wasCleared = true
			i.Logger().Debug("[DEBUG-INTERP]   CLEAR_ERROR detected within WHILE loop body on iteration %d.", iteration, "pos", posStr)
		}

		// Update the loop result with the result of the last successful block execution
		result = blockResult

	} // End loop

	// Handle errors that broke the loop
	if err != nil {
		return nil, false, false, err
	}

	// Check for max iterations exceeded
	if iteration >= maxIterations {
		errMsg := fmt.Sprintf("WHILE loop at %s exceeded max iterations (%d)", posStr, maxIterations)
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, errMsg, nil)
	}

	i.Logger().Debug("[DEBUG-INTERP]   WHILE loop finished normally.", "pos", posStr)
	// Return the result of the last successful iteration (or nil if loop never ran)
	// and the accumulated cleared state.
	return result, false, wasCleared, nil
}

// executeFor handles the "for" (for each) step.
func (i *Interpreter) executeFor(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := step.Pos.String() // Get position once
	targetVar := step.Target
	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH", "Var", targetVar, "pos", posStr)

	// Evaluate collection
	collectionVal, evalErr := i.evaluateExpression(step.Cond)
	if evalErr != nil {
		errMsg := fmt.Sprintf("evaluating collection for FOR EACH %s at %s", targetVar, posStr)
		if _, ok := evalErr.(*RuntimeError); !ok {
			evalErr = NewRuntimeError(ErrorCodeEvaluation, errMsg, fmt.Errorf("%s: %w", errMsg, evalErr))
		}
		return nil, false, false, evalErr // Return evaluation error
	}

	// Reflect on the collection value
	val := reflect.ValueOf(collectionVal)
	iteration := 0
	maxIterations := 10000 // Safety break

	// Initialize result to nil
	result = nil
	shouldBreakOuter := false // Flag to break outer switch after inner loop breaks

	// Anonymous function to execute the loop body for a single item
	executeLoopBody := func(item interface{}) (blockResult interface{}, blockReturned bool, blockCleared bool, blockErr error) {
		iteration++
		if iteration > maxIterations {
			errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", targetVar, posStr, maxIterations)
			blockErr = NewRuntimeError(ErrorCodeInternal, errMsg, nil)
			return
		}

		// Set loop variable
		if setErr := i.SetVariable(targetVar, item); setErr != nil {
			errMsg := fmt.Sprintf("iteration %d: setting loop var '%s' for FOR EACH at %s", iteration, targetVar, posStr)
			blockErr = NewRuntimeError(ErrorCodeInternal, errMsg, setErr)
			return
		}

		// Execute block
		i.Logger().Debug("[DEBUG-INTERP]   FOR EACH iteration %d", "Var", targetVar, "Value", item, "pos", posStr)
		blockResult, blockReturned, blockCleared, blockErr = i.executeBlock(step.Value, stepNum, "FOR-BODY", isInHandler, activeError)
		return // Return results from executeBlock
	}

	// Iterate based on collection type
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < val.Len(); idx++ {
			item := val.Index(idx).Interface()
			blockResult, blockReturned, blockCleared, blockErr := executeLoopBody(item)

			// --- Check for control flow signals ---
			if errors.Is(blockErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]   BREAK received in FOR EACH (Slice/Array) body on iteration %d. Exiting loop.", iteration, "pos", posStr)
				shouldBreakOuter = true // Signal to break the outer switch
				err = nil               // Consume ErrBreak signal
				break                   // Break the inner Go loop (for idx...)
			}
			if errors.Is(blockErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]   CONTINUE received in FOR EACH (Slice/Array) body on iteration %d. Skipping to next item.", iteration, "pos", posStr)
				result = blockResult // Keep last result before continue
				if blockCleared {    // Persist cleared state
					wasCleared = true
				}
				continue // Continue the inner Go loop (for idx...)
			}
			// --- End control flow check ---

			if blockErr != nil {
				err = blockErr // Assign actual error to outer err
				break          // Break the inner Go loop (for idx...)
			}
			if blockReturned {
				return blockResult, true, false, nil // Propagate return immediately
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult // Update result with last successful iteration
		}
	case reflect.Map:
		mapRange := val.MapRange()
		for mapRange.Next() {
			item := mapRange.Value().Interface() // Iterate map values
			blockResult, blockReturned, blockCleared, blockErr := executeLoopBody(item)

			// --- Check for control flow signals ---
			if errors.Is(blockErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]   BREAK received in FOR EACH (Map) body on iteration %d. Exiting loop.", iteration, "pos", posStr)
				shouldBreakOuter = true // Signal to break the outer switch
				err = nil               // Consume ErrBreak signal
				break                   // Break the inner Go loop (for mapRange...)
			}
			if errors.Is(blockErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]   CONTINUE received in FOR EACH (Map) body on iteration %d. Skipping to next item.", iteration, "pos", posStr)
				result = blockResult // Keep last result before continue
				if blockCleared {    // Persist cleared state
					wasCleared = true
				}
				continue // Continue the inner Go loop (for mapRange...)
			}
			// --- End control flow check ---

			if blockErr != nil {
				err = blockErr // Assign actual error to outer err
				break          // Break the inner Go loop (for mapRange...)
			}
			if blockReturned {
				return blockResult, true, false, nil // Propagate return immediately
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult // Update result with last successful iteration
		}
	case reflect.String:
		str := val.String()
		for _, char := range str { // Iterate runes
			item := string(char) // Convert rune to string
			blockResult, blockReturned, blockCleared, blockErr := executeLoopBody(item)

			// --- Check for control flow signals ---
			if errors.Is(blockErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]   BREAK received in FOR EACH (String) body on iteration %d. Exiting loop.", iteration, "pos", posStr)
				shouldBreakOuter = true // Signal to break the outer switch
				err = nil               // Consume ErrBreak signal
				break                   // Break the inner Go loop (for _, char...)
			}
			if errors.Is(blockErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]   CONTINUE received in FOR EACH (String) body on iteration %d. Skipping to next item.", iteration, "pos", posStr)
				result = blockResult // Keep last result before continue
				if blockCleared {    // Persist cleared state
					wasCleared = true
				}
				continue // Continue the inner Go loop (for _, char...)
			}
			// --- End control flow check ---

			if blockErr != nil {
				err = blockErr // Assign actual error to outer err
				break          // Break the inner Go loop (for _, char...)
			}
			if blockReturned {
				return blockResult, true, false, nil // Propagate return immediately
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult // Update result with last successful iteration
		}
	default:
		errMsg := fmt.Sprintf("cannot iterate over type %T for FOR EACH %s at %s", collectionVal, targetVar, posStr)
		err = NewRuntimeError(ErrorCodeType, errMsg, nil)
	}

	// Check if an error occurred that broke an inner loop
	if err != nil {
		return nil, false, false, err
	}
	// Check if break occurred within an inner loop
	if shouldBreakOuter {
		i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop terminated by BREAK.", "pos", posStr)
		// Break doesn't propagate an error; loop finishes.
		// Return the result from the iteration *before* the break.
		return result, false, wasCleared, nil
	}
	// Check for max iterations exceeded (if not already caught by error)
	if iteration >= maxIterations {
		errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", targetVar, posStr, maxIterations)
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, errMsg, nil)
	}

	i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop finished normally.", "pos", posStr)
	// Return the result of the last successful iteration and accumulated cleared state.
	return result, false, wasCleared, nil
}

// --- Placeholder for isTruthy ---
// func isTruthy(value interface{}) bool { ... } // Assume defined elsewhere

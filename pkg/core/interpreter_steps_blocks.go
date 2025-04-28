// filename: pkg/core/interpreter_steps_blocks.go
package core

import (
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
		return nil, false, false, NewRuntimeError(ErrorCodeGeneric, "evaluating IF condition", evalErr)
	}

	// isTruthy needs to be defined elsewhere
	if isTruthy(condResult) {
		i.Logger().Debug("[DEBUG-INTERP]   IF condition TRUE, executing THEN block")
		result, wasReturn, wasCleared, err = i.executeBlock(step.Value, stepNum, "IF-THEN", isInHandler, activeError)
	} else {
		i.Logger().Debug("[DEBUG-INTERP]   IF condition FALSE, executing ELSE block (if exists)")
		result, wasReturn, wasCleared, err = i.executeBlock(step.ElseValue, stepNum, "IF-ELSE", isInHandler, activeError)
	}
	return result, wasReturn, wasCleared, err
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE")
	iteration := 0
	maxIterations := 10000 // Safety break

	for iteration < maxIterations {
		iteration++
		// CORRECTED: Call evaluateExpression
		condResult, evalErr := i.evaluateExpression(step.Cond) // Pass context flags if evaluateExpression needs them
		if evalErr != nil {
			err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("iteration %d: evaluating WHILE condition", iteration), evalErr)
			break
		}
		// isTruthy needs to be defined elsewhere
		if !isTruthy(condResult) {
			i.Logger().Debug("[DEBUG-INTERP]   WHILE condition FALSE on iteration %d. Exiting loop.", iteration)
			break
		}

		i.Logger().Debug("[DEBUG-INTERP]   WHILE condition TRUE on iteration %d. Executing block.", iteration)
		var blockResult interface{}
		var blockReturned, blockCleared bool
		var blockErr error
		blockResult, blockReturned, blockCleared, blockErr = i.executeBlock(step.Value, stepNum, "WHILE-BODY", isInHandler, activeError)

		if blockErr != nil {
			err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("iteration %d: executing WHILE body", iteration), blockErr)
			break
		}
		if blockReturned {
			return blockResult, true, false, nil // Propagate return immediately
		}
		if blockCleared {
			wasCleared = true
			i.Logger().Debug("[DEBUG-INTERP]   CLEAR_ERROR detected within WHILE loop body on iteration %d.", iteration)
		}
		result = blockResult
	} // End loop

	if err != nil {
		return nil, false, false, err
	}

	if iteration >= maxIterations {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("step %d (WHILE): loop exceeded max iterations (%d)", stepNum+1, maxIterations), nil)
	}

	i.Logger().Debug("[DEBUG-INTERP]   WHILE loop finished normally.")
	return result, false, wasCleared, nil
}

// executeFor handles the "for" (for each) step.
func (i *Interpreter) executeFor(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH: Var=%s", step.Target)
	// CORRECTED: Call evaluateExpression
	collectionVal, evalErr := i.evaluateExpression(step.Cond) // Pass context flags if evaluateExpression needs them
	if evalErr != nil {
		return nil, false, false, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating collection for FOR EACH %s", step.Target), evalErr)
	}

	val := reflect.ValueOf(collectionVal)
	iteration := 0
	maxIterations := 10000 // Safety break

	executeLoopBody := func(item interface{}) (interface{}, bool, bool, error) {
		if setErr := i.SetVariable(step.Target, item); setErr != nil {
			// Wrap set error
			return nil, false, false, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("iteration %d: setting loop var '%s'", iteration, step.Target), setErr)
		}
		i.Logger().Debug("[DEBUG-INTERP]   FOR iteration %d (%s=%v). Executing block.", iteration, step.Target, item)
		// Pass context down
		return i.executeBlock(step.Value, stepNum, "FOR-BODY", isInHandler, activeError)
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < val.Len(); idx++ {
			iteration++
			if iteration > maxIterations {
				err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("step %d (FOR slice/array): loop exceeded max iterations (%d)", stepNum+1, maxIterations), nil)
				break
			}
			var blockResult interface{}
			var blockReturned, blockCleared bool
			var blockErr error
			blockResult, blockReturned, blockCleared, blockErr = executeLoopBody(val.Index(idx).Interface())
			if blockErr != nil {
				err = blockErr
				break
			}
			if blockReturned {
				return blockResult, true, false, nil
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult
		}
	case reflect.Map:
		mapRange := val.MapRange()
		for mapRange.Next() {
			iteration++
			if iteration > maxIterations {
				err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("step %d (FOR map): loop exceeded max iterations (%d)", stepNum+1, maxIterations), nil)
				break
			}
			var blockResult interface{}
			var blockReturned, blockCleared bool
			var blockErr error
			blockResult, blockReturned, blockCleared, blockErr = executeLoopBody(mapRange.Value().Interface()) // Iterate values
			if blockErr != nil {
				err = blockErr
				break
			}
			if blockReturned {
				return blockResult, true, false, nil
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult
		}
	case reflect.String:
		str := val.String()
		for _, char := range str { // Iterate runes
			iteration++
			if iteration > maxIterations {
				err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("step %d (FOR string): loop exceeded max iterations (%d)", stepNum+1, maxIterations), nil)
				break
			}
			var blockResult interface{}
			var blockReturned, blockCleared bool
			var blockErr error
			blockResult, blockReturned, blockCleared, blockErr = executeLoopBody(string(char)) // Convert rune to string
			if blockErr != nil {
				err = blockErr
				break
			}
			if blockReturned {
				return blockResult, true, false, nil
			}
			if blockCleared {
				wasCleared = true
			}
			result = blockResult
		}
	default:
		err = NewRuntimeError(ErrorCodeType, fmt.Sprintf("step %d (FOR): cannot iterate over type %T", stepNum+1, collectionVal), nil)
	}

	if err != nil {
		return nil, false, false, err
	}

	i.Logger().Debug("[DEBUG-INTERP]   FOR loop finished normally.")
	return result, false, wasCleared, nil
}

// --- Placeholder for isTruthy ---
// func isTruthy(value interface{}) bool { ... } // Assume defined elsewhere

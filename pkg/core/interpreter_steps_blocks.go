// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected executeFor to use step.LoopVarName and step.Collection.
// filename: pkg/core/interpreter_steps_blocks.go
// nlines: 200 // Approximate
// risk_rating: MEDIUM
package core

import (
	"errors"
	"fmt"
	"reflect"
	// Added for strings.ToLower if used elsewhere, or can be removed if not.
)

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing IF", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "IF step has nil Condition", nil).WithPosition(step.Pos)
	}

	condResult, evalErr := i.evaluateExpression(step.Cond)
	if evalErr != nil {
		return nil, false, false, WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating IF condition")
	}

	if isTruthy(condResult) { // isTruthy is assumed to be a helper
		i.Logger().Debug("[DEBUG-INTERP]     IF condition TRUE, executing THEN block", "pos", posStr)
		return i.executeBlock(step.Body, stepNum, "IF_THEN", isInHandler, activeError)
	} else if step.Else != nil {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, executing ELSE block", "pos", posStr)
		return i.executeBlock(step.Else, stepNum, "IF_ELSE", isInHandler, activeError)
	}

	i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, no ELSE block", "pos", posStr)
	return nil, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(step.Pos)
	}

	iteration := 0
	maxIterations := i.maxLoopIterations // Assume i.maxLoopIterations is set on the interpreter

	for {
		iteration++
		if iteration > maxIterations {
			errMsg := fmt.Sprintf("WHILE loop at %s exceeded max iterations (%d)", posStr, maxIterations)
			return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
		}

		condResult, evalErr := i.evaluateExpression(step.Cond)
		if evalErr != nil {
			return nil, false, wasCleared, WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating WHILE condition")
		}

		if !isTruthy(condResult) {
			i.Logger().Debug("[DEBUG-INTERP]     WHILE condition FALSE, exiting loop", "pos", posStr, "iterations", iteration-1)
			break // Exit loop
		}

		i.Logger().Debug("[DEBUG-INTERP]     WHILE condition TRUE, executing block", "pos", posStr, "iteration", iteration)
		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, stepNum, "WHILE_BODY", isInHandler, activeError)

		if blockCleared { // If error was cleared inside the loop body
			wasCleared = true
			activeError = nil // The error that might have triggered isInHandler is now cleared for this scope
		}

		if blockErr != nil {
			if errors.Is(blockErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]     BREAK encountered in WHILE loop body", "pos", posStr)
				return result, false, wasCleared, nil // Break the loop, not an error propagation
			}
			if errors.Is(blockErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]     CONTINUE encountered in WHILE loop body, proceeding to next iteration", "pos", posStr)
				// Any 'lastCallResult' updates from before 'continue' are preserved.
				// 'wasCleared' from this iteration of the body is also preserved.
				continue // Skip to next condition check
			}
			return nil, false, wasCleared, blockErr // Propagate other errors
		}
		if blockReturned {
			return blockResult, true, wasCleared, nil // Propagate return immediately
		}
		// If no error and no return, result might carry the last value from the block if needed by some convention
		// For now, 'result' is primarily for 'return' values. The lastCallResult is handled by executeSteps.
		result = blockResult
	}
	return result, false, wasCleared, nil // Normal loop completion
}

// executeFor handles the "for each" step.
// MODIFIED to use step.LoopVarName and step.Collection
func (i *Interpreter) executeFor(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}

	// Use LoopVarName and Collection from the Step struct
	loopVar := step.LoopVarName
	collectionExpr := step.Collection

	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH", "loopVar", loopVar, "pos", posStr)

	if collectionExpr == nil {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "FOR EACH step has nil Collection expression", nil).WithPosition(step.Pos)
	}
	if loopVar == "" {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(step.Pos)
	}

	collectionVal, evalErr := i.evaluateExpression(collectionExpr)
	if evalErr != nil {
		return nil, false, false, WrapErrorWithPosition(evalErr, collectionExpr.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", loopVar))
	}

	iteration := 0
	maxIterations := i.maxLoopIterations // Assume i.maxLoopIterations is set

	valReflection := reflect.ValueOf(collectionVal)

	switch valReflection.Kind() {
	case reflect.Slice, reflect.Array:
		for itemIdx := 0; itemIdx < valReflection.Len(); itemIdx++ {
			iteration++
			if iteration > maxIterations {
				errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", loopVar, posStr, maxIterations)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
			}

			item := valReflection.Index(itemIdx).Interface()
			if setErr := i.SetVariable(loopVar, item); setErr != nil {
				errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH", loopVar)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, setErr).WithPosition(step.Pos)
			}

			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, stepNum, "FOR_BODY", isInHandler, activeError)
			if blockCleared {
				wasCleared = true
				activeError = nil
			}

			if blockErr != nil {
				if errors.Is(blockErr, ErrBreak) {
					return result, false, wasCleared, nil
				}
				if errors.Is(blockErr, ErrContinue) {
					continue
				}
				return nil, false, wasCleared, blockErr
			}
			if blockReturned {
				return blockResult, true, wasCleared, nil
			}
			result = blockResult
		}
	case reflect.Map:
		mapKeys := valReflection.MapKeys() // Note: Iteration order is not guaranteed for maps
		for _, key := range mapKeys {
			iteration++
			if iteration > maxIterations {
				errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", loopVar, posStr, maxIterations)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
			}

			// For maps, traditionally the loop variable gets the value, not the key.
			// If key access is needed, one might use `set k = item_key` if NeuroScript supports map iteration with key and value.
			// Current NeuroScript `for each var in map` implies var gets map values.
			item := valReflection.MapIndex(key).Interface()
			if setErr := i.SetVariable(loopVar, item); setErr != nil {
				errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH (map)", loopVar)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, setErr).WithPosition(step.Pos)
			}
			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, stepNum, "FOR_BODY_MAP", isInHandler, activeError)
			if blockCleared {
				wasCleared = true
				activeError = nil
			}

			if blockErr != nil {
				if errors.Is(blockErr, ErrBreak) {
					return result, false, wasCleared, nil
				}
				if errors.Is(blockErr, ErrContinue) {
					continue
				}
				return nil, false, wasCleared, blockErr
			}
			if blockReturned {
				return blockResult, true, wasCleared, nil
			}
			result = blockResult
		}
	case reflect.String:
		strCollection := collectionVal.(string)
		for _, charRune := range strCollection { // Iterates over runes (characters)
			iteration++
			if iteration > maxIterations {
				errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", loopVar, posStr, maxIterations)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
			}

			item := string(charRune)
			if setErr := i.SetVariable(loopVar, item); setErr != nil {
				errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH (string)", loopVar)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, setErr).WithPosition(step.Pos)
			}
			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, stepNum, "FOR_BODY_STRING", isInHandler, activeError)
			if blockCleared {
				wasCleared = true
				activeError = nil
			}

			if blockErr != nil {
				if errors.Is(blockErr, ErrBreak) {
					return result, false, wasCleared, nil
				}
				if errors.Is(blockErr, ErrContinue) {
					continue
				}
				return nil, false, wasCleared, blockErr
			}
			if blockReturned {
				return blockResult, true, wasCleared, nil
			}
			result = blockResult
		}
	default:
		errMsg := fmt.Sprintf("cannot iterate over type %T for FOR EACH %s", collectionVal, loopVar)
		return nil, false, wasCleared, NewRuntimeError(ErrorCodeType, errMsg, nil).WithPosition(collectionExpr.GetPos())
	}

	i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop finished normally.", "loopVar", loopVar, "pos", posStr, "iterations", iteration)
	return result, false, wasCleared, nil
}

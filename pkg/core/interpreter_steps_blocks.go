// NeuroScript Version: 0.3.1
// File version: 5
// Purpose: Corrected conditional checks in if/while loops to use the IsTruthy() method on Value types, aligning them with the rest of the interpreter and fixing runtime errors.
// filename: pkg/core/interpreter_steps_blocks.go
// nlines: 200
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
	"reflect"
)

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step Step, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
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

	// CORRECTED: Use the IsTruthy() method for consistent behavior.
	if condResult.IsTruthy() {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition TRUE, executing THEN block", "pos", posStr)
		return i.executeBlock(step.Body, step.Pos, "IF_THEN", isInHandler, activeError)
	} else if step.Else != nil {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, executing ELSE block", "pos", posStr)
		return i.executeBlock(step.Else, step.Pos, "IF_ELSE", isInHandler, activeError)
	}

	i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, no ELSE block", "pos", posStr)
	return nil, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step Step, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(step.Pos)
	}

	iteration := 0
	maxIterations := i.maxLoopIterations

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

		// CORRECTED: Use the IsTruthy() method for consistent behavior.
		if !condResult.IsTruthy() {
			i.Logger().Debug("[DEBUG-INTERP]     WHILE condition FALSE, exiting loop", "pos", posStr, "iterations", iteration-1)
			break // Exit loop
		}

		i.Logger().Debug("[DEBUG-INTERP]     WHILE condition TRUE, executing block", "pos", posStr, "iteration", iteration)
		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.Pos, "WHILE_BODY", isInHandler, activeError)

		if blockCleared {
			wasCleared = true
			activeError = nil
		}

		if blockErr != nil {
			if errors.Is(blockErr, ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]     BREAK encountered in WHILE loop body", "pos", posStr)
				return result, false, wasCleared, nil
			}
			if errors.Is(blockErr, ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]     CONTINUE encountered in WHILE loop body, proceeding to next iteration", "pos", posStr)
				continue
			}
			return nil, false, wasCleared, blockErr
		}
		if blockReturned {
			return blockResult, true, wasCleared, nil
		}
		result = blockResult
	}
	return result, false, wasCleared, nil
}

// executeFor handles the "for each" step.
func (i *Interpreter) executeFor(step Step, isInHandler bool, activeError *RuntimeError) (result interface{}, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}

	loopVar := step.LoopVarName
	collectionExpr := step.Collection

	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH", "loopVar", loopVar, "pos", posStr)

	if collectionExpr == nil {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "FOR EACH step has nil Collection expression", nil).WithPosition(step.Pos)
	}
	if loopVar == "" {
		return nil, false, false, NewRuntimeError(ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(step.Pos)
	}

	collectionValRaw, evalErr := i.evaluateExpression(collectionExpr)
	if evalErr != nil {
		return nil, false, false, WrapErrorWithPosition(evalErr, collectionExpr.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", loopVar))
	}

	// Use unwrapValue to handle both raw types and Value types
	collectionVal := unwrapValue(collectionValRaw)

	iteration := 0
	maxIterations := i.maxLoopIterations

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

			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.Pos, "FOR_BODY", isInHandler, activeError)
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
		mapKeys := valReflection.MapKeys()
		for _, key := range mapKeys {
			iteration++
			if iteration > maxIterations {
				errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", loopVar, posStr, maxIterations)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
			}

			item := valReflection.MapIndex(key).Interface()
			if setErr := i.SetVariable(loopVar, item); setErr != nil {
				errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH (map)", loopVar)
				return nil, false, wasCleared, NewRuntimeError(ErrorCodeInternal, errMsg, setErr).WithPosition(step.Pos)
			}
			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.Pos, "FOR_BODY_MAP", isInHandler, activeError)
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
		for _, charRune := range strCollection {
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
			blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.Pos, "FOR_BODY_STRING", isInHandler, activeError)
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
		errMsg := fmt.Sprintf("cannot iterate over type %s for FOR EACH %s", TypeOf(collectionValRaw), loopVar)
		return nil, false, wasCleared, NewRuntimeError(ErrorCodeType, errMsg, nil).WithPosition(collectionExpr.GetPos())
	}

	i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop finished normally.", "loopVar", loopVar, "pos", posStr, "iterations", iteration)
	return result, false, wasCleared, nil
}

// pkg/core/interpreter_control_flow.go
package core

import (
	// Need errors package
	"fmt"
	"sort"
	"strings"
)

// --- Control Flow Statement Execution Helpers ---

// executeIf handles IF/ELSE statement (Uses specific end keyword implicitly via AST)
func (i *Interpreter) executeIf(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing IF (Step %d)", stepNum+1)
	}
	conditionNode := step.Cond
	conditionResult, evalErr := i.evaluateCondition(conditionNode)
	if evalErr != nil {
		return nil, false, fmt.Errorf("evaluating IF condition: %w", evalErr)
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]        IF condition evaluated to %t", conditionResult)
	}

	if conditionResult {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        IF condition TRUE, executing THEN block.")
		}
		blockResult, blockReturned, blockErr := i.executeBlock(step.Value, stepNum, "IF-THEN")
		if blockErr != nil {
			return nil, false, blockErr
		}
		if blockReturned {
			return blockResult, true, nil
		}
	} else if step.ElseValue != nil {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        IF condition FALSE, executing ELSE block.")
		}
		blockResult, blockReturned, blockErr := i.executeBlock(step.ElseValue, stepNum, "IF-ELSE")
		if blockErr != nil {
			return nil, false, blockErr
		}
		if blockReturned {
			return blockResult, true, nil
		}
	} else {
		if i.logger != nil {
			i.logger.Debug("-INTERP]        IF condition FALSE, no ELSE block found, skipping.")
		}
	}
	return nil, false, nil // Normal completion if no return happened
}

// executeWhile handles WHILE loops (Uses specific end keyword implicitly via AST)
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing WHILE (Step %d)", stepNum+1)
	}
	conditionNode := step.Cond
	loopCounter := 0
	maxLoops := 1000
	for loopCounter < maxLoops {
		conditionResult, evalErr := i.evaluateCondition(conditionNode)
		if evalErr != nil {
			return nil, false, fmt.Errorf("evaluating WHILE condition (iter %d): %w", loopCounter, evalErr)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        WHILE condition (iter %d) evaluated to %t", loopCounter, conditionResult)
		}
		if !conditionResult {
			break
		} // Exit loop

		if i.logger != nil {
			i.logger.Debug("-INTERP]        WHILE condition TRUE, executing block (iter %d).", loopCounter)
		}
		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter))
		if bodyErr != nil {
			return nil, false, fmt.Errorf("error in WHILE loop body (iter %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        WHILE loop body returned (iter %d). Propagating.", loopCounter)
			}
			return bodyResult, true, nil
		}
		loopCounter++
	}
	if loopCounter >= maxLoops {
		return nil, false, fmt.Errorf("WHILE loop exceeded max iterations (%d)", maxLoops)
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]      WHILE loop finished after %d iterations.", loopCounter)
	}
	return nil, false, nil
}

// executeFor handles FOR EACH loops (Uses specific end keyword implicitly via AST)
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionNode := step.Cond
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing FOR EACH %s IN ... (Step %d)", loopVar, stepNum+1)
	}
	if !isValidIdentifier(loopVar) {
		return nil, false, fmt.Errorf("FOR loop variable '%s' is not valid", loopVar)
	}
	evaluatedCollection, evalErr := i.evaluateExpression(collectionNode)
	if evalErr != nil {
		return nil, false, fmt.Errorf("evaluating collection for FOR loop: %w", evalErr)
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]        FOR evaluated collection: %v (%T)", evaluatedCollection, evaluatedCollection)
	}

	originalLoopVarValue, loopVarExists := i.variables[loopVar]
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue
		} else {
			delete(i.variables, loopVar)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Restored loop variable '%s' state after FOR.", loopVar)
		}
	}()

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}
	iterations := 0
	switch collection := evaluatedCollection.(type) {
	case []string:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over []string (len %d)...", len(collection))
		}
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item
			if i.logger != nil {
				i.logger.Debug("-INTERP]          String List Iter %d: Assign '%s' = %q (%T)", itemNum, loopVar, item, item)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Str List Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case []interface{}:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over []interface{} (len %d)...", len(collection))
		}
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item
			if i.logger != nil {
				i.logger.Debug("-INTERP]          Interface List Iter %d: Assign '%s' = %v (%T)", itemNum, loopVar, item, item)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Interface List Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case map[string]interface{}:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over map keys (size %d)...", len(collection))
		}
		keys := make([]string, 0, len(collection))
		for k := range collection {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for itemNum, key := range keys {
			iterations++
			i.variables[loopVar] = key
			if i.logger != nil {
				i.logger.Debug("-INTERP]          Map Key Iter %d: Assign '%s' = %q", itemNum, loopVar, key)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Map Key Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case string:
		if strings.Contains(collection, ",") {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        FOR iterating over comma-separated string...")
			}
			items := strings.Split(collection, ",")
			for itemNum, item := range items {
				iterations++
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Comma Iter %d: Assign '%s' = %q", itemNum, loopVar, trimmedItem)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		} else {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        FOR iterating over string characters...")
			}
			for itemNum, charRune := range collection {
				iterations++
				charStr := string(charRune)
				i.variables[loopVar] = charStr
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Char Iter %d: Assign '%s' = %q", itemNum, loopVar, charStr)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		}
	case nil:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over nil collection (0 iterations).")
		}
	default:
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR loop", evaluatedCollection)
		if i.logger != nil {
			i.logger.Error("%v", bodyErr)
		}
	}
	if bodyErr != nil {
		return nil, false, fmt.Errorf("error during FOR loop: %w", bodyErr)
	}
	if returnedFromBody {
		return resultFromLoop, true, nil
	}
	if i.logger != nil {
		i.logger.Debug("-INTERP]      FOR loop finished normally after %d iterations.", iterations)
	}
	return nil, false, nil
}

// --- NEW: executeTryCatch ---
func (i *Interpreter) executeTryCatch(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing TRY (Step %d)", stepNum+1)
	}

	var tryErr error
	var tryResult interface{}
	var tryReturned bool

	// Execute TRY block
	if i.logger != nil {
		i.logger.Debug("-INTERP]        >> Entering TRY block execution.")
	}
	tryResult, tryReturned, tryErr = i.executeBlock(step.Value, stepNum, "TRY")
	if i.logger != nil {
		i.logger.Debug("-INTERP]        << Exiting TRY block execution. Returned=%t, Err=%v", tryReturned, tryErr)
	}

	// If TRY returned or succeeded without error, execute FINALLY (if exists)
	if tryErr == nil && !tryReturned {
		// TRY succeeded, proceed to FINALLY
		if step.FinallySteps != nil {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        >> Entering FINALLY block execution (after successful TRY).")
			}
			finResult, finReturned, finErr := i.executeBlock(step.FinallySteps, stepNum, "FINALLY (after TRY success)")
			if i.logger != nil {
				i.logger.Debug("-INTERP]        << Exiting FINALLY block execution. Returned=%t, Err=%v", finReturned, finErr)
			}
			if finErr != nil {
				return nil, false, finErr
			} // Error in FINALLY overrides everything
			if finReturned {
				return finResult, true, nil
			} // RETURN in FINALLY overrides TRY result
		}
		// No FINALLY or FINALLY completed normally
		return nil, false, nil // Overall TRY-FINALLY finished without error or return
	} else if tryReturned {
		// TRY executed RETURN, execute FINALLY (if exists)
		if step.FinallySteps != nil {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        >> Entering FINALLY block execution (after TRY returned).")
			}
			finResult, finReturned, finErr := i.executeBlock(step.FinallySteps, stepNum, "FINALLY (after TRY return)")
			if i.logger != nil {
				i.logger.Debug("-INTERP]        << Exiting FINALLY block execution. Returned=%t, Err=%v", finReturned, finErr)
			}
			if finErr != nil {
				return nil, false, finErr
			} // Error in FINALLY overrides everything
			if finReturned {
				return finResult, true, nil
			} // RETURN in FINALLY overrides TRY return
		}
		// No FINALLY or FINALLY completed normally
		return tryResult, true, nil // Propagate the RETURN from TRY
	} else { // tryErr != nil
		// Error occurred in TRY block
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Error occurred in TRY block: %v", tryErr)
		}

		// Execute CATCH block (if exists)
		if step.CatchSteps != nil {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        >> Entering CATCH block execution.")
			}
			// --- Scope handling for catch variable ---
			originalCatchVarValue, catchVarExists := interface{}(nil), false
			if step.CatchVar != "" {
				originalCatchVarValue, catchVarExists = i.variables[step.CatchVar]
				// Assign error message (or error object?) to catch variable
				i.variables[step.CatchVar] = tryErr.Error() // Assign error string
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Assigned error '%v' to catch var '%s'", tryErr.Error(), step.CatchVar)
				}
			}
			// Execute catch block steps
			catchResult, catchReturned, catchErr := i.executeBlock(step.CatchSteps, stepNum, "CATCH")
			// Restore catch variable scope
			if step.CatchVar != "" {
				if catchVarExists {
					i.variables[step.CatchVar] = originalCatchVarValue
				} else {
					delete(i.variables, step.CatchVar)
				}
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Restored catch variable '%s' state.", step.CatchVar)
				}
			}
			if i.logger != nil {
				i.logger.Debug("-INTERP]        << Exiting CATCH block execution. Returned=%t, Err=%v", catchReturned, catchErr)
			}

			// Handle CATCH block outcome
			if catchErr != nil {
				// Error in CATCH block, proceed to FINALLY (if exists), error from CATCH reported later
				tryErr = catchErr // Replace original error with catch error
			} else if catchReturned {
				// CATCH block executed RETURN, proceed to FINALLY (if exists)
				tryErr = nil // Error was handled by catch return
				tryResult = catchResult
				tryReturned = true
			} else {
				// CATCH block completed normally, error is considered handled
				tryErr = nil
			}
		} // End if CATCH block exists

		// Execute FINALLY block (if exists) - always runs after TRY/CATCH phase
		if step.FinallySteps != nil {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        >> Entering FINALLY block execution (after TRY error/CATCH).")
			}
			finResult, finReturned, finErr := i.executeBlock(step.FinallySteps, stepNum, "FINALLY (after TRY error/CATCH)")
			if i.logger != nil {
				i.logger.Debug("-INTERP]        << Exiting FINALLY block execution. Returned=%t, Err=%v", finReturned, finErr)
			}
			if finErr != nil {
				return nil, false, finErr
			} // Error in FINALLY overrides everything
			if finReturned {
				return finResult, true, nil
			} // RETURN in FINALLY overrides CATCH return/handled error
		}

		// Determine final outcome after FINALLY
		if tryErr != nil {
			// Original TRY error was not handled by CATCH, or CATCH itself errored
			return nil, false, tryErr
		}
		if tryReturned {
			// RETURN from CATCH (and FINALLY didn't override)
			return tryResult, true, nil
		}
		// CATCH handled the error, FINALLY ran ok
		return nil, false, nil
	}
}

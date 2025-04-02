// pkg/core/interpreter_control_flow.go
package core

import (
	"fmt"
	"sort"
	"strings"
)

// --- Control Flow Statement Execution Helpers ---

// executeIf handles the IF statement, now checking condition evaluation error
func (i *Interpreter) executeIf(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing IF")
	}
	conditionNode := step.Cond

	// Evaluate condition and check for errors
	conditionResult, evalErr := i.evaluateCondition(conditionNode) // Uses helper from evaluation_comparison.go
	if evalErr != nil {
		// Return error immediately if condition evaluation fails
		return nil, false, fmt.Errorf("evaluating IF condition: %w", evalErr)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        IF condition evaluated to %t", conditionResult)
	}

	if conditionResult {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        IF condition TRUE, executing block.")
		}
		// Execute block and propagate its result/return/error status
		blockResult, blockReturned, blockErr := i.executeBlock(step.Value, stepNum, "IF")
		if blockErr != nil {
			return nil, false, blockErr // Propagate block error
		}
		if blockReturned {
			return blockResult, true, nil // Propagate block return
		}
		// If block finished normally, indicate success but no return
		return nil, false, nil
	} else {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        IF condition FALSE, skipping block.")
		}
		// Condition false, no error, no return
		return nil, false, nil
	}
}

// executeWhile handles the WHILE statement, now checking condition evaluation error
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing WHILE")
	}
	conditionNode := step.Cond
	loopCounter := 0
	maxLoops := 1000 // Safety break

	for loopCounter < maxLoops {
		// Evaluate condition and check for errors
		conditionResult, evalErr := i.evaluateCondition(conditionNode)
		if evalErr != nil {
			return nil, false, fmt.Errorf("evaluating WHILE condition (iteration %d): %w", loopCounter, evalErr)
		}

		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        WHILE condition (iter %d) evaluated to %t", loopCounter, conditionResult)
		}

		if !conditionResult {
			break // Exit loop if condition is false
		}

		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        WHILE condition TRUE, executing block (iter %d).", loopCounter)
		}
		// Execute loop body
		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter))
		if bodyErr != nil {
			// Propagate error from loop body
			return nil, false, fmt.Errorf("error in WHILE loop body (iteration %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			// Propagate RETURN from loop body
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]        WHILE loop body returned (iter %d). Propagating.", loopCounter)
			}
			return bodyResult, true, nil
		}
		loopCounter++
	} // End loop

	if loopCounter >= maxLoops {
		return nil, false, fmt.Errorf("WHILE loop exceeded max iterations (%d)", maxLoops)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      WHILE loop finished after %d iterations.", loopCounter)
	}
	return nil, false, nil // Normal loop completion, nil error
}

// executeFor handles FOR EACH, including list and map iteration
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionNode := step.Cond

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing FOR EACH %s IN ...", loopVar)
	}

	if !isValidIdentifier(loopVar) { // Use helper
		return nil, false, fmt.Errorf("FOR loop variable '%s' is not a valid identifier", loopVar)
	}

	// Log node details BEFORE evaluation
	if i.logger != nil {
		condStr := "<nil>"
		condType := "<nil>"
		if collectionNode != nil {
			condType = fmt.Sprintf("%T", collectionNode)
			condStr = fmt.Sprintf("%+v", collectionNode)
		}
		i.logger.Printf("[DEBUG-INTERP]        FOR evaluating collection node: (%s %s)", condType, condStr)
	}

	// Evaluate the collection expression and check for errors
	evaluatedCollection, evalErr := i.evaluateExpression(collectionNode) // Depth 0
	if evalErr != nil {
		return nil, false, fmt.Errorf("evaluating collection for FOR EACH loop: %w", evalErr)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        FOR evaluated collection: %v (%T)", evaluatedCollection, evaluatedCollection)
	}

	// Save/restore loop variable state
	originalLoopVarValue, loopVarExists := i.variables[loopVar]
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue
		} else {
			delete(i.variables, loopVar)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Restored loop variable '%s' state after FOR.", loopVar)
		}
	}() // End defer

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}
	iterations := 0

	switch collection := evaluatedCollection.(type) {
	case []interface{}: // Iterate over slice elements
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        FOR iterating over slice (len %d)...", len(collection))
		}
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item // Assign current item
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          Slice Iter %d: Assigned '%s' = %v (%T)", itemNum, loopVar, item, item)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Slice Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			} // Exit loop on error or return
		}

	case map[string]interface{}: // Iterate over map keys
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        FOR iterating over map keys (size %d)...", len(collection))
		}
		keys := make([]string, 0, len(collection))
		for k := range collection {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Iterate in deterministic order

		for itemNum, key := range keys {
			iterations++
			i.variables[loopVar] = key // Assign current KEY
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          Map Key Iter %d: Assigned '%s' = %q", itemNum, loopVar, key)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Map Key Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			} // Exit loop on error or return
		}

	case string: // Iterate over string (comma split OR characters) - Existing logic remains
		shouldCommaSplit := false
		if strings.Contains(collection, ",") {
			parts := strings.Split(collection, ",")
			if len(parts) > 1 || (len(parts) == 1 && strings.TrimSpace(parts[0]) != strings.TrimSpace(collection)) {
				shouldCommaSplit = true
			}
		}

		if shouldCommaSplit {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]        FOR iterating over comma-separated string...")
			}
			items := strings.Split(collection, ",")
			for itemNum, item := range items {
				iterations++
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Comma Iter %d: Assigned '%s' = %q", itemNum, loopVar, trimmedItem)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		} else { // Character Iteration
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]        FOR iterating over string characters...")
			}
			for itemNum, charRune := range collection {
				iterations++
				charStr := string(charRune)
				i.variables[loopVar] = charStr
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Char Iter %d: Assigned '%s' = %q", itemNum, loopVar, charStr)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		}

	case nil: // Allow iterating over nil (0 iterations)
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        FOR iterating over nil collection (0 iterations).")
		}

	default: // Cannot iterate over other types
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH loop", evaluatedCollection)
		if i.logger != nil {
			i.logger.Printf("[ERROR] %v", bodyErr)
		}
	} // End switch

	// Handle loop exit reason
	if bodyErr != nil {
		return nil, false, fmt.Errorf("error during FOR EACH loop execution: %w", bodyErr)
	}
	if returnedFromBody {
		return resultFromLoop, true, nil
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      FOR EACH loop finished normally after %d iterations.", iterations)
	}
	return nil, false, nil // Normal loop completion, nil error
}

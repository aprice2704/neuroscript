// pkg/core/interpreter_control_flow.go
package core

import (
	"fmt"
	"sort"
	"strings"
)

// --- Control Flow Statement Execution Helpers ---

// executeIf handles IF/ELSE statement, checking condition evaluation error
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
		// --- Execute THEN block (step.Value) ---
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
		return nil, false, nil
	} else {
		// --- Execute ELSE block (step.ElseValue) ---
		if step.ElseValue != nil {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        IF condition FALSE, executing ELSE block.")
			}
			// NOTE: This relies on AST Builder correctly populating ElseValue.
			blockResult, blockReturned, blockErr := i.executeBlock(step.ElseValue, stepNum, "IF-ELSE")
			if blockErr != nil {
				return nil, false, blockErr
			}
			if blockReturned {
				return blockResult, true, nil
			}
			return nil, false, nil
		} else {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        IF condition FALSE, no ELSE block found, skipping.")
			}
			return nil, false, nil
		}
	}
}

// executeWhile handles WHILE loops (remains the same)
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing WHILE (Step %d)", stepNum+1)
	}
	conditionNode := step.Cond
	loopCounter := 0
	maxLoops := 1000 // Prevent infinite loops
	for loopCounter < maxLoops {
		conditionResult, evalErr := i.evaluateCondition(conditionNode)
		if evalErr != nil {
			return nil, false, fmt.Errorf("evaluating WHILE condition (iteration %d): %w", loopCounter, evalErr)
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        WHILE condition (iter %d) evaluated to %t", loopCounter, conditionResult)
		}
		if !conditionResult {
			break // Exit loop if condition is false
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        WHILE condition TRUE, executing block (iter %d).", loopCounter)
		}
		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter))
		if bodyErr != nil {
			return nil, false, fmt.Errorf("error in WHILE loop body (iteration %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			// If the block executed a RETURN, propagate it up immediately
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
	return nil, false, nil // Normal loop completion
}

// executeFor handles FOR EACH loops (Updated for list iteration)
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionNode := step.Cond // The AST node representing the collection

	if i.logger != nil {
		i.logger.Debug("-INTERP]      Executing FOR EACH %s IN ... (Step %d)", loopVar, stepNum+1)
	}
	// Validate loop variable name
	if !isValidIdentifier(loopVar) {
		return nil, false, fmt.Errorf("FOR loop variable '%s' is not a valid identifier", loopVar)
	}
	// Log the collection node details before evaluation
	if i.logger != nil {
		condStr := "<nil>"
		condType := "<nil>"
		if collectionNode != nil {
			condType = fmt.Sprintf("%T", collectionNode)
			condStr = fmt.Sprintf("%+v", collectionNode) // Log AST node details
		}
		i.logger.Debug("-INTERP]        FOR evaluating collection node: (%s %s)", condType, condStr)
	}

	// Evaluate the collection expression
	evaluatedCollection, evalErr := i.evaluateExpression(collectionNode)
	if evalErr != nil {
		return nil, false, fmt.Errorf("evaluating collection for FOR EACH loop: %w", evalErr)
	}

	if i.logger != nil {
		i.logger.Debug("-INTERP]        FOR evaluated collection: %v (%T)", evaluatedCollection, evaluatedCollection)
	}

	// --- Scope handling for loop variable ---
	originalLoopVarValue, loopVarExists := i.variables[loopVar] // Save potential existing value
	// Defer restoring the original value or deleting the temp loop var
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue // Restore
		} else {
			delete(i.variables, loopVar) // Clean up
		}
		if i.logger != nil {
			i.logger.Debug("-INTERP]        Restored loop variable '%s' state after FOR.", loopVar)
		}
	}()

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}
	iterations := 0

	// --- Iterate based on the evaluated collection type ---
	switch collection := evaluatedCollection.(type) {
	// *** NEW: Explicit case for []string (from TOOL.ListDirectory) ***
	case []string:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over []string (len %d)...", len(collection))
		}
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item // Assign current string item
			if i.logger != nil {
				i.logger.Debug("-INTERP]          String List Iter %d: Assigned '%s' = %q (%T)", itemNum, loopVar, item, item)
			}
			// Execute the loop body block
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR String List Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break // Exit loop on error or return from body
			}
		}
	// *** END NEW CASE ***

	// *** EXISTING CASE for []interface{} (from list literals) ***
	case []interface{}:
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over []interface{} (len %d)...", len(collection))
		}
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item // Assign current element (any type)
			if i.logger != nil {
				i.logger.Debug("-INTERP]          Interface List Iter %d: Assigned '%s' = %v (%T)", itemNum, loopVar, item, item)
			}
			// Execute the loop body block
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Interface List Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break // Exit loop on error or return from body
			}
		}
	// *** END EXISTING CASE ***

	case map[string]interface{}:
		// Map iteration (keys) - logic remains the same
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over map keys (size %d)...", len(collection))
		}
		keys := make([]string, 0, len(collection))
		for k := range collection {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Iterate keys in deterministic order
		for itemNum, key := range keys {
			iterations++
			i.variables[loopVar] = key // Assign current KEY
			if i.logger != nil {
				i.logger.Debug("-INTERP]          Map Key Iter %d: Assigned '%s' = %q", itemNum, loopVar, key)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Map Key Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case string:
		// String iteration (comma split or chars) - logic remains the same
		shouldCommaSplit := false
		// Basic check for comma - could be more robust (e.g., ignore if quoted)
		if strings.Contains(collection, ",") {
			shouldCommaSplit = true
		}
		if shouldCommaSplit {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        FOR iterating over comma-separated string...")
			}
			items := strings.Split(collection, ",")
			for itemNum, item := range items {
				iterations++
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem // Assign current part
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Comma Iter %d: Assigned '%s' = %q", itemNum, loopVar, trimmedItem)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		} else {
			// Iterate over characters (runes) if no comma is detected
			if i.logger != nil {
				i.logger.Debug("-INTERP]        FOR iterating over string characters...")
			}
			for itemNum, charRune := range collection {
				iterations++
				charStr := string(charRune)
				i.variables[loopVar] = charStr // Assign current character
				if i.logger != nil {
					i.logger.Debug("-INTERP]          Char Iter %d: Assigned '%s' = %q", itemNum, loopVar, charStr)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		}
	case nil:
		// Iterating over nil results in 0 iterations
		if i.logger != nil {
			i.logger.Debug("-INTERP]        FOR iterating over nil collection (0 iterations).")
		}
	default:
		// Collection type is not iterable
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH loop", evaluatedCollection)
		if i.logger != nil {
			i.logger.Error("%v", bodyErr) // Log the error immediately
		}
	} // End switch

	// --- Handle loop termination ---
	if bodyErr != nil {
		// Return error encountered within the loop body or during type checking
		return nil, false, fmt.Errorf("error during FOR EACH loop execution: %w", bodyErr)
	}
	if returnedFromBody {
		// Propagate return if the loop body executed RETURN
		return resultFromLoop, true, nil
	}

	// Normal loop completion (finished all iterations or collection was empty/nil)
	if i.logger != nil {
		i.logger.Debug("-INTERP]      FOR EACH loop finished normally after %d iterations.", iterations)
	}
	return nil, false, nil
}

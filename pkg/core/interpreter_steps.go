package core

import (
	"fmt"
	"sort" // Import sort package for stable map key iteration
	"strings"
)

// --- Step Execution Helpers ---

// executeSet evaluates the expression node and assigns the result
func (i *Interpreter) executeSet(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing SET")
	}
	targetVar := step.Target
	valueNode := step.Value

	if !isValidIdentifier(targetVar) {
		return fmt.Errorf("SET target '%s' is not a valid variable name", targetVar)
	}

	finalValue := i.evaluateExpression(valueNode)
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        SET evaluated value: %v (%T)", finalValue, finalValue)
	}

	if targetVar == "generated_code" {
		if finalStr, isStr := finalValue.(string); isStr {
			trimmedVal := trimCodeFences(finalStr)
			if trimmedVal != finalStr {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]        Trimmed code fences for 'generated_code'")
				}
				finalValue = trimmedVal
			}
		}
	}

	i.variables[targetVar] = finalValue
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        Stored var '%s' = %v (%T)", targetVar, finalValue, finalValue)
	}
	return nil
}

// executeCall evaluates argument nodes and performs the call
func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing CALL %q", step.Target)
	}
	target := step.Target
	argNodes := step.Args

	evaluatedArgs := make([]interface{}, len(argNodes))
	for idx, argNode := range argNodes {
		evaluatedArgs[idx] = i.evaluateExpression(argNode)
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        CALL Arg %d evaluated: %v (%T)", idx, evaluatedArgs[idx], evaluatedArgs[idx])
		}
	}

	var callResultValue interface{}
	var callErr error

	if strings.HasPrefix(target, "TOOL.") {
		toolName := strings.TrimPrefix(target, "TOOL.")
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling TOOL.%s", toolName)
		}
		toolImpl, found := i.toolRegistry.GetTool(toolName)
		if !found {
			return nil, fmt.Errorf("unknown TOOL '%s'", toolName)
		}

		preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
		if validationErr != nil {
			return nil, fmt.Errorf("TOOL %s argument error: %w", toolName, validationErr)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          Prepared TOOL args: %+v", preparedArgs)
		}

		callResultValue, callErr = toolImpl.Func(i, preparedArgs)
		if callErr != nil {
			return nil, fmt.Errorf("TOOL %s execution failed: %w", toolName, callErr)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          TOOL.%s Result: %v (%T)", toolName, callResultValue, callResultValue)
		}

	} else if target == "LLM" {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling LLM")
		}
		if len(evaluatedArgs) != 1 {
			return nil, fmt.Errorf("CALL LLM expects 1 prompt arg, got %d", len(evaluatedArgs))
		}
		prompt := fmt.Sprintf("%v", evaluatedArgs[0])
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          LLM Prompt: %q", prompt)
		}

		response, llmErr := CallLLMAPI(prompt)
		if llmErr != nil {
			return nil, fmt.Errorf("CALL LLM failed: %w", llmErr)
		}
		callResultValue = response
		callErr = nil
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          LLM Result: %q", response)
		}

	} else { // Procedure Call
		procToCall := target
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling Procedure %q", procToCall)
		}

		stringArgs := make([]string, len(evaluatedArgs))
		for idx, val := range evaluatedArgs {
			stringArgs[idx] = fmt.Sprintf("%v", val)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          Procedure args (as strings): %v", stringArgs)
		}

		procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...)
		if procCallErr != nil {
			return nil, procCallErr
		}
		callResultValue = procResultValue
		callErr = nil
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          Procedure %q Result: %v (%T)", procToCall, callResultValue, callResultValue)
		}
	}

	// Return result and error from the specific call type
	return callResultValue, callErr
}

// executeIf handles the IF statement with an expression node for condition
func (i *Interpreter) executeIf(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing IF")
	}
	conditionNode := step.Cond
	conditionResult, evalErr := i.evaluateCondition(conditionNode)
	if evalErr != nil {
		return nil, false, fmt.Errorf("IF condition evaluation error: %w", evalErr)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        IF condition evaluated to %t", conditionResult)
	}

	if conditionResult {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        IF condition TRUE, executing block.")
		}
		blockResult, blockReturned, blockErr := i.executeBlock(step.Value, stepNum, "IF")
		if blockErr != nil {
			return nil, false, blockErr
		}
		if blockReturned {
			return blockResult, true, nil
		}
		return true, false, nil // Indicate block executed
	} else {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        IF condition FALSE, skipping block.")
		}
	}

	return nil, false, nil // Condition was false
}

// executeReturn handles the RETURN statement with an expression node
func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing RETURN")
	}
	valueNode := step.Value
	var returnValue interface{}

	if valueNode != nil {
		returnValue = i.evaluateExpression(valueNode)
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        RETURN evaluated value: %v (%T)", returnValue, returnValue)
		}
	} else {
		returnValue = nil
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        RETURN with no value (implicit nil)")
		}
	}

	return returnValue, true, nil // Signal RETURN occurred
}

// executeEmit handles the EMIT statement with an expression node
func (i *Interpreter) executeEmit(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing EMIT")
	}
	valueNode := step.Value
	var evaluatedValue interface{}

	if valueNode != nil {
		evaluatedValue = i.evaluateExpression(valueNode)
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        EMIT evaluated value: %v (%T)", evaluatedValue, evaluatedValue)
		}
	} else {
		evaluatedValue = ""
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        EMIT with no value (implicit empty string)")
		}
	}

	fmt.Printf("[EMIT] %v\n", evaluatedValue) // Direct print for EMIT

	return nil
}

// executeWhile handles the WHILE statement with an expression node for condition
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing WHILE")
	}
	conditionNode := step.Cond
	loopCounter := 0
	maxLoops := 1000 // Safety break

	for loopCounter < maxLoops {
		conditionResult, evalErr := i.evaluateCondition(conditionNode)
		if evalErr != nil {
			return nil, false, fmt.Errorf("WHILE condition evaluation error (iter %d): %w", loopCounter, evalErr)
		}

		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        WHILE condition (iter %d) evaluated to %t", loopCounter, conditionResult)
		}

		if !conditionResult {
			break
		}

		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        WHILE condition TRUE, executing block (iter %d).", loopCounter)
		}
		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter))
		if bodyErr != nil {
			return nil, false, fmt.Errorf("error in WHILE loop body (iter %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]        WHILE loop body returned (iter %d). Propagating.", loopCounter)
			}
			return bodyResult, true, nil // Propagate RETURN
		}
		loopCounter++
	}

	if loopCounter >= maxLoops {
		return nil, false, fmt.Errorf("WHILE loop exceeded max iterations (%d)", maxLoops)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      WHILE loop finished after %d iterations.", loopCounter)
	}
	return nil, false, nil // Normal loop completion
}

// executeFor handles FOR EACH, now supporting list and map iteration
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionNode := step.Cond

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing FOR EACH %s IN ...", loopVar)
	}

	if !isValidIdentifier(loopVar) {
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

	evaluatedCollection := i.evaluateExpression(collectionNode)

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        FOR evaluated collection: %v (%T)", evaluatedCollection, evaluatedCollection)
	}

	// Save original value of loop variable (if any) to restore after loop
	originalLoopVarValue, loopVarExists := i.variables[loopVar]
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue
		} else {
			delete(i.variables, loopVar)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Restored loop variable '%s' state.", loopVar)
		}
	}()

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}
	iterations := 0

	switch collection := evaluatedCollection.(type) {

	case []interface{}: // --- Iteration over Slices ---
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
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Exiting slice loop early. Error: %v, Returned: %t", bodyErr, returnedFromBody)
				}
				break
			}
		}

	case map[string]interface{}: // --- Iteration over Maps (Keys) ---
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        FOR iterating over map keys (size %d)...", len(collection))
		}
		keys := make([]string, 0, len(collection))
		for k := range collection {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for itemNum, key := range keys {
			iterations++
			i.variables[loopVar] = key // Assign current KEY
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          Map Key Iter %d: Assigned '%s' = %q", itemNum, loopVar, key)
			}
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Map Key Iter %d", itemNum))
			if bodyErr != nil || returnedFromBody {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Exiting map loop early. Error: %v, Returned: %t", bodyErr, returnedFromBody)
				}
				break
			}
		}

	case string: // --- Iteration over Strings ---
		shouldCommaSplit := false
		trimmedCollection := strings.TrimSpace(collection)
		if strings.Contains(collection, ",") {
			parts := strings.Split(collection, ",")
			if len(parts) > 1 || (len(parts) == 1 && strings.TrimSpace(parts[0]) != trimmedCollection) {
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
				i.variables[loopVar] = trimmedItem // Assign string part
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Comma Iter %d: Assigned '%s' = %q", itemNum, loopVar, trimmedItem)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					if i.logger != nil {
						i.logger.Printf("[DEBUG-INTERP]          Exiting comma loop early. Error: %v, Returned: %t", bodyErr, returnedFromBody)
					}
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
				i.variables[loopVar] = charStr // Assign character string
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Char Iter %d: Assigned '%s' = %q", itemNum, loopVar, charStr)
				}
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum))
				if bodyErr != nil || returnedFromBody {
					if i.logger != nil {
						i.logger.Printf("[DEBUG-INTERP]          Exiting char loop early. Error: %v, Returned: %t", bodyErr, returnedFromBody)
					}
					break
				}
			}
		}

	default: // --- Error Case ---
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Iteration failed: cannot iterate over type %T", evaluatedCollection)
		}
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH", evaluatedCollection)
	} // End switch

	// Handle loop exit reason
	if bodyErr != nil {
		return nil, false, fmt.Errorf("error in FOR EACH loop body: %w", bodyErr)
	}
	if returnedFromBody {
		return resultFromLoop, true, nil // Propagate RETURN
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      FOR EACH loop finished normally after %d iterations.", iterations)
	}
	return nil, false, nil // Normal loop completion
}

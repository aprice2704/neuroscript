package core

import (
	"fmt"
	"strings"
)

// --- Step Execution Helpers ---

func (i *Interpreter) executeSet(step Step, stepNum int) error {
	targetVar := step.Target
	valueExpr, ok := step.Value.(string)
	if !ok {
		return fmt.Errorf("SET value is not string (type %T)", step.Value)
	}
	if !isValidIdentifier(targetVar) {
		return fmt.Errorf("SET target '%s' is not a valid variable name", targetVar)
	}

	// Evaluate the expression first
	finalValue := i.evaluateExpression(valueExpr)

	// Special handling for generated_code to trim fences
	if targetVar == "generated_code" {
		if finalStr, isStr := finalValue.(string); isStr {
			trimmedVal := trimCodeFences(finalStr) // Use utility function
			if trimmedVal != finalStr {
				// fmt.Printf("      [Debug] Trimmed code fences for %s\n", targetVar) // Optional Debug
				finalValue = trimmedVal
			}
		}
	}

	// Store the final value
	i.variables[targetVar] = finalValue
	// fmt.Printf("      [Debug] SET %s = %v (%T)\n", targetVar, finalValue, finalValue) // DEBUG
	return nil
}

func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) {
	target := step.Target
	var callResultValue interface{}
	// fmt.Printf("    Step %d: CALL %s with %d raw args: %v\n", stepNum+1, target, len(step.Args), step.Args) // DEBUG

	// Evaluate all arguments BEFORE making the call
	evaluatedArgs := make([]interface{}, len(step.Args))
	for idx, argExpr := range step.Args {
		evaluatedArgs[idx] = i.evaluateExpression(argExpr)
		// fmt.Printf("      Arg %d evaluated: %v (%T)\n", idx, evaluatedArgs[idx], evaluatedArgs[idx]) // DEBUG
	}

	if strings.HasPrefix(target, "TOOL.") {
		toolName := strings.TrimPrefix(target, "TOOL.")
		toolImpl, found := i.toolRegistry.GetTool(toolName)
		if !found {
			return nil, fmt.Errorf("unknown TOOL '%s'", toolName)
		}

		// Validate and convert arguments based on the tool's spec
		preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs) // Assumes in tools.go or tools_validation.go
		if validationErr != nil {
			return nil, fmt.Errorf("TOOL %s argument error: %w", toolName, validationErr)
		}

		// fmt.Printf("      Calling TOOL.%s with prepared args: %v\n", toolName, preparedArgs) // DEBUG
		var toolErr error
		callResultValue, toolErr = toolImpl.Func(i, preparedArgs) // Pass interpreter and prepared args
		if toolErr != nil {
			return nil, fmt.Errorf("TOOL %s execution failed: %w", toolName, toolErr)
		}
		// fmt.Printf("      TOOL.%s returned: %v (%T)\n", toolName, callResultValue, callResultValue) // DEBUG

	} else if target == "LLM" {
		if len(evaluatedArgs) != 1 {
			return nil, fmt.Errorf("CALL LLM expects 1 prompt arg, got %d", len(evaluatedArgs))
		}
		// Ensure prompt is a string
		prompt := fmt.Sprintf("%v", evaluatedArgs[0])

		// fmt.Printf("      Calling LLM with prompt (first 50 chars): %s\n", truncateString(prompt, 50)) // DEBUG
		response, llmErr := CallLLMAPI(prompt) // Assumes in llm.go
		if llmErr != nil {
			return nil, fmt.Errorf("CALL LLM failed: %w", llmErr)
		}
		callResultValue = response
		// fmt.Printf("      LLM returned (first 50 chars): %s\n", truncateString(response, 50)) // DEBUG

	} else { // Procedure Call
		// Convert evaluated arguments to strings for procedure call signature
		stringArgs := make([]string, len(evaluatedArgs))
		for idx, val := range evaluatedArgs {
			stringArgs[idx] = fmt.Sprintf("%v", val)
		}

		procToCall := target
		// fmt.Printf("      Calling procedure '%s' with string args: %v\n", procToCall, stringArgs) // DEBUG
		procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...) // Recursively call RunProcedure
		if procCallErr != nil {
			return nil, fmt.Errorf("CALL to proc '%s' failed: %w", procToCall, procCallErr)
		}
		callResultValue = procResultValue
		// fmt.Printf("      Procedure '%s' returned: %v (%T)\n", procToCall, callResultValue, callResultValue) // DEBUG
	}

	// Store result (handled by caller in executeSteps)
	return callResultValue, nil
}

// executeIf handles the IF statement
func (i *Interpreter) executeIf(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	conditionStr := step.Cond // Condition is stored directly as string in AST Step
	conditionResult, evalErr := i.evaluateCondition(conditionStr)
	if evalErr != nil {
		// Treat condition evaluation error as false, but log it? Or return error?
		// For now, treat as false and continue (no block execution)
		// fmt.Printf("      [Warning] Step %d: IF condition '%s' evaluation error: %v. Treating as false.\n", stepNum+1, conditionStr, evalErr)
		conditionResult = false
	}

	// fmt.Printf("    Step %d: IF condition '%s' evaluated to %t\n", stepNum+1, conditionStr, conditionResult) // DEBUG

	if conditionResult {
		// fmt.Printf("      Executing IF block...\n") // DEBUG
		blockResult, blockReturned, blockErr := i.executeBlock(step.Value, stepNum, "IF") // Use executeBlock from interpreter.go
		if blockErr != nil {
			return nil, false, blockErr // Propagate error from block
		}
		if blockReturned {
			// fmt.Printf("      RETURN encountered inside IF block. Returning: %v\n", blockResult) // DEBUG
			return blockResult, true, nil // Signal RETURN occurred
		}
		// fmt.Printf("      Finished IF block normally.\n") // DEBUG
		return true, false, nil // Signal true condition ran (no RETURN)
	}

	// fmt.Printf("      Skipping IF block (condition false).\n") // DEBUG
	return nil, false, nil // Condition was false, no error
}

// executeReturn handles the RETURN statement
func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	valueExpr, ok := step.Value.(string)
	if !ok && step.Value != nil { // Allow nil value for RETURN without expression
		return nil, false, fmt.Errorf("RETURN value must be an expression string or nil (type %T)", step.Value)
	}

	var returnValue interface{}
	if valueExpr != "" { // Only evaluate if there's an expression
		returnValue = i.evaluateExpression(valueExpr)
	} else {
		returnValue = nil // RETURN with no expression evaluates to nil
	}

	// fmt.Printf("    Step %d: RETURN encountered. Value: %v (%T)\n", stepNum+1, returnValue, returnValue) // DEBUG
	return returnValue, true, nil // Signal RETURN occurred
}

// --- NEW: EMIT Step Execution ---
// executeEmit handles the EMIT statement for simple output/debugging.
func (i *Interpreter) executeEmit(step Step, stepNum int) error {
	valueExpr, ok := step.Value.(string)
	if !ok {
		// EMIT's value should always be the expression string from the AST builder
		return fmt.Errorf("step %d: EMIT value expected string, got %T", stepNum+1, step.Value)
	}

	evaluatedValue := i.evaluateExpression(valueExpr) // Evaluate the expression

	// Simple print to console for debugging the EMIT functionality
	fmt.Printf("[EMIT] %v\n", evaluatedValue)

	// Decide if EMIT should have a "result" and set i.lastCallResult.
	// For now, let's treat it like SET and not modify lastCallResult.
	// i.lastCallResult = evaluatedValue

	return nil // Indicate success
}

// executeWhile handles the WHILE statement
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	conditionStr := step.Cond
	loopCounter := 0
	maxLoops := 1000 // Safety break

	// fmt.Printf("    Step %d: WHILE %s DO\n", stepNum+1, conditionStr) // DEBUG

	for loopCounter < maxLoops {
		conditionResult, evalErr := i.evaluateCondition(conditionStr)
		if evalErr != nil {
			// fmt.Printf("      [Error] WHILE condition evaluation error (iter %d): %v\n", loopCounter, evalErr) // DEBUG
			return nil, false, evalErr // Propagate condition error
		}

		// fmt.Printf("      WHILE condition '%s' (iter %d) evaluated to %t\n", conditionStr, loopCounter, conditionResult) // DEBUG
		if !conditionResult {
			break // Exit loop if condition is false
		}

		// fmt.Printf("      Executing WHILE block (iter %d)...\n", loopCounter) // DEBUG
		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter)) // Use executeBlock from interpreter.go
		if bodyErr != nil {
			return nil, false, fmt.Errorf("error in WHILE loop body (iter %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			// fmt.Printf("      RETURN encountered inside WHILE block (iter %d). Returning: %v\n", loopCounter, bodyResult) // DEBUG
			return bodyResult, true, nil // Propagate RETURN
		}
		loopCounter++
	}

	if loopCounter >= maxLoops {
		return nil, false, fmt.Errorf("WHILE loop exceeded max iterations (%d)", maxLoops)
	}

	// fmt.Printf("      Finished WHILE loop normally after %d iterations.\n", loopCounter) // DEBUG
	return nil, false, nil // Normal loop completion
}

// executeFor handles the FOR EACH statement
// ** FIX: Removed exitLoop flag, using standard break **
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionExpr := step.Cond // Collection expression is stored in Cond field for FOR steps
	// fmt.Printf("    Step %d: FOR EACH %s IN %s DO\n", stepNum+1, loopVar, collectionExpr) // DEBUG

	if !isValidIdentifier(loopVar) {
		return nil, false, fmt.Errorf("FOR loop variable '%s' is not a valid identifier", loopVar)
	}

	evaluatedCollection := i.evaluateExpression(collectionExpr)
	// fmt.Printf("      Evaluated collection: %v (%T)\n", evaluatedCollection, evaluatedCollection) // DEBUG

	// Save original value of loop variable (if any) to restore after loop
	originalLoopVarValue, loopVarExists := i.variables[loopVar]
	// Ensure variable is restored even if errors occur or RETURN happens
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue
		} else {
			delete(i.variables, loopVar)
		}
		// fmt.Printf("      Restored loop variable '%s' state.\n", loopVar) // DEBUG
	}()

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}
	iterations := 0

	switch collection := evaluatedCollection.(type) {
	// --- Iteration over Slices (requires interpreter support for list literals/results) ---
	case []string:
		// fmt.Printf("      Iterating over string slice (len %d)...\n", len(collection)) // DEBUG
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item // Assign current item to loop variable
			// fmt.Printf("        Iter %d: SET %s = %q\n", itemNum, loopVar, item) // DEBUG
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR String Slice Iter %d", itemNum)) // Use executeBlock from interpreter.go
			if bodyErr != nil || returnedFromBody {
				break
			} // Exit inner loop on error or RETURN
		}
	case []interface{}: // Handle slices returned by tools more generically
		// fmt.Printf("      Iterating over interface slice (len %d)...\n", len(collection)) // DEBUG
		for itemNum, item := range collection {
			iterations++
			i.variables[loopVar] = item // Assign current item to loop variable
			// fmt.Printf("        Iter %d: SET %s = %v (%T)\n", itemNum, loopVar, item, item) // DEBUG
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Interface Slice Iter %d", itemNum)) // Use executeBlock from interpreter.go
			if bodyErr != nil || returnedFromBody {
				break
			} // Exit inner loop on error or RETURN
		}
	// --- Iteration over Strings ---
	case string:
		// Determine if it's comma-separated or character iteration
		// Treat as comma-separated if it contains ',' AND splitting yields more than one part
		// OR if splitting yields one part that is different after trimming (e.g., " a ,")
		shouldCommaSplit := false
		trimmedCollection := strings.TrimSpace(collection)
		if strings.Contains(collection, ",") {
			parts := strings.Split(collection, ",")
			if len(parts) > 1 {
				shouldCommaSplit = true
			} else if len(parts) == 1 && strings.TrimSpace(parts[0]) != trimmedCollection {
				// Handles cases like "a," or ",a" or " a , "
				shouldCommaSplit = true
			}
		}

		if shouldCommaSplit {
			// fmt.Printf("      Iterating over comma-separated string...\n") // DEBUG
			items := strings.Split(collection, ",")
			for itemNum, item := range items {
				iterations++
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem
				// fmt.Printf("        Iter %d: SET %s = %q\n", itemNum, loopVar, trimmedItem) // DEBUG
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum)) // Use executeBlock from interpreter.go
				if bodyErr != nil || returnedFromBody {
					break
				} // Exit inner loop on error or RETURN
			}
		} else { // Character Iteration
			// fmt.Printf("      Iterating over string characters...\n") // DEBUG
			for itemNum, charRune := range collection { // Iterates over runes correctly
				iterations++
				charStr := string(charRune)
				i.variables[loopVar] = charStr
				// fmt.Printf("        Iter %d: SET %s = %q\n", itemNum, loopVar, charStr) // DEBUG
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum)) // Use executeBlock from interpreter.go
				if bodyErr != nil || returnedFromBody {
					break
				} // Exit inner loop on error or RETURN
			}
		}
	default:
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH", evaluatedCollection)
	} // End switch

	// Handle loop exit reason (error or return from body) outside the switch
	if bodyErr != nil {
		return nil, false, fmt.Errorf("error in FOR EACH loop body: %w", bodyErr)
	}
	if returnedFromBody {
		// fmt.Printf("      RETURN encountered inside FOR EACH block. Returning: %v\n", resultFromLoop) // DEBUG
		return resultFromLoop, true, nil // Propagate RETURN
	}

	// fmt.Printf("      Finished FOR EACH loop normally after %d iterations.\n", iterations) // DEBUG
	return nil, false, nil // Normal loop completion
}

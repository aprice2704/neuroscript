package core

import (
	"fmt"
	"os"
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

	finalValue := i.evaluateExpression(valueExpr)

	if targetVar == "generated_code" {
		if finalStr, isStr := finalValue.(string); isStr {
			trimmedVal := trimCodeFences(finalStr)
			if trimmedVal != finalStr {
				finalValue = trimmedVal
			}
		}
	}
	i.variables[targetVar] = finalValue
	return nil
}

func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) {
	target := step.Target
	var callResultValue interface{}

	evaluatedArgs := make([]interface{}, len(step.Args))
	for idx, argExpr := range step.Args {
		evaluatedArgs[idx] = i.evaluateExpression(argExpr)
	}

	if strings.HasPrefix(target, "TOOL.") {
		toolName := strings.TrimPrefix(target, "TOOL.")
		toolImpl, found := i.toolRegistry.GetTool(toolName)
		if !found {
			return nil, fmt.Errorf("unknown TOOL '%s'", toolName)
		}
		preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs) // Assumes in tools.go or tools_validation.go
		if validationErr != nil {
			return nil, fmt.Errorf("TOOL %s argument error: %w", toolName, validationErr)
		}
		var toolErr error
		callResultValue, toolErr = toolImpl.Func(i, preparedArgs)
		if toolErr != nil {
			return nil, fmt.Errorf("TOOL %s execution failed: %w", toolName, toolErr)
		}
	} else if target == "LLM" {
		if len(evaluatedArgs) != 1 {
			return nil, fmt.Errorf("CALL LLM expects 1 prompt arg, got %d", len(evaluatedArgs))
		}
		prompt := fmt.Sprintf("%v", evaluatedArgs[0])
		response, llmErr := CallLLMAPI(prompt) // Assumes in llm.go
		if llmErr != nil {
			return nil, fmt.Errorf("CALL LLM failed: %w", llmErr)
		}
		callResultValue = response
	} else { // Procedure Call
		stringArgs := make([]string, len(evaluatedArgs))
		for idx, val := range evaluatedArgs {
			stringArgs[idx] = fmt.Sprintf("%v", val)
		}
		procToCall := target
		procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...)
		if procCallErr != nil {
			return nil, fmt.Errorf("CALL to proc '%s' failed: %w", procToCall, procCallErr)
		}
		callResultValue = procResultValue
	}
	return callResultValue, nil
}

func (i *Interpreter) executeIf(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	conditionStr := step.Cond
	conditionResult, evalErr := i.evaluateCondition(conditionStr)
	if evalErr != nil {
		conditionResult = false
	} // Treat condition error as false

	if conditionResult {
		blockResult, blockReturned, blockErr := i.executeBlock(step.Value, stepNum, "IF") // Use executeBlock from interpreter.go
		if blockErr != nil {
			return nil, false, blockErr
		}
		if blockReturned {
			return blockResult, true, nil
		}
		return true, false, nil // Signal true condition ran
	}
	return nil, false, nil // Condition was false
}

func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	valueExpr, ok := step.Value.(string)
	if !ok {
		return nil, false, fmt.Errorf("RETURN value must be an expression string (type %T)", step.Value)
	}
	returnValue := i.evaluateExpression(valueExpr)
	return returnValue, true, nil
}

// executeWhile handles the WHILE statement
func (i *Interpreter) executeWhile(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	conditionStr := step.Cond
	loopCounter := 0
	maxLoops := 1000

	for loopCounter < maxLoops {
		conditionResult, evalErr := i.evaluateCondition(conditionStr)
		if evalErr != nil {
			return nil, false, evalErr
		}
		if !conditionResult {
			break
		}

		bodyResult, bodyReturned, bodyErr := i.executeBlock(step.Value, stepNum, fmt.Sprintf("WHILE Iter %d", loopCounter)) // Use executeBlock from interpreter.go
		if bodyErr != nil {
			return nil, false, fmt.Errorf("error in WHILE loop body (iter %d): %w", loopCounter, bodyErr)
		}
		if bodyReturned {
			return bodyResult, true, nil
		} // Propagate RETURN
		loopCounter++
	}

	if loopCounter >= maxLoops {
		return nil, false, fmt.Errorf("WHILE loop exceeded max iterations (%d)", maxLoops)
	}
	return nil, false, nil // Normal loop completion
}

// executeFor handles the FOR EACH statement
// ** FIX: Removed exitLoop flag **
func (i *Interpreter) executeFor(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	loopVar := step.Target
	collectionExpr := step.Cond
	evaluatedCollection := i.evaluateExpression(collectionExpr)

	_, cwdErr := os.Getwd()
	if cwdErr != nil {
		return nil, false, fmt.Errorf("FOR loop cwd error: %w", cwdErr)
	}
	originalLoopVarValue, loopVarExists := i.variables[loopVar]
	defer func() {
		if loopVarExists {
			i.variables[loopVar] = originalLoopVarValue
		} else {
			delete(i.variables, loopVar)
		}
	}()

	var bodyErr error
	var returnedFromBody bool
	var resultFromLoop interface{}

	switch collection := evaluatedCollection.(type) {
	case []string:
		for itemNum, item := range collection {
			i.variables[loopVar] = item
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR String Slice Iter %d", itemNum)) // Use executeBlock from interpreter.go
			// *** FIX: Check error/return immediately and break inner loop ***
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case []interface{}:
		for itemNum, item := range collection {
			i.variables[loopVar] = item
			resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Interface Slice Iter %d", itemNum)) // Use executeBlock from interpreter.go
			// *** FIX: Check error/return immediately and break inner loop ***
			if bodyErr != nil || returnedFromBody {
				break
			}
		}
	case string:
		shouldCommaSplit := false
		if strings.Contains(collection, ",") {
			parts := strings.Split(collection, ",")
			if len(parts) > 1 || (len(parts) == 1 && strings.TrimSpace(parts[0]) != collection) {
				shouldCommaSplit = true
			}
		}
		if shouldCommaSplit {
			items := strings.Split(collection, ",")
			for itemNum, item := range items {
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Comma Iter %d", itemNum)) // Use executeBlock from interpreter.go
				// *** FIX: Check error/return immediately and break inner loop ***
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		} else { // Character Iteration
			for itemNum, charRune := range collection {
				charStr := string(charRune)
				i.variables[loopVar] = charStr
				resultFromLoop, returnedFromBody, bodyErr = i.executeBlock(step.Value, stepNum, fmt.Sprintf("FOR Char Iter %d", itemNum)) // Use executeBlock from interpreter.go
				// *** FIX: Check error/return immediately and break inner loop ***
				if bodyErr != nil || returnedFromBody {
					break
				}
			}
		}
	default:
		bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH", evaluatedCollection)
		// No need for exitLoop flag here, error is handled below
	} // End switch

	// Handle loop exit reason (error or return from body)
	if bodyErr != nil {
		return nil, false, fmt.Errorf("error in FOR EACH loop body: %w", bodyErr)
	}
	if returnedFromBody {
		return resultFromLoop, true, nil
	} // Propagate RETURN
	return nil, false, nil // Normal loop completion
}

// --- REMOVED duplicate executeBlock ---

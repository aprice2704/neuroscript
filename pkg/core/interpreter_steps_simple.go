// NeuroScript Version: 0.3.1
// File version: 0.0.7 // Implemented truthy/falsy check for MUST statements.
// Purpose: Defines interpreter execution for simple (non-control-flow) steps.
// filename: pkg/core/interpreter_simple_steps.go
// nlines: 390 // Approximate
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt" // For isTruthy check on maps/slices
	"strings"
)

// toInt64Coerce attempts to convert an interface{} to int64.
// This is a helper for list indexing.
func toInt64Coerce(val interface{}) (int64, bool) {
	// ... (content as previously provided) ...
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float32:
		if float32(int64(v)) == v { // Check if conversion is exact
			return int64(v), true
		}
	case float64:
		if float64(int64(v)) == v { // Check if conversion is exact
			return int64(v), true
		}
	}
	return 0, false
}

// executeSet handles the "set" step, now supporting simple and complex lvalues.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	// ... (content as previously provided and corrected) ...
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}

	if step.LValue == nil {
		i.Logger().Error("[DEBUG-INTERP] CRITICAL: executeSet called with nil LValue.", "pos", posStr)
		return nil, NewRuntimeError(ErrorCodeInternal, "SetStep LValue is nil in executeSet", nil).WithPosition(step.Pos)
	}

	baseVarName := step.LValue.Identifier
	i.Logger().Debug("[DEBUG-INTERP] Executing SET", "LValue", step.LValue.String(), "pos", posStr)

	rhsValue, evalErr := i.evaluateExpression(step.Value)
	if evalErr != nil {
		errMsg := fmt.Sprintf("evaluating value for SET %s", step.LValue.String())
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), errMsg)
	}

	if isInHandler && (baseVarName == "err_code" || baseVarName == "err_msg") && len(step.LValue.Accessors) == 0 {
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler", baseVarName)
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, ErrReadOnlyViolation).WithPosition(step.Pos)
	}

	if len(step.LValue.Accessors) == 0 {
		if err := i.SetVariable(baseVarName, rhsValue); err != nil {
			errMsg := fmt.Sprintf("setting variable '%s'", baseVarName)
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeInternal, errMsg, err)
			}
			return nil, WrapErrorWithPosition(err, step.Pos, "")
		}
		i.Logger().Debug("[DEBUG-INTERP] Set simple variable", "variable", baseVarName, "value", rhsValue)
		return rhsValue, nil
	}

	currentVal, varExists := i.GetVariable(baseVarName)
	if !varExists {
		if len(step.LValue.Accessors) > 0 {
			i.Logger().Debugf("Interpreter: Base variable '%s' for complex set not found, auto-creating as map.", baseVarName)
			newMap := make(map[string]interface{})
			if err := i.SetVariable(baseVarName, newMap); err != nil {
				return nil, WrapErrorWithPosition(err, step.LValue.Pos, fmt.Sprintf("auto-creating map for '%s'", baseVarName))
			}
			currentVal = newMap
		} else {
			return nil, NewRuntimeError(ErrorCodeInternal, "internal logic error in executeSet for LValue handling", nil).WithPosition(step.LValue.Pos)
		}
	}

	for accessorIdx, accessor := range step.LValue.Accessors {
		isFinalAccessor := accessorIdx == len(step.LValue.Accessors)-1

		switch accessor.Type {
		case BracketAccess:
			indexOrKeyVal, errIdx := i.evaluateExpression(accessor.IndexOrKey)
			if errIdx != nil {
				return nil, WrapErrorWithPosition(errIdx, accessor.Pos, fmt.Sprintf("evaluating index/key for '%s'", baseVarName))
			}

			switch collection := currentVal.(type) {
			case map[string]interface{}:
				keyStr, ok := indexOrKeyVal.(string)
				if !ok {
					return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("map key for '%s' must be a string, got %T (%v)", baseVarName, indexOrKeyVal, indexOrKeyVal), nil).WithPosition(accessor.Pos)
				}
				if isFinalAccessor {
					collection[keyStr] = rhsValue
					i.Logger().Debugf("Interpreter: Set map key '%s' for '%s' to: %v", keyStr, baseVarName, rhsValue)
					return rhsValue, nil
				}
				if nextVal, found := collection[keyStr]; found {
					mapVal, isMap := nextVal.(map[string]interface{})
					listVal, isList := nextVal.([]interface{})
					if isMap {
						currentVal = mapVal
					} else if isList {
						currentVal = listVal
					} else {
						i.Logger().Debugf("Overwriting non-collection at '%s' with new map for further assignment in '%s'", keyStr, baseVarName)
						newMap := make(map[string]interface{})
						collection[keyStr] = newMap
						currentVal = newMap
					}
				} else {
					newMap := make(map[string]interface{})
					collection[keyStr] = newMap
					currentVal = newMap
					i.Logger().Debugf("Interpreter: Auto-created nested map for key '%s' in '%s'", keyStr, baseVarName)
				}

			case []interface{}:
				index, ok := toInt64Coerce(indexOrKeyVal)
				if !ok {
					return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("list index for '%s' must be an integer, got %T (%v)", baseVarName, indexOrKeyVal, indexOrKeyVal), nil).WithPosition(accessor.Pos)
				}
				if index < 0 {
					return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("list index for '%s' cannot be negative, got %d", baseVarName, index), nil).WithPosition(accessor.Pos)
				}

				tempSlice := collection
				if isFinalAccessor {
					for int64(len(tempSlice)) <= index {
						tempSlice = append(tempSlice, nil)
					}
					tempSlice[index] = rhsValue
					if err := i.SetVariable(baseVarName, tempSlice); err != nil {
						return nil, WrapErrorWithPosition(err, step.Pos, fmt.Sprintf("updating list variable '%s' after indexed set", baseVarName))
					}
					i.Logger().Debugf("Interpreter: Set list '%s' index %d to: %v", baseVarName, index, rhsValue)
					return rhsValue, nil
				}
				if index >= int64(len(tempSlice)) {
					return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("intermediate list index %d out of bounds for '%s' (len %d)", index, baseVarName, len(tempSlice)), nil).WithPosition(accessor.Pos)
				}
				currentVal = tempSlice[index]

			default:
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("variable '%s' is not a map or list for bracket assignment, it's a %T", baseVarName, currentVal), nil).WithPosition(accessor.Pos)
			}

		case DotAccess:
			mapCollection, ok := currentVal.(map[string]interface{})
			if !ok {
				i.Logger().Debugf("Interpreter: Path for dot access on '%s' is not a map (is %T), auto-creating as map.", baseVarName, currentVal)
				newMap := make(map[string]interface{})
				currentVal = newMap
				mapCollection = newMap
				if accessorIdx == 0 {
					if err := i.SetVariable(baseVarName, newMap); err != nil {
						return nil, WrapErrorWithPosition(err, step.LValue.GetPos(), fmt.Sprintf("auto-creating map for dot access on '%s'", baseVarName))
					}
				}
			}

			fieldName := accessor.FieldName
			if isFinalAccessor {
				mapCollection[fieldName] = rhsValue
				i.Logger().Debugf("Interpreter: Set map field '%s' for '%s' to: %v", fieldName, baseVarName, rhsValue)
				return rhsValue, nil
			}
			if nextVal, found := mapCollection[fieldName]; found {
				mapVal, isMap := nextVal.(map[string]interface{})
				listVal, isList := nextVal.([]interface{})
				if isMap {
					currentVal = mapVal
				} else if isList {
					currentVal = listVal
				} else {
					i.Logger().Debugf("Overwriting non-collection at field '%s' with new map for further assignment in '%s'", fieldName, baseVarName)
					newMap := make(map[string]interface{})
					mapCollection[fieldName] = newMap
					currentVal = newMap
				}
			} else {
				newMap := make(map[string]interface{})
				mapCollection[fieldName] = newMap
				currentVal = newMap
				i.Logger().Debugf("Interpreter: Auto-created nested map for field '%s' in '%s'", fieldName, baseVarName)
			}
		default:
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("unknown accessor type %v for '%s'", accessor.Type, baseVarName), nil).WithPosition(accessor.Pos)
		}
	}
	i.Logger().Error("Interpreter: executeSet complex LValue assignment logic fell through without assignment.", "lvalue", step.LValue.String())
	return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("indexed/field assignment did not complete for '%s'", step.LValue.String()), nil).WithPosition(step.Pos)
}

// executeReturn handles the "return" step.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	// ... (content as previously provided) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing RETURN", "pos", posStr)

	if len(step.Values) > 0 {
		i.Logger().Debug("[DEBUG-INTERP] Return has multiple expressions", "count", len(step.Values), "pos", posStr)
		results := make([]interface{}, len(step.Values))
		for idx, exprNode := range step.Values {
			evaluatedValue, err := i.evaluateExpression(exprNode)
			if err != nil {
				exprPosStr := "<unknown>"
				if exprNode != nil {
					nodePos := exprNode.GetPos()
					if nodePos != nil {
						exprPosStr = nodePos.String()
					}
				}
				errMsg := fmt.Sprintf("evaluating return expression %d at %s", idx+1, exprPosStr)
				return nil, true, WrapErrorWithPosition(err, exprNode.GetPos(), errMsg)
			}
			results[idx] = evaluatedValue
		}
		return results, true, nil
	}

	if step.Value != nil {
		i.Logger().Debug("[DEBUG-INTERP] Return has a single expression", "pos", posStr)
		evaluatedValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			exprPosStr := "<unknown>"
			if step.Value != nil {
				nodePos := step.Value.GetPos()
				if nodePos != nil {
					exprPosStr = nodePos.String()
				}
			}
			errMsg := fmt.Sprintf("evaluating return expression at %s", exprPosStr)
			return nil, true, WrapErrorWithPosition(err, step.Value.GetPos(), errMsg)
		}
		return evaluatedValue, true, nil
	}

	i.Logger().Debug("[DEBUG-INTERP] Return has no value (implicit nil)", "pos", posStr)
	return nil, true, nil
}

// executeEmit handles the "emit" step.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	// ... (content as previously provided and corrected) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing EMIT", "pos", posStr)

	var valueToEmit interface{}
	var evalErr error

	if step.Value != nil {
		valueToEmit, evalErr = i.evaluateExpression(step.Value)
	} else {
		return nil, NewRuntimeError(ErrorCodeSyntax, "EMIT statement requires an expression", nil).WithPosition(step.Pos)
	}

	if evalErr != nil {
		errMsg := fmt.Sprintf("evaluating value for EMIT at %s", posStr)
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), errMsg)
	}

	formattedOutput := fmt.Sprintf("%v", valueToEmit)

	if i.stdout == nil {
		i.Logger().Error("executeEmit: Interpreter stdout is nil! This is a critical setup error. Falling back to os.Stdout.")
		fmt.Println(formattedOutput)
	} else {
		if _, err := fmt.Fprintln(i.stdout, formattedOutput); err != nil {
			i.Logger().Error("Failed to write EMIT output via i.stdout", "error", err)
			return nil, NewRuntimeError(ErrorCodeIOFailed, "failed to emit output", err).WithPosition(step.Pos)
		}
	}
	return valueToEmit, nil
}

// executeMust handles "must" and "mustbe" steps.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type)
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing MUST/MUSTBE", "type", strings.ToUpper(stepType), "pos", posStr)

	var conditionMet bool
	var valueEvaluated interface{}
	var evalErr error                     // Stores error from direct evaluation or type assertion for mustbe
	var errorNodePos *Position = step.Pos // Default position

	if stepType == "must" {
		if step.Value == nil {
			return nil, NewRuntimeError(ErrorCodeSyntax, "must step has nil condition expression", nil).WithPosition(step.Pos)
		}
		errorNodePos = step.Value.GetPos()
		var err error
		valueEvaluated, err = i.evaluateExpression(step.Value)
		if err != nil {
			// Error during evaluation of the 'must' condition expression itself.
			return nil, WrapErrorWithPosition(err, errorNodePos, "evaluating condition for must")
		}
		conditionMet = isTruthy(valueEvaluated) // Use truthy check for "must"
		if !conditionMet {
			evalErr = ErrMustConditionFailed // Store this to indicate the nature of failure
		}
	} else if stepType == "mustbe" {
		if step.Call == nil {
			errorNodePos = step.Pos
			return nil, NewRuntimeError(ErrorCodeSyntax, "mustbe step has nil callable expression", nil).WithPosition(errorNodePos)
		}
		errorNodePos = step.Call.GetPos()
		var errCall error
		valueEvaluated, errCall = i.evaluateExpression(step.Call) // step.Call is *CallableExprNode

		if errCall != nil {
			// Error executing the check function (e.g., not found, wrong args, runtime error in func).
			// This should become ErrMustConditionFailed, wrapping the original error.
			evalErr = fmt.Errorf("%w: check function %s call failed: %w", ErrMustConditionFailed, step.Call.Target.String(), errCall)
			conditionMet = false
		} else {
			// Check function executed, now check if its result is a boolean true.
			boolVal, ok := valueEvaluated.(bool) // MUSTBE requires a strict boolean from its check function.
			if !ok {
				typeErrMessage := fmt.Sprintf("mustbe check function %s did not return a boolean, got %T", step.Call.Target.String(), valueEvaluated)
				evalErr = fmt.Errorf("%w: %s", ErrMustConditionFailed, typeErrMessage)
				conditionMet = false
			} else {
				conditionMet = boolVal
				if !conditionMet { // Successfully returned false
					evalErr = fmt.Errorf("%w: check function %s returned false", ErrMustConditionFailed, step.Call.Target.String())
				}
			}
		}
	} else {
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("executeMust called with invalid step type: %s", step.Type), ErrInternal).WithPosition(step.Pos)
	}

	// If the condition was not met (either by being falsy for 'must', or any failure/false for 'mustbe')
	if !conditionMet {
		// If evalErr is already set (e.g., from mustbe logic, or must's falsiness), use it.
		// It will already be (or wrap) ErrMustConditionFailed.
		if evalErr != nil {
			// Ensure it's a RuntimeError with the correct code etc.
			if _, ok := evalErr.(*RuntimeError); ok {
				// If it's already a runtime error that wraps ErrMustConditionFailed, good.
				// If not, we might need to re-wrap here, but evalErr for mustbe should already be correctly formatted.
				// For must, evalErr is set to ErrMustConditionFailed if isTruthy was false.
				if errors.Is(evalErr, ErrMustConditionFailed) {
					// Construct a new RuntimeError if it's just the sentinel, or use existing if already detailed
					if evalErr == ErrMustConditionFailed { // raw sentinel
						detailMsg := fmt.Sprintf("must condition evaluated to false (value was %T: %v)", valueEvaluated, valueEvaluated)
						return nil, NewRuntimeError(ErrorCodeMustFailed, detailMsg, ErrMustConditionFailed).WithPosition(errorNodePos)
					}
					// If evalErr is already a fmt.Errorf wrapping ErrMustConditionFailed (from mustbe's call error)
					return nil, NewRuntimeError(ErrorCodeMustFailed, evalErr.Error(), evalErr).WithPosition(errorNodePos)
				}
				// If it's some other RuntimeError not wrapping ErrMustConditionFailed (shouldn't happen for failed condition)
				return nil, evalErr // Should be positioned.
			}
			// If evalErr is a plain error wrapping ErrMustConditionFailed (e.g. from mustbe)
			return nil, NewRuntimeError(ErrorCodeMustFailed, evalErr.Error(), evalErr).WithPosition(errorNodePos)

		}
		// This case should ideally not be reached if evalErr is always set when !conditionMet
		// due to the above logic (e.g. must will set evalErr = ErrMustConditionFailed if !isTruthy)
		// But as a fallback:
		return nil, NewRuntimeError(ErrorCodeMustFailed, "must condition failed", ErrMustConditionFailed).WithPosition(errorNodePos)
	}

	i.Logger().Debug("[DEBUG-INTERP] MUST/MUSTBE condition TRUE", "type", strings.ToUpper(stepType), "pos", posStr)
	return valueEvaluated, nil
}

// executeFail handles the "fail" step.
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	// ... (content as previously provided and corrected) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing FAIL", "pos", posStr)
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var wrappedErr error = ErrFailStatement
	var finalPos = step.Pos

	if step.Value != nil {
		finalPos = step.Value.GetPos()
		failValue, err := i.evaluateExpression(step.Value)
		if err != nil {
			evalFailMsg := fmt.Sprintf("error evaluating message/code for FAIL statement: %s", err.Error())
			return NewRuntimeError(errCode, evalFailMsg, err).WithPosition(finalPos)
		}

		switch v := failValue.(type) {
		case string:
			errMsg = v
		case int:
			errCode = ErrorCode(v)
			errMsg = fmt.Sprintf("fail with code %d", errCode)
		case int64:
			errCode = ErrorCode(int(v))
			errMsg = fmt.Sprintf("fail with code %d", errCode)
		case float64:
			errCode = ErrorCode(int(v))
			errMsg = fmt.Sprintf("fail with code %d (from float %v)", errCode, v)
		default:
			errMsg = fmt.Sprintf("%v", failValue)
		}
	}
	return NewRuntimeError(errCode, errMsg, wrappedErr).WithPosition(finalPos)
}

// executeOnError handles the "on_error" step setup.
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	// ... (content as previously provided) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing ON_ERROR - Handler now active.", "pos", posStr)
	handlerStep := step
	return &handlerStep, nil
}

// executeClearError handles the "clear_error" step.
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	// ... (content as previously provided) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CLEAR_ERROR", "pos", posStr)
	if !isInHandler {
		errMsg := fmt.Sprintf("'clear_error' can only be used inside an on_error block")
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, ErrClearViolation).WithPosition(step.Pos)
	}
	return true, nil
}

// executeAsk handles the "ask" step.
func (i *Interpreter) executeAsk(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	// ... (content as previously provided and corrected) ...
	posStr := step.Pos.String()
	targetVar := step.AskIntoVar
	i.Logger().Debug("[DEBUG-INTERP] Executing ASK", "pos", posStr, "target_var", targetVar)

	if step.Value == nil {
		return nil, NewRuntimeError(ErrorCodeSyntax, "ASK step has nil Value field for prompt", nil).WithPosition(step.Pos)
	}
	promptValue, err := i.evaluateExpression(step.Value)
	if err != nil {
		errMsg := fmt.Sprintf("evaluating prompt for ASK")
		return nil, WrapErrorWithPosition(err, step.Value.GetPos(), errMsg)
	}

	promptStr, ok := promptValue.(string)
	if !ok {
		errMsg := fmt.Sprintf("prompt for ASK evaluated to non-string type (%T)", promptValue)
		return nil, NewRuntimeError(ErrorCodeType, errMsg, ErrInvalidOperandTypeString).WithPosition(step.Value.GetPos())
	}

	if i.llmClient == nil {
		i.Logger().Error("ASK step: LLM client not configured in interpreter.", "pos", posStr)
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM client not configured", ErrLLMNotConfigured).WithPosition(step.Pos)
	}

	llmResult := fmt.Sprintf("LLM Response placeholder for: %s", promptStr)
	var llmErr error = nil

	if llmErr != nil {
		errMsg := fmt.Sprintf("LLM interaction failed for ASK: %s", llmErr.Error())
		return nil, NewRuntimeError(ErrorCodeLLMError, errMsg, llmErr).WithPosition(step.Pos)
	}

	if targetVar != "" {
		if setErr := i.SetVariable(targetVar, llmResult); setErr != nil {
			errMsg := fmt.Sprintf("setting variable '%s' for ASK result", targetVar)
			if _, ok := setErr.(*RuntimeError); !ok {
				setErr = NewRuntimeError(ErrorCodeInternal, errMsg, setErr)
			}
			return nil, WrapErrorWithPosition(setErr, step.Pos, "")
		}
		i.Logger().Debug("[DEBUG-INTERP] Stored ASK result in variable", "variable", targetVar)
	}

	return llmResult, nil
}

// executeBreak handles the "break" step by returning ErrBreak.
func (i *Interpreter) executeBreak(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	// ... (content as previously provided) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing BREAK", "pos", posStr)
	return nil, ErrBreak
}

// executeContinue handles the "continue" step by returning ErrContinue.
func (i *Interpreter) executeContinue(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	// ... (content as previously provided) ...
	posStr := step.Pos.String()
	i.Logger().Debug("[DEBUG-INTERP] Executing CONTINUE", "pos", posStr)
	return nil, ErrContinue
}

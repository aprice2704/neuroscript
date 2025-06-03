// NeuroScript Version: 0.3.1
// File version: 0.0.9 (Corrected parent updates for list reallocation, fixed unused vars)
// Purpose: Defines interpreter execution for assignment ("set") steps.
// filename: pkg/core/interpreter_assignment.go
// nlines: 275 // Approximate
// risk_rating: MEDIUM

package core

import (
	"fmt"
)

// toInt64Coerce attempts to convert an interface{} to int64.
func toInt64Coerce(val interface{}) (int64, bool) {
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
		if float32(int64(v)) == v {
			return int64(v), true
		}
	case float64:
		if float64(int64(v)) == v {
			return int64(v), true
		}
	}
	return 0, false
}

// isCollection checks if an interface is a map or a slice.
func isCollection(v interface{}) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

// isList checks if an interface is a slice of interfaces.
func isList(v interface{}) bool {
	if v == nil {
		return false
	}
	_, ok := v.([]interface{})
	return ok
}

// isMap checks if an interface is a map of string to interface.
func isMap(v interface{}) bool {
	if v == nil {
		return false
	}
	_, ok := v.(map[string]interface{})
	return ok
}

// Ternary helper
func Ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// processListElementAssignment handles list-specific logic for 'set' operations.
// Returns: the (potentially new) list, the next value to work on, whether the list header itself changed, and any error.
func (i *Interpreter) processListElementAssignment(
	listBeingModified []interface{},
	indexVal interface{},
	isFinalAccessor bool,
	rhsValue interface{},
	createNextAsList bool,
	accessorPos *Position,
	baseVarNameForLog string,
) (updatedList []interface{}, nextCurrentValue interface{}, listHeaderDidChange bool, err *RuntimeError) {

	index, indexIsIntLike := toInt64Coerce(indexVal)
	if !indexIsIntLike {
		return listBeingModified, nil, false, NewRuntimeError(ErrorCodeType, fmt.Sprintf("list index for '%s' path must be an integer, got %T (%v)", baseVarNameForLog, indexVal, indexVal), ErrListInvalidIndexType).WithPosition(accessorPos)
	}
	if index < 0 {
		return listBeingModified, nil, false, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("list index for '%s' path cannot be negative, got %d", baseVarNameForLog, index), ErrListIndexOutOfBounds).WithPosition(accessorPos)
	}

	headerChanged := false

	if isFinalAccessor {
		originalLen := len(listBeingModified)
		for int64(len(listBeingModified)) <= index {
			listBeingModified = append(listBeingModified, nil)
		}
		if len(listBeingModified) > originalLen {
			headerChanged = true
		}
		listBeingModified[index] = rhsValue
		return listBeingModified, rhsValue, headerChanged, nil
	}

	originalLen := len(listBeingModified)
	isOutOfBounds := index >= int64(len(listBeingModified))
	var elementAtAutocreateIndex interface{}
	if !isOutOfBounds {
		elementAtAutocreateIndex = listBeingModified[index]
	}

	needsCreationOrOverwrite := isOutOfBounds || elementAtAutocreateIndex == nil ||
		!isCollection(elementAtAutocreateIndex) ||
		(createNextAsList && !isList(elementAtAutocreateIndex)) ||
		(!createNextAsList && !isMap(elementAtAutocreateIndex))

	if needsCreationOrOverwrite {
		for int64(len(listBeingModified)) <= index {
			listBeingModified = append(listBeingModified, nil)
		}
		if len(listBeingModified) > originalLen {
			headerChanged = true
		}

		var newCollection interface{}
		logAction := Ternary(isOutOfBounds || elementAtAutocreateIndex == nil, "Auto-creating", "Overwriting")
		logType := Ternary(createNextAsList, "LIST", "MAP")
		i.Logger().Debugf("Interpreter: %s nested %s at index %d in list path of '%s'", logAction, logType, index, baseVarNameForLog)

		if createNextAsList {
			newCollection = make([]interface{}, 0)
		} else {
			newCollection = make(map[string]interface{})
		}

		listBeingModified[index] = newCollection
		return listBeingModified, newCollection, headerChanged, nil
	}
	return listBeingModified, elementAtAutocreateIndex, headerChanged, nil
}

// executeSet handles the "set" step.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	if step.LValue == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "SetStep LValue is nil", nil).WithPosition(step.Pos)
	}
	baseVarName := step.LValue.Identifier
	rhsValue, evalErr := i.evaluateExpression(step.Value)
	if evalErr != nil {
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), fmt.Sprintf("evaluating value for SET %s", baseVarName))
	}

	if isInHandler && (baseVarName == "err_code" || baseVarName == "err_msg") && len(step.LValue.Accessors) == 0 {
		return nil, NewRuntimeError(ErrorCodeReadOnly, fmt.Sprintf("cannot assign to read-only var '%s' in handler", baseVarName), ErrReadOnlyViolation).WithPosition(step.Pos)
	}

	if len(step.LValue.Accessors) == 0 { // Simple assignment
		if err := i.SetVariable(baseVarName, rhsValue); err != nil {
			return nil, WrapErrorWithPosition(err, step.Pos, fmt.Sprintf("setting var '%s'", baseVarName))
		}
		return rhsValue, nil
	}

	// --- Complex Assignment ---
	currentValue, varExists := i.GetVariable(baseVarName)
	const (
		_ = iota
		MapTypeT
		ListTypeT
	)
	determinedBaseType := MapTypeT
	firstAccessor := step.LValue.Accessors[0]

	if firstAccessor.Type == BracketAccess {
		accVal, err := i.evaluateExpression(firstAccessor.IndexOrKey)
		if err != nil {
			return nil, WrapErrorWithPosition(err, firstAccessor.Pos, "evaluating first accessor")
		}
		if _, isInt := toInt64Coerce(accVal); isInt {
			determinedBaseType = ListTypeT
		}
	}

	if !varExists || (determinedBaseType == ListTypeT && !isList(currentValue)) || (determinedBaseType == MapTypeT && !isMap(currentValue)) {
		action := Ternary(!varExists, "auto-creating", "overwriting")
		newTypeStr := Ternary(determinedBaseType == ListTypeT, "LIST", "MAP")
		logMsg := fmt.Sprintf("Interpreter: %s variable '%s' (type %%T) as %s for assignment.", action, baseVarName, newTypeStr)
		i.Logger().Debugf(logMsg, currentValue)

		if determinedBaseType == ListTypeT {
			currentValue = make([]interface{}, 0)
		} else {
			currentValue = make(map[string]interface{})
		}
		if err := i.SetVariable(baseVarName, currentValue); err != nil {
			return nil, WrapErrorWithPosition(err, step.LValue.Pos, fmt.Sprintf("%s '%s'", action, baseVarName))
		}
	}

	var parentOfValueBeingAccessed interface{} = nil
	var keyUsedToAccessValue string
	var indexUsedToAccessValue int64 = -1

	for accessorIdx, accessor := range step.LValue.Accessors {
		isFinalAccessor := accessorIdx == len(step.LValue.Accessors)-1
		valueBeingAccessed := currentValue // This is the collection we will apply the current accessor to.

		createNextAsList := false
		if !isFinalAccessor {
			nextAcc := step.LValue.Accessors[accessorIdx+1]
			if nextAcc.Type == BracketAccess {
				nextKey, err := i.evaluateExpression(nextAcc.IndexOrKey)
				if err == nil {
					if _, isInt := toInt64Coerce(nextKey); isInt {
						createNextAsList = true
					}
				}
			}
		}

		// Before processing the current accessor, 'valueBeingAccessed' is the collection we are working on.
		// 'parentOfValueBeingAccessed', 'keyUsedToAccessValue', 'indexUsedToAccessValue' refer to how 'valueBeingAccessed'
		// was obtained from *its* parent (i.e., the 'valueBeingAccessed' of the previous iteration).

		currentIterationParent := parentOfValueBeingAccessed
		currentIterationKey := keyUsedToAccessValue
		currentIterationIndex := indexUsedToAccessValue

		// Setup for the *next* iteration: current 'valueBeingAccessed' will be the parent.
		parentOfValueBeingAccessed = valueBeingAccessed

		switch accessor.Type {
		case BracketAccess:
			indexOrKeyVal, errVal := i.evaluateExpression(accessor.IndexOrKey)
			if errVal != nil {
				return nil, WrapErrorWithPosition(errVal, accessor.Pos, "evaluating bracket accessor")
			}

			if actualMap, ok := valueBeingAccessed.(map[string]interface{}); ok {
				keyStr, isStr := indexOrKeyVal.(string)
				if !isStr {
					return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("map key string, got %T", indexOrKeyVal), ErrMapKeyNotFound).WithPosition(accessor.Pos)
				}

				keyUsedToAccessValue = keyStr // This key accesses the next element from actualMap
				indexUsedToAccessValue = -1

				if isFinalAccessor {
					actualMap[keyStr] = rhsValue
					return rhsValue, nil
				}
				nextVal, found := actualMap[keyStr]
				if !found || !isCollection(nextVal) || (createNextAsList && !isList(nextVal)) || (!createNextAsList && !isMap(nextVal)) {
					var newColl interface{}
					// Logging for creation/overwrite
					if createNextAsList {
						newColl = make([]interface{}, 0)
					} else {
						newColl = make(map[string]interface{})
					}
					actualMap[keyStr] = newColl
					currentValue = newColl
				} else {
					currentValue = nextVal
				}
			} else if listBeingAccessed, ok := valueBeingAccessed.([]interface{}); ok {
				idxToUseInList, isIntLike := toInt64Coerce(indexOrKeyVal)
				// This error is now handled inside processListElementAssignment
				if !isIntLike { /* already handled by helper */
				}

				indexUsedToAccessValue = idxToUseInList
				keyUsedToAccessValue = ""

				modifiedList, nextElement, listHeaderDidChange, errHelper := i.processListElementAssignment(listBeingAccessed, indexOrKeyVal, isFinalAccessor, rhsValue, createNextAsList, accessor.Pos, baseVarName)
				if errHelper != nil {
					return nil, errHelper
				}

				if listHeaderDidChange {
					// The list's header might have changed. Update its reference in its parent.
					if currentIterationParent == nil { // Implies accessorIdx == 0, listBeingAccessed was baseVar
						if errSet := i.SetVariable(baseVarName, modifiedList); errSet != nil {
							return nil, WrapErrorWithPosition(errSet, step.LValue.Pos, "updating base list variable")
						}
					} else if pMap, pOk := currentIterationParent.(map[string]interface{}); pOk {
						pMap[currentIterationKey] = modifiedList // Update the parent map
					} else if pList, pOk := currentIterationParent.([]interface{}); pOk {
						if currentIterationIndex >= 0 && currentIterationIndex < int64(len(pList)) {
							pList[currentIterationIndex] = modifiedList // Update the parent list
						} else {
							i.Logger().Errorf("Error updating parent list: index %d out of bounds for len %d", currentIterationIndex, len(pList))
							// Potentially return an error here if this state is considered critical
						}
					}
					currentValue = modifiedList // Ensure 'currentValue' for the next step (if any) or for final state refers to the new list
				} else {
					// No header change, but content might have. 'modifiedList' is the correct reference.
					currentValue = modifiedList
				}

				if isFinalAccessor {
					return nextElement, nil
				}
				currentValue = nextElement
			} else {
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("var '%s' path %T not collection for bracket", baseVarName, valueBeingAccessed), ErrCannotAccessType).WithPosition(accessor.Pos)
			}
		case DotAccess:
			fieldName := accessor.FieldName
			actualMap, ok := valueBeingAccessed.(map[string]interface{})
			if !ok {
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("var '%s' path %T not map for field '%s'", baseVarName, valueBeingAccessed, fieldName), ErrCannotAccessType).WithPosition(accessor.Pos)
			}

			keyUsedToAccessValue = fieldName
			indexUsedToAccessValue = -1

			if isFinalAccessor {
				actualMap[fieldName] = rhsValue
				return rhsValue, nil
			}
			nextVal, found := actualMap[fieldName]
			if !found || !isCollection(nextVal) || (createNextAsList && !isList(nextVal)) || (!createNextAsList && !isMap(nextVal)) {
				var newColl interface{}
				// Logging for creation/overwrite
				if createNextAsList {
					newColl = make([]interface{}, 0)
				} else {
					newColl = make(map[string]interface{})
				}
				actualMap[fieldName] = newColl
				currentValue = newColl
			} else {
				currentValue = nextVal
			}
		default:
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("unknown accessor %v for '%s'", accessor.Type, baseVarName), ErrInternal).WithPosition(accessor.Pos)
		}
	}
	return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("assignment did not complete for '%s'", step.LValue.String()), ErrInternal).WithPosition(step.Pos)
}

// NeuroScript Version: 0.4.2
// File version: 1.0.0
// Purpose: Defines interpreter execution for assignment ("set") steps, now fully Value-aware.
// filename: pkg/core/interpreter_assignment.go
// nlines: 215
// risk_rating: HIGH

package core

import (
	"fmt"
)

// executeSet handles the "set" step. It is now fully aware of Value types for both
// the collections being modified and the accessors being used.
func (i *Interpreter) executeSet(step Step) (interface{}, error) {
	if step.LValue == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "SetStep LValue is nil", nil).WithPosition(step.Pos)
	}

	baseVarName := step.LValue.Identifier
	rhsValueRaw, evalErr := i.evaluateExpression(step.Value)
	if evalErr != nil {
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), fmt.Sprintf("evaluating value for SET %s", baseVarName))
	}

	// Ensure the RHS is a proper Value type.
	rhsValue, ok := rhsValueRaw.(Value)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("RHS of assignment to '%s' did not evaluate to a Value type, got %T", baseVarName, rhsValueRaw), nil).WithPosition(step.Value.GetPos())
	}

	// Simple assignment: set x = ...
	if len(step.LValue.Accessors) == 0 {
		if err := i.SetVariable(baseVarName, rhsValue); err != nil {
			return nil, WrapErrorWithPosition(err, step.Pos, fmt.Sprintf("setting var '%s'", baseVarName))
		}
		return rhsValue, nil
	}

	// --- Complex Assignment: set x[0].key = ... ---

	// Get the base variable, which might not exist yet.
	currentVal, varExists := i.GetVariable(baseVarName)
	if !varExists {
		// Determine if the base should be a list or map based on the first accessor.
		firstAccessor := step.LValue.Accessors[0]
		if firstAccessor.Type == BracketAccess {
			accVal, err := i.evaluateExpression(firstAccessor.IndexOrKey)
			if err != nil {
				return nil, WrapErrorWithPosition(err, firstAccessor.Pos, "evaluating first accessor")
			}
			if _, isInt := toInt64(accVal); isInt {
				currentVal = NewListValue(nil) // It's a list access, so create a ListValue
			} else {
				currentVal = NewMapValue(nil) // Otherwise, it's a map access
			}
		} else { // Dot access implies the base must be a map.
			currentVal = NewMapValue(nil)
		}
		i.SetVariable(baseVarName, currentVal)
	}

	var parentVal Value
	var lastKey string
	var lastIndex int64

	// Traverse the accessor chain up to the second-to-last one.
	for accessorIdx, accessor := range step.LValue.Accessors {
		isFinalAccessor := accessorIdx == len(step.LValue.Accessors)-1

		if cv, ok := currentVal.(MapValue); ok {
			// --- Current level is a MAP ---
			var key string
			if accessor.Type == DotAccess {
				key = accessor.FieldName
			} else { // Bracket access on a map
				keyVal, err := i.evaluateExpression(accessor.IndexOrKey)
				if err != nil {
					return nil, WrapErrorWithPosition(err, accessor.Pos, "evaluating map key")
				}
				key, _ = toString(keyVal)
			}

			if isFinalAccessor {
				cv.Value[key] = rhsValue
				return rhsValue, nil
			}

			parentVal = cv
			lastKey = key
			nextVal, found := cv.Value[key]
			if !found || (!isList(nextVal) && !isMap(nextVal)) {
				// Auto-create next level if it doesn't exist or is not a collection.
				createNextAsList := !isFinalAccessor && step.LValue.Accessors[accessorIdx+1].Type == BracketAccess
				if _, isInt := toInt64(step.LValue.Accessors[accessorIdx+1].IndexOrKey); isInt {
					createNextAsList = true
				}

				if createNextAsList {
					nextVal = NewListValue(nil)
				} else {
					nextVal = NewMapValue(nil)
				}
				cv.Value[key] = nextVal
			}
			currentVal = nextVal

		} else if cv, ok := currentVal.(ListValue); ok {
			// --- Current level is a LIST ---
			if accessor.Type == DotAccess {
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("cannot use dot notation on a list for '%s'", baseVarName), ErrCannotAccessType).WithPosition(accessor.Pos)
			}
			indexVal, err := i.evaluateExpression(accessor.IndexOrKey)
			if err != nil {
				return nil, WrapErrorWithPosition(err, accessor.Pos, "evaluating list index")
			}
			index, isInt := toInt64(indexVal)
			if !isInt {
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("list index must be an integer, got %s", TypeOf(indexVal)), ErrListInvalidIndexType).WithPosition(accessor.Pos)
			}
			if index < 0 {
				return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("list index cannot be negative, got %d", index), ErrListIndexOutOfBounds).WithPosition(accessor.Pos)
			}

			// Pad list with NilValue if index is out of bounds
			for int64(len(cv.Value)) <= index {
				cv.Value = append(cv.Value, NilValue{})
				// If the list itself was grown, we need to update its reference in the parent
				if pmap, ok := parentVal.(MapValue); ok {
					pmap.Value[lastKey] = cv
				} else if plist, ok := parentVal.(ListValue); ok {
					plist.Value[lastIndex] = cv
				}
			}

			if isFinalAccessor {
				cv.Value[index] = rhsValue
				return rhsValue, nil
			}

			parentVal = cv
			lastIndex = index
			nextVal := cv.Value[index]

			if !isList(nextVal) && !isMap(nextVal) {
				// Auto-create next level if it doesn't exist or is not a collection.
				createNextAsList := !isFinalAccessor && step.LValue.Accessors[accessorIdx+1].Type == BracketAccess
				if _, isInt := toInt64(step.LValue.Accessors[accessorIdx+1].IndexOrKey); isInt {
					createNextAsList = true
				}

				if createNextAsList {
					nextVal = NewListValue(nil)
				} else {
					nextVal = NewMapValue(nil)
				}
				cv.Value[index] = nextVal
			}
			currentVal = nextVal

		} else {
			// Current value is not a collection, but we need to descend further.
			// This implies we need to overwrite it.
			var newColl Value
			createNextAsList := !isFinalAccessor && step.LValue.Accessors[accessorIdx+1].Type == BracketAccess
			if _, isInt := toInt64(step.LValue.Accessors[accessorIdx+1].IndexOrKey); isInt {
				createNextAsList = true
			}

			if createNextAsList {
				newColl = NewListValue(nil)
			} else {
				newColl = NewMapValue(nil)
			}

			// Update the parent to point to this new collection
			if pmap, ok := parentVal.(MapValue); ok {
				pmap.Value[lastKey] = newColl
			} else if plist, ok := parentVal.(ListValue); ok {
				plist.Value[lastIndex] = newColl
			}

			currentVal = newColl
			// Retry the current accessor with the newly created collection (this is complex, for now we will fail)
			// A simpler model is to assume the script creates structures correctly.
			// For now, we will error if we hit a non-collection mid-path.
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("attempted to access through a non-collection type (%s) during assignment", TypeOf(currentVal)), ErrCannotAccessType).WithPosition(accessor.Pos)

		}
	}

	return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("assignment did not complete for '%s'", step.LValue.String()), ErrInternal).WithPosition(step.Pos)
}

// isList checks if an interface is a ListValue.
func isList(v interface{}) bool {
	_, ok := v.(ListValue)
	return ok
}

// isMap checks if an interface is a MapValue.
func isMap(v interface{}) bool {
	_, ok := v.(MapValue)
	return ok
}

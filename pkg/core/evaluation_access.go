// pkg/core/evaluation_access.go
package core

import (
	"fmt"
	"strconv"
)

// evaluateElementAccess handles the logic for accessing elements within collections.
// It's called by evaluateExpression when an ElementAccessNode is encountered.
func (i *Interpreter) evaluateElementAccess(n *ElementAccessNode) (interface{}, error) {
	// 1. Evaluate the collection part
	collectionVal, errColl := i.evaluateExpression(n.Collection)
	if errColl != nil {
		return nil, fmt.Errorf("evaluating collection for element access: %w", errColl)
	}
	// 2. Evaluate the accessor part
	accessorVal, errAcc := i.evaluateExpression(n.Accessor)
	if errAcc != nil {
		return nil, fmt.Errorf("evaluating accessor for element access: %w", errAcc)
	}

	if i.logger != nil {
		posStr := ""
		if n.Pos != nil {
			posStr = fmt.Sprintf(" (at %s)", n.Pos.String())
		}
		i.logger.Debug("[DEBUG-INTERP] Evaluating Element Access%s: Collection=%T, Accessor=%T (%v)", posStr, collectionVal, accessorVal, accessorVal)
	}

	if collectionVal == nil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "collection evaluated to nil", ErrCollectionIsNil)
	}
	if accessorVal == nil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "accessor evaluated to nil", ErrAccessorIsNil)
	}

	// 3. Perform access based on the evaluated collection type
	switch coll := collectionVal.(type) {

	// *** FIX STARTS HERE: Handle specific Value wrapper types instead of raw Go types ***
	case MapValue:
		key, ok := accessorVal.(string)
		if !ok {
			key = fmt.Sprintf("%v", accessorVal)
		}
		value, found := coll.Value[key]
		if !found {
			// A key not being found in a map returns an error, unlike in event data.
			return nil, NewRuntimeError(ErrorCodeKeyNotFound,
				fmt.Sprintf("key '%s' not found in map", key),
				fmt.Errorf("%w: key '%s'", ErrMapKeyNotFound, key))
		}
		return value, nil

	case ListValue:
		index, err := i.coerceToInt(accessorVal)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeType,
				fmt.Sprintf("list index must be an integer, but got %T", accessorVal),
				ErrListInvalidIndexType)
		}
		listLen := len(coll.Value)
		if index < 0 || int(index) >= listLen {
			return nil, NewRuntimeError(ErrorCodeBounds,
				fmt.Sprintf("list index %d is out of bounds for list of length %d", index, listLen),
				fmt.Errorf("%w: index %d, length %d", ErrListIndexOutOfBounds, index, listLen))
		}
		return coll.Value[int(index)], nil

	case EventValue:
		key, ok := accessorVal.(string)
		if !ok {
			key = fmt.Sprintf("%v", accessorVal)
		}
		value, found := coll.Value[key]
		if !found {
			// A key not being found in event data returns nil (e.g., if payload is optional).
			return nil, nil
		}
		return value, nil
	// *** FIX ENDS HERE ***

	default:
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("cannot perform element access using [...] on type %T", collectionVal),
			ErrCannotAccessType)
	}
}

// coerceToInt is a helper to convert an interface{} to an int64 if possible.
func (i *Interpreter) coerceToInt(val interface{}) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		if v == float64(int64(v)) { // Check if it's a whole number
			return int64(v), nil
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("cannot coerce type %T to an integer", val)
}

// evaluateListElementAccess is no longer directly called but kept for reference if needed.
func (i *Interpreter) evaluateListElementAccess(list []interface{}, accessorVal interface{}) (interface{}, error) {
	index, err := i.coerceToInt(accessorVal)
	if err != nil {
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("list index must evaluate to an integer, but got %T (%v)", accessorVal, accessorVal),
			ErrListInvalidIndexType)
	}

	listLen := len(list)
	if index < 0 || int(index) >= listLen {
		return nil, NewRuntimeError(ErrorCodeBounds,
			fmt.Sprintf("list index %d is out of bounds", index),
			fmt.Errorf("%w: index %d, length %d", ErrListIndexOutOfBounds, index, listLen))
	}
	element := list[int(index)]
	if i.logger != nil {
		i.logger.Debug("[DEBUG-INTERP]   List access successful: Index=%d, Value=%v", index, element)
	}
	return element, nil
}

// evaluateMapElementAccess is no longer directly called but kept for reference if needed.
func (i *Interpreter) evaluateMapElementAccess(m map[string]interface{}, accessorVal interface{}) (interface{}, error) {
	key, ok := accessorVal.(string)
	if !ok {
		key = fmt.Sprintf("%v", accessorVal)
		if i.logger != nil {
			i.logger.Debug("[INFO-INTERP] Map key was not a string (%T), converted to string key '%s' for access", accessorVal, key)
		}
	}

	value, found := m[key]
	if !found {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound,
			fmt.Sprintf("key '%s' not found", key),
			fmt.Errorf("%w: key '%s'", ErrMapKeyNotFound, key))
	}
	if i.logger != nil {
		i.logger.Debug("[DEBUG-INTERP]   Map access successful: Key='%s', Value=%v", key, value)
	}
	return value, nil
}

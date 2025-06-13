// NeuroScript Version: 0.3.5
// File version: 9
// Purpose: Handles element access for both wrapped Value types and raw Go types, properly unwrapping accessors.
// filename: pkg/core/evaluation_access.go
// nlines: 101
// risk_rating: HIGH

package core

import (
	"fmt"
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

	if collectionVal == nil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "collection evaluated to nil", ErrCollectionIsNil).WithPosition(n.GetPos())
	}
	if accessorVal == nil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "accessor evaluated to nil", ErrAccessorIsNil).WithPosition(n.GetPos())
	}

	// 3. Perform access based on the evaluated collection type
	switch coll := collectionVal.(type) {
	case ListValue:
		index, ok := toInt64(accessorVal)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeType,
				fmt.Sprintf("list index must be an integer, but got %s", TypeOf(accessorVal)),
				ErrListInvalidIndexType).WithPosition(n.GetPos())
		}
		listLen := len(coll.Value)
		if index < 0 || int(index) >= listLen {
			return nil, NewRuntimeError(ErrorCodeBounds,
				fmt.Sprintf("list index %d is out of bounds for list of length %d", index, listLen),
				fmt.Errorf("%w: index %d, length %d", ErrListIndexOutOfBounds, index, listLen)).WithPosition(n.GetPos())
		}
		return coll.Value[int(index)], nil

	case MapValue:
		// Map keys must be strings. We convert the accessor to a string.
		key, _ := toString(accessorVal)
		value, found := coll.Value[key]
		if !found {
			return nil, NewRuntimeError(ErrorCodeKeyNotFound,
				fmt.Sprintf("key '%s' not found in map", key),
				fmt.Errorf("%w: key '%s'", ErrMapKeyNotFound, key)).WithPosition(n.GetPos())
		}
		return value, nil

	case EventValue:
		key, _ := toString(accessorVal)
		value, found := coll.Value[key]
		if !found {
			// A key not being found in event data returns nil (e.g., if payload is optional).
			return NilValue{}, nil
		}
		return value, nil

	// FIX: Add cases for raw Go types from variables (for legacy test support and flexibility)
	case []interface{}:
		index, ok := toInt64(accessorVal)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeType,
				fmt.Sprintf("list index must be an integer, but got %s", TypeOf(accessorVal)),
				ErrListInvalidIndexType).WithPosition(n.GetPos())
		}
		listLen := len(coll)
		if index < 0 || int(index) >= listLen {
			return nil, NewRuntimeError(ErrorCodeBounds,
				fmt.Sprintf("list index %d is out of bounds for list of length %d", index, listLen),
				fmt.Errorf("%w: index %d, length %d", ErrListIndexOutOfBounds, index, listLen)).WithPosition(n.GetPos())
		}
		return coll[int(index)], nil

	case map[string]interface{}:
		key, _ := toString(accessorVal)
		value, found := coll[key]
		if !found {
			// This path is for raw maps where a key not found is an error, not nil.
			return nil, NewRuntimeError(ErrorCodeKeyNotFound,
				fmt.Sprintf("key '%s' not found in map", key),
				fmt.Errorf("%w: key '%s'", ErrMapKeyNotFound, key)).WithPosition(n.GetPos())
		}
		return value, nil

	default:
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("cannot perform element access using [...] on type %T", collectionVal),
			ErrCannotAccessType).WithPosition(n.GetPos())
	}
}

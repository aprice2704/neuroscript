// NeuroScript Version: 0.4.0
// File version: 10
// Purpose: Corrects element access to work exclusively with Value wrapper types (ListValue, MapValue, etc.).
// filename: pkg/core/evaluation_access.go
// nlines: 77
// risk_rating: HIGH

package core

import (
	"fmt"
)

// evaluateElementAccess handles the logic for accessing elements within collections.
// It now expects the collection to be a Value type and returns a Value.
func (i *Interpreter) evaluateElementAccess(n *ElementAccessNode) (Value, error) {
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

	if _, isNil := collectionVal.(NilValue); isNil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "collection evaluated to nil", ErrCollectionIsNil).WithPosition(n.GetPos())
	}
	if _, isNil := accessorVal.(NilValue); isNil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "accessor evaluated to nil", ErrAccessorIsNil).WithPosition(n.GetPos())
	}

	// 3. Perform access based on the collection's wrapper type.
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
			return NilValue{}, nil
		}
		return value, nil

	default:
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("cannot perform element access using [...] on type %s", coll.Type()),
			ErrCannotAccessType).WithPosition(n.GetPos())
	}
}

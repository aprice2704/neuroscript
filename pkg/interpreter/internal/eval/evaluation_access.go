// NeuroScript Version: 0.4.0
// File version: 10
// Purpose: Corrects element access to work exclusively with Value wrapper types (ListValue, MapValue, etc.).
// filename: pkg/interpreter/internal/eval/evaluation_access.go
// nlines: 77
// risk_rating: HIGH

package eval

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// evaluateElementAccess handles the logic for accessing elements within collections.
// It now expects the collection to be a Value type and returns a Value.
func (i *Interpreter) evaluateElementAccess(n *ast.ElementAccessNode) (lang.Value, error) {
	// 1. Evaluate the collection part
	collectionVal, errColl := i.evaluate.Expression(n.Collection)
	if errColl != nil {
		return nil, fmt.Errorf("evaluating collection for element access: %w", errColl)
	}
	// 2. Evaluate the accessor part
	accessorVal, errAcc := i.evaluate.Expression(n.Accessor)
	if errAcc != nil {
		return nil, fmt.Errorf("evaluating accessor for element access: %w", errAcc)
	}

	if _, isNil := collectionVal.(NilValue); isNil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeEvaluation, "collection evaluated to nil", lang.ErrCollectionIsNil).WithPosition(n.GetPos())
	}
	if _, isNil := accessorVal.(NilValue); isNil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeEvaluation, "accessor evaluated to nil", lang.ErrAccessorIsNil).WithPosition(n.GetPos())
	}

	// 3. Perform access based on the collection's wrapper type.
	switch coll := collectionVal.(type) {
	case ListValue:
		index, ok := lang.toInt64(accessorVal)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType,
				fmt.Sprintf("list index must be an integer, but got %s", lang.TypeOf(accessorVal)),
				lang.ErrListInvalidIndexType).WithPosition(n.GetPos())
		}
		listLen := len(coll.Value)
		if index < 0 || int(index) >= listLen {
			return nil, lang.NewRuntimeError(lang.ErrorCodeBounds,
				fmt.Sprintf("list index %d is out of bounds for list of length %d", index, listLen),
				fmt.Errorf("%w: index %d, length %d", lang.ErrListIndexOutOfBounds, index, listLen)).WithPosition(n.GetPos())
		}
		return coll.Value[int(index)], nil

	case MapValue:
		key, _ := lang.toString(accessorVal)
		value, found := coll.Value[key]
		if !found {
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound,
				fmt.Sprintf("key '%s' not found in map", key),
				fmt.Errorf("%w: key '%s'", lang.ErrMapKeyNotFound, key)).WithPosition(n.GetPos())
		}
		return value, nil

	case EventValue:
		key, _ := lang.toString(accessorVal)
		value, found := coll.Value[key]
		if !found {
			return lang.NilValue{}, nil
		}
		return value, nil

	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeType,
			fmt.Sprintf("cannot perform element access using [...] on type %s", coll.Type()),
			lang.ErrCannotAccessType).WithPosition(n.GetPos())
	}
}
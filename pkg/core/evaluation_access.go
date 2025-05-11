// pkg/core/evaluation_access.go
package core

import (
	"fmt"
	"strconv"
)

// evaluateElementAccess handles the logic for accessing elements within lists or maps.
// It's called by evaluateExpression when an ElementAccessNode is encountered.
func (i *Interpreter) evaluateElementAccess(n *ElementAccessNode) (interface{}, error) {
	// Input 'n' is a pointer, access fields normally.

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

	// Handle case where collection evaluated to nil before attempting access
	if collectionVal == nil {
		// Use RuntimeError - consistent error type
		return nil, NewRuntimeError(ErrorCodeEvaluation, "collection evaluated to nil", ErrCollectionIsNil)
	}
	if accessorVal == nil {
		// Use RuntimeError - consistent error type
		return nil, NewRuntimeError(ErrorCodeEvaluation, "accessor evaluated to nil", ErrAccessorIsNil)

	}

	// 3. Perform access based on the evaluated collection type
	switch coll := collectionVal.(type) {
	case []interface{}: // List Access
		return i.evaluateListElementAccess(coll, accessorVal)
	case map[string]interface{}: // Map Access
		return i.evaluateMapElementAccess(coll, accessorVal)
	default:
		// Return RuntimeError for attempting access on unsupported type
		// Use ErrorCodeType for this kind of error
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("cannot perform element access using [...] on type %T", collectionVal),
			ErrCannotAccessType) // Use the pre-defined error variable
	}
}

// evaluateListElementAccess handles accessing an element within a slice.
func (i *Interpreter) evaluateListElementAccess(list []interface{}, accessorVal interface{}) (interface{}, error) {
	var index int64
	converted := false
	// Attempt to convert the accessor to an integer index
	switch acc := accessorVal.(type) {
	case int64:
		index = acc
		converted = true
	case int: // Allow Go int type
		index = int64(acc)
		converted = true
	case float64: // Allow float if it represents a whole number
		if acc == float64(int64(acc)) {
			index = int64(acc)
			converted = true
		}
	case string: // Allow numeric strings
		parsedIndex, err := strconv.ParseInt(acc, 10, 64)
		if err == nil {
			index = parsedIndex
			converted = true
		}
	}

	if !converted {
		// Use RuntimeError - Use ErrorCodeType
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("list index must evaluate to an integer, but got %T (%v)", accessorVal, accessorVal),
			ErrListInvalidIndexType) // Use pre-defined error
	}

	// Bounds check
	listLen := len(list)
	if index < 0 || int(index) >= listLen {
		// Use RuntimeError for user-facing boundary errors
		// --- MODIFIED: Use ErrorCodeBounds ---
		return nil, NewRuntimeError(ErrorCodeBounds, // Corrected constant
			fmt.Sprintf("list index %d is out of bounds", index),
			// Use pre-defined error variable for wrapping if desired, or keep specific fmt.Errorf
			// Using ErrListIndexOutOfBounds here for consistency if Is/As checks are used elsewhere
			fmt.Errorf("%w: index %d, length %d", ErrListIndexOutOfBounds, index, listLen))
	}
	// Return the element and nil error on success
	element := list[int(index)]
	if i.logger != nil {
		i.logger.Debug("[DEBUG-INTERP]   List access successful: Index=%d, Value=%v", index, element)
	}
	return element, nil
}

// evaluateMapElementAccess handles accessing an element within a map.
func (i *Interpreter) evaluateMapElementAccess(m map[string]interface{}, accessorVal interface{}) (interface{}, error) {
	// Map keys must be strings (or convertible)
	key, ok := accessorVal.(string)
	if !ok {
		// Lenient: convert accessor to string representation
		key = fmt.Sprintf("%v", accessorVal)
		if i.logger != nil {
			i.logger.Debug("[INFO-INTERP] Map key was not a string (%T), converted to string key '%s' for access", accessorVal, key)
		}
	}

	value, found := m[key]
	if !found {
		// Use RuntimeError for user-facing key errors
		// ErrorCodeKeyNotFound is correctly defined in errors.go
		return nil, NewRuntimeError(ErrorCodeKeyNotFound,
			fmt.Sprintf("key '%s' not found", key),
			// Use pre-defined error variable
			fmt.Errorf("%w: key '%s'", ErrMapKeyNotFound, key))
	}
	// Return the found value and nil error
	if i.logger != nil {
		i.logger.Debug("[DEBUG-INTERP]   Map access successful: Key='%s', Value=%v", key, value)
	}
	return value, nil
}

// NeuroScript Version: 0.4.2
// File version: 8.0.0
// Purpose: Implements a fully robust recursive 'set' execution with correct auto-vivification logic.
// filename: pkg/core/interpreter_assignment.go
// nlines: 195
// risk_rating: HIGH

package core

import (
	"fmt"
)

// executeSet handles the "set" step, including complex assignments with auto-creation
// of nested lists and maps (auto-vivification).
func (i *Interpreter) executeSet(step Step) (interface{}, error) {
	if step.LValue == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "SetStep LValue is nil", nil).WithPosition(step.Pos)
	}

	rhsValueRaw, evalErr := i.evaluateExpression(step.Value)
	if evalErr != nil {
		return nil, WrapErrorWithPosition(evalErr, step.Value.GetPos(), fmt.Sprintf("evaluating value for SET %s", step.LValue.Identifier))
	}

	rhsValue, ok := rhsValueRaw.(Value)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("RHS of assignment to '%s' did not evaluate to a Value type, got %T", step.LValue.Identifier, rhsValueRaw), nil).WithPosition(step.Value.GetPos())
	}

	// Simple assignment: set x = ...
	if len(step.LValue.Accessors) == 0 {
		return rhsValue, i.SetVariable(step.LValue.Identifier, rhsValue)
	}

	// Complex Assignment: set x[0].key = ...
	baseVarName := step.LValue.Identifier

	// 1. Get the root container, creating it if it doesn't exist.
	root, err := i.getOrCreateRootContainer(baseVarName, step.LValue.Accessors[0])
	if err != nil {
		return nil, err
	}

	// 2. Recursively traverse the path, modifying the data structure.
	modifiedRoot, err := i.vivifyAndSet(root, step.LValue.Accessors, rhsValue)
	if err != nil {
		return nil, err
	}

	// 3. Commit the modified root back to the variable scope.
	if err := i.SetVariable(baseVarName, modifiedRoot); err != nil {
		return nil, err
	}

	return rhsValue, nil
}

// getOrCreateRootContainer retrieves the top-level variable for a complex assignment,
// creating it if it doesn't exist based on the first accessor.
func (i *Interpreter) getOrCreateRootContainer(name string, firstAccessor AccessorNode) (Value, error) {
	rawVal, varExists := i.GetVariable(name)
	if varExists {
		if container, ok := rawVal.(Value); ok && (isMap(container) || isList(container)) {
			return container, nil
		}
	}
	return i.determineInitialContainer(firstAccessor)
}

// vivifyAndSet recursively traverses the accessor path, creating nested containers
// as needed, and returns the (potentially modified) container.
func (i *Interpreter) vivifyAndSet(current Value, accessors []AccessorNode, rhsValue Value) (Value, error) {
	accessor := accessors[0]
	isFinal := len(accessors) == 1

	// --- Handle Map ---
	if m, ok := current.(MapValue); ok {
		key, err := i.evaluateAccessorKey(accessor)
		if err != nil {
			return nil, err
		}

		if isFinal {
			m.Value[key] = rhsValue
			return m, nil
		}

		child, exists := m.Value[key]
		if !exists || (!isMap(child) && !isList(child)) {
			child, err = i.determineInitialContainer(accessors[1])
			if err != nil {
				return nil, err
			}
		}

		modifiedChild, err := i.vivifyAndSet(child, accessors[1:], rhsValue)
		if err != nil {
			return nil, err
		}
		m.Value[key] = modifiedChild // Update the child reference
		return m, nil
	}

	// --- Handle List ---
	if l, ok := current.(ListValue); ok {
		index, err := i.evaluateAccessorIndex(accessor)
		if err != nil {
			return nil, err
		}

		l.Value = padList(l.Value, index)

		if isFinal {
			l.Value[index] = rhsValue
			return l, nil
		}

		child := l.Value[index]
		if !isMap(child) && !isList(child) {
			child, err = i.determineInitialContainer(accessors[1])
			if err != nil {
				return nil, err
			}
		}

		modifiedChild, err := i.vivifyAndSet(child, accessors[1:], rhsValue)
		if err != nil {
			return nil, err
		}
		l.Value[index] = modifiedChild // Update the child reference
		return l, nil
	}

	// --- Handle Primitive ---
	// If it's a primitive, it must be overwritten by a new container.
	newContainer, err := i.determineInitialContainer(accessor)
	if err != nil {
		return nil, err
	}
	return i.vivifyAndSet(newContainer, accessors, rhsValue)
}

// Helper functions for evaluation and type checking.

func (i *Interpreter) determineInitialContainer(accessor AccessorNode) (Value, error) {
	if accessor.Type == DotAccess {
		return NewMapValue(nil), nil
	}
	key, err := i.evaluateExpression(accessor.IndexOrKey)
	if err != nil {
		return nil, WrapErrorWithPosition(err, accessor.Pos, "evaluating accessor key")
	}
	if _, isInt := toInt64(key); isInt {
		return NewListValue(nil), nil
	}
	return NewMapValue(nil), nil
}

func (i *Interpreter) evaluateAccessorKey(accessor AccessorNode) (string, error) {
	if accessor.Type == DotAccess {
		return accessor.FieldName, nil
	}
	keyVal, err := i.evaluateExpression(accessor.IndexOrKey)
	if err != nil {
		return "", WrapErrorWithPosition(err, accessor.Pos, "evaluating map key")
	}
	key, _ := toString(keyVal)
	return key, nil
}

func (i *Interpreter) evaluateAccessorIndex(accessor AccessorNode) (int64, error) {
	indexVal, err := i.evaluateExpression(accessor.IndexOrKey)
	if err != nil {
		return 0, WrapErrorWithPosition(err, accessor.Pos, "evaluating list index")
	}
	index, isInt := toInt64(indexVal)
	if !isInt {
		return 0, NewRuntimeError(ErrorCodeType, fmt.Sprintf("list index must be an integer, got %s", TypeOf(indexVal)), ErrListInvalidIndexType).WithPosition(accessor.Pos)
	}
	if index < 0 {
		return 0, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("list index cannot be negative, got %d", index), ErrListIndexOutOfBounds).WithPosition(accessor.Pos)
	}
	return index, nil
}

func padList(list []Value, requiredIndex int64) []Value {
	for int64(len(list)) <= requiredIndex {
		list = append(list, NilValue{})
	}
	return list
}

func isList(v Value) bool { _, ok := v.(ListValue); return ok }
func isMap(v Value) bool  { _, ok := v.(MapValue); return ok }

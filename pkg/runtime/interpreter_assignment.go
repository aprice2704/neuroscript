// NeuroScript Version: 0.4.2
// File version: 11.0.0
// Purpose: Refactored set execution to support multiple assignment targets.
// filename: pkg/runtime/interpreter_assignment.go

package runtime

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeSet handles the "set" step. It now dispatches to a helper for each LValue.
func (i *Interpreter) executeSet(step ast.Step) (Value, error) {
	if len(step.LValues) == 0 {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, "SetStep LValues is empty", nil).WithPosition(step.Pos)
	}

	rhsValue, evalErr := i.evaluate.Expression(step.Value)
	if evalErr != nil {
		// Use the lang.Position of the first LValue for more accurate error reporting
		return nil, WrapErrorWithPosition(evalErr, step.LValues[0].GetPos(), "evaluating value for SET statement")
	}

	// Case 1: Multiple assignment (e.g., set a, b = myList)
	if len(step.LValues) > 1 {
		list, ok := rhsValue.(ListValue)
		if !ok {
			return nil, lang.NewRuntimeError(ErrorCodeType, "multiple assignment requires a list on the right-hand side", ErrMultiAssignNonList).WithPosition(step.Value.GetPos())
		}
		if len(step.LValues) != len(list.Value) {
			return nil, lang.NewRuntimeError(ErrorCodeCountMismatch, fmt.Sprintf("assignment mismatch: %d variables but %d values", len(step.LValues), len(list.Value)), ErrAssignCountMismatch).WithPosition(step.Pos)
		}
		for idx, lval := range step.LValues {
			if err := i.setSingleLValue(lval, list.Value[idx]); err != nil {
				return nil, err
			}
		}
		return list, nil // Return the original list for multiple assignments
	}

	// Case 2: Single assignment (e.g., set a = 1 or set a.b[0] = 1)
	if err := i.setSingleLValue(step.LValues[0], rhsValue); err != nil {
		return nil, err
	}
	return rhsValue, nil
}

// setSingleLValue handles the logic for assigning a value to a single, potentially complex LValue.
// This includes auto-vivification of maps and lists.
func (i *Interpreter) setSingleLValue(lvalueExpr ast.Expression, rhsValue Value) error {
	lval, ok := lvalueExpr.(*ast.LValueNode)
	if !ok {
		return lang.NewRuntimeError(ErrorCodeInternal, "LValue expression is not an ast.LValueNode", nil).WithPosition(lvalueExpr.GetPos())
	}

	// Simple assignment: set x = ...
	if len(lval.Accessors) == 0 {
		return i.SetVariable(lval.Identifier, rhsValue)
	}

	// Complex Assignment: set x[0].key = ...
	baseVarName := lval.Identifier

	// 1. Get the root container, creating it if it doesn't exist.
	root, err := i.getOrCreateRootContainer(baseVarName, lval.Accessors[0])
	if err != nil {
		return err
	}

	// 2. Recursively traverse the path, modifying the data structure.
	modifiedRoot, err := i.vivifyAndSet(root, lval.Accessors, rhsValue)
	if err != nil {
		return err
	}

	// 3. Commit the modified root back to the variable scope.
	return i.SetVariable(baseVarName, modifiedRoot)
}

// getOrCreateRootContainer retrieves the top-level variable for a complex assignment,
// creating it if it doesn't exist based on the first accessor.
func (i *Interpreter) getOrCreateRootContainer(name string, firstAccessor ast.AccessorNode) (Value, error) {
	container, varExists := i.GetVariable(name)
	if varExists {
		if isMap(container) || isList(container) {
			return container, nil
		}
	}
	return i.determineInitialContainer(firstAccessor)
}

// vivifyAndSet recursively traverses the accessor path, creating nested containers
// as needed, and returns the (potentially modified) container.
func (i *Interpreter) vivifyAndSet(current Value, accessors []ast.AccessorNode, rhsValue Value) (Value, error) {
	if len(accessors) == 0 {
		return rhsValue, nil
	}

	accessor := accessors[0]
	isFinal := len(accessors) == 1

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
		m.Value[key] = modifiedChild
		return m, nil
	}

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
		if child == nil || (!isMap(child) && !isList(child)) {
			child, err = i.determineInitialContainer(accessors[1])
			if err != nil {
				return nil, err
			}
		}
		modifiedChild, err := i.vivifyAndSet(child, accessors[1:], rhsValue)
		if err != nil {
			return nil, err
		}
		l.Value[index] = modifiedChild
		return l, nil
	}

	newContainer, err := i.determineInitialContainer(accessor)
	if err != nil {
		return nil, err
	}
	return i.vivifyAndSet(newContainer, accessors, rhsValue)
}

// ... (rest of the helper functions: determineInitialContainer, evaluateAccessorKey, etc. are unchanged) ...
func (i *Interpreter) determineInitialContainer(accessor ast.AccessorNode) (Value, error) {
	if accessor.Type == ast.DotAccess {
		return NewMapValue(nil), nil
	}
	key, err := i.evaluate.Expression(accessor.IndexOrKey)
	if err != nil {
		return nil, WrapErrorWithPosition(err, accessor.Pos, "evaluating accessor key")
	}
	if _, isInt := toInt64(key); isInt {
		return NewListValue(nil), nil
	}
	return NewMapValue(nil), nil
}

func (i *Interpreter) evaluateAccessorKey(accessor ast.AccessorNode) (string, error) {
	if accessor.Type == ast.DotAccess {
		return accessor.FieldName, nil
	}
	keyVal, err := i.evaluate.Expression(accessor.IndexOrKey)
	if err != nil {
		return "", WrapErrorWithPosition(err, accessor.Pos, "evaluating map key")
	}
	key, _ := toString(keyVal)
	return key, nil
}

func (i *Interpreter) evaluateAccessorIndex(accessor ast.AccessorNode) (int64, error) {
	indexVal, err := i.evaluate.Expression(accessor.IndexOrKey)
	if err != nil {
		return 0, WrapErrorWithPosition(err, accessor.Pos, "evaluating list index")
	}
	index, isInt := toInt64(indexVal)
	if !isInt {
		return 0, lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("list index must be an integer, got %s", TypeOf(indexVal)), ErrListInvalidIndexType).WithPosition(accessor.Pos)
	}
	if index < 0 {
		return 0, lang.NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("list index cannot be negative, got %d", index), ErrListIndexOutOfBounds).WithPosition(accessor.Pos)
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

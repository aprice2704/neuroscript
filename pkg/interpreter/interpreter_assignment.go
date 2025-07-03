// NeuroScript Version: 0.5.2
// File version: 12.0.1
// Purpose: Refactored set execution to support multiple assignment targets and align with the Value contract.
// filename: pkg/interpreter/interpreter_assignment.go
// nlines: 215
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeSet handles the "set" step. It now dispatches to a helper for each LValue.
func (i *Interpreter) executeSet(step ast.Step) (lang.Value, error) {
	if len(step.LValues) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "SetStep LValues is empty", nil).WithPosition(&step.Position)
	}

	// Corrected: ast.Step has a slice field 'Values', not a single 'Value'.
	if len(step.Values) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "SetStep Values slice is empty", nil).WithPosition(&step.Position)
	}
	rhsValue, evalErr := i.evaluate.Expression(step.Values[0])
	if evalErr != nil {
		// Use the lang.Position of the first LValue for more accurate error reporting
		return nil, lang.WrapErrorWithPosition(evalErr, step.LValues[0].GetPos(), "evaluating value for SET statement")
	}

	// Case 1: Multiple assignment (e.g., set a, b = myList)
	if len(step.LValues) > 1 {
		list, ok := rhsValue.(lang.ListValue)
		if !ok {
			// Corrected: Use Values[0] to get the position of the expression.
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "multiple assignment requires a list on the right-hand side", lang.ErrMultiAssignNonList).WithPosition(step.Values[0].GetPos())
		}
		if len(step.LValues) != len(list.Value) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeCountMismatch, fmt.Sprintf("assignment mismatch: %d variables but %d values", len(step.LValues), len(list.Value)), lang.ErrAssignCountMismatch).WithPosition(&step.Position)
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
func (i *Interpreter) setSingleLValue(lvalueExpr ast.Expression, rhsValue lang.Value) error {
	lval, ok := lvalueExpr.(*ast.LValueNode)
	if !ok {
		return lang.NewRuntimeError(lang.ErrorCodeInternal, "LValue expression is not an ast.LValueNode", nil).WithPosition(lvalueExpr.GetPos())
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

// getOrCreateRootContainer retrieves the top-level variable for a complex assignment.
func (i *Interpreter) getOrCreateRootContainer(name string, firstAccessor *ast.AccessorNode) (lang.Value, error) {
	container, varExists := i.GetVariable(name)
	if varExists {
		if isList(container) || isMap(container) {
			return container, nil
		}
	}
	// If the variable doesn't exist or isn't a collection, create a new one.
	return i.determineInitialContainer(firstAccessor)
}

// vivifyAndSet recursively traverses the accessor path, creating nested containers as needed.
func (i *Interpreter) vivifyAndSet(current lang.Value, accessors []*ast.AccessorNode, rhsValue lang.Value) (lang.Value, error) {
	if len(accessors) == 0 {
		return rhsValue, nil
	}

	accessor := accessors[0]
	isFinal := len(accessors) == 1

	if m, ok := current.(lang.MapValue); ok {
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

	if l, ok := current.(lang.ListValue); ok {
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

func (i *Interpreter) determineInitialContainer(accessor *ast.AccessorNode) (lang.Value, error) {
	if accessor.Type == ast.DotAccess {
		return lang.NewMapValue(nil), nil
	}
	key, err := i.evaluate.Expression(accessor.Key)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating accessor key")
	}
	if _, isInt := lang.ToInt64(key); isInt {
		return lang.NewListValue(nil), nil
	}
	return lang.NewMapValue(nil), nil
}

func (i *Interpreter) evaluateAccessorKey(accessor *ast.AccessorNode) (string, error) {
	if accessor.Type == ast.DotAccess {
		// This needs to be resolved based on the LValueNode's structure.
		// For now, assuming it's handled differently or needs a field name.
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "dot access key evaluation not fully implemented", nil).WithPosition(accessor.Key.GetPos())
	}
	keyVal, err := i.evaluate.Expression(accessor.Key)
	if err != nil {
		return "", lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating map key")
	}
	key, _ := lang.ToString(keyVal)
	return key, nil
}

func (i *Interpreter) evaluateAccessorIndex(accessor *ast.AccessorNode) (int64, error) {
	indexVal, err := i.evaluate.Expression(accessor.Key)
	if err != nil {
		return 0, lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating list index")
	}
	index, isInt := lang.ToInt64(indexVal)
	if !isInt {
		return 0, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("list index must be an integer, got %s", lang.TypeOf(indexVal)), lang.ErrListInvalidIndexType).WithPosition(accessor.Key.GetPos())
	}
	if index < 0 {
		return 0, lang.NewRuntimeError(lang.ErrorCodeBounds, fmt.Sprintf("list index cannot be negative, got %d", index), lang.ErrListIndexOutOfBounds).WithPosition(accessor.Key.GetPos())
	}
	return index, nil
}

func padList(list []lang.Value, requiredIndex int64) []lang.Value {
	for int64(len(list)) <= requiredIndex {
		list = append(list, &lang.NilValue{})
	}
	return list
}

func isList(v lang.Value) bool { _, ok := v.(lang.ListValue); return ok }
func isMap(v lang.Value) bool  { _, ok := v.(lang.MapValue); return ok }

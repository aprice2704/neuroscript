// NeuroScript Version: 0.5.2
// File version: 33
// Purpose: Corrected vivifyAndSet to fail when accessing a sub-element of an explicit 'nil', rather than incorrectly auto-creating a container.
// filename: pkg/interpreter/interpreter_assignment.go
// nlines: 240
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeSet handles the "set" step.
func (i *Interpreter) executeSet(step ast.Step) (lang.Value, error) {
	if len(step.LValues) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "SetStep LValues is empty", nil).WithPosition(&step.Position)
	}

	var rhsExpr ast.Expression
	if len(step.Values) > 0 {
		rhsExpr = step.Values[0]
	} else if step.Call != nil {
		rhsExpr = step.Call
	} else {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "SetStep has no RHS expression (neither Value nor Call)", nil).WithPosition(&step.Position)
	}

	rhsValue, evalErr := i.evaluate.Expression(rhsExpr)
	if evalErr != nil {
		return nil, lang.WrapErrorWithPosition(evalErr, step.LValues[0].GetPos(), "evaluating value for SET statement")
	}

	if len(step.LValues) > 1 {
		list, ok := rhsValue.(lang.ListValue)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "multiple assignment requires a list on the right-hand side", lang.ErrMultiAssignNonList).WithPosition(rhsExpr.GetPos())
		}
		if len(step.LValues) != len(list.Value) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeCountMismatch, fmt.Sprintf("assignment mismatch: %d variables but %d values", len(step.LValues), len(list.Value)), lang.ErrAssignCountMismatch).WithPosition(&step.Position)
		}
		for idx, lval := range step.LValues {
			if err := i.setSingleLValue(lval, list.Value[idx]); err != nil {
				return nil, err
			}
		}
		return list, nil
	}

	if err := i.setSingleLValue(step.LValues[0], rhsValue); err != nil {
		return nil, err
	}
	return rhsValue, nil
}

func (i *Interpreter) setSingleLValue(lvalueExpr ast.Expression, rhsValue lang.Value) error {
	lval, ok := lvalueExpr.(*ast.LValueNode)
	if !ok {
		return lang.NewRuntimeError(lang.ErrorCodeInternal, "LValue expression is not an ast.LValueNode", nil).WithPosition(lvalueExpr.GetPos())
	}

	if len(lval.Accessors) == 0 {
		return i.SetVariable(lval.Identifier, rhsValue)
	}

	baseVarName := lval.Identifier
	root, err := i.getOrCreateRootContainer(baseVarName, lval.Accessors[0])
	if err != nil {
		return lang.WrapErrorWithPosition(err, lval.GetPos(), fmt.Sprintf("cannot assign to '%s'", baseVarName))
	}

	modifiedRoot, err := i.vivifyAndSet(root, lval.Accessors, rhsValue)
	if err != nil {
		return err
	}

	return i.SetVariable(baseVarName, modifiedRoot)
}

func (i *Interpreter) getOrCreateRootContainer(name string, firstAccessor *ast.AccessorNode) (lang.Value, error) {
	container, varExists := i.GetVariable(name)
	if !varExists || isNil(container) {
		return i.determineInitialContainer(firstAccessor)
	}

	if isList(container) || isMap(container) {
		return container, nil
	}
	return nil, lang.NewRuntimeError(lang.ErrorCodeType,
		fmt.Sprintf("cannot use element access on type %s", container.Type()),
		lang.ErrCannotAccessType)
}

func (i *Interpreter) vivifyAndSet(current lang.Value, accessors []*ast.AccessorNode, rhsValue lang.Value) (lang.Value, error) {
	if len(accessors) == 0 {
		return rhsValue, nil
	}

	accessor := accessors[0]
	isFinal := len(accessors) == 1

	// FIX: This is the critical change. If we encounter a nil value *during* the
	// traversal (i.e., not at the root), it's an error to try to access its members.
	if isNil(current) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType,
			"cannot perform element access on a nil value",
			lang.ErrCollectionIsNil,
		).WithPosition(accessor.Key.GetPos())
	}

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
		// Vivify only if the key does not exist. Do not replace an existing explicit nil.
		if !exists {
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
		originalLen := int64(len(l.Value))
		l.Value = padList(l.Value, index)
		child := l.Value[index]

		if isFinal {
			l.Value[index] = rhsValue
			return l, nil
		}

		// Vivify only if the index was newly created by padding.
		// Do not replace an existing explicit nil.
		if index >= originalLen {
			newChild, err := i.determineInitialContainer(accessors[1])
			if err != nil {
				return nil, err
			}
			l.Value[index] = newChild
			child = newChild
		}

		modifiedChild, err := i.vivifyAndSet(child, accessors[1:], rhsValue)
		if err != nil {
			return nil, err
		}
		l.Value[index] = modifiedChild
		return l, nil
	}

	return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("cannot perform element access on type %s", current.Type()), lang.ErrCannotAccessType).WithPosition(accessor.Key.GetPos())
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
		if strLiteral, ok := accessor.Key.(*ast.StringLiteralNode); ok {
			return strLiteral.Value, nil
		}
		return strings.TrimPrefix(accessor.Key.String(), "."), nil
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
func isNil(v lang.Value) bool  { _, ok := v.(*lang.NilValue); return v == nil || ok }

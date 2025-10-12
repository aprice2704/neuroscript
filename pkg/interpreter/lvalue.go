// NeuroScript Version: 0.8.0
// File version: 27
// Purpose: Re-plumbed all expression evaluation to use the external 'eval' package.
// filename: pkg/interpreter/lvalue.go
// nlines: 230
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// LHSType defines whether a future container should be a map or a list.
type LHSType int

const (
	LHS_MAP LHSType = iota
	LHS_LIST
)

// executeSet handles a `set … = …` step at policy.
func (i *Interpreter) executeSet(step ast.Step) (lang.Value, error) {
	if len(step.LValues) == 0 {
		return nil, lang.NewRuntimeError(
			lang.ErrorCodeInternal, "SetStep LValues is empty", nil,
		).WithPosition(step.GetPos())
	}

	var rhsExpr ast.Expression
	if len(step.Values) > 0 {
		rhsExpr = step.Values[0]
	} else if step.Call != nil {
		rhsExpr = step.Call
	} else {
		return nil, lang.NewRuntimeError(
			lang.ErrorCodeInternal,
			"SetStep has no RHS expression (neither Value nor Call)", nil,
		).WithPosition(step.GetPos())
	}

	rhsValue, evalErr := eval.Expression(i, rhsExpr)
	if evalErr != nil {
		return nil, lang.WrapErrorWithPosition(evalErr, rhsExpr.GetPos(),
			"evaluating value for SET statement")
	}

	if len(step.LValues) > 1 {
		list, ok := rhsValue.(lang.ListValue)
		if !ok {
			if lp, okp := rhsValue.(*lang.ListValue); okp {
				list = *lp
			} else {
				return nil, lang.NewRuntimeError(
					lang.ErrorCodeType,
					"multiple LHS names but RHS is not list",
					lang.ErrMultiAssignNonList,
				).WithPosition(rhsExpr.GetPos())
			}
		}
		if len(list.Value) != len(step.LValues) {
			return nil, lang.NewRuntimeError(
				lang.ErrorCodeCountMismatch,
				fmt.Sprintf("LHS count %d doesn't match RHS list length %d",
					len(step.LValues), len(list.Value)),
				lang.ErrAssignCountMismatch,
			).WithPosition(step.GetPos())
		}
		for idx, lval := range step.LValues {
			if err := i.setSingleLValue(lval, list.Value[idx]); err != nil {
				return nil, err
			}
		}
		return rhsValue, nil
	}

	if err := i.setSingleLValue(step.LValues[0], rhsValue); err != nil {
		return nil, err
	}
	return rhsValue, nil
}

// setSingleLValue handles the core logic for all assignments.
func (i *Interpreter) setSingleLValue(lvalueExpr ast.Expression, rhsValue lang.Value) error {
	lval, ok := lvalueExpr.(*ast.LValueNode)
	if !ok {
		return lang.NewRuntimeError(
			lang.ErrorCodeInternal,
			"LValue expression is not an ast.LValueNode", nil,
		).WithPosition(lvalueExpr.GetPos())
	}

	if len(lval.Accessors) == 0 {
		return i.SetVariable(lval.Identifier, rhsValue)
	}

	baseVarName := lval.Identifier
	orig, exists := i.GetVariable(baseVarName)

	if !exists || isNil(orig) || !isContainer(orig) {
		newStructure, err := i.buildNestedStructure(lval.Accessors, rhsValue)
		if err != nil {
			return err
		}
		return i.SetVariable(baseVarName, newStructure)
	}

	clone := deepCloneValue(orig)
	if err := i.traverseAndSet(clone, lval.Accessors, rhsValue); err != nil {
		return err
	}
	return i.SetVariable(baseVarName, clone)
}

// traverseAndSet navigates a pre-existing container structure.
func (i *Interpreter) traverseAndSet(current lang.Value, accessors []*ast.AccessorNode, valueToPlace lang.Value) error {
	// ... (implementation remains the same)
	accessor := accessors[0]
	remainingAccessors := accessors[1:]
	isLast := len(remainingAccessors) == 0

	switch c := current.(type) {
	case *lang.MapValue:
		key, err := i.evaluateAccessorKey(accessor)
		if err != nil {
			return err
		}
		if isLast {
			c.Value[key] = valueToPlace
			return nil
		}
		child, exists := c.Value[key]
		if !exists || !isContainer(child) {
			newBranch, err := i.buildNestedStructure(remainingAccessors, valueToPlace)
			if err != nil {
				return err
			}
			c.Value[key] = newBranch
			return nil
		}
		return i.traverseAndSet(child, remainingAccessors, valueToPlace)

	case *lang.ListValue:
		index, err := i.evaluateAccessorIndex(accessor)
		if err != nil {
			return err
		}
		c.Value = padList(c.Value, index)
		if isLast {
			c.Value[index] = valueToPlace
			return nil
		}
		child := c.Value[index]
		if !isContainer(child) {
			newBranch, err := i.buildNestedStructure(remainingAccessors, valueToPlace)
			if err != nil {
				return err
			}
			c.Value[index] = newBranch
			return nil
		}
		return i.traverseAndSet(child, remainingAccessors, valueToPlace)
	}
	return lang.NewRuntimeError(lang.ErrorCodeInternal, "traverseAndSet called on a non-container", nil)
}

// buildNestedStructure recursively creates the necessary maps and lists for vivification.
func (i *Interpreter) buildNestedStructure(accessors []*ast.AccessorNode, finalValue lang.Value) (lang.Value, error) {
	// ... (implementation remains the same)
	if len(accessors) == 0 {
		return finalValue, nil
	}
	accessor := accessors[0]
	innerStructure, err := i.buildNestedStructure(accessors[1:], finalValue)
	if err != nil {
		return nil, err
	}
	containerType, err := i.determineContainerType(accessor)
	if err != nil {
		return nil, err
	}
	if containerType == LHS_MAP {
		key, err := i.evaluateAccessorKey(accessor)
		if err != nil {
			return nil, err
		}
		return lang.NewMapValue(map[string]lang.Value{key: innerStructure}), nil
	}
	index, err := i.evaluateAccessorIndex(accessor)
	if err != nil {
		return nil, err
	}
	list := padList(make([]lang.Value, 0), index)
	list[index] = innerStructure
	return &lang.ListValue{Value: list}, nil
}

// --- Helpers ---

func (i *Interpreter) determineContainerType(accessor *ast.AccessorNode) (LHSType, error) {
	if accessor.Type == ast.DotAccess {
		return LHS_MAP, nil
	}
	keyVal, err := eval.Expression(i, accessor.Key)
	if err != nil {
		return 0, lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating accessor key for type determination")
	}
	if _, isInt := lang.ToInt64(keyVal); isInt {
		return LHS_LIST, nil
	}
	return LHS_MAP, nil
}

func (i *Interpreter) evaluateAccessorKey(accessor *ast.AccessorNode) (string, error) {
	if accessor.Type == ast.DotAccess {
		if strLiteral, ok := accessor.Key.(*ast.StringLiteralNode); ok {
			return strLiteral.Value, nil
		}
		return strings.TrimPrefix(accessor.Key.String(), "."), nil
	}
	keyVal, err := eval.Expression(i, accessor.Key)
	if err != nil {
		return "", lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating map key")
	}
	key, _ := lang.ToString(keyVal)
	return key, nil
}

func (i *Interpreter) evaluateAccessorIndex(accessor *ast.AccessorNode) (int64, error) {
	indexVal, err := eval.Expression(i, accessor.Key)
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

// ... other helpers (padList, deepCloneValue, isNil, isContainer) remain the same ...
func padList(list []lang.Value, requiredIndex int64) []lang.Value {
	for int64(len(list)) <= requiredIndex {
		list = append(list, &lang.NilValue{})
	}
	return list
}

func deepCloneValue(v lang.Value) lang.Value {
	switch t := v.(type) {
	case *lang.MapValue:
		if t == nil {
			return nil
		}
		nm := make(map[string]lang.Value, len(t.Value))
		for k, val := range t.Value {
			nm[k] = deepCloneValue(val)
		}
		return lang.NewMapValue(nm)
	case *lang.ListValue:
		if t == nil {
			return nil
		}
		nl := make([]lang.Value, len(t.Value))
		for i, val := range t.Value {
			nl[i] = deepCloneValue(val)
		}
		return &lang.ListValue{Value: nl}
	default:
		return v
	}
}

func isNil(v lang.Value) bool {
	_, ok := v.(*lang.NilValue)
	return v == nil || ok
}

func isContainer(v lang.Value) bool {
	switch v.(type) {
	case *lang.MapValue, lang.MapValue, *lang.ListValue, lang.ListValue:
		return true
	default:
		return false
	}
}

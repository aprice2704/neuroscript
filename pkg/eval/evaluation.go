// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 8
// :: description: Updated evaluateMapLiteral to evaluate key expressions at runtime.
// :: latestChange: Support for dynamic keys in map literals.
// :: filename: pkg/eval/evaluation.go
// :: serialization: go

package eval

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// evaluation holds the state for an evaluation pass.
type evaluation struct {
	rt Runtime
}

// Expression is the main entry point for evaluating any expression node.
func (e *evaluation) Expression(node ast.Expression) (lang.Value, error) {
	if node == nil {
		return &lang.NilValue{}, nil
	}
	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return lang.StringValue{Value: n.Value}, nil
	case *ast.NumberLiteralNode:
		val, _ := lang.ToFloat64(n.Value)
		return lang.NumberValue{Value: val}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return &lang.NilValue{}, nil
	case *ast.ListLiteralNode:
		return e.evaluateListLiteral(n)
	case *ast.MapLiteralNode:
		return e.evaluateMapLiteral(n)
	case *ast.VariableNode:
		return e.evaluateIdentifier(n)
	case *ast.BinaryOpNode:
		return e.evaluateBinaryOp(n)
	case *ast.UnaryOpNode:
		return e.evaluateUnaryOp(n)
	case *ast.CallableExprNode:
		return e.evaluateCall(n)
	case *ast.LValueNode:
		return e.evaluateLValue(n)
	case *ast.ElementAccessNode: // FIX: Added missing case
		return e.evaluateElementAccess(n)
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("unhandled expression node type: %T", n), lang.ErrInternal).WithPosition(n.GetPos())
	}
}

// evaluateElementAccess handles expressions like `myList[0]` or `myMap["key"]`.
func (e *evaluation) evaluateElementAccess(n *ast.ElementAccessNode) (lang.Value, error) {
	collectionVal, err := e.Expression(n.Collection)
	if err != nil {
		return nil, err
	}

	accessorVal, err := e.Expression(n.Accessor)
	if err != nil {
		return nil, err
	}

	switch coll := collectionVal.(type) {
	case lang.ListValue:
		index, ok := lang.ToInt64(accessorVal)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "list index must be an integer", lang.ErrListInvalidIndexType).WithPosition(n.GetPos())
		}
		if index < 0 || int(index) >= len(coll.Value) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeBounds, "list index out of bounds", lang.ErrListIndexOutOfBounds).WithPosition(n.GetPos())
		}
		return coll.Value[int(index)], nil
	case *lang.ListValue: // ADDED
		index, ok := lang.ToInt64(accessorVal)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "list index must be an integer", lang.ErrListInvalidIndexType).WithPosition(n.GetPos())
		}
		if coll == nil || index < 0 || int(index) >= len(coll.Value) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeBounds, "list index out of bounds", lang.ErrListIndexOutOfBounds).WithPosition(n.GetPos())
		}
		return coll.Value[int(index)], nil
	case lang.MapValue: // ADDED
		key, _ := lang.ToString(accessorVal)
		val, ok := coll.Value[key]
		if !ok {
			return &lang.NilValue{}, nil // Accessing a non-existent key returns nil
		}
		return val, nil
	case *lang.MapValue:
		key, _ := lang.ToString(accessorVal)
		if coll == nil {
			return &lang.NilValue{}, nil // Accessing key on nil map returns nil
		}
		val, ok := coll.Value[key]
		if !ok {
			return &lang.NilValue{}, nil // Accessing a non-existent key returns nil
		}
		return val, nil
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("cannot access elements on type %s", lang.TypeOf(coll)), lang.ErrInvalidOperation).WithPosition(n.GetPos())
	}
}

func (e *evaluation) evaluateListLiteral(node *ast.ListLiteralNode) (lang.Value, error) {
	items := make([]lang.Value, len(node.Elements))
	for i, itemExpr := range node.Elements {
		val, err := e.Expression(itemExpr)
		if err != nil {
			return nil, err
		}
		items[i] = val
	}
	return lang.ListValue{Value: items}, nil
}

func (e *evaluation) evaluateMapLiteral(node *ast.MapLiteralNode) (lang.Value, error) {
	m := make(map[string]lang.Value)
	for _, pair := range node.Entries {
		// Evaluate Key
		keyVal, err := e.Expression(pair.Key)
		if err != nil {
			return nil, lang.WrapErrorWithPosition(err, pair.Key.GetPos(), "evaluating map key")
		}
		keyStr, _ := lang.ToString(keyVal)

		// Evaluate Value
		val, err := e.Expression(pair.Value)
		if err != nil {
			return nil, lang.WrapErrorWithPosition(err, pair.Value.GetPos(), "evaluating map value")
		}

		m[keyStr] = val
	}
	return lang.MapValue{Value: m}, nil
}

func (e *evaluation) evaluateIdentifier(node *ast.VariableNode) (lang.Value, error) {
	val, exists := e.rt.GetVariable(node.Name)
	if !exists {
		// As per our discussion, tool.context.get_payload() returns a MapValue.
		// If a variable is not found, it should be nil, not an error.
		return &lang.NilValue{}, nil
	}
	return val, nil
}

func (e *evaluation) evaluateBinaryOp(node *ast.BinaryOpNode) (lang.Value, error) {
	left, err := e.Expression(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := e.Expression(node.Right)
	if err != nil {
		return nil, err
	}
	return lang.PerformBinaryOperation(node.Operator, left, right)
}

func (e *evaluation) evaluateUnaryOp(node *ast.UnaryOpNode) (lang.Value, error) {
	operand, err := e.Expression(node.Operand)
	if err != nil {
		return nil, err
	}
	return lang.PerformUnaryOperation(node.Operator, operand)
}

func (e *evaluation) evaluateCall(node *ast.CallableExprNode) (lang.Value, error) {
	args := make([]lang.Value, len(node.Arguments))
	for i, argExpr := range node.Arguments {
		val, err := e.Expression(argExpr)
		if err != nil {
			return nil, lang.WrapErrorWithPosition(err, argExpr.GetPos(), fmt.Sprintf("evaluating argument %d for call to '%s'", i+1, node.Target.Name))
		}
		args[i] = val
	}

	if isBuiltInFunction(node.Target.Name) {
		return e.evaluateBuiltInFunction(node.Target.Name, args, node.GetPos())
	}

	if node.Target.IsTool {
		toolName, err := resolveToolName(node)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "resolving tool name failed", err).WithPosition(node.GetPos())
		}
		namedArgs, err := e.mapArgsToSpec(toolName, args, node)
		if err != nil {
			return nil, err
		}
		return e.rt.ExecuteTool(toolName, namedArgs)
	}

	return e.rt.RunProcedure(node.Target.Name, args...)
}

func (e *evaluation) evaluateLValue(lval *ast.LValueNode) (lang.Value, error) {
	baseVar, exists := e.rt.GetVariable(lval.Identifier)
	if !exists {
		return &lang.NilValue{}, nil
	}

	current := baseVar
	for _, accessor := range lval.Accessors {
		switch c := current.(type) {
		case lang.MapValue: // ADDED
			key, err := e.evaluateAccessorKey(accessor)
			if err != nil {
				return nil, err
			}
			child, ok := c.Value[key]
			if !ok {
				return &lang.NilValue{}, nil
			}
			current = child
		case *lang.MapValue:
			key, err := e.evaluateAccessorKey(accessor)
			if err != nil {
				return nil, err
			}
			if c == nil {
				return &lang.NilValue{}, nil
			}
			child, ok := c.Value[key]
			if !ok {
				return &lang.NilValue{}, nil
			}
			current = child
		case lang.ListValue: // Corrected to handle value type
			index, err := e.evaluateAccessorIndex(accessor)
			if err != nil {
				return nil, err
			}
			if index < 0 || index >= int64(len(c.Value)) {
				return &lang.NilValue{}, nil
			}
			current = c.Value[index]
		case *lang.ListValue:
			index, err := e.evaluateAccessorIndex(accessor)
			if err != nil {
				return nil, err
			}
			if c == nil || index < 0 || index >= int64(len(c.Value)) {
				return &lang.NilValue{}, nil
			}
			current = c.Value[index]
		default:
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("cannot access field on non-container type %s", lang.TypeOf(current)), lang.ErrInvalidOperation).WithPosition(lval.GetPos())
		}
	}
	return current, nil
}

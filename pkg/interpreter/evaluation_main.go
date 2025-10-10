// NeuroScript Version: 0.8.0
// File version: 68
// Purpose: FIX: Uses the ToolRegistry() accessor instead of the removed 'tools' field.
// filename: pkg/interpreter/evaluation_main.go
// nlines: 153
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var placeholderRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// evaluation holds the methods for evaluating AST expression nodes.
type evaluation struct {
	i *Interpreter
}

// Expression evaluates an AST node representing an expression.
func (e *evaluation) Expression(node ast.Expression) (lang.Value, error) {
	if node == nil {
		return &lang.NilValue{}, nil
	}

	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return e.evaluateStringLiteral(n)
	case *ast.NumberLiteralNode:
		val, _ := lang.ToFloat64(n.Value)
		return lang.NumberValue{Value: val}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return &lang.NilValue{}, nil
	case *ast.VariableNode:
		return e.i.resolveVariable(n)
	case *ast.PlaceholderNode:
		return e.i.resolvePlaceholder(n)
	case *ast.LastNode:
		return e.i.lastCallResult, nil
	case *ast.EvalNode:
		return e.i.lastCallResult, nil
	case *ast.ListLiteralNode:
		return e.evaluateListLiteral(n)
	case *ast.MapLiteralNode:
		return e.evaluateMapLiteral(n)
	case *ast.UnaryOpNode:
		operandVal, err := e.Expression(n.Operand)
		if err != nil {
			return nil, err
		}
		return e.i.EvaluateUnaryOp(n.Operator, operandVal)
	case *ast.BinaryOpNode:
		leftVal, errL := e.Expression(n.Left)
		if errL != nil {
			return nil, errL
		}
		rightVal, errR := e.Expression(n.Right)
		if errR != nil {
			return nil, errR
		}
		return e.i.EvaluateBinaryOp(leftVal, rightVal, n.Operator)
	case *ast.TypeOfNode:
		return e.evaluateTypeOf(n)
	case *ast.CallableExprNode:
		return e.evaluateCall(n)
	case *ast.ElementAccessNode:
		return e.i.evaluateElementAccess(n)
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("unhandled expression type: %T", node), nil)
	}
}

func (e *evaluation) evaluateStringLiteral(n *ast.StringLiteralNode) (lang.Value, error) {
	if n.IsRaw {
		resolvedStr, resolveErr := e.i.resolvePlaceholdersWithError(n.Value)
		if resolveErr != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeEvaluation, "evaluating raw string literal", resolveErr).WithPosition(n.StartPos)
		}
		return lang.StringValue{Value: resolvedStr}, nil
	}
	return lang.StringValue{Value: n.Value}, nil
}

func (i *Interpreter) resolveVariable(n *ast.VariableNode) (lang.Value, error) {
	if val, exists := i.GetVariable(n.Name); exists {
		return val, nil
	}

	if proc, procExists := i.KnownProcedures()[n.Name]; procExists {
		return lang.FunctionValue{Value: proc}, nil
	}
	if tool, toolExists := i.ToolRegistry().GetTool(types.FullName(n.Name)); toolExists {
		return lang.ToolValue{Value: &tool}, nil
	}
	if typeVal, typeExists := GetTypeConstant(n.Name); typeExists {
		return lang.StringValue{Value: typeVal}, nil
	}
	if isBuiltInFunction(n.Name) {
		return lang.StringValue{Value: fmt.Sprintf("<built-in function: %s>", n.Name)}, nil
	}
	return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("variable or function '%s' not found", n.Name), lang.ErrVariableNotFound).WithPosition(n.StartPos)
}

func (i *Interpreter) resolvePlaceholder(n *ast.PlaceholderNode) (lang.Value, error) {
	var refValue lang.Value
	var exists bool
	if n.Name == "LAST" {
		refValue = i.lastCallResult
		exists = refValue != nil
	} else {
		refValue, exists = i.GetVariable(n.Name)
	}
	if !exists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' for placeholder not found", n.Name), lang.ErrVariableNotFound).WithPosition(n.StartPos)
	}
	return refValue, nil
}

func (e *evaluation) evaluateListLiteral(n *ast.ListLiteralNode) (lang.Value, error) {
	evaluatedElements := make([]lang.Value, len(n.Elements))
	for idx, elemNode := range n.Elements {
		var err error
		evaluatedElements[idx], err = e.Expression(elemNode)
		if err != nil {
			return nil, err
		}
	}
	return lang.ListValue{Value: evaluatedElements}, nil
}

func (e *evaluation) evaluateMapLiteral(n *ast.MapLiteralNode) (lang.Value, error) {
	evaluatedMap := make(map[string]lang.Value)
	for _, entry := range n.Entries {
		mapKey := entry.Key.Value
		elemVal, err := e.Expression(entry.Value)
		if err != nil {
			return nil, err
		}
		evaluatedMap[mapKey] = elemVal
	}
	return lang.NewMapValue(evaluatedMap), nil
}

func (e *evaluation) evaluateTypeOf(n *ast.TypeOfNode) (lang.Value, error) {
	argValue, err := e.Expression(n.Argument)
	if err != nil {
		if errors.Is(err, lang.ErrVariableNotFound) {
			argValue = &lang.NilValue{}
		} else {
			return nil, err
		}
	}
	return lang.StringValue{Value: string(lang.TypeOf(argValue))}, nil
}

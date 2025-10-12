// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/eval_test.go
// nlines: 100+
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestExpressionASTGeneral(t *testing.T) {
	vars := map[string]lang.Value{
		"name":     lang.StringValue{Value: "World"},
		"greeting": lang.StringValue{Value: "Hello World"},
		"numVar":   lang.NumberValue{Value: 123},
		"boolProp": lang.BoolValue{Value: true},
		"listVar": lang.ListValue{Value: []lang.Value{
			lang.StringValue{Value: "x"},
			lang.NumberValue{Value: 99},
		}},
		"mapVar": lang.NewMapValue(map[string]lang.Value{
			"mKey": lang.StringValue{Value: "mVal World"},
			"mNum": lang.NumberValue{Value: 1},
		}),
		"nilVar": &lang.NilValue{},
		"numStr": lang.StringValue{Value: "456"},
	}

	dummyPos := &types.Position{Line: 1, Column: 1, File: "test"}
	bp := ast.BaseNode{StartPos: dummyPos}

	tests := []localEvalTestCase{
		{
			Name:        "String Literal",
			InputNode:   &ast.StringLiteralNode{BaseNode: bp, Value: "Hello World"},
			InitialVars: vars,
			Expected:    lang.StringValue{Value: "Hello World"},
		},
		{
			Name:        "Variable String",
			InputNode:   &ast.VariableNode{BaseNode: bp, Name: "greeting"},
			InitialVars: vars,
			Expected:    lang.StringValue{Value: "Hello World"},
		},
		{
			Name: "Concat with Number",
			InputNode: &ast.BinaryOpNode{
				BaseNode: bp,
				Left:     &ast.StringLiteralNode{BaseNode: bp, Value: "Count: "},
				Operator: "+",
				Right:    &ast.VariableNode{BaseNode: bp, Name: "numVar"},
			},
			InitialVars: vars,
			Expected:    lang.StringValue{Value: "Count: 123"},
		},
		{
			Name:        "Concat Error Operand",
			InputNode:   &ast.BinaryOpNode{BaseNode: bp, Left: &ast.StringLiteralNode{BaseNode: bp, Value: "Val: "}, Operator: "+", Right: &ast.VariableNode{BaseNode: bp, Name: "missing"}},
			InitialVars: vars,
			WantErr:     false, // Undefined variables are nil, which is a valid concat operand
			Expected:    lang.StringValue{Value: "Val: "},
		},
		{
			Name: "Add Numbers",
			InputNode: &ast.BinaryOpNode{
				BaseNode: bp,
				Left:     &ast.NumberLiteralNode{Value: 5.0},
				Operator: "+",
				Right:    &ast.NumberLiteralNode{Value: 3.0},
			},
			InitialVars: vars,
			Expected:    lang.NumberValue{Value: 8},
		},
		{
			Name: "Simple List",
			InputNode: &ast.ListLiteralNode{
				BaseNode: bp,
				Elements: []ast.Expression{
					&ast.NumberLiteralNode{Value: 1.0},
					&ast.StringLiteralNode{Value: "World"},
					&ast.VariableNode{Name: "boolProp"},
				},
			},
			InitialVars: vars,
			Expected: lang.ListValue{Value: []lang.Value{
				lang.NumberValue{Value: 1},
				lang.StringValue{Value: "World"},
				lang.BoolValue{Value: true},
			}},
		},
	}

	for _, tt := range tests {
		runLocalExpressionTest(t, tt)
	}
}

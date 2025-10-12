// NeuroScript Version: 0.8.0
// File version: 2.0.0
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/unary_test.go
// nlines: 75
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestUnaryOpSomeAndNo(t *testing.T) {
	dummyPos := &types.Position{Line: 1, Column: 1, File: "test"}
	bp := ast.BaseNode{StartPos: dummyPos}

	vars := map[string]lang.Value{
		"nonEmptyString": lang.StringValue{Value: "hello"},
		"emptyString":    lang.StringValue{Value: ""},
		"zeroNum":        lang.NumberValue{Value: 0},
		"nonZeroNum":     lang.NumberValue{Value: 5},
		"trueBool":       lang.BoolValue{Value: true},
		"falseBool":      lang.BoolValue{Value: false},
		"nilVal":         &lang.NilValue{},
		"emptyList":      lang.ListValue{Value: []lang.Value{}},
		"nonEmptyList":   lang.ListValue{Value: []lang.Value{lang.NumberValue{Value: 1}}},
	}

	testCases := []localEvalTestCase{
		{
			Name:        "some non-empty string",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "nonEmptyString"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name:        "some empty string",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "emptyString"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},
		{
			Name:        "no nil",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "no", Operand: &ast.VariableNode{Name: "nilVal"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name:        "no non-empty string",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "no", Operand: &ast.VariableNode{Name: "nonEmptyString"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},
	}

	for _, tt := range testCases {
		runLocalExpressionTest(t, tt)
	}
}

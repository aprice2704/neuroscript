// NeuroScript Version: 0.5.2
// File version: 1.0.0
// Purpose: Adds dedicated tests for the 'some' and 'no' unary operators, including checks for nil inputs.
// filename: pkg/interpreter/evaluation_unary_test.go
// nlines: 75
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
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
		"emptyList":      lang.NewListValue([]lang.Value{}),
		"nonEmptyList":   lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}}),
	}

	testCases := []testutil.EvalTestCase{
		// --- 'some' operator ---
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
			Name:        "some non-zero number",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "nonZeroNum"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name:        "some zero number",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "zeroNum"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},
		{
			Name:        "some nil",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "nilVal"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},
		{
			Name:        "some non-empty list",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "nonEmptyList"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name:        "some empty list",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "some", Operand: &ast.VariableNode{Name: "emptyList"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},

		// --- 'no' operator ---
		{
			Name:        "no nil",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "no", Operand: &ast.VariableNode{Name: "nilVal"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name:        "no empty string",
			InputNode:   &ast.UnaryOpNode{BaseNode: bp, Operator: "no", Operand: &ast.VariableNode{Name: "emptyString"}},
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
		t.Run(tt.Name, func(t *testing.T) {
			testutil.ExpressionTest(t, tt)
		})
	}
}

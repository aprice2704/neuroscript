// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: "Tiger Tests" to attack weak points in the eval package.
// filename: pkg/eval/tiger_test.go
// nlines: 125
// risk_rating: HIGH

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestAccessTiger_IncompatibleAccess attacks element access with wrong accessor types.
func TestAccessTiger_IncompatibleAccess(t *testing.T) {
	initialVars := map[string]lang.Value{
		"myList": lang.ListValue{Value: []lang.Value{
			lang.StringValue{Value: "apple"},
		}},
		"myMap": lang.MapValue{Value: map[string]lang.Value{
			"0":    lang.StringValue{Value: "zero_key"},
			"key1": lang.StringValue{Value: "value1_val"},
			"true": lang.StringValue{Value: "true_key"},
			"":     lang.StringValue{Value: "nil_key"},
		}},
		"nilVar": &lang.NilValue{},
	}

	testCases := []localEvalTestCase{
		{
			Name:            "Access List with String Key",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.StringLiteralNode{Value: "key"}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrListInvalidIndexType,
		},
		{
			Name:        "Access Map with Numeric Key",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMap"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "zero_key"}, // Should convert 0 to "0"
		},
		{
			Name:        "Access Map with Boolean Key",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMap"}, Accessor: &ast.BooleanLiteralNode{Value: true}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "true_key"}, // Should convert true to "true"
		},
		{
			Name:        "Access Map with Nil Key",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMap"}, Accessor: &ast.NilLiteralNode{}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "nil_key"}, // Should convert nil to ""
		},
		{
			Name:            "Access Nil as List",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "nilVar"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrInvalidOperation,
		},
		{
			Name:            "Access Nil as Map",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "nilVar"}, Accessor: &ast.StringLiteralNode{Value: "key"}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrInvalidOperation,
		},
		{
			Name:            "Access String as List",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.StringLiteralNode{Value: "hello"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrInvalidOperation,
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}

// TestBuiltInTiger_Len attacks the 'len' built-in with non-collection types.
func TestBuiltInTiger_Len(t *testing.T) {
	testCases := []localEvalTestCase{
		{
			Name:      "len(123)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "len"}, Arguments: []ast.Expression{&ast.NumberLiteralNode{Value: 123}}},
			Expected:  lang.NumberValue{Value: 1}, // len of a scalar is 1
		},
		{
			Name:      "len(true)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "len"}, Arguments: []ast.Expression{&ast.BooleanLiteralNode{Value: true}}},
			Expected:  lang.NumberValue{Value: 1}, // len of a scalar is 1
		},
		{
			Name:      "len(nil)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "len"}, Arguments: []ast.Expression{&ast.NilLiteralNode{}}},
			Expected:  lang.NumberValue{Value: 0}, // len of nil is 0
		},
	}
	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}

// TestBuiltInTiger_TypeChecks attacks the 'is_...' functions with tricky inputs.
func TestBuiltInTiger_TypeChecks(t *testing.T) {
	testCases := []localEvalTestCase{
		{
			Name:      "is_list(nil)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "is_list"}, Arguments: []ast.Expression{&ast.NilLiteralNode{}}},
			Expected:  lang.BoolValue{Value: false},
		},
		{
			Name:      "is_map(nil)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "is_map"}, Arguments: []ast.Expression{&ast.NilLiteralNode{}}},
			Expected:  lang.BoolValue{Value: false},
		},
		{
			Name:      "is_nil(123)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "is_nil"}, Arguments: []ast.Expression{&ast.NumberLiteralNode{Value: 123}}},
			Expected:  lang.BoolValue{Value: false},
		},
		{
			Name:      "is_nil(nil)",
			InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "is_nil"}, Arguments: []ast.Expression{&ast.NilLiteralNode{}}},
			Expected:  lang.BoolValue{Value: true},
		},
	}
	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}

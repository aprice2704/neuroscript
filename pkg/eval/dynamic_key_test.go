// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: New test suite specifically for dynamic map keys (variables, expressions) to prevent regression.
// :: filename: pkg/eval/dynamic_key_test.go
// :: serialization: go

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestDynamicMapKeys(t *testing.T) {
	// Setup variables for the test context
	vars := map[string]lang.Value{
		"myKeyVar": lang.StringValue{Value: "resolved_key"},
		"myValVar": lang.NumberValue{Value: 42},
		"prefix":   lang.StringValue{Value: "item_"},
		"id":       lang.NumberValue{Value: 5},
	}

	testCases := []localEvalTestCase{
		{
			Name: "Variable as Map Key",
			// { myKeyVar: "constant_value" } -> { "resolved_key": "constant_value" }
			InputNode: &ast.MapLiteralNode{
				Entries: []*ast.MapEntryNode{
					{
						Key:   &ast.VariableNode{Name: "myKeyVar"},
						Value: &ast.StringLiteralNode{Value: "constant_value"},
					},
				},
			},
			InitialVars: vars,
			Expected: lang.MapValue{Value: map[string]lang.Value{
				"resolved_key": lang.StringValue{Value: "constant_value"},
			}},
		},
		{
			Name: "Expression as Map Key",
			// { prefix + id : myValVar } -> { "item_5": 42 }
			InputNode: &ast.MapLiteralNode{
				Entries: []*ast.MapEntryNode{
					{
						Key: &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "prefix"},
							Operator: "+",
							Right:    &ast.VariableNode{Name: "id"},
						},
						Value: &ast.VariableNode{Name: "myValVar"},
					},
				},
			},
			InitialVars: vars,
			Expected: lang.MapValue{Value: map[string]lang.Value{
				"item_5": lang.NumberValue{Value: 42},
			}},
		},
		{
			Name: "Mixed Static and Dynamic Keys",
			// { "static": 1, myKeyVar: 2 }
			InputNode: &ast.MapLiteralNode{
				Entries: []*ast.MapEntryNode{
					{
						Key:   &ast.StringLiteralNode{Value: "static"},
						Value: &ast.NumberLiteralNode{Value: 1},
					},
					{
						Key:   &ast.VariableNode{Name: "myKeyVar"},
						Value: &ast.NumberLiteralNode{Value: 2},
					},
				},
			},
			InitialVars: vars,
			Expected: lang.MapValue{Value: map[string]lang.Value{
				"static":       lang.NumberValue{Value: 1},
				"resolved_key": lang.NumberValue{Value: 2},
			}},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}

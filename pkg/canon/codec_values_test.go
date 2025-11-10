// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Fixes 'KindUnknown' error in unit test by initializing all nil slices in test data.
// filename: pkg/canon/codec_values_test.go
// nlines: 120+

package canon

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestValueToNodeToValueRoundtrip verifies that complex lang.Value types
// can be converted to AST nodes and back without loss of data.
func TestValueToNodeToValueRoundtrip(t *testing.T) {
	originalValue := lang.MapValue{Value: map[string]lang.Value{
		"string_key": lang.StringValue{Value: "hello"},
		"number_key": lang.NumberValue{Value: 123.45},
		"bool_key":   lang.BoolValue{Value: true},
		"nil_key":    lang.NilValue{},
		"list_key": lang.ListValue{Value: []lang.Value{
			lang.StringValue{Value: "nested"},
			lang.NumberValue{Value: 99},
		}},
		"map_key": lang.MapValue{Value: map[string]lang.Value{
			"nested_key": lang.StringValue{Value: "deep"},
		}},
	}}

	// 1. Convert Value -> Node
	node, err := ValueToNode(originalValue)
	if err != nil {
		t.Fatalf("valueToNode failed: %v", err)
	}

	// 2. Convert Node -> Value
	roundtrippedValue, err := NodeToValue(node)
	if err != nil {
		t.Fatalf("nodeToValue failed: %v", err)
	}

	// 3. Compare
	if diff := cmp.Diff(originalValue, roundtrippedValue, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Value-to-Node roundtrip failed (-original +roundtripped):\n%s", diff)
	}
}

// TestCanonicaliseNodeRoundtrip verifies that minimal AST nodes can be
// serialized and deserialized correctly.
func TestCanonicaliseNodeRoundtrip(t *testing.T) {
	t.Run("StringLiteralNode", func(t *testing.T) {
		originalNode := &ast.StringLiteralNode{
			BaseNode: ast.BaseNode{NodeKind: types.KindStringLiteral},
			Value:    "hello world",
		}

		blob, _, err := CanonicaliseNode(originalNode)
		if err != nil {
			t.Fatalf("CanonicaliseNode failed: %v", err)
		}

		decodedNode, err := DecodeNode(blob)
		if err != nil {
			t.Fatalf("DecodeNode failed: %v", err)
		}

		cmpOpts := []cmp.Option{
			cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
		}
		if diff := cmp.Diff(originalNode, decodedNode, cmpOpts...); diff != "" {
			t.Errorf("StringLiteralNode roundtrip failed (-original +decoded):\n%s", diff)
		}
	})

	t.Run("ProcedureNode (Minimal)", func(t *testing.T) {
		// --- FIX: Initialize all nil slices to empty slices ---
		// This ensures consistency, as the decoder always creates empty slices.
		originalNode := &ast.Procedure{
			BaseNode:       ast.BaseNode{NodeKind: types.KindProcedureDecl},
			Metadata:       make(map[string]string),
			Comments:       make([]*ast.Comment, 0),
			RequiredParams: []string{"a"},
			OptionalParams: make([]*ast.ParamSpec, 0),
			ReturnVarNames: make([]string, 0),
			ErrorHandlers:  make([]*ast.Step, 0),
			Steps: []ast.Step{
				{
					BaseNode: ast.BaseNode{NodeKind: types.KindStep},
					Type:     "return",
					Values: []ast.Expression{&ast.VariableNode{
						BaseNode: ast.BaseNode{NodeKind: types.KindVariable},
						Name:     "a",
					}},
					// Initialize nil slices for Step
					LValues:  make([]*ast.LValueNode, 0),
					Body:     make([]ast.Step, 0),
					ElseBody: make([]ast.Step, 0),
					Comments: make([]*ast.Comment, 0),
				},
			},
		}
		originalNode.SetName("my_func")
		// --- END FIX ---

		blob, _, err := CanonicaliseNode(originalNode)
		if err != nil {
			t.Fatalf("CanonicaliseNode failed: %v", err)
		}

		decodedNode, err := DecodeNode(blob)
		if err != nil {
			t.Fatalf("DecodeNode failed: %v", err)
		}

		cmpOpts := []cmp.Option{
			cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
			cmpopts.IgnoreUnexported(ast.Procedure{}, ast.Step{}),
			cmpopts.EquateEmpty(),
		}
		if diff := cmp.Diff(originalNode, decodedNode, cmpOpts...); diff != "" {
			t.Errorf("ProcedureNode roundtrip failed (-original +decoded):\n%s", diff)
		}
	})
}

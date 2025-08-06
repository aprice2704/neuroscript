// filename: pkg/canon/step_serialization_test.go
// NeuroScript Version: 0.6.2
// File version: 3
// Purpose: FIX: Correctly initializes NodeKind for all manually created AST nodes to prevent serialization errors.
// nlines: 60+
// risk_rating: HIGH

package canon

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/crypto/blake2b"
)

// TestStepSerializationForLoop proves that the encoder/decoder cycle for an
// ast.Step of type "for" is now handled correctly. It constructs the node,
// serializes it, deserializes it, and confirms it matches.
func TestStepSerializationForLoop(t *testing.T) {
	// 1. Manually construct the 'for' loop AST node with proper NodeKinds.
	originalStep := &ast.Step{
		BaseNode:    ast.BaseNode{NodeKind: types.KindStep},
		Type:        "for",
		LoopVarName: "item",
		Collection:  &ast.VariableNode{BaseNode: ast.BaseNode{NodeKind: types.KindVariable}, Name: "items"},
		Body: []ast.Step{
			{
				BaseNode: ast.BaseNode{NodeKind: types.KindStep},
				Type:     "emit",
				Values: []ast.Expression{&ast.VariableNode{
					BaseNode: ast.BaseNode{NodeKind: types.KindVariable},
					Name:     "item",
				}},
			},
		},
	}

	// 2. Canonicalize it.
	var buf bytes.Buffer
	hasher, _ := blake2b.New256(nil)
	encoder := &canonVisitor{w: &buf, hasher: hasher}
	if err := encoder.visit(originalStep); err != nil {
		t.Fatalf("encoder.visit() failed: %v", err)
	}

	// 3. Decode it.
	decoder := &canonReader{r: bytes.NewReader(buf.Bytes())}
	decodedNode, err := decoder.readNode()
	if err != nil {
		t.Fatalf("decoder.readNode() failed: %v", err)
	}
	decodedStep, ok := decodedNode.(*ast.Step)
	if !ok {
		t.Fatalf("Decoded node is not *ast.Step, but %T", decodedNode)
	}

	// 4. Compare the original and decoded steps.
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreUnexported(ast.Step{}, ast.BaseNode{}, ast.VariableNode{}),
		cmpopts.EquateEmpty(),
	}
	if diff := cmp.Diff(originalStep, decodedStep, cmpOpts...); diff != "" {
		t.Errorf("FAIL: The decoded step does not match the original:\n%s", diff)
	}
}

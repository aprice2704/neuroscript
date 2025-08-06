// filename: pkg/canon/step_serialization_test.go
// NeuroScript Version: 0.6.2
// File version: 2
// Purpose: Provides a direct unit test for the serialization of ast.Step, proving the "for" case is not handled. Corrected compiler error.
// nlines: 60
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
// ast.Step of type "for" is lossy. It directly constructs the Step node,
// serializes it, and deserializes it, confirming that critical fields like
// Collection, LoopVarName, and Body are dropped.
//
// This provides irrefutable, targeted proof that the canonVisitor.visitStep
// and canonReader.readStep functions are missing the required logic for 'for' loops.
func TestStepSerializationForLoop(t *testing.T) {
	// 1. Manually construct the 'for' loop AST node.
	originalStep := &ast.Step{
		BaseNode:    ast.BaseNode{NodeKind: types.KindStep},
		Type:        "for",
		LoopVarName: "item",
		Collection:  &ast.VariableNode{BaseNode: ast.BaseNode{NodeKind: types.KindVariable}, Name: "items"},
		Body: []ast.Step{
			{
				BaseNode: ast.BaseNode{NodeKind: types.KindStep},
				Type:     "emit",
				Values:   []ast.Expression{&ast.VariableNode{Name: "item"}},
			},
		},
	}

	// 2. Canonicalize it.
	var buf bytes.Buffer
	// FIX: Use the actual hasher, not a non-existent helper.
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
	}
	if diff := cmp.Diff(originalStep, decodedStep, cmpOpts...); diff == "" {
		t.Fatal("FAIL: Expected a difference between original and decoded step, but they were identical. The bug may have been fixed.")
	} else {
		t.Logf("SUCCESS: Found expected difference, proving bug. Diff (-original +decoded):\n%s", diff)
	}

	// 5. Explicitly assert the decoded fields are nil/empty.
	if decodedStep.Collection != nil {
		t.Error("BUG NOT REPRODUCED: Collection was not nil.")
	}
	if decodedStep.LoopVarName != "" {
		t.Error("BUG NOT REPRODUCED: LoopVarName was not empty.")
	}
	if len(decodedStep.Body) > 0 {
		t.Error("BUG NOT REPRODUCED: Body was not empty.")
	}
}

// filename: pkg/canon/control_flow_serialization_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Proves the canonicalization process drops all control flow and other unhandled statements.
// nlines: 75
// risk_rating: HIGH

package canon_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestUnhandledStatementSerialization proves that the canonicalization process is
// lossy for several critical statement types, including 'if', 'while', and 'must'.
// The visitStep and readStep functions are missing cases for these types, meaning
// their essential data (conditions, bodies, etc.) is never serialized.
//
// This test will fail until the encoder and decoder are updated to handle these
// and other missing statement types.
func TestUnhandledStatementSerialization(t *testing.T) {
	src := `
func main() means
    set x = 10
    if x > 5
        emit "x is greater than 5"
    endif

    while x > 0
        set x = x - 1
    endwhile

    must x == 0
endfunc
`
	// 1. Parse to get a known-good AST.
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse() failed unexpectedly: %v", err)
	}

	// 2. Run the full Canonicalise/Decode cycle.
	blob, _, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("api.Canonicalise() failed unexpectedly: %v", err)
	}
	decodedTree, _, err := api.Decode(blob)
	if err != nil {
		t.Fatalf("api.Decode() failed unexpectedly: %v", err)
	}

	// 3. Inspect the decoded AST.
	program := decodedTree.Root.(*ast.Program)
	proc := program.Procedures["main"]
	if len(proc.Steps) != 4 {
		t.Fatalf("Expected 4 steps in decoded procedure, but got %d", len(proc.Steps))
	}

	// 4. Assert that the decoded steps are corrupted (missing their data).
	ifStep := proc.Steps[1]
	if ifStep.Type != "if" {
		t.Errorf("Expected step 2 to be 'if', got '%s'", ifStep.Type)
	}
	if ifStep.Cond == nil {
		t.Error("BUG REPRODUCED: 'if' statement Cond is nil after decoding.")
	}
	if len(ifStep.Body) == 0 {
		t.Error("BUG REPRODUCED: 'if' statement Body is empty after decoding.")
	}

	whileStep := proc.Steps[2]
	if whileStep.Type != "while" {
		t.Errorf("Expected step 3 to be 'while', got '%s'", whileStep.Type)
	}
	if whileStep.Cond == nil {
		t.Error("BUG REPRODUCED: 'while' statement Cond is nil after decoding.")
	}
	if len(whileStep.Body) == 0 {
		t.Error("BUG REPRODUCED: 'while' statement Body is empty after decoding.")
	}

	mustStep := proc.Steps[3]
	if mustStep.Type != "must" {
		t.Errorf("Expected step 4 to be 'must', got '%s'", mustStep.Type)
	}
	if mustStep.Cond == nil {
		t.Error("BUG REPRODUCED: 'must' statement Cond is nil after decoding.")
	}
}

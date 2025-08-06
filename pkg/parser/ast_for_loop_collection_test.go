// filename: pkg/parser/for_loop_collection_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Adds a regression test to ensure that the collection in a 'for each' loop is correctly parsed as an expression, preventing a nil pointer in the AST.
// nlines: 45
// risk_rating: LOW

package parser

import (
	"testing"
)

// TestForEachLoop_CollectionIsExpression ensures that when a variable is used
// as the collection in a for-each loop, the AST builder correctly creates
// a non-nil 'Collection' expression node. A failure here indicates a regression
// of the bug where an *ast.LValueNode was not being converted to an
// *ast.VariableNode, leading to a nil field and a runtime panic.
func TestForEachLoop_CollectionIsExpression(t *testing.T) {
	script := `
func main() means
    set my_list = [1, 2, 3]
    for each item in my_list
        emit item
    endfor
endfunc
`
	prog := testParseAndBuild(t, script)
	if prog == nil {
		t.Fatal("testParseAndBuild returned a nil program")
	}

	proc, ok := prog.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in AST")
	}

	if len(proc.Steps) != 2 {
		t.Fatalf("Expected 2 steps in procedure, got %d", len(proc.Steps))
	}

	// The second step is the for loop
	loopStep := proc.Steps[1]
	if loopStep.Type != "for" {
		t.Fatalf("Expected the second step to be a 'for' loop, but got '%s'", loopStep.Type)
	}

	// THIS IS THE CRITICAL ASSERTION
	if loopStep.Collection == nil {
		t.Fatal("REGRESSION: The 'Collection' field of the 'for' loop step is nil.")
	}
}

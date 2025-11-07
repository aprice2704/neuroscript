// NeuroScript Version: 0.6.0
// File version: 6
// Purpose: Corrects test assertions to align with the new "next-node" comment association algorithm.
// filename: pkg/parser/file_header_comments_test.go
// nlines: 60
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestFileHeaderCommentAssociation(t *testing.T) {
	script := `
# This is a file-level comment.
-- This is also a file-level comment.

:: file-key: file-value

# This comment immediately precedes the function.
func main() means
	# This is a comment inside the function.
	set x = 1
endfunc
`
	t.Logf("--- Source Code for TestFileHeaderCommentAssociation ---\n%s\n------------------", script)
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)
	tree, tokenStream, errs := parserAPI.parseInternal("header_comment_test.ns", script)
	if len(errs) > 0 {
		t.Fatalf("Parse failed with errors: %v", errs)
	}

	builder := NewASTBuilder(logger)
	program, _, err := builder.BuildFromParseResult(tree, tokenStream)
	if err != nil {
		t.Fatalf("Build failed unexpectedly: %v", err)
	}

	// 1. Verify that file-level comments (before metadata) are no longer attached to the Program.
	if len(program.Comments) != 0 {
		t.Errorf("Expected 0 file-level comments on Program, but got %d", len(program.Comments))
	}

	// 2. Verify all comments before the first step belong to the Procedure node.
	proc, ok := program.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in AST")
	}

	if len(proc.Comments) != 3 {
		t.Errorf("Expected 3 comments associated with the procedure, but got %d", len(proc.Comments))
	}

	// 3. Verify that the step has the comment preceding it.
	if len(proc.Steps) != 1 {
		t.Fatalf("Expected 1 step in the procedure, got %d", len(proc.Steps))
	}
	step := proc.Steps[0]
	if len(step.Comments) != 1 {
		t.Errorf("Expected 1 comment associated with the first step, got %d", len(step.Comments))
	}
}

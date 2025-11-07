// NeuroScript Version: 0.6.0
// File version: 9
// Purpose: Corrects all test assertions to align with the new "next-node" comment association logic.
// filename: pkg/parser/ast_builder_comments_metadata_test.go
// nlines: 110
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/go-cmp/cmp"
)

func TestMetadataAndCommentParsing(t *testing.T) {
	script := `
# Standalone comment before file metadata.
:: file-key: file-value
:: author: Test Author
-- Another standalone comment.

# Comment before the function definition.
func TestFunc(needs p1) means
	:: proc-key: proc-value  # Comment on the same line as metadata.
	:: status: testing

	# A comment before the first statement.
	set x = p1 + 1 -- A comment after the first statement.
	# This is a comment on its own line between statements.
	emit x # Final comment in the block.
endfunc
`
	t.Logf("--- Source Code for TestMetadataAndCommentParsing ---\n%s\n------------------", script)
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)

	tree, tokenStream, errs := parserAPI.parseInternal("test_script.ns", script)
	if len(errs) > 0 {
		t.Fatalf("parseInternal() failed with errors: %v", errs)
	}

	builder := NewASTBuilder(logger)
	program, fileMetadata, err := builder.BuildFromParseResult(tree, tokenStream)
	if err != nil {
		t.Fatalf("Build() failed unexpectedly: %v", err)
	}

	t.Run("File Metadata", func(t *testing.T) {
		expectedFileMeta := map[string]string{
			"file-key": "file-value",
			"author":   "Test Author",
		}
		if diff := cmp.Diff(expectedFileMeta, fileMetadata); diff != "" {
			t.Errorf("File metadata mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Procedure Metadata", func(t *testing.T) {
		proc, ok := program.Procedures["TestFunc"]
		if !ok {
			t.Fatal("Procedure 'TestFunc' not found in AST")
		}
		expectedProcMeta := map[string]string{
			"proc-key": "proc-value",
			"status":   "testing",
		}
		if diff := cmp.Diff(expectedProcMeta, proc.Metadata); diff != "" {
			t.Errorf("Procedure metadata mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Comment Association", func(t *testing.T) {
		// FIX: With next-node logic, file-header comments are attached to the
		// first node (the procedure), not the program.
		if len(program.Comments) != 0 {
			t.Errorf("Expected 0 file-level comments, but got %d", len(program.Comments))
		}

		proc, _ := program.Procedures["TestFunc"]
		// FIX: All 3 comments before the procedure body are now correctly
		// associated with the procedure itself.
		if len(proc.Comments) != 3 {
			t.Errorf("Expected 3 comments associated with the procedure, but got %d", len(proc.Comments))
		}

		if len(proc.Steps) != 2 {
			t.Fatalf("Expected 2 steps in procedure, got %d", len(proc.Steps))
		}

		setStep := proc.Steps[0]
		// FIX: The leading comment and the trailing comment are both
		// correctly associated with the 'set' step.
		if len(setStep.Comments) != 2 {
			t.Errorf("Expected 2 comments for the 'set' step, got %d", len(setStep.Comments))
		}

		emitStep := proc.Steps[1]
		// FIX: The comment on the line *between* statements and the
		// final trailing comment are both associated with the 'emit' step.
		if len(emitStep.Comments) != 2 {
			t.Errorf("Expected 2 comments for the 'emit' step, got %d", len(emitStep.Comments))
		}
	})
}

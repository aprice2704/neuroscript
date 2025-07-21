// NeuroScript Version: 0.6.0
// File version: 6
// Purpose: Corrects test assertions to align with the new "always preceding" comment association algorithm.
// filename: pkg/parser/ast_builder_comments_metadata_test.go
// nlines: 105
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
		// All 3 comments before the function belong to the preceding Program node.
		if len(program.Comments) != 3 {
			t.Errorf("Expected 3 file-level comments, but got %d", len(program.Comments))
		}

		proc, _ := program.Procedures["TestFunc"]
		// The comment before the first statement belongs to the preceding Procedure node.
		if len(proc.Comments) != 1 {
			t.Errorf("Expected 1 comment associated with the procedure, but got %d", len(proc.Comments))
		}

		if len(proc.Steps) != 2 {
			t.Fatalf("Expected 2 steps in procedure, got %d", len(proc.Steps))
		}

		setStep := proc.Steps[0]
		// Trailing comment on same line + comment on next line belong to the 'set' step.
		if len(setStep.Comments) != 2 {
			t.Errorf("Expected 2 comments for the 'set' step, got %d", len(setStep.Comments))
		}

		emitStep := proc.Steps[1]
		// The final trailing comment belongs to the 'emit' step.
		if len(emitStep.Comments) != 1 {
			t.Errorf("Expected 1 comment for the 'emit' step, got %d", len(emitStep.Comments))
		}
	})
}

// filename: pkg/parser/ast_builder_positions_test.go
// NeuroScript Version: 0.6.0
// File version: 16
// Purpose: Updated tests to align with the new simple comment association algorithm by checking for total comment preservation instead of brittle counts and removing blank-line checks.

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/go-cmp/cmp"
)

// countTotalComments traverses the AST and counts all attached comments.
func countTotalComments(prog *ast.Program) int {
	count := len(prog.Comments)
	for _, proc := range prog.Procedures {
		count += len(proc.Comments)
		for _, step := range proc.Steps {
			count += len(step.Comments)
		}
	}
	for _, cmd := range prog.Commands {
		count += len(cmd.Comments)
		for _, step := range cmd.Body {
			count += len(step.Comments)
		}
	}
	for _, event := range prog.Events {
		count += len(event.Comments)
		for _, step := range event.Body {
			count += len(step.Comments)
		}
	}
	return count
}

func TestComprehensiveBlockParsing(t *testing.T) {
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)

	t.Run("Command Block Formatting", func(t *testing.T) {
		script := `
:: name: test_command

// File-level comment


// 2 blank lines before command
# Comment right before command
command // trailing command comment
    set x = 1
endcommand
`
		tree, tokenStream, errs := parserAPI.parseInternal("command_test.ns", script)
		if len(errs) > 0 {
			t.Fatalf("parseInternal() for command failed with errors: %v", errs)
		}

		builder := NewASTBuilder(logger)
		program, fileMetadata, err := builder.BuildFromParseResult(tree, tokenStream)
		if err != nil {
			t.Fatalf("Build() for command failed unexpectedly: %v", err)
		}

		expectedFileMeta := map[string]string{"name": "test_command"}
		if diff := cmp.Diff(expectedFileMeta, fileMetadata); diff != "" {
			t.Errorf("File metadata mismatch (-want +got):\n%s", diff)
		}

		expectedCommentCount := strings.Count(script, "//") + strings.Count(script, "#")
		actualCommentCount := countTotalComments(program)
		if actualCommentCount != expectedCommentCount {
			t.Errorf("Expected total comment count to be %d, but got %d", expectedCommentCount, actualCommentCount)
		}
	})

	t.Run("Function Block Formatting", func(t *testing.T) {
		script := `
:: a: b

# Standalone comment before func

// 1 blank line before func
func MyFunction(needs p1) means
    emit p1
endfunc
`
		tree, tokenStream, errs := parserAPI.parseInternal("function_test.ns", script)
		if len(errs) > 0 {
			t.Fatalf("parseInternal() for function failed: %v", errs)
		}
		builder := NewASTBuilder(logger)
		program, _, err := builder.BuildFromParseResult(tree, tokenStream)
		if err != nil {
			t.Fatalf("Build() for function failed: %v", err)
		}
		expectedCommentCount := strings.Count(script, "//") + strings.Count(script, "#")
		actualCommentCount := countTotalComments(program)
		if actualCommentCount != expectedCommentCount {
			t.Errorf("Expected total comment count to be %d, but got %d", expectedCommentCount, actualCommentCount)
		}
	})

	t.Run("Event Block Formatting", func(t *testing.T) {
		script := `
# another comment between blocks



// 3 blank lines before event
on event "test.event" do
    call MyFunction("hello")
endon
`
		tree, tokenStream, errs := parserAPI.parseInternal("event_test.ns", script)
		if len(errs) > 0 {
			t.Fatalf("parseInternal() for event failed: %v", errs)
		}
		builder := NewASTBuilder(logger)
		program, _, err := builder.BuildFromParseResult(tree, tokenStream)
		if err != nil {
			t.Fatalf("Build() for event failed: %v", err)
		}
		expectedCommentCount := strings.Count(script, "//") + strings.Count(script, "#")
		actualCommentCount := countTotalComments(program)
		if actualCommentCount != expectedCommentCount {
			t.Errorf("Expected total comment count to be %d, but got %d", expectedCommentCount, actualCommentCount)
		}
	})
}

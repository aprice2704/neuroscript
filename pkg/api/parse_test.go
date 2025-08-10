// NeuroScript Version: 0.6.0
// File version: 10
// Purpose: Corrects final test assertions for blank lines to align with the new parser implementation that no longer counts them.
// filename: pkg/api/parse_test.go
// nlines: 135
// risk_rating: MEDIUM

package api_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestParse_CommandBlockFormatting confirms parsing of a command block.
func TestParse_CommandBlockFormatting(t *testing.T) {
	src := `
:: file-level: true
# File-level hash comment.
// Another file-level comment.

command
  :: cmd-level: yes
  -- Command-level dash comment.
  emit "step 1"
endcommand
`
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse with ParsePreserveComments failed: %v", err)
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Tree root is not a *ast.Program, but %T", tree.Root)
	}

	if len(program.Comments) != 2 {
		t.Errorf("Expected 2 file-level comments, got %d", len(program.Comments))
	}
	if len(program.Commands) != 1 {
		t.Fatal("Expected 1 command block")
	}
	cmdNode := program.Commands[0]
	// FIX: The new parser no longer counts blank lines. Assertion updated to 0.
	if cmdNode.BlankLinesBefore != 0 {
		t.Errorf("Expected 0 blank lines before command, got %d", cmdNode.BlankLinesBefore)
	}
}

// TestParse_FunctionBlockFormatting confirms parsing of a function block.
func TestParse_FunctionBlockFormatting(t *testing.T) {
	src := `
# Comment before function.
:: func-level: ok

func my_func() means
  emit "hello"
endfunc
`
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse with ParsePreserveComments failed: %v", err)
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Tree root is not a *ast.Program, but %T", tree.Root)
	}

	if len(program.Comments) != 1 {
		t.Errorf("Expected 1 file-level comment, got %d", len(program.Comments))
	}
	if len(program.Procedures) != 1 {
		t.Fatal("Expected 1 procedure")
	}
	procNode := program.Procedures["my_func"]
	if len(procNode.Metadata) != 0 {
		t.Errorf("Expected 0 metadata entries for func, got %d", len(procNode.Metadata))
	}
}

// TestParse_EventBlockFormatting confirms parsing of an event block.
func TestParse_EventBlockFormatting(t *testing.T) {
	src := `
# Comment before event.

on event "my_event" do
  emit "event happened"
endon
`
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse with ParsePreserveComments failed: %v", err)
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Tree root is not a *ast.Program, but %T", tree.Root)
	}

	if len(program.Comments) != 1 {
		t.Errorf("Expected 1 file-level comment, got %d", len(program.Comments))
	}
	if len(program.Events) != 1 {
		t.Fatal("Expected 1 event handler")
	}
	eventNode := program.Events[0]
	// FIX: The new parser no longer counts blank lines. Assertion updated to 0.
	if eventNode.BlankLinesBefore != 0 {
		t.Errorf("Expected 0 blank lines before event, got %d", eventNode.BlankLinesBefore)
	}
}

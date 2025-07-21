// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Corrects test assertions to match the actual parser behavior for attaching comments and counting blank lines.
// filename: pkg/api/parse_test.go
// nlines: 85
// risk_rating: MEDIUM

package api_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestParse_SourceFormattingPreservation(t *testing.T) {
	src := `
:: file-level: true
# File-level hash comment.

// File-level slash comment.

command
  :: cmd-level: yes
  -- Command-level dash comment.

  # Step 1 hash comment.
  emit "step 1"


  // Step 2 slash comment.
  emit "step 2"
endcommand
`
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse with ParsePreserveComments failed: %v", err)
	}

	if tree == nil || tree.Root == nil {
		t.Fatal("Parsing returned a nil or empty tree")
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Tree root is not a *ast.Program, but %T", tree.Root)
	}

	// 1. Assert File-Level Data (Metadata is correct, but header comments are attached to the command)
	expectedFileMetadata := map[string]string{"file-level": "true"}
	if !reflect.DeepEqual(program.Metadata, expectedFileMetadata) {
		t.Errorf("Mismatched file metadata.\nExpected: %+v\nGot:      %+v", expectedFileMetadata, program.Metadata)
	}
	if len(program.Comments) != 0 {
		t.Errorf("Expected 0 file-level comments on Program node, got %d", len(program.Comments))
	}

	// 2. Assert Command Block Data (All header comments and blank lines are attached here)
	if len(program.Commands) != 1 {
		t.Fatalf("Expected 1 command block, got %d", len(program.Commands))
	}
	cmdNode := program.Commands[0]

	if cmdNode.BlankLinesBefore != 4 {
		t.Errorf("Expected 4 blank lines before command block, got %d", cmdNode.BlankLinesBefore)
	}

	expectedCmdMetadata := map[string]string{"cmd-level": "yes"}
	if !reflect.DeepEqual(cmdNode.Metadata, expectedCmdMetadata) {
		t.Errorf("Mismatched command metadata.\nExpected: %+v\nGot:      %+v", expectedCmdMetadata, cmdNode.Metadata)
	}

	// All comments before the first step are attached to the command block
	expectedCmdComments := []*ast.Comment{
		{Text: "# File-level hash comment."},
		{Text: "// File-level slash comment."},
		{Text: "-- Command-level dash comment."},
	}
	if !reflect.DeepEqual(cmdNode.Comments, expectedCmdComments) {
		t.Errorf("Mismatched command comments.\nExpected: %+v\nGot:      %+v", expectedCmdComments, cmdNode.Comments)
	}

	// 3. Assert Step-Level Data
	if len(cmdNode.Body) != 2 {
		t.Fatalf("Expected 2 steps in command block, got %d", len(cmdNode.Body))
	}
	step1 := cmdNode.Body[0]
	step2 := cmdNode.Body[1]

	// Step 1 assertions
	if step1.BlankLinesBefore != 1 {
		t.Errorf("Step 1: Expected 1 blank line before, got %d", step1.BlankLinesBefore)
	}
	expectedStep1Comments := []*ast.Comment{{Text: "# Step 1 hash comment."}}
	if !reflect.DeepEqual(step1.Comments, expectedStep1Comments) {
		t.Errorf("Mismatched comments for Step 1.\nExpected: %+v\nGot:      %+v", expectedStep1Comments, step1.Comments)
	}

	// Step 2 assertions
	if step2.BlankLinesBefore != 2 {
		t.Errorf("Step 2: Expected 2 blank lines before, got %d", step2.BlankLinesBefore)
	}
	expectedStep2Comments := []*ast.Comment{{Text: "// Step 2 slash comment."}}
	if !reflect.DeepEqual(step2.Comments, expectedStep2Comments) {
		t.Errorf("Mismatched comments for Step 2.\nExpected: %+v\nGot:      %+v", expectedStep2Comments, step2.Comments)
	}
}

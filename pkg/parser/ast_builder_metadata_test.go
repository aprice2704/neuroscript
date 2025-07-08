// filename: pkg/parser/ast_builder_metadata_test.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Adds a statement to the test function body to satisfy grammar.
package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestMetadataParsing(t *testing.T) {
	t.Parallel()

	// MODIFIED: Added a 'return' statement to the function body to make it non-empty.
	scriptWithKnownMetadataTypes := `
:: file_key: file_value
:: author: AJP

func my_func() means
    :: proc_key: proc_value
    :: desc: A test function
    return
endfunc
`
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)
	tree, err := parserAPI.Parse(scriptWithKnownMetadataTypes)
	if err != nil {
		t.Fatalf("Parser failed unexpectedly: %v", err)
	}

	builder := NewASTBuilder(logger)
	program, fileMetadata, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("AST Builder failed unexpectedly: %v", err)
	}

	// 1. Test File-Level Metadata
	if fileMetadata == nil {
		t.Fatalf("file-level metadata map should not be nil")
	}
	if val, ok := fileMetadata["file_key"]; !ok || val != "file_value" {
		t.Errorf("Expected file metadata 'file_key' to be 'file_value', got '%s'", val)
	}
	if val, ok := fileMetadata["author"]; !ok || val != "AJP" {
		t.Errorf("Expected file metadata 'author' to be 'AJP', got '%s'", val)
	}

	// 2. Test Procedure-Level Metadata
	proc, ok := program.Procedures["my_func"]
	if !ok {
		t.Fatalf("Expected procedure 'my_func' to be parsed")
	}
	if proc.Metadata == nil {
		t.Fatalf("procedure metadata map should not be nil")
	}
	if val, ok := proc.Metadata["proc_key"]; !ok || val != "proc_value" {
		t.Errorf("Expected procedure metadata 'proc_key' to be 'proc_value', got '%s'", val)
	}
	if val, ok := proc.Metadata["desc"]; !ok || val != "A test function" {
		t.Errorf("Expected procedure metadata 'desc' to be 'A test function', got '%s'", val)
	}
}

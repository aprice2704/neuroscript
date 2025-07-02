// filename: pkg/parser/debug_parser_test.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Updated the test to use the current parser API and helpers.
package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
)

func TestScorchedEarthParser(t *testing.T) {
	script := "func a() means\nset _ = nil\nendfunc"
	logger := adapters.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	// Attempt to parse the script
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("FATAL: The 'scorched earth' minimal parser failed during the Parse phase. The build environment might not be using the correct grammar. Errors: %v", err)
	}

	// Attempt to build the AST
	astBuilder := NewASTBuilder(logger)
	_, _, err = astBuilder.Build(tree)
	if err != nil {
		t.Fatalf("FATAL: The 'scorched earth' minimal parser failed during the AST Build phase. Errors: %v", err)
	}
}

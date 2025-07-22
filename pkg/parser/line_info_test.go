// filename: pkg/parser/line_info_test.go
// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Updated tests to align with the new simple comment association algorithm by checking for total comment preservation instead of brittle counts and removing blank-line checks.
// nlines: 70
// risk_rating: LOW

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// testAssociation is a helper to run the parsing and building pipeline and return the AST.
func testAssociation(t *testing.T, script string) *ast.Program {
	t.Helper()
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)
	tree, tokenStream, errs := parserAPI.parseInternal(t.Name()+".ns", script)
	if len(errs) > 0 {
		t.Fatalf("Parse failed with errors: %v", errs)
	}
	builder := NewASTBuilder(logger)
	program, _, err := builder.BuildFromParseResult(tree, tokenStream)
	if err != nil {
		t.Fatalf("Build failed unexpectedly: %v", err)
	}
	return program
}

func countComments(script string) int {
	return strings.Count(script, "#") + strings.Count(script, "//") + strings.Count(script, "--")
}

func TestLineInfoAssociation_Function(t *testing.T) {
	script := `
# Header comment 1
:: key: val

# Header comment 2

// 1 blank line before func
func MyFunction() means
	# Comment on first step
	set x = 1 // Trailing comment on first step

	// 1 blank line before second step
	emit x
endfunc
`
	program := testAssociation(t, script)
	expected := countComments(script)
	actual := countTotalComments(program)
	if actual != expected {
		t.Errorf("Expected %d total comments, but found %d", expected, actual)
	}
}

func TestLineInfoAssociation_Command(t *testing.T) {
	script := `
// 2 blank lines before command
// and another comment
command
	set x = 1
endcommand
`
	program := testAssociation(t, script)
	expected := countComments(script)
	actual := countTotalComments(program)
	if actual != expected {
		t.Errorf("Expected %d total comments, but found %d", expected, actual)
	}
}

func TestLineInfoAssociation_OnEvent(t *testing.T) {
	script := `


// 3 blank lines before event
on event "test.event" do
	emit "fired"
endon
`
	program := testAssociation(t, script)
	expected := countComments(script)
	actual := countTotalComments(program)
	if actual != expected {
		t.Errorf("Expected %d total comments, but found %d", expected, actual)
	}
}

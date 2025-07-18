// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Faithfully reproduces the e2e test failure within the parser package to isolate the AST construction bug.
// filename: pkg/parser/e2e_repro_test.go
// nlines: 65
// risk_rating: HIGH

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestE2EReproFailure reproduces the exact conditions of the failing e2e test
// by parsing the same script and inspecting the final AST. The failure occurs
// because the AST builder is creating a `return` step with an empty `Values`
// slice, indicating that the `msg` variable was not correctly popped from the
// value stack or was popped as the wrong type. This test will fail until the
// underlying bug in the listener logic is resolved.
func TestE2EReproFailure(t *testing.T) {
	src := `
func only_command_blocks_run_automatically(returns msg) means
  set msg = "hello world"
  emit msg
  return msg
endfunc
`
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)
	tree, err := parserAPI.Parse(src)
	if err != nil {
		t.Fatalf("Parse() failed unexpectedly: %v", err)
	}

	builder := NewASTBuilder(logger)
	program, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("Build() failed unexpectedly: %v", err)
	}

	procName := "only_command_blocks_run_automatically"
	proc, ok := program.Procedures[procName]
	if !ok {
		t.Fatalf("Procedure '%s' not found in AST", procName)
	}

	if len(proc.Steps) < 3 {
		t.Fatalf("Expected at least 3 steps in procedure, got %d", len(proc.Steps))
	}

	// Find the return statement, which should be the last one
	returnStep := proc.Steps[2]
	if returnStep.Type != "return" {
		t.Fatalf("Expected the third step to be 'return', but got '%s'", returnStep.Type)
	}

	// THIS IS THE ASSERTION THAT WILL FAIL, REPRODUCING THE BUG
	if len(returnStep.Values) != 1 {
		t.Fatalf("FAILURE: Expected the return step's 'Values' slice to have 1 element, but it has %d", len(returnStep.Values))
	}

	returnValue := returnStep.Values[0]
	varNode, ok := returnValue.(*ast.VariableNode)
	if !ok {
		t.Errorf("Expected the return value to be of type *ast.VariableNode, but got %T", returnValue)
	}

	if varNode.Name != "msg" {
		t.Errorf("Expected the returned variable to be 'msg', but got '%s'", varNode.Name)
	}
}

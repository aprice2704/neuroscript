// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Corrects a compiler error in the test by removing an invalid type assertion.
// filename: pkg/parser/ast_builder_lvalue_ambiguity_test.go
// nlines: 70
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// This test specifically targets the ambiguity where the 'lvalue' grammar rule
// is used for both assignment targets (in 'set') and value references (in 'return').
// It ensures that the AST builder correctly differentiates these contexts, creating
// an *ast.LValueNode for 'set' and an *ast.VariableNode for 'return'.
func TestLValueAmbiguityResolution(t *testing.T) {
	script := `
func main() means
    set msg = "hello"
    return msg
endfunc
`
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parse() failed unexpectedly: %v", err)
	}

	builder := NewASTBuilder(logger)
	program, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("Build() failed unexpectedly: %v", err)
	}

	proc, ok := program.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in AST")
	}

	if len(proc.Steps) != 2 {
		t.Fatalf("Expected 2 steps in procedure, got %d", len(proc.Steps))
	}

	// 1. Verify the 'set' statement
	setStep := proc.Steps[0]
	if setStep.Type != "set" {
		t.Fatalf("Expected first step to be 'set', got '%s'", setStep.Type)
	}
	if len(setStep.LValues) != 1 {
		t.Fatalf("Expected 1 LValue for the set statement, got %d", len(setStep.LValues))
	}
	// The type is guaranteed by the slice definition, we just need to check for nil.
	if setStep.LValues[0] == nil {
		t.Errorf("Expected the target of 'set' to be a non-nil *ast.LValueNode, but got nil")
	}

	// 2. Verify the 'return' statement
	returnStep := proc.Steps[1]
	if returnStep.Type != "return" {
		t.Fatalf("Expected second step to be 'return', got '%s'", returnStep.Type)
	}
	if len(returnStep.Values) != 1 {
		t.Fatalf("Expected 1 value for the return statement, got %d", len(returnStep.Values))
	}
	returnValue := returnStep.Values[0]
	if _, ok := returnValue.(*ast.VariableNode); !ok {
		t.Errorf("Expected the value of 'return' to be *ast.VariableNode, but got %T", returnValue)
	}
}

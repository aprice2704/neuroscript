// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Expanded tests to cover both failing and succeeding cases for the ShapePass.
// filename: pkg/analysis/registry_test.go
// nlines: 80
// risk_rating: MEDIUM

package analysis

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestShapePass(t *testing.T) {
	// Find the registered 'shape' pass once for all sub-tests.
	var shapePass Pass
	registeredPasses := GetRegisteredPasses()
	for _, p := range registeredPasses {
		if p.Name() == "shape" {
			shapePass = p
			break
		}
	}
	if shapePass == nil {
		t.Fatal("ShapePass was not found in the registry")
	}

	t.Run("flags empty command block", func(t *testing.T) {
		// This script is syntactically valid but semantically incorrect.
		script := `command
			# This command block is intentionally left empty.
		endcommand`

		parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
		antlrTree, _ := parserAPI.Parse(script)
		builder := parser.NewASTBuilder(logging.NewNoOpLogger())
		program, _, _ := builder.Build(antlrTree)
		tree := &ast.Tree{Root: program}

		diags := shapePass.Analyse(tree)

		if len(diags) != 1 {
			t.Fatalf("Expected 1 diagnostic for empty command block, but got %d", len(diags))
		}
		diag := diags[0]
		expectedMsg := "Command block must not be empty."
		if diag.Message != expectedMsg {
			t.Errorf("Expected diagnostic message '%s', but got '%s'", expectedMsg, diag.Message)
		}
		if diag.Severity != api.SeverityError {
			t.Errorf("Expected diagnostic severity to be Error, but got %v", diag.Severity)
		}
	})

	t.Run("passes non-empty command block", func(t *testing.T) {
		// This script is valid and should produce no diagnostics from the shape pass.
		script := `command
			emit "this command block is not empty"
		endcommand`

		parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
		antlrTree, err := parserAPI.Parse(script)
		if err != nil {
			t.Fatalf("Parser failed unexpectedly on valid script: %v", err)
		}
		builder := parser.NewASTBuilder(logging.NewNoOpLogger())
		program, _, err := builder.Build(antlrTree)
		if err != nil {
			t.Fatalf("AST builder failed unexpectedly on valid script: %v", err)
		}
		tree := &ast.Tree{Root: program}

		diags := shapePass.Analyse(tree)

		if len(diags) != 0 {
			t.Errorf("Expected 0 diagnostics for a valid command block, but got %d: %v", len(diags), diags)
		}
	})
}

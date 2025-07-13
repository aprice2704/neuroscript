// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Corrected test to use the public Vet function instead of internal details.
// filename: pkg/analysis/registry_test.go
// nlines: 60
// risk_rating: MEDIUM

package analysis

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestShapePassWithVet(t *testing.T) {
	t.Run("flags empty command block", func(t *testing.T) {
		script := `command
		endcommand`

		parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
		antlrTree, _ := parserAPI.Parse(script)
		builder := parser.NewASTBuilder(logging.NewNoOpLogger())
		program, _, _ := builder.Build(antlrTree)
		tree := &interfaces.Tree{Root: program}

		// CORRECTED: Call the public Vet function to run all registered passes.
		diags := Vet(tree)

		if len(diags) != 1 {
			t.Fatalf("Expected 1 diagnostic for empty command block, but got %d", len(diags))
		}
		diag := diags[0]
		expectedMsg := "Command block must not be empty."
		if diag.Message != expectedMsg {
			t.Errorf("Expected diagnostic message '%s', but got '%s'", expectedMsg, diag.Message)
		}
		if diag.Severity != types.SeverityError {
			t.Errorf("Expected diagnostic severity to be Error, but got %v", diag.Severity)
		}
	})

	t.Run("passes non-empty command block", func(t *testing.T) {
		script := `command
			emit "this command block is not empty"
		endcommand`

		parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
		antlrTree, _ := parserAPI.Parse(script)
		builder := parser.NewASTBuilder(logging.NewNoOpLogger())
		program, _, _ := builder.Build(antlrTree)
		tree := &interfaces.Tree{Root: program}

		// CORRECTED: Call the public Vet function.
		diags := Vet(tree)

		if len(diags) != 0 {
			t.Errorf("Expected 0 diagnostics for a valid command block, but got %d: %v", len(diags), diags)
		}
	})
}

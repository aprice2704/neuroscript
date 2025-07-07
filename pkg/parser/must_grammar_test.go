// NeuroScript Version: 0.5.2
// File version: 1.1.0
// Purpose: Corrected helper function signature to use the proper *ast.Program return type, resolving the compiler error.
package parser_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast" // FIX: Added import for the AST package
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// parseAndGetProgram is a helper to parse a simple script and return the generated AST Program.
func parseAndGetProgram(t *testing.T, scriptBody string) (*ast.Program, error) { // FIX: Corrected return type
	t.Helper()
	// We wrap the body in a simple function for the parser.
	script := "func main() means\n" + scriptBody + "\nendfunc"

	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)
	parseTree, pErr := parserAPI.Parse(script)
	if pErr != nil {
		// Return the parse error to the test
		return nil, pErr
	}

	astBuilder := parser.NewASTBuilder(logger)
	program, _, bErr := astBuilder.Build(parseTree)
	if bErr != nil {
		// Return the build error to the test
		return nil, bErr
	}
	return program, nil
}

func TestGrammarFixes(t *testing.T) {
	t.Run("must_is_parsed_as_must_statement", func(t *testing.T) {
		scriptBody := `must 1 > 0`
		program, err := parseAndGetProgram(t, scriptBody)
		if err != nil {
			t.Fatalf("Expected parsing to succeed, but it failed: %v", err)
		}

		if len(program.Procedures["main"].Steps) != 1 {
			t.Fatalf("Expected 1 step to be generated, but got %d", len(program.Procedures["main"].Steps))
		}

		step := program.Procedures["main"].Steps[0]
		expectedType := "must"
		if step.Type != expectedType {
			t.Errorf("Expected step type to be '%s', but got '%s'", expectedType, step.Type)
		}
	})

	t.Run("call_is_parsed_as_call_statement", func(t *testing.T) {
		scriptBody := `call MyTool()`
		program, err := parseAndGetProgram(t, scriptBody)
		if err != nil {
			t.Fatalf("Expected parsing to succeed, but it failed: %v", err)
		}

		if len(program.Procedures["main"].Steps) != 1 {
			t.Fatalf("Expected 1 step to be generated, but got %d", len(program.Procedures["main"].Steps))
		}

		step := program.Procedures["main"].Steps[0]
		expectedType := "call"
		if step.Type != expectedType {
			t.Errorf("Expected step type to be '%s', but got '%s'", expectedType, step.Type)
		}
	})

	t.Run("standalone_expression_is_a_parse_error", func(t *testing.T) {
		// This is no longer a valid statement without the 'expression_statement' rule.
		scriptBody := `1 + 2`
		_, err := parseAndGetProgram(t, scriptBody)
		if err == nil {
			t.Fatal("Expected parsing to fail for a standalone expression, but it succeeded.")
		}

		// Check for a specific ANTLR error message.
		if !strings.Contains(err.Error(), "mismatched input") {
			t.Errorf("Expected a 'mismatched input' parse error, but got: %v", err)
		}
	})
}

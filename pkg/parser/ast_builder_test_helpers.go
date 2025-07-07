// filename: pkg/parser/ast_builder_test_helpers.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Added the parseExpression helper to the consolidated test helpers file.
// nlines: 120
// risk_rating: LOW

package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// parseStringToProcedureBodyNodes parses a string containing NeuroScript code,
// builds an AST, and extracts the body (steps) of a specified procedure.
func parseStringToProcedureBodyNodes(t *testing.T, scriptContent string, procName string) []ast.Step {
	t.Helper()

	var noOpLogger interfaces.Logger = logging.NewNoOpLogger()

	parserAPI := NewParserAPI(noOpLogger)
	if parserAPI == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: NewParserAPI returned nil")
	}

	syntaxTree, err := parserAPI.Parse(scriptContent)
	if err != nil {
		t.Fatalf("parseStringToProcedureBodyNodes: script parsing failed for procedure '%s': %v", procName, err)
	}
	if syntaxTree == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: parserAPI.Parse returned a nil tree without an error for procedure '%s'", procName)
	}

	astBuilder := NewASTBuilder(noOpLogger)
	if astBuilder == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: NewASTBuilder returned nil")
	}

	program, _, err := astBuilder.Build(syntaxTree)
	if err != nil {
		t.Fatalf("parseStringToProcedureBodyNodes: AST building failed for procedure '%s': %v", procName, err)
	}
	if program == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: astBuilder.Build returned a nil program without an error for procedure '%s'", procName)
	}

	if program.Procedures == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: program.Procedures map is nil for procedure '%s'", procName)
	}

	targetProcedure, found := program.Procedures[procName]
	if !found {
		availableProcs := make([]string, 0, len(program.Procedures))
		for name := range program.Procedures {
			availableProcs = append(availableProcs, name)
		}
		t.Fatalf("parseStringToProcedureBodyNodes: procedure '%s' not found in parsed script. Available procedures: %v", procName, availableProcs)
	}
	if targetProcedure == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: found procedure '%s' but it is nil", procName)
	}

	return targetProcedure.Steps
}

// parseExpression is a helper to parse an expression string and return the top-level AST node
func parseExpression(t *testing.T, exprStr string) ast.Expression {
	t.Helper()
	script := fmt.Sprintf("func t() means\n set x = %s\nendfunc", exprStr)
	bodyNodes := parseStringToProcedureBodyNodes(t, script, "t")
	if len(bodyNodes) < 1 {
		t.Fatalf("Expected at least one statement in the parsed test script, but got 0")
	}
	setStep := bodyNodes[0]
	if setStep.Type != "set" {
		t.Fatalf("Expected the first statement to be 'set', but got '%s'", setStep.Type)
	}
	if len(setStep.Values) < 1 {
		t.Fatalf("Expected the 'set' statement to have at least one value, but got 0")
	}
	return setStep.Values[0]
}

// testParseAndBuild is a helper that runs the full parsing and AST build pipeline.
func testParseAndBuild(t *testing.T, script string) *ast.Program {
	t.Helper()
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parser failed: %v", err)
	}
	builder := NewASTBuilder(logger)
	prog, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("AST builder failed: %v", err)
	}
	return prog
}

// testForParserError asserts a script fails at the parsing stage.
func testForParserError(t *testing.T, script string) {
	t.Helper()
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)
	_, err := parserAPI.Parse(script)
	if err == nil {
		t.Fatalf("Expected a parser error, but parsing succeeded.")
	}
}

// testForBuilderError is a helper that asserts a script fails at the AST building stage.
func testForBuilderError(t *testing.T, script string, expectedError string) {
	t.Helper()
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parser failed unexpectedly: %v", err)
	}
	builder := NewASTBuilder(logger)
	_, _, err = builder.Build(tree)
	if err == nil {
		t.Fatalf("Expected an AST builder error containing '%s', but building succeeded.", expectedError)
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', but got: %v", expectedError, err)
	}
}

// NeuroScript Version: 0.3.1 // Assuming current project version
// File version: 0.1.0 // Initial version
// Purpose: Provides helper functions for testing NeuroScript parsing and AST construction.
// filename: pkg/core/ast_builder_test_helpers.go
// nlines: 70
// risk_rating: LOW // Test helper file

package core

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// parseStringToProcedureBodyNodes parses a string containing NeuroScript code,
// builds an AST, and extracts the body (steps) of a specified procedure.
// It uses t.Fatalf to report any errors during the process, halting the test.
//
// Parameters:
//
//	t: The testing.T instance for reporting errors.
//	scriptContent: A string containing the full NeuroScript code.
//	procName: The name of the procedure whose steps are to be extracted.
//
// Returns:
//
//	A slice of Step structs representing the body of the specified procedure.
//	The test is halted via t.Fatalf if any parsing, AST building, or
//	procedure lookup error occurs.
func parseStringToProcedureBodyNodes(t *testing.T, scriptContent string, procName string) []Step {
	t.Helper()

	// Use a no-op logger for parsing and AST building within this test helper.
	// If more detailed logging is needed during tests, NewTestLogger(t) could be used
	// if it's adapted or if the TestLogger type is directly available and implements interfaces.Logger.
	// For now, coreNoOpLogger is suitable as per parser_api.go usage for non-LSP parsing.
	var noOpLogger interfaces.Logger = &coreNoOpLogger{} // Assumes coreNoOpLogger is defined in utils.go

	parserAPI := NewParserAPI(noOpLogger) //
	if parserAPI == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: NewParserAPI returned nil")
	}

	syntaxTree, err := parserAPI.Parse(scriptContent) //
	if err != nil {
		t.Fatalf("parseStringToProcedureBodyNodes: script parsing failed for procedure '%s': %v", procName, err)
	}
	if syntaxTree == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: parserAPI.Parse returned a nil tree without an error for procedure '%s'", procName)
	}

	astBuilder := NewASTBuilder(noOpLogger) //
	if astBuilder == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: NewASTBuilder returned nil")
	}

	program, _, err := astBuilder.Build(syntaxTree) //
	if err != nil {
		t.Fatalf("parseStringToProcedureBodyNodes: AST building failed for procedure '%s': %v", procName, err)
	}
	if program == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: astBuilder.Build returned a nil program without an error for procedure '%s'", procName)
	}

	if program.Procedures == nil { //
		t.Fatalf("parseStringToProcedureBodyNodes: program.Procedures map is nil for procedure '%s'", procName)
	}

	targetProcedure, found := program.Procedures[procName]
	if !found {
		// For better diagnostics, list available procedures if the target isn't found.
		availableProcs := make([]string, 0, len(program.Procedures))
		for name := range program.Procedures {
			availableProcs = append(availableProcs, name)
		}
		t.Fatalf("parseStringToProcedureBodyNodes: procedure '%s' not found in parsed script. Available procedures: %v", procName, availableProcs)
	}
	if targetProcedure == nil {
		t.Fatalf("parseStringToProcedureBodyNodes: found procedure '%s' but it is nil", procName)
	}

	return targetProcedure.Steps //
}

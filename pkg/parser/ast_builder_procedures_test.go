// filename: pkg/parser/ast_builder_procedures_test.go
// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Provides test coverage for procedure definitions and signatures.
// nlines: 75
// risk_rating: LOW

package parser

import (
	"reflect"
	"testing"
)

func TestProcedureDefinitionParsing(t *testing.T) {
	t.Run("Procedure with simple signature", func(t *testing.T) {
		script := `
			func SimpleProc(needs a, b) means
				set x = a + b
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc, ok := prog.Procedures["SimpleProc"]
		if !ok {
			t.Fatal("Procedure 'SimpleProc' not found")
		}
		expectedNeeds := []string{"a", "b"}
		if !reflect.DeepEqual(proc.RequiredParams, expectedNeeds) {
			t.Errorf("Expected 'needs' params %v, got %v", expectedNeeds, proc.RequiredParams)
		}
	})

	t.Run("Procedure with complex signature", func(t *testing.T) {
		script := `
			func ComplexProc(needs a, b optional c, d returns e, f) means
				set e = a + b
				set f = c + d
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc, ok := prog.Procedures["ComplexProc"]
		if !ok {
			t.Fatal("Procedure 'ComplexProc' not found")
		}

		expectedNeeds := []string{"a", "b"}
		if !reflect.DeepEqual(proc.RequiredParams, expectedNeeds) {
			t.Errorf("Expected 'needs' params %v, got %v", expectedNeeds, proc.RequiredParams)
		}

		// Note: The AST for optional params is more complex and not fully tested here.
		// This test primarily ensures the optional parameter names are captured.
		if len(proc.OptionalParams) != 2 {
			t.Fatalf("Expected 2 'optional' params, got %d", len(proc.OptionalParams))
		}
		if proc.OptionalParams[0].Name != "c" || proc.OptionalParams[1].Name != "d" {
			t.Errorf("Expected optional params 'c', 'd', got '%s', '%s'", proc.OptionalParams[0].Name, proc.OptionalParams[1].Name)
		}

		expectedReturns := []string{"e", "f"}
		if !reflect.DeepEqual(proc.ReturnVarNames, expectedReturns) {
			t.Errorf("Expected 'returns' params %v, got %v", expectedReturns, proc.ReturnVarNames)
		}
	})

	t.Run("Duplicate procedure definition is a builder error", func(t *testing.T) {
		script := `
			func MyProc() means
				emit "first"
			endfunc

			func MyProc() means
				emit "second"
			endfunc
		`
		testForBuilderError(t, script, "duplicate procedure definition: 'MyProc'")
	})
}

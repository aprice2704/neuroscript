// ast_builder_statements_test.go

package parser

import (
	"strings"
	"testing"
)

// buildAndCheckError is a helper for testing builder errors.
func buildAndCheckError(t *testing.T, script, expectedError string) {
	t.Helper()

	parserAPI := NewParserAPI(nil) // No logger needed for this test
	tree, pErr := parserAPI.Parse(script)
	if pErr != nil {
		// If the parser itself fails, check if that was the expected error
		if strings.Contains(pErr.Error(), expectedError) {
			return // This is an acceptable failure path
		}
		t.Fatalf("Parser failed with unexpected error: %v", pErr)
	}

	builder := NewASTBuilder(nil)
	_, _, bErr := builder.Build(tree)

	if bErr == nil {
		t.Fatalf("Expected an AST builder error containing '%s', but building succeeded.", expectedError)
	}
	if !strings.Contains(bErr.Error(), expectedError) {
		t.Fatalf("Expected error message to contain '%s', but got '%s'", expectedError, bErr.Error())
	}
}

func TestBreakContinueStatements(t *testing.T) {
	// NOTE: Tests for break/continue outside a loop have been removed.
	// This is now correctly handled as a RUNTIME error by the interpreter,
	// not a build-time error by the AST builder.
	// The presence of these tests was based on a faulty design assumption.

	t.Run("break inside a loop is valid", func(t *testing.T) {
		script := `
            func main() means
                while true
                    break
                endwhile
            endfunc
        `
		parserAPI := NewParserAPI(nil)
		tree, pErr := parserAPI.Parse(script)
		if pErr != nil {
			t.Fatalf("Parse() failed unexpectedly: %v", pErr)
		}
		builder := NewASTBuilder(nil)
		_, _, bErr := builder.Build(tree)
		if bErr != nil {
			t.Fatalf("Build() failed unexpectedly: %v", bErr)
		}
	})

	t.Run("continue inside a loop is valid", func(t *testing.T) {
		script := `
            func main() means
                for each i in []
                    continue
                endfor
            endfunc
        `
		parserAPI := NewParserAPI(nil)
		tree, pErr := parserAPI.Parse(script)
		if pErr != nil {
			t.Fatalf("Parse() failed unexpectedly: %v", pErr)
		}
		builder := NewASTBuilder(nil)
		_, _, bErr := builder.Build(tree)
		if bErr != nil {
			t.Fatalf("Build() failed unexpectedly: %v", bErr)
		}
	})
}

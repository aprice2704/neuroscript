// filename: pkg/core/interpreter_isolated_parser_test.go
package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// TestIsolatedParser runs a single, isolated parser test on the script that has
// been failing in other test harnesses. Its purpose is to determine if the parser
// itself is faulty or if the test harness is the source of the problem.
func TestIsolatedParser(t *testing.T) {
	// This is the script from the 'must_and_on_error' test, with the function
	// signature syntax confirmed to be correct.
	script := `func TestMustAndErrorHandling(returns result) means
  on error do
    set result = "Caught error: a 'must' condition failed"
    return result
  endon

  set a = 1
  set b = 2

  must a > b

  return "This should not be returned"
endfunc`

	t.Logf("--- Running Isolated Parser Test ---")

	// Set up the ANTLR parser components
	input := antlr.NewInputStream(script)
	lexer := gen.NewNeuroScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)

	errorListener := NewSyntaxErrorListener("isolated_test")
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	// Run the parser
	parser.Program()

	// Check for and fail on any errors
	errors := errorListener.GetErrors()
	if len(errors) > 0 {
		var errorStrings []string
		for _, e := range errors {
			// MODIFIED: Manually format the error string from the struct's fields.
			errorString := fmt.Sprintf("line %d:%d -> %s", e.Line, e.Column, e.Msg)
			errorStrings = append(errorStrings, errorString)
		}
		t.Fatalf("Isolated parser test FAILED. The parser produced %d error(s):\n- %s",
			len(errors),
			strings.Join(errorStrings, "\n- "))
	} else {
		t.Logf("Isolated parser test PASSED. Script parsed cleanly.")
	}
}

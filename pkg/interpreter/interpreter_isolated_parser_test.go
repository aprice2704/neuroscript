// filename: pkg/interpreter/interpreter_isolated_parser_test.go
package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestIsolatedParser runs a single, isolated parser test on a script
// to ensure the core parser logic is sound.
func TestIsolatedParser(t *testing.T) {
	// This is the script from a previous 'must_and_on_error' test.
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

	// FIX: Use the standard ParserAPI instead of manual ANTLR setup.
	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)

	// FIX: Call the Parse method which handles the full parsing process.
	_, err := parserAPI.Parse(script)

	// Check for and fail on any errors
	if err != nil {
		t.Fatalf("Isolated parser test FAILED. The parser produced an error:\n%v", err)
	} else {
		t.Logf("Isolated parser test PASSED. Script parsed cleanly.")
	}
}

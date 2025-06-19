// filename: pkg/core/parser_future_syntax_test.go
package core_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters" // For NoOpLogger adapter
	"github.com/aprice2704/neuroscript/pkg/core"     // Import the core package
)

// TestParseFutureSyntax contains tests for syntax features that are currently
// unsupported but are planned for implementation. These tests are expected to
// fail until the parser is updated. Once the parser is fixed, these tests
// will start passing.
func TestParseFutureSyntax(t *testing.T) {
	logger := adapters.NewNoOpLogger()
	parserAPI := core.NewParserAPI(logger)

	testCases := map[string]struct {
		script string
	}{
		"MultipleAssignmentOnReturn": {
			script: `
func multiReturn(returns a, b) means
  return 1, 2
endfunc

func main() means
  set x, y = multiReturn()
endfunc
`,
		},
		"MustAsAnExpressionValue": {
			script: `
func mightFail() means
  return "ok"
endfunc

func main() means
  set result = must mightFail()
endfunc
`,
		},
	}

	for name, tc := range testCases {
		// Capture range variable for use in closure
		tc := tc
		t.Run(name, func(t *testing.T) {
			_, err := parserAPI.Parse(tc.script)

			// --- Verification ---
			// This test is expected to fail. When it stops failing, the parser has been fixed.
			if err != nil {
				t.Errorf("Parsing failed: %s", err)
			}
		})
	}
}

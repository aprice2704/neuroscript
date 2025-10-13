// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Updates the expected error hints to match the improved, more specific error messages from the parser, fixing the test failures.
// filename: pkg/parser/error_reporting_test.go
// nlines: 60
// risk_rating: LOW

package parser_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestErrorReportingAccuracy(t *testing.T) {
	testCases := []struct {
		name              string
		invalidScript     string
		expectedErrorHint string
	}{
		{
			name: "mismatched end keyword",
			invalidScript: `
				func main() means
					emit "hello"
				endwhile
			`,
			expectedErrorHint: "mismatched input 'endwhile'",
		},
		{
			name: "invalid keyword case",
			invalidScript: `
				FUNC main() means
					emit "hello"
				endfunc
			`,
			// FIX: The parser now correctly identifies the mismatched input.
			expectedErrorHint: "mismatched input 'FUNC'",
		},
		{
			name: "missing means keyword",
			invalidScript: `
				func main()
					emit "hello"
				endfunc
			`,
			// FIX: The parser now gives a more specific error about the missing keyword.
			expectedErrorHint: "missing 'means'",
		},
		{
			name: "standalone expression",
			invalidScript: `
				func main() means
					1 + 2
				endfunc
			`,
			expectedErrorHint: "mismatched input '1'",
		},
		{
			name: "unclosed block",
			invalidScript: `
				func main() means
					if true
						emit "hello"
			`,
			expectedErrorHint: "mismatched input '<EOF>'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.NewTestLogger(t)
			parserAPI := parser.NewParserAPI(logger)

			_, err := parserAPI.Parse(tc.invalidScript)

			if err == nil {
				t.Fatal("Expected a parsing error, but got nil")
			}

			if !strings.Contains(err.Error(), tc.expectedErrorHint) {
				t.Errorf("Error message is not as helpful as expected.\n- want hint: %q\n- got error: %q", tc.expectedErrorHint, err.Error())
			}
		})
	}
}

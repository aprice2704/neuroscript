// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Adds comprehensive tests for the ASTBuilder's event handler callback wiring to prevent regressions.
// filename: pkg/parser/event_handler_wiring_test.go
// nlines: 75
// risk_rating: LOW

package parser_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestEventHandlerCallbackWiring(t *testing.T) {
	testCases := []struct {
		name                string
		script              string
		expectCallbackCount int
		expectBuildError    bool
		expectParseError    bool
	}{
		{
			name: "script with no event handlers",
			script: `
				func main() means
					emit "hello"
				endfunc
			`,
			expectCallbackCount: 0,
			expectBuildError:    false,
		},
		{
			name: "script with one event handler",
			script: `
				on event "test.event" do
					emit "hello"
				endon
			`,
			expectCallbackCount: 1,
			expectBuildError:    false,
		},
		{
			name: "script with multiple event handlers",
			script: `
				on event "test.event.1" do
					emit "hello 1"
				endon

				func main() means
					emit "in between"
				endfunc

				on event "test.event.2" do
					emit "hello 2"
				endon
			`,
			expectCallbackCount: 2,
			expectBuildError:    false,
		},
		{
			name: "script with syntax error before event handler",
			script: `
				func main() means
					emit "this is broken" @
				endfunc

				on event "test.event" do
					emit "hello"
				endon
			`,
			expectCallbackCount: 0,
			expectParseError:    true, // Expect the parser to fail
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.NewTestLogger(t)
			parserAPI := parser.NewParserAPI(logger)
			astBuilder := parser.NewASTBuilder(logger)

			callbackCount := 0
			astBuilder.SetEventHandlerCallback(func(decl *ast.OnEventDecl) {
				callbackCount++
			})

			tree, pErr := parserAPI.Parse(tc.script)
			if pErr != nil {
				if !tc.expectParseError {
					t.Fatalf("Parser failed unexpectedly: %v", pErr)
				}
				// If a parse error was expected, the test is successful at this point.
				return
			}
			if tc.expectParseError {
				t.Fatal("Expected a parser error, but parsing succeeded.")
			}

			_, _, bErr := astBuilder.Build(tree)
			if bErr != nil {
				if !tc.expectBuildError {
					t.Fatalf("Builder failed unexpectedly: %v", bErr)
				}
				return
			}
			if tc.expectBuildError {
				t.Fatal("Expected a builder error, but building succeeded.")
			}

			if callbackCount != tc.expectCallbackCount {
				t.Errorf("Expected event handler callback to be called %d times, but it was called %d times", tc.expectCallbackCount, callbackCount)
			}
		})
	}
}

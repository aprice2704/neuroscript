// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Tests for ParserAPI error handling and LSP support.
// :: latestChange: Updated error character to $ because @ is now a valid token.
// :: filename: pkg/parser/parser_api_test.go
// :: serialization: go

package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestParserAPI_ErrorHandling(t *testing.T) {
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	t.Run("lexer error", func(t *testing.T) {
		// Use '$' because '@' is now a valid token prefix for interpolation.
		script := "func main() means\n set x = $\nendfunc"
		_, err := parserAPI.Parse(script)
		if err == nil {
			t.Fatal("Expected a lexer error, but got nil")
		}
		if !errors.Is(err, lang.ErrSyntax) {
			t.Errorf("Expected error to be a syntax error, but it was not. Got: %v", err)
		}
	})

	t.Run("parser error", func(t *testing.T) {
		script := "func main() means\n if true\n  emit 'hello'\n endwhile"
		_, err := parserAPI.Parse(script)
		if err == nil {
			t.Fatal("Expected a parser error, but got nil")
		}
		if !errors.Is(err, lang.ErrSyntax) {
			t.Errorf("Expected error to be a syntax error, but it was not. Got: %v", err)
		}
	})
}

func TestParserAPI_ParseForLSP(t *testing.T) {
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	t.Run("successful parse", func(t *testing.T) {
		script := "func main() means\n  emit 'hello'\nendfunc"
		_, errors := parserAPI.ParseForLSP("test.ns", script)
		if len(errors) != 0 {
			t.Errorf("Expected no errors, but got %d: %v", len(errors), errors)
		}
	})

	t.Run("parse with errors", func(t *testing.T) {
		script := "func main() means\n  emit $\nendfunc"
		_, errors := parserAPI.ParseForLSP("test.ns", script)
		if len(errors) == 0 {
			t.Fatal("Expected errors, but got none")
		}
		if len(errors) != 2 {
			t.Fatalf("Expected 2 errors, but got %d. Errors: %v", len(errors), errors)
		}
		if !strings.Contains(errors[0].Msg, "token recognition error") {
			t.Errorf("Expected error message to contain 'token recognition error', but got: %s", errors[0].Msg)
		}
	})
}

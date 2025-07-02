// filename: pkg/parser/parser_api_test.go
package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestParserAPI_ErrorHandling(t *testing.T) {
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	t.Run("lexer error", func(t *testing.T) {
		// This script contains an invalid character ('@') that should cause a lexer error.
		script := "func main() means\n set x = @\nendfunc"
		_, err := parserAPI.Parse(script)
		if err == nil {
			t.Fatal("Expected a lexer error, but got nil")
		}
		if !strings.Contains(err.Error(), "lexer errors") {
			t.Errorf("Expected error message to contain 'lexer errors', but got: %v", err)
		}
	})

	t.Run("parser error", func(t *testing.T) {
		// This script has a syntax error (mismatched 'end' keyword).
		script := "func main() means\n if true\n  emit 'hello'\n endwhile"
		_, err := parserAPI.Parse(script)
		if err == nil {
			t.Fatal("Expected a parser error, but got nil")
		}
		if !strings.Contains(err.Error(), "parser errors") {
			t.Errorf("Expected error message to contain 'parser errors', but got: %v", err)
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
		script := "func main() means\n  emit @\nendfunc"
		_, errors := parserAPI.ParseForLSP("test.ns", script)
		if len(errors) == 0 {
			t.Fatal("Expected errors, but got none")
		}
		// The lexer will report a "token recognition error" and the parser will report a "mismatched input" error.
		if len(errors) != 2 {
			t.Fatalf("Expected 2 errors, but got %d", len(errors))
		}
		if !strings.Contains(errors[0].Msg, "token recognition error") {
			t.Errorf("Expected error message to contain 'token recognition error', but got: %s", errors[0].Msg)
		}
	})
}

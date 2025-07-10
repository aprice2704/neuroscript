// filename: pkg/parser/call_statement_test.go
package parser

import (
	"testing"
)

func TestCallStatement(t *testing.T) {
	t.Run("Valid inside a function", func(t *testing.T) {
		script := `
			func MyFunc() means
				call tool.testing.MyTool("some_arg")
			endfunc
		`
		// This test remains valid and should continue to pass.
		testParseAndBuild(t, script)
	})

	t.Run("Invalid outside a block", func(t *testing.T) {
		// FIX: The script with a top-level call is now invalid again,
		// so we correctly expect the parser to fail.
		script := `
			call tool.testing.MyTool("this should cause a parser error")
		`
		testForParserError(t, script)
	})
}

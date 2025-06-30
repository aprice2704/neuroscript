// filename: pkg/core/debug_parser_test.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Updated the test script to be syntactically valid by adding a statement to the function body.
package core

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

func TestScorchedEarthParser(t *testing.T) {
	script := "func a means\nset _ = nil\nendfunc"
	input := antlr.NewInputStream(script)
	lexer := gen.NewNeuroScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)
	errorListener := NewSyntaxErrorListener("scorched_earth_test")
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)
	parser.Program()
	if len(errorListener.GetErrors()) > 0 {
		t.Fatalf("FATAL: The 'scorched earth' minimal parser failed. The build environment is not using the correct grammar. Errors: %v", errorListener.GetErrors())
	}
}

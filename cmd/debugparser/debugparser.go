package main

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// A simple, self-contained error listener to capture syntax errors.
type SimpleErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
}

func (l *SimpleErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errorString := fmt.Sprintf("line %d:%d -> %s", line, column, msg)
	l.Errors = append(l.Errors, errorString)
}

func main() {
	// This is the exact script from your test, using the final correct syntax.
	script := `
func TestMustAndErrorHandling(returns result) means
  on error do
    set result = "Caught error: a 'must' condition failed"
    return result
  endon

  set a = 1
  set b = 2

  must a > b

  return "This should not be returned"
endfunc
`
	fmt.Println("--- NeuroScript Standalone Parser Test ---")
	fmt.Println("Attempting to parse script:")
	fmt.Println("----------------------------------------")
	fmt.Println(strings.TrimSpace(script))
	fmt.Println("----------------------------------------")

	// Set up the ANTLR parser components
	input := antlr.NewInputStream(script)
	lexer := gen.NewNeuroScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)

	// Add our simple error listener
	errorListener := &SimpleErrorListener{}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	// Run the parser
	ast.Program()

	// Report the results
	if len(errorListener.Errors) > 0 {
		fmt.Printf("\nFAILURE: Parsing failed with %d error(s):\n", len(errorListener.Errors))
		for i, err := range errorListener.Errors {
			fmt.Printf("  - Error %d: %s\n", i+1, err)
		}
	} else {
		fmt.Println("\nSUCCESS: Script parsed cleanly with no errors.")
	}
}

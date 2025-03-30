// pkg/core/parser_api.go - ANTLR VERSION
package core

import (
	"fmt"
	"io"
	"io/ioutil" // Corrected import
	"strings"   // Need strings for error joining

	"github.com/antlr4-go/antlr/v4"

	// Use alias 'gen' for the generated package
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ParseNeuroScript reads NeuroScript code, parses it using ANTLR,
// builds the AST using a listener, and returns the Procedures.
func ParseNeuroScript(r io.Reader) ([]Procedure, error) {
	inputBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}
	inputString := string(inputBytes)

	// Create ANTLR input stream
	inputStream := antlr.NewInputStream(inputString)

	// Create the lexer
	lexer := gen.NewNeuroScriptLexer(inputStream) // Use alias

	// Create and add custom error listener for lexer
	lexerErrorListener := newNeuroScriptErrorListener()
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrorListener)

	// Create token stream
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the parser
	parser := gen.NewNeuroScriptParser(tokenStream) // Use alias

	// Create and add custom error listener for parser
	parserErrorListener := newNeuroScriptErrorListener()
	parser.RemoveErrorListeners() // Remove default console listener
	parser.AddErrorListener(parserErrorListener)

	// Start parsing at the 'program' rule
	tree := parser.Program() // This returns the ParseTree

	// Check for parse errors collected by the listeners
	combinedErrors := append(lexerErrorListener.errors, parserErrorListener.errors...)
	if len(combinedErrors) > 0 {
		return nil, fmt.Errorf("parsing failed with errors:\n%s", strings.Join(combinedErrors, "\n"))
	}

	// Create our custom listener to build the AST
	listener := newNeuroScriptListener() // Assumes func is defined in ast_builder.go

	// *** This is the Walk call you mentioned ***
	// It walks the 'tree' using your 'listener' implementation
	// *** FIX: Use antlr.ParseTreeWalkerDefault instead of antlr.ParseTreeWalker.Default ***
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Return the result built by the listener
	return listener.GetResult(), nil // Assumes GetResult func exists on listener
}

// --- Custom Error Listener --- (Keep this part as well)

type neuroScriptErrorListener struct {
	*antlr.DefaultErrorListener // Embed default listener
	errors                      []string
}

func newNeuroScriptErrorListener() *neuroScriptErrorListener {
	return &neuroScriptErrorListener{
		errors: make([]string, 0),
	}
}

func (l *neuroScriptErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errMsg := fmt.Sprintf("line %d:%d %s", line, column, msg)
	l.errors = append(l.errors, errMsg)
}

func (l *neuroScriptErrorListener) HasErrors() bool {
	return len(l.errors) > 0
}

func (l *neuroScriptErrorListener) GetErrors() string {
	return strings.Join(l.errors, "\n")
}

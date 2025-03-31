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

	// --- Error Handling Setup ---
	// ** V14: Use exported names **
	lexerErrorListener := NewDiagnosticErrorListener()
	lexer.RemoveErrorListeners() // Remove default console listener
	lexer.AddErrorListener(lexerErrorListener)

	// Create token stream
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the parser
	parser := gen.NewNeuroScriptParser(tokenStream) // Use alias

	// --- Add Diagnostic Error Listener to Parser ---
	// ** V14: Use exported names **
	parserErrorListener := NewDiagnosticErrorListener()
	parser.RemoveErrorListeners()                // Remove default console listener
	parser.AddErrorListener(parserErrorListener) // Add our custom one

	// Optional: Enable ANTLR Trace Listener for extreme detail (prints every step)
	// parser.SetTrace(true) // Uncomment for trace output

	// Start parsing at the 'program' rule
	tree := parser.Program() // This returns the ParseTree

	// Check for parse errors collected by the listeners
	// Combine lexer and parser errors
	combinedErrors := append(lexerErrorListener.Errors, parserErrorListener.Errors...) // Use exported Errors field
	if len(combinedErrors) > 0 {
		// Provide combined error message
		return nil, fmt.Errorf("parsing failed with errors:\n%s", strings.Join(combinedErrors, "\n"))
	}

	// --- AST Building ---
	// Create our custom listener AFTER checking for fundamental parse errors
	listener := newNeuroScriptListener() // Assumes func is defined in ast_builder.go

	// Walk the tree using our listener only if no syntax errors occurred
	// *** FIX: Use antlr.ParseTreeWalkerDefault instead of antlr.ParseTreeWalker.Default ***
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Return the result built by the listener
	return listener.GetResult(), nil // Assumes GetResult func exists on listener
}

// --- Custom Diagnostic Error Listener ---

// ** V14: Renamed struct to be exported **
type DiagnosticErrorListener struct {
	*antlr.DefaultErrorListener          // Embed default listener
	Errors                      []string // ** V14: Renamed field to be exported **
}

// ** V14: Renamed constructor to be exported **
func NewDiagnosticErrorListener() *DiagnosticErrorListener {
	return &DiagnosticErrorListener{
		Errors: make([]string, 0), // Use exported Errors field
	}
}

// Override SyntaxError for more detailed reporting
// ** V14: Renamed receiver variable 'l' to receiver type 'd' for convention **
func (d *DiagnosticErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	// Basic error message
	basicErrMsg := fmt.Sprintf("line %d:%d %s", line, column, msg)

	// Add more details if possible
	var extraInfo []string
	if p, ok := recognizer.(antlr.Parser); ok {
		// --- Cannot reliably get Expected Tokens due to unexported methods ---
		extraInfo = append(extraInfo, "Expected: (Unavailable)")

		// Get parser rule stack (Optional, can be very verbose)
		// extraInfo = append(extraInfo, fmt.Sprintf("RuleStack: %v", p.GetRuleInvocationStack()))

		// Get offending token details
		if offendingToken, ok := offendingSymbol.(antlr.Token); ok {
			tokenName := "Unknown"
			// Ensure token type index is valid before accessing slice
			symbolicNames := p.GetSymbolicNames()
			tokenType := offendingToken.GetTokenType()

			if tokenType >= 0 && tokenType < len(symbolicNames) {
				tokenName = symbolicNames[tokenType]
			} else if tokenType == antlr.TokenEOF {
				tokenName = "EOF"
			}
			extraInfo = append(extraInfo, fmt.Sprintf("OffendingToken: %s (%q)", tokenName, offendingToken.GetText()))
		} else {
			extraInfo = append(extraInfo, fmt.Sprintf("OffendingSymbol: %v (Not an antlr.Token)", offendingSymbol))
		}

	} else if _, ok := recognizer.(antlr.Lexer); ok {
		// Handle Lexer errors
		extraInfo = append(extraInfo, "(Lexer Error)")
		if offendingSymbol != nil {
			extraInfo = append(extraInfo, fmt.Sprintf("OffendingSymbol: %v", offendingSymbol))
		}
		// Exception 'e' for lexer errors might be nil or less informative
		if e != nil {
			extraInfo = append(extraInfo, fmt.Sprintf("Lexer Exception: %T", e))
		}
	}

	detailedErrMsg := basicErrMsg
	if len(extraInfo) > 0 {
		detailedErrMsg += " [" + strings.Join(extraInfo, ", ") + "]"
	}

	// ** V14: Use exported Errors field **
	d.Errors = append(d.Errors, detailedErrMsg)
}

// ** V14: Renamed receiver variable 'l' to receiver type 'd' for convention **
func (d *DiagnosticErrorListener) HasErrors() bool {
	// ** V14: Use exported Errors field **
	return len(d.Errors) > 0
}

// ** V14: Renamed receiver variable 'l' to receiver type 'd' for convention **
func (d *DiagnosticErrorListener) GetErrors() string {
	// ** V14: Use exported Errors field **
	return strings.Join(d.Errors, "\n")
}

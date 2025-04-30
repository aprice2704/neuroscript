// filename: pkg/core/parser_api.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	// Use a shorter alias for the generated package
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/interfaces" // Ensure this path is correct relative to your project structure
)

// ParserAPI provides a simplified interface to the ANTLR parser.
type ParserAPI struct {
	logger interfaces.Logger
}

// NewParserAPI creates a new ParserAPI instance.
func NewParserAPI(logger interfaces.Logger) *ParserAPI {
	if logger == nil {
		// Fallback to a no-op logger if nil is provided
		logger = &interfaces.NoOpLogger{} // Use the exported NoOpLogger
	}
	return &ParserAPI{logger: logger}
}

// ErrorListener captures syntax errors during lexing and parsing.
type ErrorListener struct {
	*antlr.DefaultErrorListener                   // Embed default listener behavior
	Errors                      []string          // Store collected error messages
	logger                      interfaces.Logger // Logger for reporting errors
}

// SyntaxError is called by ANTLR when a syntax error is encountered.
// It formats the error message and stores it.
func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errorMsg := fmt.Sprintf("line %d:%d: %s", line, column, msg)
	l.Errors = append(l.Errors, errorMsg)
	if l.logger != nil {
		// Log the error using the provided logger
		l.logger.Error("[SYNTAX ERROR] %s", errorMsg)
	} else {
		// Fallback to standard output if no logger is available (should not happen ideally)
		fmt.Printf("[SYNTAX ERROR] %s\n", errorMsg)
	}
}

// Parse takes NeuroScript source code as input and returns the ANTLR parse tree (antlr.Tree).
// It handles setting up the lexer, parser, error listeners, and initiating the parsing process.
func (p *ParserAPI) Parse(source string) (antlr.Tree, error) {
	p.logger.Debug("Parsing source code (length: %d)", len(source))

	// 1. Create an ANTLR input stream from the source string.
	inputStream := antlr.NewInputStream(source)

	// 2. Create the lexer using the generated lexer type.
	lexer := gen.NewNeuroScriptLexer(inputStream)

	// 3. Remove the default console error listener from the lexer.
	// We use our custom listener to collect errors.
	lexer.RemoveErrorListeners()

	// 4. Add our custom error listener to the lexer.
	lexerErrorListener := &ErrorListener{logger: p.logger}
	lexer.AddErrorListener(lexerErrorListener)

	// 5. Create a token stream from the lexer's output.
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// 6. Create the parser using the generated parser type.
	parser := gen.NewNeuroScriptParser(stream)

	// 7. Remove the default console error listener from the parser.
	parser.RemoveErrorListeners()

	// 8. Add our custom error listener to the parser.
	parserErrorListener := &ErrorListener{logger: p.logger}
	parser.AddErrorListener(parserErrorListener)

	// 9. Set the error handling strategy. DefaultErrorStrategy attempts recovery.
	// BailErrorStrategy would stop at the first error.
	parser.SetErrorHandler(antlr.NewDefaultErrorStrategy())

	// 10. Start parsing using the root rule defined in the grammar ('program').
	tree := parser.Program() // Use the 'program' rule as the entry point

	// 11. Check for errors collected by the lexer's listener.
	if len(lexerErrorListener.Errors) > 0 {
		p.logger.Error("Lexer Errors encountered during parsing.")
		// Combine lexer errors into a single error message.
		return nil, fmt.Errorf("lexer errors: %s", strings.Join(lexerErrorListener.Errors, "; "))
	}

	// 12. Check for errors collected by the parser's listener.
	// This is now the primary way we detect parser-level syntax errors.
	if len(parserErrorListener.Errors) > 0 {
		p.logger.Error("Parser Errors encountered during parsing.")
		// Combine parser errors into a single error message.
		return nil, fmt.Errorf("parser errors: %s", strings.Join(parserErrorListener.Errors, "; "))
	}

	// --- REMOVED: The check for parser.GetNumberOfSyntaxErrors() ---
	// The custom ErrorListener (parserErrorListener) should reliably capture
	// any syntax errors reported by the parser during the parsing process.

	// 13. If no errors were found by the listeners, log success and return the parse tree.
	p.logger.Debug("Parsing completed successfully.")
	return tree, nil
}

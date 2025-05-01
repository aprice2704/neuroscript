// filename: pkg/core/parser_api.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	// Use a shorter alias for the generated package

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/logging" // Ensure this path is correct relative to your project structure
)

// ParserAPI provides a simplified interface to the ANTLR parser.
type ParserAPI struct {
	logger logging.Logger
}

// NewParserAPI creates a new ParserAPI instance.
func NewParserAPI(logger logging.Logger) *ParserAPI {
	if logger == nil {
		// Fallback to a no-op logger if nil is provided
		logger = &coreNoOpLogger{}
	}
	return &ParserAPI{logger: logger}
}

// ErrorListener captures syntax errors during lexing and parsing.
type ErrorListener struct {
	*antlr.DefaultErrorListener                // Embed default listener behavior
	Errors                      []string       // Store collected error messages
	logger                      logging.Logger // Logger for reporting errors
}

// NewErrorListener creates a new ErrorListener instance.
// It requires a logger; if nil is passed, it uses a no-op logger.
func NewErrorListener(logger logging.Logger) *ErrorListener {
	if logger == nil {
		logger = &coreNoOpLogger{} // Ensure logger is never nil
	}
	return &ErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		Errors:               make([]string, 0),
		logger:               logger,
	}
}

// SyntaxError is called by ANTLR when a syntax error is encountered.
// It formats the error message and stores it.
func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errorMsg := fmt.Sprintf("line %d:%d: %s", line, column, msg)
	l.Errors = append(l.Errors, errorMsg)
	if l.logger != nil {
		// Log the error using structured logging - CORRECTED
		l.logger.Error("Syntax Error", "line", line, "column", column, "message", msg)
	} else {
		// Fallback to standard output if no logger is available (should not happen with NewErrorListener fix)
		fmt.Printf("[SYNTAX ERROR] %s\n", errorMsg)
	}
}

// Parse takes NeuroScript source code as input and returns the ANTLR parse tree (antlr.Tree).
// It handles setting up the lexer, parser, error listeners, and initiating the parsing process.
func (p *ParserAPI) Parse(source string) (antlr.Tree, error) {
	// Use structured logging - CORRECTED
	p.logger.Debug("Parsing source code", "length", len(source))

	// 1. Create an ANTLR input stream from the source string.
	inputStream := antlr.NewInputStream(source)

	// 2. Create the lexer using the generated lexer type.
	lexer := gen.NewNeuroScriptLexer(inputStream)

	// 3. Remove the default console error listener from the lexer.
	lexer.RemoveErrorListeners()

	// 4. Add our custom error listener to the lexer.
	lexerErrorListener := NewErrorListener(p.logger) // Pass logger
	lexer.AddErrorListener(lexerErrorListener)

	// 5. Create a token stream from the lexer's output.
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// 6. Create the parser using the generated parser type.
	parser := gen.NewNeuroScriptParser(stream)

	// 7. Remove the default console error listener from the parser.
	parser.RemoveErrorListeners()

	// 8. Add our custom error listener to the parser.
	parserErrorListener := NewErrorListener(p.logger) // Pass logger
	parser.AddErrorListener(parserErrorListener)

	// 9. Set the error handling strategy.
	parser.SetErrorHandler(antlr.NewDefaultErrorStrategy())

	// 10. Start parsing using the root rule defined in the grammar ('program').
	tree := parser.Program() // Use the 'program' rule as the entry point

	// 11. Check for errors collected by the lexer's listener.
	if len(lexerErrorListener.Errors) > 0 {
		// Use structured logging - CORRECTED
		p.logger.Error("Lexer Errors encountered during parsing.", "count", len(lexerErrorListener.Errors))
		// Combine lexer errors into a single error message.
		return nil, fmt.Errorf("lexer errors: %s", strings.Join(lexerErrorListener.Errors, "; "))
	}

	// 12. Check for errors collected by the parser's listener.
	if len(parserErrorListener.Errors) > 0 {
		// Use structured logging - CORRECTED
		p.logger.Error("Parser Errors encountered during parsing.", "count", len(parserErrorListener.Errors))
		// Combine parser errors into a single error message.
		return nil, fmt.Errorf("parser errors: %s", strings.Join(parserErrorListener.Errors, "; "))
	}

	// 13. If no errors were found by the listeners, log success and return the parse tree.
	// Use structured logging
	p.logger.Debug("Parsing completed successfully.")
	return tree, nil
}

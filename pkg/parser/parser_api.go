// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Utilized existing NoOpLogger from pkg/adapters.
// Purpose: Provides a simplified interface to the ANTLR parser for NeuroScript,
//          including capabilities for structured error reporting for LSP.
// filename: pkg/parser/parser_api.go
// nlines: 185 // Approximate
// risk_rating: MEDIUM // Parser interactions are core to functionality.

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// StructuredSyntaxError holds detailed information about a single syntax error.
// This structure is primarily for LSP and detailed diagnostic consumers.
type StructuredSyntaxError struct {
	Line            int    // 1-based line number from ANTLR
	Column          int    // 0-based character types.Position in line from ANTLR
	OffendingSymbol string // Text of the offending token, if available
	Msg             string // The error message from the parser/lexer
	SourceName      string // e.g., file URI or filename, for context
}

// ParserAPI provides a simplified interface to the ANTLR parser.
type ParserAPI struct {
	logger interfaces.Logger
}

// NewParserAPI creates a new ParserAPI instance.
func NewParserAPI(logger interfaces.Logger) *ParserAPI {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &ParserAPI{logger: logger}
}

// ErrorListener captures syntax errors during lexing and parsing.
// It now collects both raw formatted strings and structured error details.
type ErrorListener struct {
	*antlr.DefaultErrorListener
	RawErrors        []string // Stores formatted error strings for general use
	StructuredErrors []StructuredSyntaxError
	logger           interfaces.Logger
	sourceName       string // Optional: for context in structured errors, set by specific constructors
}

// NewErrorListener creates a new ErrorListener instance, primarily for raw error string collection.
// This constructor is kept for compatibility with the existing Parse method.
func NewErrorListener(logger interfaces.Logger) *ErrorListener {
	return newErrorListenerWithSource("", logger) // Call common constructor with no sourceName
}

// NewLSPErrorListener creates a new ErrorListener with a sourceName, suitable for LSP.
// This is a new routine.
func NewLSPErrorListener(sourceName string, logger interfaces.Logger) *ErrorListener {
	return newErrorListenerWithSource(sourceName, logger)
}

// common constructor for ErrorListener
func newErrorListenerWithSource(sourceName string, logger interfaces.Logger) *ErrorListener {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &ErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		RawErrors:            make([]string, 0),
		StructuredErrors:     make([]StructuredSyntaxError, 0),
		logger:               logger,
		sourceName:           sourceName,
	}
}

// SyntaxError formats and stores syntax errors.
// It populates both RawErrors (formatted string) and StructuredErrors.
func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	finalLine := line     // ANTLR line is 1-based
	finalColumn := column // ANTLR column is 0-based character types.Position in line
	finalMsg := msg
	offendingTokenText := ""

	if token, ok := offendingSymbol.(antlr.Token); ok {
		finalLine = token.GetLine()
		finalColumn = token.GetColumn()
		offendingTokenText = token.GetText()
	}

	// Populate StructuredError
	structuredErr := StructuredSyntaxError{
		Line:            finalLine,
		Column:          finalColumn,
		OffendingSymbol: offendingTokenText,
		Msg:             finalMsg,
		SourceName:      l.sourceName, // Will be empty if NewErrorListener was used without sourceName
	}
	l.StructuredErrors = append(l.StructuredErrors, structuredErr)

	// Populate RawError string (using 1-based column for user messages)
	displayOffendingText := offendingTokenText
	if len(displayOffendingText) > 40 {
		displayOffendingText = displayOffendingText[:37] + "..."
	}
	errorMsgForLog := fmt.Sprintf("line %d:%d: %s%s", finalLine, finalColumn+1, finalMsg, fmt.Sprintf(" near token '%s'", displayOffendingText))
	l.RawErrors = append(l.RawErrors, errorMsgForLog)

	effectiveSourceName := l.sourceName
	if effectiveSourceName == "" {
		effectiveSourceName = "unknown_source"
	}
	if l.logger != nil {
		l.logger.Error("Syntax Error Reported by Listener", "source", effectiveSourceName, "line", finalLine, "column", finalColumn, "message", finalMsg, "token", strings.TrimSpace(offendingTokenText))
	} else {
		fmt.Printf("[SYNTAX ERROR LISTENER - %s] %s\n", effectiveSourceName, errorMsgForLog)
	}
}

// GetStructuredErrors returns the list of structured syntax errors. (New routine)
func (l *ErrorListener) GetStructuredErrors() []StructuredSyntaxError {
	return l.StructuredErrors
}

// GetRawErrors returns the list of raw formatted error strings. (New routine for explicit access)
func (l *ErrorListener) GetRawErrors() []string {
	return l.RawErrors
}

// Parse performs lexing and parsing. (Original signature and primary behavior maintained)
// It now uses an ErrorListener that also collects structured errors internally,
// but this Parse method continues to return errors as a single formatted string.
// For structured errors, use ParseForLSP.
func (p *ParserAPI) Parse(source string) (antlr.Tree, error) {
	p.logger.Debug("Parsing source code (original Parse method)", "length", len(source))
	inputStream := antlr.NewInputStream(source)

	lexer := gen.NewNeuroScriptLexer(inputStream)
	lexer.RemoveErrorListeners()
	lexerErrorListener := NewErrorListener(p.logger)
	lexer.AddErrorListener(lexerErrorListener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)
	parser.RemoveErrorListeners()
	parserErrorListener := NewErrorListener(p.logger)
	parser.AddErrorListener(parserErrorListener)

	parser.SetErrorHandler(antlr.NewDefaultErrorStrategy())
	p.logger.Debug("Using DefaultErrorStrategy for parsing.")

	tree := parser.Program()

	numListenerLexerErrors := len(lexerErrorListener.GetRawErrors())
	numListenerParserErrors := len(parserErrorListener.GetRawErrors())
	p.logger.Debug("Parse attempt completed.", "lexerListenerErrors", numListenerLexerErrors, "parserListenerErrors", numListenerParserErrors)

	if numListenerLexerErrors > 0 {
		p.logger.Error("Lexer Errors encountered.", "count", numListenerLexerErrors)
		return nil, fmt.Errorf("lexer errors: %s", strings.Join(lexerErrorListener.GetRawErrors(), "; "))
	}
	if numListenerParserErrors > 0 {
		p.logger.Error("Parser Listener reported errors.", "count", numListenerParserErrors)
		return tree, fmt.Errorf("parser errors: %s", strings.Join(parserErrorListener.GetRawErrors(), "; "))
	}

	p.logger.Debug("Parsing successful (based on listener raw errors).")
	return tree, nil
}

// ParseForLSP performs lexing and parsing, returning the AST and a slice of structured errors.
// This is a new routine, more suitable for LSP diagnostics.
func (p *ParserAPI) ParseForLSP(sourceName string, sourceContent string) (antlr.Tree, []StructuredSyntaxError) {
	p.logger.Debug("Parsing for LSP", "sourceName", sourceName, "length", len(sourceContent))
	inputStream := antlr.NewInputStream(sourceContent)

	lexer := gen.NewNeuroScriptLexer(inputStream)
	lexer.RemoveErrorListeners()
	errorListener := NewLSPErrorListener(sourceName, p.logger)
	lexer.AddErrorListener(errorListener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	parser.SetErrorHandler(antlr.NewDefaultErrorStrategy())
	tree := parser.Program()

	structuredErrors := errorListener.GetStructuredErrors()

	if len(structuredErrors) > 0 {
		p.logger.Debug("Syntax errors found during LSP parse", "sourceName", sourceName, "count", len(structuredErrors))
	} else {
		p.logger.Debug("LSP Parse successful (no syntax errors).", "sourceName", sourceName)
	}

	return tree, structuredErrors
}

// filename: pkg/core/parser_api.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// ParserAPI provides a simplified interface to the ANTLR parser.
type ParserAPI struct {
	logger logging.Logger
}

// NewParserAPI creates a new ParserAPI instance.
func NewParserAPI(logger logging.Logger) *ParserAPI {
	if logger == nil {
		logger = &coreNoOpLogger{}
	}
	return &ParserAPI{logger: logger}
}

// ErrorListener captures syntax errors during lexing and parsing.
type ErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
	logger logging.Logger
}

// NewErrorListener creates a new ErrorListener instance.
func NewErrorListener(logger logging.Logger) *ErrorListener {
	if logger == nil {
		logger = &coreNoOpLogger{}
	}
	return &ErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		Errors:               make([]string, 0),
		logger:               logger,
	}
}

// SyntaxError formats and stores syntax errors.
func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	finalLine := line
	finalColumn := column
	finalMsg := msg
	offendingTokenText := ""

	if token, ok := offendingSymbol.(antlr.Token); ok {
		finalLine = token.GetLine()
		finalColumn = token.GetColumn()
		offendingTokenText = fmt.Sprintf(" near token '%s'", token.GetText())
	}

	// Use 1-based column for user messages
	errorMsg := fmt.Sprintf("line %d:%d: %s%s", finalLine, finalColumn+1, finalMsg, offendingTokenText)
	l.Errors = append(l.Errors, errorMsg)

	if l.logger != nil {
		// Log 0-based column internally
		l.logger.Error("Syntax Error Reported by Listener", "line", finalLine, "column", finalColumn, "message", finalMsg, "token", strings.TrimSpace(offendingTokenText))
	} else {
		fmt.Printf("[SYNTAX ERROR LISTENER] %s\n", errorMsg)
	}
}

// Parse performs lexing and parsing.
func (p *ParserAPI) Parse(source string) (antlr.Tree, error) {
	p.logger.Debug("Parsing source code", "length", len(source))
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

	// *** Use DefaultErrorStrategy again ***
	parser.SetErrorHandler(antlr.NewDefaultErrorStrategy())
	p.logger.Debug("Using DefaultErrorStrategy for parsing.")

	tree := parser.Program()

	// Check errors (reverted to trusting listener)
	numListenerLexerErrors := len(lexerErrorListener.Errors)
	numListenerParserErrors := len(parserErrorListener.Errors)
	p.logger.Debug("Parse attempt completed.", "lexerListenerErrors", numListenerLexerErrors, "parserListenerErrors", numListenerParserErrors)

	if numListenerLexerErrors > 0 {
		p.logger.Error("Lexer Errors encountered.", "count", numListenerLexerErrors)
		return nil, fmt.Errorf("lexer errors: %s", strings.Join(lexerErrorListener.Errors, "; "))
	}
	if numListenerParserErrors > 0 {
		p.logger.Error("Parser Listener reported errors.", "count", numListenerParserErrors)
		return nil, fmt.Errorf("parser errors: %s", strings.Join(parserErrorListener.Errors, "; "))
	}

	p.logger.Debug("Parsing successful (based on listener).")
	return tree, nil
}

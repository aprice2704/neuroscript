// pkg/core/parser_api.go
package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"log" // Import log package
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ParseOptions structure to pass flags/logger
type ParseOptions struct {
	DebugAST bool
	Logger   *log.Logger // Use standard log.Logger
}

// ParseNeuroScript reads NeuroScript code, parses it using ANTLR,
// builds the AST using a listener, and returns the Procedures.
// Updated signature to accept filename and options.
func ParseNeuroScript(r io.Reader, sourceName string, options ParseOptions) ([]Procedure, error) {
	// Use logger from options, default to discard if nil
	logger := options.Logger
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	inputBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading input '%s': %w", sourceName, err)
	}
	inputString := string(inputBytes)

	inputStream := antlr.NewInputStream(inputString)
	lexer := gen.NewNeuroScriptLexer(inputStream)

	// Use DiagnosticErrorListener for detailed errors
	lexerErrorListener := NewDiagnosticErrorListener(sourceName) // Pass source name
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrorListener)

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(tokenStream)

	parserErrorListener := NewDiagnosticErrorListener(sourceName) // Pass source name
	parser.RemoveErrorListeners()
	parser.AddErrorListener(parserErrorListener)

	// Start parsing
	tree := parser.Program()

	// Combine lexer and parser errors
	combinedErrors := append(lexerErrorListener.Errors, parserErrorListener.Errors...)
	if len(combinedErrors) > 0 {
		return nil, fmt.Errorf("parsing '%s' failed:\n%s", sourceName, strings.Join(combinedErrors, "\n"))
	}

	// AST Building - Pass logger and debug flag to the listener
	listener := newNeuroScriptListener(logger, options.DebugAST) // Pass logger/flag
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Return the result built by the listener
	return listener.GetResult(), nil
}

// --- Custom Diagnostic Error Listener ---

type DiagnosticErrorListener struct {
	*antlr.DefaultErrorListener
	Errors     []string
	SourceName string // Store the source filename
}

// NewDiagnosticErrorListener constructor takes the source name.
func NewDiagnosticErrorListener(sourceName string) *DiagnosticErrorListener {
	return &DiagnosticErrorListener{
		Errors:     make([]string, 0),
		SourceName: sourceName, // Store filename
	}
}

// Override SyntaxError for more detailed reporting including filename.
func (d *DiagnosticErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	// Include filename in the error message
	detailedErrMsg := fmt.Sprintf("%s:%d:%d %s", d.SourceName, line, column, msg)

	// Add ANTLR specific details (optional, can be verbose)
	/*
		var extraInfo []string
		if p, ok := recognizer.(antlr.Parser); ok {
			// ... (existing code to get expected tokens, rule stack, etc.) ...
		}
		if len(extraInfo) > 0 {
			detailedErrMsg += " [" + strings.Join(extraInfo, ", ") + "]"
		}
	*/

	d.Errors = append(d.Errors, detailedErrMsg)
}

// HasErrors and GetErrors remain the same conceptually

func (d *DiagnosticErrorListener) HasErrors() bool {
	return len(d.Errors) > 0
}

func (d *DiagnosticErrorListener) GetErrors() string {
	return strings.Join(d.Errors, "\n")
}

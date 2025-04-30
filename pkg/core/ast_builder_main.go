// filename: pkg/core/ast_builder_main.go
package core

import (
	"errors" // Import errors package
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Position Helper ---

// tokenToPosition converts an ANTLR token to a core.Position.
func tokenToPosition(token antlr.Token) *Position {
	if token == nil {
		return nil // Or return a default position?
	}
	return &Position{
		Line:   token.GetLine(),
		Column: token.GetColumn() + 1, // ANTLR columns are 0-based, prefer 1-based
		File:   token.GetInputStream().GetSourceName(),
	}
}

// --- ASTBuilder (Exported Constructor and Build Method) ---

// ASTBuilder encapsulates the logic for building the NeuroScript AST using a listener.
type ASTBuilder struct {
	logger   logging.Logger
	debugAST bool // Option to enable detailed AST construction logging
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger logging.Logger) *ASTBuilder {
	if logger == nil {
		// Using the NoOpLogger from adapters for safety
		logger = &coreNoOpLogger{} // Assumes NoOpLogger is available in core or adapters package
		logger.Warn("ASTBuilder created with nil logger, using NoOpLogger.")
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: false, // Default to false, could be configurable
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
// *** MODIFIED: Checks for errors collected by the listener. ***
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, error) {
	if tree == nil {
		return nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST)

	// Walk the parse tree with the listener.
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		// Return the first error encountered
		return nil, listener.errors[0]
	}

	// Assemble Program AST using collected metadata and procedures.
	programAST := listener.program // Get the assembled program

	if programAST == nil {
		// Should not happen if no errors were reported, but check defensively
		return nil, errors.New("AST build completed without errors, but resulted in a nil program AST")
	}

	b.logger.Debug("AST build process completed successfully.")
	return programAST, nil
}

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---

// neuroScriptListenerImpl builds the AST using the Listener pattern.
// *** MODIFIED: Added program field and errors slice. ***
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program        *Program          // Field to hold the final program AST
	fileMetadata   map[string]string // For file-level metadata
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{}
	currentMapKey  *StringLiteralNode // Keep for map literal building
	// blockSteps     map[antlr.ParserRuleContext][]Step // REMOVED: Direct building preferred over map
	logger   logging.Logger
	debugAST bool
	errors   []error // Slice to collect errors during build
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		program:        &Program{}, // Initialize program struct
		fileMetadata:   make(map[string]string),
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10),
		// blockSteps:     make(map[antlr.ParserRuleContext][]Step), // REMOVED
		logger:   logger,
		debugAST: debugAST,
		errors:   make([]error, 0), // Initialize errors slice
	}
}

// --- Listener Error Handling ---

// addError records an error encountered during AST building.
func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	pos := tokenToPosition(ctx.GetStart())
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	l.errors = append(l.errors, err)
	l.logger.Error(err.Error()) // Log the error immediately as well
}

// addErrorf creates and adds an error with file/line info.
func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token)
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	l.errors = append(l.errors, err)
	l.logger.Error(err.Error())
}

// --- Listener Getters ---

// GetFileMetadata returns the collected file-level metadata.
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	// Return from the program struct
	if l.program != nil {
		return l.program.Metadata
	}
	return l.fileMetadata // Fallback just in case
}

// GetResult returns the collected procedures (now part of Program).
func (l *neuroScriptListenerImpl) GetResult() []Procedure {
	// Return from the program struct
	if l.program != nil {
		return l.program.Procedures
	}
	return l.procedures // Fallback
}

// --- Listener Stack Helpers ---
// *** MODIFIED: Add error reporting on stack issues ***

func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	// Use more detailed logging if debugAST is on
	if l.debugAST {
		l.logDebugAST("      Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
	}
}

func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
		// Add error to listener errors - use a generic context if none available
		l.errors = append(l.errors, errors.New("AST builder internal error: attempted pop from empty value stack"))
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	if l.debugAST {
		l.logDebugAST("      Popped Value: %T %+v (Stack size: %d)", value, value, len(l.valueStack))
	}
	return value, true
}

func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if n < 0 {
		l.logger.Error("AST Builder: popNValues called with negative count", "n", n)
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: popNValues called with negative count %d", n))
		return nil, false
	}
	if len(l.valueStack) < n {
		l.logger.Error("AST Builder: Stack underflow pop %d, have %d.", n, len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: stack underflow, needed %d values, only have %d", n, len(l.valueStack)))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	// Safe copy
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	l.logDebugAST("      Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true
}

// --- Listener Logging Helper (Unchanged) ---

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(format, v...)
	}
}

// --- Listener ANTLR Method Implementations ---

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.program = &Program{ // Initialize the program AST node
		Metadata:   make(map[string]string),
		Procedures: make([]Procedure, 0),
		Pos:        tokenToPosition(ctx.GetStart()), // Record start position
	}
	l.fileMetadata = l.program.Metadata // Use program's metadata map
	l.procedures = l.program.Procedures // Use program's procedures slice
	l.errors = make([]error, 0)         // Reset errors for this build
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	// Update program fields after visiting children
	l.program.Metadata = l.fileMetadata
	l.program.Procedures = l.procedures
	l.logDebugAST("<<< Exit Program (Metadata Keys: %v, Procedures: %d, Errors: %d)", MapKeysListener(l.program.Metadata), len(l.program.Procedures), len(l.errors))
}

// --- MODIFIED: Metadata Handling via file_header ---

// EnterFile_header processes all metadata lines found at the start of the file.
func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  >> Enter File Header")
	// Iterate through all children of the header context
	for _, child := range ctx.GetChildren() {
		// Check if the child is a METADATA_LINE terminal node
		if termNode, ok := child.(antlr.TerminalNode); ok && termNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMETADATA_LINE {
			lineText := termNode.GetText()
			token := termNode.GetSymbol()
			l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
			// Parse the line
			lineText = strings.TrimSpace(lineText)
			if strings.HasPrefix(lineText, "::") {
				trimmedLine := strings.TrimSpace(lineText[2:])
				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1]) // TODO: Handle potential quoting/escaping
					if key != "" {
						// Use program's metadata map directly
						l.program.Metadata[key] = value
						l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
					} else {
						l.addErrorf(token, "Ignoring file metadata line with empty key")
					}
				} else {
					l.addErrorf(token, "Ignoring malformed file metadata line (missing or misplaced ':'?)")
				}
			} else {
				// Should not happen if lexer rule is correct
				l.addErrorf(token, "Unexpected line format in file_header (missing '::'?)")
			}
		}
		// Ignore NEWLINE tokens within the file_header
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  << Exit File Header")
}

// --- END MODIFIED Methods ---

// MapKeysListener is a helper function (consider moving to utils)
func MapKeysListener(m map[string]string) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- Methods to be implemented in other ast_builder_*.go files ---
// EnterProcedure_definition, ExitProcedure_definition
// EnterSet_stmt, ExitSet_stmt
// EnterReturn_stmt, ExitReturn_stmt
// EnterIf_stmt, ExitIf_stmt
// EnterWhile_stmt, ExitWhile_stmt
// EnterFor_stmt, ExitFor_stmt
// EnterMust_stmt, ExitMust_stmt
// EnterMustbe_stmt, ExitMustbe_stmt
// EnterFail_stmt, ExitFail_stmt
// EnterOn_error_handler, ExitOn_error_handler
// EnterCall_expr, ExitCall_expr
// EnterLiteral_expr, ExitLiteral_expr
// EnterVariable_expr, ExitVariable_expr
// EnterBinary_op_expr, ExitBinary_op_expr
// etc...

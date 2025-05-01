// filename: pkg/core/ast_builder_main.go
package core

import (
	"errors"
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
	// REMOVED: listener field - it's local to Build
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
// It now returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) { // NEW Signature
	if tree == nil {
		// Return nils for program and metadata on initial error
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("Starting AST build process using Listener.")

	// Create the listener instance.
	// Assumes newNeuroScriptListener is defined correctly and returns *neuroScriptListenerImpl
	listener := newNeuroScriptListener(b.logger, b.debugAST)

	// Walk the parse tree with the listener.
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)

	// Get metadata *after* the walk, before returning on error.
	// Assumes listener has GetFileMetadata method returning map[string]string.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		// Ensure metadata map is not nil, even if empty
		fileMetadata = make(map[string]string)
		b.logger.Warn("Listener returned nil metadata map, initialized empty map.")
	}

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		// Combine errors if multiple exist
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil { // Ensure error is not nil before calling Error()
				errorMessages = append(errorMessages, err.Error())
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
			}
		}
		combinedError := errors.New(strings.Join(errorMessages, "; "))
		// Return potentially partial program AST and metadata even on error
		// It's often useful to have the metadata even if procedures fail to build.
		// Return listener.program here which might be partially built or nil.
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	// Assumes listener has a 'program' field of type *Program
	programAST := listener.program

	if programAST == nil {
		// This case suggests an internal listener error not caught previously
		b.logger.Error("AST build completed without explicit errors, but resulted in a nil program AST")
		// Return metadata even on this internal error
		return nil, fileMetadata, errors.New("AST build completed without errors, but resulted in a nil program AST")
	}

	// Ensure the program's metadata field matches what the listener collected/returned.
	// This should already be the case as the listener modifies the program's map directly.
	// programAST.Metadata = fileMetadata // This line is likely redundant now

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	// Initialize the map (now correct type based on modified ast.go)
	if programAST.Procedures == nil {
		programAST.Procedures = make(map[string]*Procedure) // Correct type now
	}
	duplicateProcs := false
	// Loop iterates over []*Procedure now
	for _, proc := range listener.procedures {
		if proc != nil { // Comparison with nil is now valid for *Procedure
			// Map key check is now valid
			if _, exists := programAST.Procedures[proc.Name]; exists {
				// Log duplicate, and ensure an error is returned by adding to listener.errors
				errorMsg := fmt.Sprintf("duplicate procedure definition: %s", proc.Name)
				b.logger.Error(errorMsg)
				listener.errors = append(listener.errors, errors.New(errorMsg)) // Add error
				duplicateProcs = true
			}
			// Map assignment is now valid
			programAST.Procedures[proc.Name] = proc // Add pointer to map
		}
	}

	// If duplicates were found, report the error state
	if duplicateProcs {
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				errorMessages = append(errorMessages, err.Error())
			}
		}
		// Return the program (with duplicates possibly overwritten) and the error
		return programAST, fileMetadata, errors.New(strings.Join(errorMessages, "; "))
	}

	b.logger.Debug("AST build process completed successfully.")
	// Return program, metadata, and nil error on success
	return programAST, fileMetadata, nil
}

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---

// neuroScriptListenerImpl builds the AST using the Listener pattern.
// *** MODIFIED: Added program field and errors slice. ***
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program      *Program          // Field to hold the final program AST
	fileMetadata map[string]string // Points to program.Metadata
	// procedures     []Procedure    // OLD: Slice of Procedure values
	procedures     []*Procedure // NEW: Temporary slice of Procedure pointers
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{}
	currentMapKey  *StringLiteralNode // Keep for map literal building
	logger         logging.Logger
	debugAST       bool
	errors         []error // Slice to collect errors during build
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	// Initialize program struct with the correct Procedures map type
	prog := &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure), // Initialize the map
		Pos:        nil,                         // Will be set in EnterProgram
	}
	return &neuroScriptListenerImpl{
		program:        prog,
		fileMetadata:   prog.Metadata,         // Point to program's map
		procedures:     make([]*Procedure, 0), // NEW: Initialize slice of pointers
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10),
		logger:         logger,
		debugAST:       debugAST,
		errors:         make([]error, 0), // Initialize errors slice
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
	// Fallback shouldn't be necessary but kept for safety
	return l.fileMetadata
}

// GetResult returns the temporary slice of collected procedures.
// Note: The final result is in program.Procedures map. This might be obsolete.
func (l *neuroScriptListenerImpl) GetResult() []*Procedure { // UPDATED Return Type
	// Return from the temporary slice used during building
	return l.procedures
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
	// Initialize or reset the program struct correctly
	l.program = &Program{ // Initialize the program AST node
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),     // Initialize map
		Pos:        tokenToPosition(ctx.GetStart()), // Record start position
	}
	l.fileMetadata = l.program.Metadata // Use program's metadata map
	// Reset temporary procedures slice for this build
	l.procedures = make([]*Procedure, 0) // Use slice of pointers
	l.errors = make([]error, 0)          // Reset errors for this build
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	// Metadata is already part of l.program
	// Procedures map is populated directly in Build method now
	// Log final state
	procCount := 0
	if l.program != nil && l.program.Procedures != nil {
		procCount = len(l.program.Procedures)
	}
	metaKeys := []string{}
	if l.program != nil && l.program.Metadata != nil {
		metaKeys = MapKeysListener(l.program.Metadata)
	}
	l.logDebugAST("<<< Exit Program (Metadata Keys: %v, Procedures: %d, Errors: %d)", metaKeys, procCount, len(l.errors))
}

// --- MODIFIED: Metadata Handling via file_header ---

// EnterFile_header processes all metadata lines found at the start of the file.
func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  >> Enter File Header")
	if l.program == nil || l.program.Metadata == nil {
		l.logger.Error("EnterFile_header called with nil program or metadata map!")
		l.errors = append(l.errors, errors.New("internal AST builder error: program/metadata nil in EnterFile_header"))
		return // Avoid panic
	}
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
					// Store the whole line as key if no ':' found? Or error? Currently errors.
					l.addErrorf(token, "Ignoring malformed file metadata line (missing ':'?)")
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

// REMOVED ASTBuilder.GetFileMetadata method as it's redundant

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

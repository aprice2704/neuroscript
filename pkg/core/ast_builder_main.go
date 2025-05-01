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
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger logging.Logger) *ASTBuilder {
	if logger == nil {
		logger = &coreNoOpLogger{}
		logger.Warn("ASTBuilder created with nil logger, using NoOpLogger.")
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: true, // Force debug logging for this specific issue
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
// It now returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) {
	if tree == nil {
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST) // Pass debug flag

	// Walk the parse tree with the listener.
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)

	// Get metadata *after* the walk, before returning on error.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
		b.logger.Warn("Listener returned nil metadata map, initialized empty map.")
	}

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				errorMessages = append(errorMessages, err.Error())
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
			}
		}
		combinedError := errors.New(strings.Join(errorMessages, "; "))
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	programAST := listener.program

	if programAST == nil {
		b.logger.Error("AST build completed without explicit errors, but resulted in a nil program AST")
		return nil, fileMetadata, errors.New("AST build completed without errors, but resulted in a nil program AST")
	}

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	if programAST.Procedures == nil {
		programAST.Procedures = make(map[string]*Procedure)
	}
	duplicateProcs := false
	for _, proc := range listener.procedures {
		if proc != nil {
			if _, exists := programAST.Procedures[proc.Name]; exists {
				errorMsg := fmt.Sprintf("duplicate procedure definition: %s", proc.Name)
				b.logger.Error(errorMsg)
				listener.errors = append(listener.errors, errors.New(errorMsg)) // Add error
				duplicateProcs = true
			}
			programAST.Procedures[proc.Name] = proc
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
		return programAST, fileMetadata, errors.New(strings.Join(errorMessages, "; "))
	}

	b.logger.Debug("AST build process completed successfully.")
	return programAST, fileMetadata, nil
}

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program        *Program
	fileMetadata   map[string]string
	procedures     []*Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{}
	currentMapKey  *StringLiteralNode
	logger         logging.Logger
	debugAST       bool
	errors         []error
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	prog := &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Pos:        nil,
	}
	return &neuroScriptListenerImpl{
		program:        prog,
		fileMetadata:   prog.Metadata,
		procedures:     make([]*Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10),
		logger:         logger,
		debugAST:       debugAST, // Use passed-in value
		errors:         make([]error, 0),
	}
}

// --- Listener Error Handling --- (Unchanged)
func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	pos := tokenToPosition(ctx.GetStart())
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	l.errors = append(l.errors, err)
	l.logger.Error(err.Error()) // Log the error immediately as well
}
func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token)
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	l.errors = append(l.errors, err)
	l.logger.Error(err.Error())
}

// --- Listener Getters --- (Unchanged)
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	if l.program != nil {
		return l.program.Metadata
	}
	return l.fileMetadata
}
func (l *neuroScriptListenerImpl) GetResult() []*Procedure {
	return l.procedures
}

// --- Listener Stack Helpers ---
// *** ADDED More Verbose Logging for Stack Ops ***
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	// Always log push operations for debugging this issue
	l.logger.Debug("[DEBUG-AST-STACK] --> PUSH", "value_type", fmt.Sprintf("%T", v), "value", fmt.Sprintf("%+v", v), "new_stack_size", len(l.valueStack)+1)
	l.valueStack = append(l.valueStack, v)
}

func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
		l.errors = append(l.errors, errors.New("AST builder internal error: attempted pop from empty value stack"))
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	// Always log pop operations
	l.logger.Debug("[DEBUG-AST-STACK] <-- POP", "value_type", fmt.Sprintf("%T", value), "value", fmt.Sprintf("%+v", value), "new_stack_size", len(l.valueStack))
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
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	// Always log popN operations
	l.logger.Debug("[DEBUG-AST-STACK] <-- POP N", "count", n, "new_stack_size", len(l.valueStack))
	return values, true
}

// --- Listener Logging Helper (Unchanged) ---
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(format, v...)
	}
}

// --- Listener ANTLR Method Implementations --- (Unchanged)

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.program = &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Pos:        tokenToPosition(ctx.GetStart()),
	}
	l.fileMetadata = l.program.Metadata
	l.procedures = make([]*Procedure, 0)
	l.errors = make([]error, 0)
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
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

func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  >> Enter File Header")
	if l.program == nil || l.program.Metadata == nil {
		l.logger.Error("EnterFile_header called with nil program or metadata map!")
		l.errors = append(l.errors, errors.New("internal AST builder error: program/metadata nil in EnterFile_header"))
		return
	}
	for _, child := range ctx.GetChildren() {
		if termNode, ok := child.(antlr.TerminalNode); ok && termNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMETADATA_LINE {
			lineText := termNode.GetText()
			token := termNode.GetSymbol()
			l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
			lineText = strings.TrimSpace(lineText)
			if strings.HasPrefix(lineText, "::") {
				trimmedLine := strings.TrimSpace(lineText[2:])
				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if key != "" {
						l.program.Metadata[key] = value
						l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
					} else {
						l.addErrorf(token, "Ignoring file metadata line with empty key")
					}
				} else {
					l.addErrorf(token, "Ignoring malformed file metadata line (missing ':'?)")
				}
			} else {
				l.addErrorf(token, "Unexpected line format in file_header (missing '::'?)")
			}
		}
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

// Implementations for other Enter/Exit methods are in other ast_builder_*.go files

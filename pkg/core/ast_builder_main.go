// filename: pkg/core/ast_builder_main.go
package core

import (
	"errors"
	"fmt" // Ensure fmt is imported
	"strings"

	"github.com/antlr4-go/antlr/v4"                            // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated" // Corrected path
	"github.com/aprice2704/neuroscript/pkg/logging"            // Keep logging import for interface
)

// --- Position Helper ---

// tokenToPosition converts an ANTLR token to a core.Position.
// It sets the exported fields Line, Column, and File.
func tokenToPosition(token antlr.Token) *Position {
	if token == nil {
		// *** CORRECTED LINE BELOW: Removed unexported 'token' field ***
		return &Position{Line: 0, Column: 0, File: "<nil token>"} // Return a default invalid position
	}
	// Handle potential nil InputStream or SourceName gracefully
	sourceName := "<unknown>"
	if token.GetInputStream() != nil {
		sourceName = token.GetInputStream().GetSourceName()
		if sourceName == "<INVALID>" { // Use a more descriptive name if ANTLR provides one
			sourceName = "<input stream>"
		}
	}
	// *** CORRECTED LINE BELOW: Removed unexported 'token' field ***
	return &Position{
		Line:   token.GetLine(),
		Column: token.GetColumn() + 1, // ANTLR columns are 0-based, prefer 1-based
		File:   sourceName,
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
		// Use the existing coreNoOpLogger from this package (defined in helpers.go)
		logger = &coreNoOpLogger{}
	}
	// Forcing debugAST true here for continued debugging.
	return &ASTBuilder{
		logger:   logger,
		debugAST: true,
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
// It now returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("AST Builder: Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST) // Pass debug flag

	// Walk the parse tree with the listener.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	// Get metadata *after* the walk, before returning on error.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
		b.logger.Warn("AST Builder: Listener returned nil metadata map, initialized empty map.")
	}

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				errorMessages = append(errorMessages, err.Error())
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
				b.logger.Error("AST Builder: Listener recorded a nil error object.")
			}
		}
		combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Errors detected during ANTLR walk.", "error", combinedError)
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	programAST := listener.program

	if programAST == nil {
		b.logger.Error("AST Builder: Build completed without explicit errors, but resulted in a nil program AST")
		return nil, fileMetadata, errors.New("AST builder internal error: program AST is unexpectedly nil after successful walk")
	}

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	if programAST.Procedures == nil {
		b.logger.Warn("AST Builder: Program AST Procedures map was nil, initializing.")
		programAST.Procedures = make(map[string]*Procedure)
	}

	// --- Debug Logging Added Here ---
	b.logger.Debug("AST Builder: Assembling procedures found by listener.", "count", len(listener.procedures))
	// --- End Debug Logging ---

	duplicateProcs := false
	var firstDuplicateName string

	for i, proc := range listener.procedures { // Iterate listener's temporary slice

		// --- Debug Logging Added Here ---
		b.logger.Debug("AST Builder: Checking procedure from listener list.", "index", i, "proc_pointer", fmt.Sprintf("%p", proc))
		// --- End Debug Logging ---

		if proc != nil {
			// Check for duplicates *before* assigning.
			if existingProc, exists := programAST.Procedures[proc.Name]; exists {
				if !duplicateProcs { // Record the first duplicate found
					firstDuplicateName = proc.Name
				}
				duplicateProcs = true
				// Use the String() method of the Position struct
				errMsg := fmt.Sprintf("duplicate procedure definition: '%s' found at %s conflicts with existing at %s",
					proc.Name, proc.Pos.String(), existingProc.Pos.String())

				b.logger.Error("AST Builder: Duplicate procedure.", "name", proc.Name, "new_pos", proc.Pos.String(), "existing_pos", existingProc.Pos.String())
				listener.errors = append(listener.errors, errors.New(errMsg))
				continue // Skip adding this duplicate procedure
			}
			// Assign the non-nil proc pointer to the map only if it doesn't exist yet
			programAST.Procedures[proc.Name] = proc

		} else {
			b.logger.Warn("AST Builder encountered a nil procedure pointer in the temporary list.", "index", i)
			// Potentially make this fatal:
			// return nil, fileMetadata, fmt.Errorf("internal AST builder error: found nil procedure pointer at index %d", i)
		}
	}

	// If duplicates were found OR other errors exist, aggregate errors and return.
	if duplicateProcs || len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		hasNonDuplicateError := false
		uniqueErrors := make(map[string]struct{}) // Avoid adding same error msg multiple times

		currentErrors := listener.errors
		for _, err := range currentErrors {
			if err != nil {
				msg := err.Error()
				if _, seen := uniqueErrors[msg]; !seen {
					errorMessages = append(errorMessages, msg)
					uniqueErrors[msg] = struct{}{}
					if !strings.Contains(msg, "duplicate procedure definition") {
						hasNonDuplicateError = true
					}
				}
			}
		}
		errorPrefix := "AST build failed"
		if duplicateProcs && !hasNonDuplicateError {
			errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s')", firstDuplicateName)
		} else if duplicateProcs {
			errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s') and other errors", firstDuplicateName)
		}

		combinedError := fmt.Errorf("%s: %s", errorPrefix, strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Build failed.", "error", combinedError)
		return programAST, fileMetadata, combinedError
	}

	b.logger.Debug("AST Builder: Build process completed successfully.")
	return programAST, fileMetadata, nil // Return nil error on success
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
		Pos:        nil, // Position set later in EnterProgram
	}
	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 prog,
		fileMetadata:            prog.Metadata,
		procedures:              make([]*Procedure, 0, 10),
		blockStepStack:          make([]*[]Step, 0, 5),
		valueStack:              make([]interface{}, 0, 20),
		logger:                  logger,
		debugAST:                debugAST,
		errors:                  make([]error, 0),
	}
}

// --- Listener Error Handling ---
func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	pos := tokenToPosition(ctx.GetStart())
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}
func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token)
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}

// --- Listener Getters ---
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	if l.program != nil && l.program.Metadata != nil {
		return l.program.Metadata
	}
	if l.fileMetadata == nil {
		l.logger.Warn("GetFileMetadata called when listener.fileMetadata is nil.")
		l.fileMetadata = make(map[string]string)
	}
	return l.fileMetadata
}
func (l *neuroScriptListenerImpl) GetResult() []*Procedure {
	l.logger.Warn("GetResult called on listener; this returns the temporary slice, not the final program map.")
	return l.procedures
}

// --- Listener Stack Helpers ---
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	if l.debugAST {
		valueStr := fmt.Sprintf("%+v", v)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		l.logger.Debug("[DEBUG-AST-STACK] --> PUSH", "value_type", fmt.Sprintf("%T", v), "value_preview", valueStr, "new_stack_size", len(l.valueStack)+1)
	}
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
	if l.debugAST {
		valueStr := fmt.Sprintf("%+v", value)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP", "value_type", fmt.Sprintf("%T", value), "value_preview", valueStr, "new_stack_size", len(l.valueStack))
	}
	return value, true
}

func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if n < 0 {
		l.logger.Error("AST Builder: popNValues called with negative count", "n", n)
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: popNValues called with negative count %d", n))
		return nil, false
	}
	if n == 0 {
		return []interface{}{}, true
	}
	if len(l.valueStack) < n {
		l.logger.Error("AST Builder: Stack underflow", "needed", n, "available", len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: stack underflow, needed %d values, only have %d", n, len(l.valueStack)))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP N", "count", n, "new_stack_size", len(l.valueStack))
	}
	return values, true
}

// --- Listener Logging Helper ---
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

// --- Listener ANTLR Method Implementations ---
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.program = &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Pos:        tokenToPosition(ctx.GetStart()),
	}
	l.fileMetadata = l.program.Metadata
	l.procedures = make([]*Procedure, 0, 10)
	l.errors = make([]error, 0)
	l.valueStack = make([]interface{}, 0, 20)
	l.blockStepStack = make([]*[]Step, 0, 5)
	l.currentProc = nil
	l.currentSteps = nil
	l.currentMapKey = nil
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	finalProcCount := 0
	if l.program != nil && l.program.Procedures != nil {
		finalProcCount = len(l.program.Procedures)
	} else if l.program == nil {
		l.logger.Error("ExitProgram: l.program is nil!")
	} else {
		l.logger.Error("ExitProgram: l.program.Procedures is nil!")
	}
	metaCount := 0
	if l.fileMetadata != nil {
		metaCount = len(l.fileMetadata)
	} else {
		l.logger.Error("ExitProgram: l.fileMetadata is nil!")
	}
	l.logDebugAST("<<< Exit Program (Metadata Count: %d, Final Procedure Count: %d, Listener Errors: %d, Final Stack Size: %d)",
		metaCount, finalProcCount, len(l.errors), len(l.valueStack))
	if len(l.valueStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: value stack size is %d at end of program", len(l.valueStack))
		l.logger.Error("ExitProgram: Value stack not empty!", "size", len(l.valueStack), "top_value_type", fmt.Sprintf("%T", l.valueStack[len(l.valueStack)-1]))
		l.errors = append(l.errors, errors.New(errMsg))
	}
	if len(l.blockStepStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack))
		l.logger.Error("ExitProgram: Block step stack not empty!", "size", len(l.blockStepStack))
		l.errors = append(l.errors, errors.New(errMsg))
	}
}

func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  >> Enter File Header")
	if l.program == nil || l.program.Metadata == nil {
		l.logger.Error("EnterFile_header called with nil program or metadata map!")
		l.errors = append(l.errors, errors.New("internal AST builder error: program/metadata nil in EnterFile_header"))
		return
	}
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		lineText := metaLineNode.GetText()
		token := metaLineNode.GetSymbol()
		l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
		lineText = strings.TrimSpace(lineText)
		if strings.HasPrefix(lineText, "::") {
			trimmedLine := strings.TrimSpace(lineText[2:])
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					if _, exists := l.program.Metadata[key]; exists {
						l.logDebugAST("     Overwriting File Metadata: '%s'", key)
					}
					l.program.Metadata[key] = value
					l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
				} else {
					l.addErrorf(token, "Ignoring file metadata line with empty key: '%s'", lineText)
				}
			} else {
				l.addErrorf(token, "Ignoring malformed file metadata line (missing ':' separator?): '%s'", lineText)
			}
		} else {
			l.addErrorf(token, "Unexpected line format in file_header (expected '::' prefix): '%s'", lineText)
		}
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  << Exit File Header")
}

// --- END MODIFIED Methods ---

func MapKeysListener(m map[string]string) []string { // Renamed to avoid conflict if core.MapKeys exists
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort.Strings(keys) // Optional sort
	return keys
}

// Note: Implementations for other Enter/Exit methods are expected to be in other ast_builder_*.go files.

// pkg/core/ast_builder_main.go
package core

import (
	"io"
	"log"
	"strconv" // Needed for Unquote

	// "strings" // Not needed directly here
	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// neuroScriptListenerImpl builds the AST.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	fileVersion    string // <-- ADDED: Store parsed file version
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step            // Pointer to the current list of steps being built
	blockStepStack []*[]Step          // Stack for managing nested block contexts (stores pointers to parent step lists)
	valueStack     []interface{}      // Stack holds expression AST nodes
	currentMapKey  *StringLiteralNode // Temp storage for map key node

	blockSteps map[antlr.ParserRuleContext][]Step // For IF/ELSE step collection

	logger   *log.Logger
	debugAST bool
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger *log.Logger, debugAST bool) *neuroScriptListenerImpl {
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Default to discarding logs if none provided
	}
	return &neuroScriptListenerImpl{
		fileVersion:    "", // <-- ADDED: Initialize
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10), // Initialize with some capacity
		logger:         logger,
		debugAST:       debugAST,
		blockSteps:     make(map[antlr.ParserRuleContext][]Step),
	}
}

// --- Getter for File Version ---
func (l *neuroScriptListenerImpl) GetFileVersion() string { // <-- ADDED
	return l.fileVersion
}

// --- Stack Helper Methods (pushValue, popValue, popNValues) remain the same ---
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("    Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
}
func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Println("[ERROR] AST Builder: Pop from empty value stack!")
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	l.logDebugAST("    Popped Value: %T %+v (Stack size: %d)", value, value, len(l.valueStack))
	return value, true
}
func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if len(l.valueStack) < n {
		l.logger.Printf("[ERROR] AST Builder: Stack underflow pop %d, have %d.", n, len(l.valueStack))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	l.logDebugAST("    Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true
}

// --- Core Listener Methods ---
func (l *neuroScriptListenerImpl) GetResult() []Procedure { return l.procedures }
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Printf(format, v...)
	}
}
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
	l.fileVersion = "" // Reset on new program
}
func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	l.logDebugAST("<<< Exit Program (File Version: %q)", l.fileVersion)
}

// --- ADDED: Handle File Version Declaration ---
func (l *neuroScriptListenerImpl) ExitFile_version_decl(ctx *gen.File_version_declContext) {
	if ctx.STRING_LIT() != nil {
		versionStr := ctx.STRING_LIT().GetText()
		unquotedVersion, err := strconv.Unquote(versionStr)
		if err != nil {
			l.logger.Printf("[WARN] Failed to unquote FILE_VERSION string literal: %q - %v", versionStr, err)
			// Optionally store the raw string or report error more formally
			l.fileVersion = versionStr // Store raw as fallback
		} else {
			l.fileVersion = unquotedVersion
			l.logDebugAST("    Captured FILE_VERSION: %q", l.fileVersion)
		}
	} else {
		l.logger.Printf("[WARN] FILE_VERSION keyword found but missing string literal value.")
	}
}

// Ensure other Enter/Exit methods for different rules are in their respective files
// (ast_builder_procedures.go, ast_builder_statements.go, etc.)

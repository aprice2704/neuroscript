// pkg/core/ast_builder_main.go
package core

import (
	// Added fmt import
	"strconv" // Needed for Unquote

	// "strings" // Not needed directly here
	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// neuroScriptListenerImpl builds the AST using the Listener pattern.
// It uses a stack-based approach for expressions and manages block scopes.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	fileVersion    string
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{}
	currentMapKey  *StringLiteralNode

	// blockSteps map is needed for IF statement's THEN/ELSE branch collection
	// because their ExitStatement_list happens before ExitIf_statement.
	blockSteps map[antlr.ParserRuleContext][]Step

	logger   interfaces.Logger
	debugAST bool
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	if logger == nil {
		panic("NeuroScript listener must have valid logger")
	}
	return &neuroScriptListenerImpl{
		fileVersion:    "",
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10),
		logger:         logger,
		debugAST:       debugAST,
		blockSteps:     make(map[antlr.ParserRuleContext][]Step),
	}
}

func (l *neuroScriptListenerImpl) GetFileVersion() string {
	return l.fileVersion
}

func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("    Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
}

func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
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
		l.logger.Error("AST Builder: Stack underflow pop %d, have %d.", n, len(l.valueStack))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	l.logDebugAST("    Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true
}

func (l *neuroScriptListenerImpl) GetResult() []Procedure { return l.procedures }

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(format, v...)
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

func (l *neuroScriptListenerImpl) ExitFile_version_decl(ctx *gen.File_version_declContext) {
	if ctx.STRING_LIT() != nil {
		versionStr := ctx.STRING_LIT().GetText()
		unquotedVersion, err := strconv.Unquote(versionStr)
		if err != nil {
			l.logger.Warn("Failed to unquote FILE_VERSION string literal: %q - %v", versionStr, err)
			l.fileVersion = versionStr // Store raw as fallback
		} else {
			l.fileVersion = unquotedVersion
			l.logDebugAST("    Captured FILE_VERSION: %q", l.fileVersion)
		}
	} else {
		l.logger.Warn("FILE_VERSION keyword found but missing string literal value.")
	}
}

// Note: This AST Builder uses the Listener pattern. Visitor methods like
// VisitStatement are not part of the Listener pattern. The logic is distributed
// across EnterXxx and ExitXxx methods triggered by the ANTLR walker.
// Statement processing happens within the Exit methods of specific statement types
// (e.g., ExitSet_statement, ExitCall_statement) defined in ast_builder_statements.go
// or block types (e.g., ExitIf_statement) in ast_builder_blocks.go.
// Therefore, no changes are needed in this file for VisitStatement.

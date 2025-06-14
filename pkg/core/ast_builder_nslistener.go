// NeuroScript Version: 0.3.0
// File version: 4
// Purpose: Re-instates the sentinel var to ensure full interface implementation.
// filename: pkg/core/ast_builder_nslistener.go
// nlines: 95
// risk_rating: LOW

package core

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

type neuroScriptListenerImpl struct {
	program              *Program
	fileMetadata         map[string]string
	procedures           []*Procedure
	events               []*OnEventDecl
	currentProc          *Procedure
	currentSteps         *[]Step
	blockStepStack       []*[]Step
	valueStack           []interface{}
	currentMapKey        *StringLiteralNode
	logger               interfaces.Logger
	debugAST             bool
	errors               []error
	loopDepth            int
	blockValueDepthStack []int
}

// EnterEveryRule implements core.NeuroScriptListener.
func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {
}

// ExitEveryRule implements core.NeuroScriptListener.
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext) {
}

// VisitErrorNode implements core.NeuroScriptListener.
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode) {
}

// VisitTerminal implements core.NeuroScriptListener.
func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode) {
}

// Sentinel variable to ensure neuroScriptListenerImpl implements the full interface.
var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		program:        &Program{Procedures: make(map[string]*Procedure), Events: make([]*OnEventDecl, 0)},
		fileMetadata:   make(map[string]string),
		procedures:     make([]*Procedure, 0),
		events:         make([]*OnEventDecl, 0),
		valueStack:     make([]interface{}, 0),
		blockStepStack: make([]*[]Step, 0),
		logger:         logger,
		debugAST:       debugAST,
	}
}

func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	return l.fileMetadata
}

func (l *neuroScriptListenerImpl) GetResult() []*Procedure {
	return l.procedures
}

func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	token := ctx.GetStart()
	l.addErrorf(token, format, args...)
}

func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pos := tokenToPosition(token)
	l.errors = append(l.errors, fmt.Errorf("AST build error near %s: %s", pos.String(), msg))
	l.logger.Error("AST Builder Error", "pos", pos.String(), "message", msg)
}

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

func (l *neuroScriptListenerImpl) isInsideLoop() bool {
	return l.loopDepth > 0
}

func (l *neuroScriptListenerImpl) EnterProgram(c *gen.ProgramContext) {
	l.logDebugAST("EnterProgram")
}

func (l *neuroScriptListenerImpl) ExitProgram(c *gen.ProgramContext) {
	l.logDebugAST("ExitProgram: Finalizing program...")
	if l.program != nil {
		l.program.Metadata = l.fileMetadata
		l.program.Events = l.events
	}

	if len(l.valueStack) > 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: value stack size is %d at end of program", len(l.valueStack))
		l.logger.Error("ExitProgram: value stack not empty", "size", len(l.valueStack))
	}
	if len(l.blockStepStack) != 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack))
		l.logger.Error("ExitProgram: blockStepStack not empty", "size", len(l.blockStepStack))
	}
}

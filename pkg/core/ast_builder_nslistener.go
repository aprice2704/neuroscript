// filename: pkg/core/ast_builder_nslistener.go
// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Bugfix: Corrected the type of the 'commands' slice to []*CommandNode.
// nlines: 194
// risk_rating: HIGH

package core

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program              *Program
	fileMetadata         map[string]string
	procedures           []*Procedure
	events               []*OnEventDecl
	commands             []*CommandNode // CORRECTED
	currentProc          *Procedure
	currentCommand       *CommandNode
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

// --- Block Management Helpers ---

// pushNewStepBlock saves the current step context and creates a new one.
func (l *neuroScriptListenerImpl) pushNewStepBlock() {
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	}
	newSteps := make([]Step, 0)
	l.currentSteps = &newSteps
}

// popCurrentStepBlock finalizes the current step block and restores the parent context.
func (l *neuroScriptListenerImpl) popCurrentStepBlock() []Step {
	completedSteps := *l.currentSteps
	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.currentSteps = nil
	} else {
		l.currentSteps = l.blockStepStack[stackSize-1]
		l.blockStepStack = l.blockStepStack[:stackSize-1]
	}
	return completedSteps
}

// --- Standard Listener Implementation ---

func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext)  {}
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode)        {}

func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode) {
	if node.GetSymbol().GetTokenType() == gen.NeuroScriptParserMETADATA_LINE && l.currentProc == nil {
		l.processMetadataLine(l.fileMetadata, node.GetSymbol())
	}
}

var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	initialSteps := make([]Step, 0)

	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 &Program{Procedures: make(map[string]*Procedure), Events: make([]*OnEventDecl, 0)},
		fileMetadata:            make(map[string]string),
		procedures:              make([]*Procedure, 0),
		events:                  make([]*OnEventDecl, 0),
		commands:                make([]*CommandNode, 0), // CORRECTED
		currentSteps:            &initialSteps,
		valueStack:              make([]interface{}, 0),
		blockStepStack:          make([]*[]Step, 0),
		logger:                  logger,
		debugAST:                debugAST,
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
	if len(l.valueStack) > 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: value stack size is %d at end of program", len(l.valueStack))
	}
	if len(l.blockStepStack) != 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack))
	}
}

func (l *neuroScriptListenerImpl) EnterLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("EnterLvalue_list")
}

func (l *neuroScriptListenerImpl) ExitLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("ExitLvalue_list: Collecting assignable targets.")
	numLvalues := len(ctx.AllLvalue())
	if numLvalues == 0 {
		l.pushValue([]Expression{})
		return
	}
	lvalues := make([]Expression, numLvalues)
	for i := numLvalues - 1; i >= 0; i-- {
		value, popok := l.popValue()
		if !popok {
			l.addError(ctx, "internal error: failed to pop value for lvalue")
			l.pushValue([]Expression{})
			return
		}
		expr, ok := value.(Expression)
		if !ok {
			l.addError(ctx, "internal error: value for lvalue is not an Expression, got %T", value)
			l.pushValue([]Expression{})
			return
		}
		lvalues[i] = expr
	}
	l.pushValue(lvalues)
}

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("ExitSet_statement: Building set step.")
	rhsVal, ok := l.popValue()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop RHS")
		return
	}
	rhsExpr, ok := rhsVal.(Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: RHS value is not an Expression, but %T", rhsVal)
		return
	}
	lhsVal, ok := l.popValue()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop LHS")
		return
	}
	lhsExprs, ok := lhsVal.([]Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: LHS value is not []Expression, but %T", lhsVal)
		return
	}
	step := Step{
		Pos:     tokenToPosition(ctx.GetStart()),
		Type:    "set",
		LValues: lhsExprs,
		Value:   rhsExpr,
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

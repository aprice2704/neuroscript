// NeuroScript Version: 0.3.0
// File version: 10
// Purpose: Corrected ANTLR method call in VisitTerminal from GetType to GetTokenType.
// filename: pkg/core/ast_builder_nslistener.go
// nlines: 174
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

func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext)  {}
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode)        {}

// VisitTerminal is called for every terminal node (token) in the parse tree.
func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode) {
	// MODIFIED: Changed GetType() to the correct ANTLR method, GetTokenType().
	if node.GetSymbol().GetTokenType() == gen.NeuroScriptParserMETADATA_LINE && l.currentProc == nil {
		l.processMetadataLine(l.fileMetadata, node.GetSymbol())
	}
}

// Sentinel variable to ensure neuroScriptListenerImpl implements the full interface.
var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 &Program{Procedures: make(map[string]*Procedure), Events: make([]*OnEventDecl, 0)},
		fileMetadata:            make(map[string]string),
		procedures:              make([]*Procedure, 0),
		events:                  make([]*OnEventDecl, 0),
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

func (l *neuroScriptListenerImpl) EnterLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("EnterLvalue_list")
}

func (l *neuroScriptListenerImpl) ExitLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("ExitLvalue_list: Collecting assignable targets.")
	numLvalues := len(ctx.AllLvalue())
	if numLvalues == 0 {
		l.addError(ctx, "lvalue_list is empty, which should not be possible based on grammar")
		l.pushValue([]Expression{})
		return
	}
	lvalues := make([]Expression, numLvalues)
	for i := numLvalues - 1; i >= 0; i-- {
		value, popok := l.popValue()
		if !popok {
			l.addError(ctx, "internal AST builder error: failed to pop value for lvalue from stack; stack is likely empty")
			l.pushValue([]Expression{})
			return
		}
		expr, ok := value.(Expression)
		if !ok {
			l.addError(ctx, "internal AST builder error: value for lvalue is not an Expression node, got %T", value)
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
		l.addError(ctx, "internal error in set_statement: could not pop RHS expression from stack")
		return
	}
	rhsExpr, ok := rhsVal.(Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: RHS value is not an Expression, but %T", rhsVal)
		return
	}

	lhsVal, ok := l.popValue()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop LHS expressions from stack")
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

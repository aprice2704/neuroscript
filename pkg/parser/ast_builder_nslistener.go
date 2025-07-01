// filename: pkg/core/ast_builder_nslistener.go
// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Bugfix: Corrected the type of the 'commands' slice to []*CommandNode.
// nlines: 194
// risk_rating: HIGH

package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program              *ast.Program
	fileMetadata         map[string]string
	procedures           []*ast.Procedure
	events               []*ast.OnEventDecl
	commands             []*ast.CommandNode // CORRECTED
	currentProc          *ast.Procedure
	currentCommand       *ast.CommandNode
	currentSteps         *[]ast.Step
	blockStepStack       []*[]ast.Step
	ValueStack           []interface{}
	currentMapKey        *ast.StringLiteralNode
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
	newSteps := make([]ast.Step, 0)
	l.currentSteps = &newSteps
}

// popCurrentStepBlock finalizes the current step block and restores the parent context.
func (l *neuroScriptListenerImpl) popCurrentStepBlock() []ast.Step {
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
	if node.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMETADATA_LINE && l.currentProc == nil {
		l.processMetadataLine(l.fileMetadata, node.GetSymbol())
	}
}

var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	initialSteps := make([]ast.Step, 0)

	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 &ast.Program{Procedures: make(map[string]*ast.Procedure), Events: make([]*ast.OnEventDecl, 0)},
		fileMetadata:            make(map[string]string),
		procedures:              make([]*ast.Procedure, 0),
		events:                  make([]*ast.OnEventDecl, 0),
		commands:                make([]*ast.CommandNode, 0), // CORRECTED
		currentSteps:            &initialSteps,
		ValueStack:              make([]interface{}, 0),
		blockStepStack:          make([]*[]ast.Step, 0),
		logger:                  logger,
		debugAST:                debugAST,
	}
}

func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	return l.fileMetadata
}

func (l *neuroScriptListenerImpl) GetResult() []*ast.Procedure {
	return l.procedures
}

func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	token := ctx.GetStart()
	l.addErrorf(token, format, args...)
}

func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pos := tokenTolang.Position(token)
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
	if len(l.ValueStack) > 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: value stack size is %d at end of program", len(l.ValueStack))
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
	numLValues := len(ctx.AllLvalue())
	if numLValues == 0 {
		l.push([]*ast.LValueNode{})
		return
	}
	lValues := make([]*ast.LValueNode, numLValues)
	for i := numLValues - 1; i >= 0; i-- {
		value, popok := l.poplang.Value()
		if !popok {
			l.addError(ctx, "internal error: failed to pop value for lvalue")
			l.push([]*ast.LValueNode{})
			return
		}
		expr, ok := value.(*ast.LValueNode)
		if !ok {
			l.addError(ctx, "internal error: value for lvalue is not an ast.LValueNode, got %T", value)
			l.push([]*ast.LValueNode{})
			return
		}
		lValues[i] = expr
	}
	l.push(lValues)
}

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("ExitSet_statement: Building set step.")
	rhsVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop RHS")
		return
	}
	rhsExpr, ok := rhsVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: RHS value is not an ast.Expression, but %T", rhsVal)
		return
	}
	lhsVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop LHS")
		return
	}
	lhsExprs, ok := lhsVal.([]*ast.LValueNode)
	if !ok {
		l.addError(ctx, "internal error in set_statement: LHS value is not []*ast.LValueNode, but %T", lhsVal)
		return
	}
	step := ast.Step{
		Position: tokenTolang.Position(ctx.GetStart()),
		Type:     "set",
		LValues:  lhsExprs,
		Values:   []ast.Expression{rhsExpr},
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

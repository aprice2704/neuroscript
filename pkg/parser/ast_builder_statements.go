// NeuroScript Version: 0.5.2
// File version: 11
// Purpose: Implemented the definitive fix for return-value order using a robust prepend-to-slice pattern.
// filename: pkg/parser/ast_builder_statements.go
// nlines: 150
// risk_rating: HIGH

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// addStep is a helper to append a new step to the correct part of the AST.
func (l *neuroScriptListenerImpl) addStep(step ast.Step) {
	if len(l.blockStack) > 0 {
		currentBlock := l.blockStack[len(l.blockStack)-1]
		currentBlock.steps = append(currentBlock.steps, step)
	} else {
		l.addError(nil, "internal AST error: addStep called with no active block context")
	}
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(c *gen.Emit_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in emit_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in emit_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "emit",
		Values:   []ast.Expression{expr},
	})
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(c *gen.Return_statementContext) {
	var returnValues []ast.Expression
	if c.Expression_list() != nil {
		numExpr := len(c.Expression_list().AllExpression())
		// The expressions are pushed onto the stack from left to right,
		// so we pop them off in reverse order. To restore the original
		// order, we prepend each popped item to the front of our slice.
		for i := 0; i < numExpr; i++ {
			val, ok := l.pop()
			if !ok {
				l.addError(c, "internal error in return_statement: could not pop value")
				return // Stop processing to avoid further errors
			}
			expr, ok := val.(ast.Expression)
			if !ok {
				l.addError(c, "internal error in return_statement: value is not an ast.Expression, but %T", val)
				return // Stop processing
			}
			// Prepending to the slice effectively reverses the stack order.
			returnValues = append([]ast.Expression{expr}, returnValues...)
		}
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "return",
		Values:   returnValues,
	})
}

func (l *neuroScriptListenerImpl) ExitCall_statement(c *gen.Call_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in call_statement: could not pop value")
		return
	}
	callExpr, ok := val.(*ast.CallableExprNode)
	if !ok {
		l.addError(c, "internal error in call_statement: value is not a *ast.CallableExprNode, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "call",
		Call:     callExpr,
	})
}

func (l *neuroScriptListenerImpl) ExitFail_statement(c *gen.Fail_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in fail_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in fail_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "fail",
		Values:   []ast.Expression{expr},
	})
}

func (l *neuroScriptListenerImpl) ExitAsk_stmt(c *gen.Ask_stmtContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in ask_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in ask_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position:   pos,
		Type:       "ask",
		Values:     []ast.Expression{expr},
		AskIntoVar: c.IDENTIFIER().GetText(),
	})
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(c *gen.ClearErrorStmtContext) {
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "clear_error",
	})
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(c *gen.Continue_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'continue' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "continue",
	})
}

func (l *neuroScriptListenerImpl) ExitBreak_statement(c *gen.Break_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'break' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position: pos,
		Type:     "break",
	})
}

func (l *neuroScriptListenerImpl) ExitExpression_statement(ctx *gen.Expression_statementContext) {
	rawExpr, ok := l.pop()
	if !ok {
		l.addError(ctx, "internal error in expression_statement: could not pop expression")
		return
	}

	expr, ok := rawExpr.(ast.Expression)
	if !ok {
		l.addError(ctx, "internal error in expression_statement: value is not an ast.Expression, but %T", rawExpr)
		return
	}

	pos := tokenToPosition(ctx.GetStart())

	stmtNode := &ast.ExpressionStatementNode{
		Pos:        &pos,
		Expression: expr,
	}

	l.addStep(ast.Step{
		Position:       pos,
		Type:           "expression_statement",
		ExpressionStmt: stmtNode,
	})
}

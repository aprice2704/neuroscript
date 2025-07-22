// filename: pkg/parser/ast_builder_statements.go
// NeuroScript Version: 0.6.0
// File version: 22
// Purpose: Removed obsolete blank line counting logic. Association is now handled by the LineInfo algorithm.
// nlines: 130
// risk_rating: LOW

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
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
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "emit",
		Values:   []ast.Expression{expr},
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(c *gen.Return_statementContext) {
	var returnValues []ast.Expression
	if c.Expression_list() != nil {
		numExpr := len(c.Expression_list().AllExpression())
		if numExpr > 0 {
			popped, ok := l.popN(numExpr)
			if !ok {
				l.addError(c, "internal error in return_statement: could not pop values")
				return
			}
			for _, val := range popped {
				var expr ast.Expression
				isExpr := false

				expr, isExpr = val.(ast.Expression)
				if !isExpr {
					if lval, isLval := val.(*ast.LValueNode); isLval && len(lval.Accessors) == 0 {
						expr = &ast.VariableNode{
							BaseNode: lval.BaseNode,
							Name:     lval.Identifier,
						}
						isExpr = true
					}
				}

				if isExpr {
					returnValues = append(returnValues, expr)
				} else {
					l.addError(c, "internal error in return_statement: value is not an ast.Expression, but %T", val)
					return
				}
			}
		}
	}
	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "return",
		Values:   returnValues,
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
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
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "call",
		Call:     callExpr,
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitMust_statement(c *gen.Must_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in must_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in must_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "must",
		Cond:     expr,
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitFail_statement(c *gen.Fail_statementContext) {
	var failValue ast.Expression
	if c.Expression() != nil {
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
		failValue = expr
	}

	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "fail",
	}
	if failValue != nil {
		step.Values = []ast.Expression{failValue}
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
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
	step := ast.Step{
		BaseNode:   ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:       "ask",
		Values:     []ast.Expression{expr},
		AskIntoVar: c.IDENTIFIER().GetText(),
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(c *gen.ClearErrorStmtContext) {
	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "clear_error",
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(c *gen.Continue_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'continue' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "continue",
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitBreak_statement(c *gen.Break_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'break' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "break",
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.GetStop())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("ExitSet_statement: Building set step.")
	rhsVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop RHS")
		return
	}
	rhsExpr, ok := rhsVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: RHS value is not an ast.Expression, but %T", rhsVal)
		return
	}
	lhsVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop LHS")
		return
	}
	lhsExprs, ok := lhsVal.([]*ast.LValueNode)
	if !ok {
		l.addError(ctx, "internal error in set_statement: LHS value is not []*ast.LValueNode, but %T", lhsVal)
		return
	}

	pos := tokenToPosition(ctx.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "set",
		LValues:  lhsExprs,
		Values:   []ast.Expression{rhsExpr},
	}
	SetEndPos(&step, ctx.GetStop())
	l.addStep(step)
}

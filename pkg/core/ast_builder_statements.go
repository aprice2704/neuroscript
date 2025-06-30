// filename: pkg/core/ast_builder_statements.go
// version: 9
// purpose: Removed duplicate on_stmt listener logic, which is now centralized in ast_builder_events.go.
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitEmit_statement handles the 'emit' statement, which can have an optional expression.
func (l *neuroScriptListenerImpl) ExitEmit_statement(c *gen.Emit_statementContext) {
	l.logDebugAST("<<< ExitEmit_statement")
	var values []Expression

	if c.Expression() != nil {
		rawExpr, ok := l.popValue()
		if !ok {
			l.addError(c, "stack underflow: could not pop expression for emit statement")
			return
		}

		expr, castOk := rawExpr.(Expression)
		if !castOk {
			l.addError(c, "internal error: value on stack for emit was not an Expression, but %T", rawExpr)
			l.pushValue(rawExpr)
			return
		}
		values = []Expression{expr}
	}

	stmt := Step{
		Pos:    tokenToPosition(c.GetStart()),
		Type:   "emit",
		Values: values,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitReturn_statement handles the 'return' statement, correctly packaging single or multiple return values.
func (l *neuroScriptListenerImpl) ExitReturn_statement(c *gen.Return_statementContext) {
	l.logDebugAST("<<< ExitReturn_statement")
	var values []Expression

	if exprListCtx := c.Expression_list(); exprListCtx != nil {
		numExprs := len(exprListCtx.AllExpression())
		if numExprs > 0 {
			rawExprs, ok := l.popNValues(numExprs)
			if !ok {
				l.addError(c, "stack underflow: could not pop %d expressions for return statement", numExprs)
				return
			}

			values = make([]Expression, numExprs)
			for i, rawExpr := range rawExprs {
				expr, castOk := rawExpr.(Expression)
				if !castOk {
					l.addError(c, "internal error: value on stack for return statement was not an Expression, but %T", rawExpr)
					l.pushValue(rawExpr)
					return
				}
				values[i] = expr
			}
		}
	}

	stmt := Step{
		Pos:    tokenToPosition(c.GetStart()),
		Type:   "return",
		Values: values,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitBreak_statement handles the 'break' statement.
func (l *neuroScriptListenerImpl) ExitBreak_statement(c *gen.Break_statementContext) {
	l.logDebugAST("<<< ExitBreak_statement")
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "break",
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitContinue_statement handles the 'continue' statement.
func (l *neuroScriptListenerImpl) ExitContinue_statement(c *gen.Continue_statementContext) {
	l.logDebugAST("<<< ExitContinue_statement")
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "continue",
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitFail_statement handles the 'fail' statement, which halts execution with an error.
func (l *neuroScriptListenerImpl) ExitFail_statement(c *gen.Fail_statementContext) {
	l.logDebugAST("<<< ExitFail_statement")
	var values []Expression
	if c.Expression() != nil {
		val, ok := l.pop().(Expression)
		if !ok {
			l.addError(c, "fail statement has invalid expression: %T", val)
		} else {
			values = []Expression{val}
		}
	}

	stmt := Step{
		Pos:    tokenToPosition(c.GetStart()),
		Type:   "fail",
		Values: values,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitMust_statement handles the 'must' and 'mustbe' statements, which assert a condition.
func (l *neuroScriptListenerImpl) ExitMust_statement(c *gen.Must_statementContext) {
	l.logDebugAST("<<< ExitMust_statement")
	val, ok := l.pop().(Expression)
	if !ok {
		l.addError(c, "must statement requires a valid expression")
		return
	}

	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "must",
		Cond: val,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitCall_statement handles a standalone 'call' statement for executing a procedure or tool for its side effects.
func (l *neuroScriptListenerImpl) ExitCall_statement(c *gen.Call_statementContext) {
	l.logDebugAST("<<< ExitCall_statement")
	val, ok := l.pop().(Expression)
	if !ok {
		l.addError(c, "call statement requires a valid expression")
		return
	}

	callExpr, ok := val.(*CallableExprNode)
	if !ok {
		l.addError(c, "call statement expression is not a callable expression node, but %T", val)
		return
	}

	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "call",
		Call: callExpr,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitClearErrorStmt handles the 'clear_error' statement within an on_error block.
func (l *neuroScriptListenerImpl) ExitClearErrorStmt(c *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< ExitClearErrorStmt")
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "clear_error",
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitAsk_stmt handles the 'ask' statement for interacting with an AI.
func (l *neuroScriptListenerImpl) ExitAsk_stmt(c *gen.Ask_stmtContext) {
	l.logDebugAST("<<< ExitAsk_stmt")
	var target string
	if c.IDENTIFIER() != nil {
		target = c.IDENTIFIER().GetText()
	}

	prompt, ok := l.pop().(Expression)
	if !ok {
		l.addError(c, "ask statement requires a valid prompt expression")
		return
	}

	stmt := Step{
		Pos:        tokenToPosition(c.GetStart()),
		Type:       "ask",
		PromptExpr: prompt,
		AskIntoVar: target,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

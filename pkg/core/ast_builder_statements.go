// filename: pkg/core/ast_builder_statements.go
// version: 3
// purpose: Implements listener methods for simple statements using the correct Step struct fields.
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitEmit_statement handles the 'emit' statement.
func (l *neuroScriptListenerImpl) ExitEmit_statement(c *gen.Emit_statementContext) {
	l.logDebugAST("<<< ExitEmit_statement")
	val, ok := l.pop().(Expression)
	if !ok {
		l.addError(c, "emit statement requires a valid expression")
		return
	}

	// For emit, the single expression is stored in the 'Value' field.
	stmt := Step{
		Pos:   tokenToPosition(c.GetStart()),
		Type:  "emit",
		Value: val,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitReturn_statement handles the 'return' statement.
func (l *neuroScriptListenerImpl) ExitReturn_statement(c *gen.Return_statementContext) {
	l.logDebugAST("<<< ExitReturn_statement")
	var values []Expression

	if c.Expression_list() != nil {
		val := l.pop()
		if exprs, ok := val.([]Expression); ok {
			values = exprs
		} else if expr, ok := val.(Expression); ok {
			values = []Expression{expr}
		} else {
			l.addError(c, "return statement has invalid expression list value on stack: %T", val)
			return
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

// ExitFail_statement handles the 'fail' statement.
func (l *neuroScriptListenerImpl) ExitFail_statement(c *gen.Fail_statementContext) {
	l.logDebugAST("<<< ExitFail_statement")
	var msg Expression
	if c.Expression() != nil {
		val, ok := l.pop().(Expression)
		if !ok {
			l.addError(c, "fail statement has invalid expression: %T", val)
		} else {
			msg = val
		}
	}
	// 'fail' uses the single 'Value' field for its optional message.
	stmt := Step{
		Pos:   tokenToPosition(c.GetStart()),
		Type:  "fail",
		Value: msg,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitMust_statement handles the 'must' and 'mustbe' statements.
func (l *neuroScriptListenerImpl) ExitMust_statement(c *gen.Must_statementContext) {
	l.logDebugAST("<<< ExitMust_statement")
	val, ok := l.pop().(Expression)
	if !ok {
		l.addError(c, "must statement requires a valid expression")
		return
	}
	// 'must' uses the 'Cond' field for its condition.
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "must",
		Cond: val,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitCall_statement handles the 'call' statement.
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

// ExitClearErrorStmt handles the 'clear_error' statement.
func (l *neuroScriptListenerImpl) ExitClearErrorStmt(c *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< ExitClearErrorStmt")
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "clear_error",
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// ExitAsk_stmt handles the 'ask' statement.
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

	// MINIMAL CHANGE: Using the correct field names from the Step struct definition.
	stmt := Step{
		Pos:        tokenToPosition(c.GetStart()),
		Type:       "ask",
		PromptExpr: prompt,
		AskIntoVar: target,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

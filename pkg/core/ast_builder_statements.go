// NeuroScript Version: 0.3.0 // Keep user's version marker
// Last Modified: 2025-05-01 12:37:21 PDT // Keep user's timestamp
package core

import (
	"fmt" // Import fmt

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// "log" // Use logger from listener impl instead
)

// --- Simple Statement Exit Handlers ---
// *** MODIFIED: Added type assertions for Expression, position setting, and error handling ***

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop value for SET statement")
		return
	}
	valueNode, ok := valueRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Value for SET statement is not an Expression (got %T)", valueRaw)
		return
	}
	varName := ctx.IDENTIFIER().GetText()
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Set_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "set", Target: varName, Value: valueNode, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended SET Step: Target=%s, Value=%T", varName, valueNode)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var returnNodes []Expression = nil
	if exprListCtx := ctx.Expression_list(); exprListCtx != nil {
		numExpr := len(exprListCtx.AllExpression())
		if numExpr > 0 {
			nodesPoppedRaw, ok := l.popNValues(numExpr)
			if !ok {
				l.addError(ctx, "Internal error: Failed to pop %d value(s) for RETURN statement", numExpr)
				return
			}
			returnNodes = make([]Expression, 0, numExpr)
			for i, nodeRaw := range nodesPoppedRaw {
				nodeExpr, ok := nodeRaw.(Expression)
				if !ok {
					pos := tokenToPosition(exprListCtx.Expression(i).GetStart())
					l.errors = append(l.errors, fmt.Errorf("AST build error at %s: RETURN argument %d is not an Expression (got %T)", pos.String(), i+1, nodeRaw))
					return
				}
				returnNodes = append(returnNodes, nodeExpr)
			}
			l.logDebugAST("         Popped and asserted %d return nodes", len(returnNodes))
		} else {
			l.logDebugAST("         RETURN statement has empty Expression_list (value will be empty list).")
			returnNodes = []Expression{}
		}
	} else {
		l.logDebugAST("         RETURN statement has no expression list (value will be nil)")
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Return_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "return", Value: returnNodes, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended RETURN Step")
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode Expression = nil
	if ctx.Expression() != nil {
		valueRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop value for EMIT statement")
			return
		}
		valueNode, ok = valueRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Value for EMIT statement is not an Expression (got %T)", valueRaw)
			return
		}
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Emit_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "emit", Value: valueNode, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended EMIT Step")
}

func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Enter ExitMust_statement: %q", ctx.GetText()) // Log entry
	var valueNode Expression
	var target string
	stepType := "must"
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop node for MUST/MUSTBE statement")
		return
	}
	if ctx.Callable_expr() != nil {
		stepType = "mustbe"
		callNode, ok := valueRaw.(*CallableExprNode)
		if !ok {
			l.addError(ctx, "Internal error: Expected CallableExprNode for MUSTBE statement, got %T", valueRaw)
			return
		}
		target = callNode.Target.Name
		if callNode.Target.IsTool {
			target = "tool." + target
		}
		valueNode = callNode
		l.logDebugAST("         Interpreting as MUSTBE, Target=%s", target)
	} else if ctx.Expression() != nil {
		stepType = "must"
		valueNode, ok = valueRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Condition for MUST statement is not an Expression (got %T)", valueRaw)
			return
		}
		target = ""
		l.logDebugAST("         Interpreting as MUST")
	} else {
		l.addError(ctx, "Internal error: Invalid structure for Must_statementContext")
		return
	}

	// +++ Add Debug Logging +++
	procName := "(nil)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	stackSize := len(l.blockStepStack)
	l.logger.Debug("ExitMust_statement Check", "proc", procName, "currentStepsIsNil", l.currentSteps == nil, "stackSize", stackSize)
	if l.currentSteps != nil {
		// Avoid panic if currentSteps is nil
		l.logger.Debug("ExitMust_statement Check", "currentStepsLen", len(*l.currentSteps))
	} else {
		l.logger.Debug("ExitMust_statement Check: l.currentSteps IS NIL!")
	}
	// +++ End Debug Logging +++

	if l.currentSteps == nil { // <<< ERROR CHECK HERE
		l.addError(ctx, "Internal error: Must_statement exited with nil currentSteps")
		return
	}

	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: stepType, Target: target, Value: valueNode, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended %s Step: Value=%T", stepType, valueNode)
}

func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", ctx.GetText())
	var valueNode Expression = nil
	if ctx.Expression() != nil {
		valueRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop value for FAIL statement")
			return
		}
		valueNode, ok = valueRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Value for FAIL statement is not an Expression (got %T)", valueRaw)
			return
		}
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Fail_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "fail", Value: valueNode, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended FAIL Step")
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: ClearErrorStmt exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "clear_error", Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CLEAR_ERROR Step")
}

func (l *neuroScriptListenerImpl) ExitAsk_stmt(ctx *gen.Ask_stmtContext) {
	l.logDebugAST("<<< Exit Ask_stmt: %q", ctx.GetText())
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop prompt expression for ASK statement")
		return
	}
	promptExpr, ok := valueRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Prompt for ASK statement is not an Expression (got %T)", valueRaw)
		return
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Ask_stmt exited with nil currentSteps")
		return
	}
	targetVar := ""
	if ctx.IDENTIFIER() != nil {
		targetVar = ctx.IDENTIFIER().GetText()
		l.logDebugAST("         Ask target variable: %s", targetVar)
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "ask", Target: targetVar, Value: promptExpr, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended ASK Step: Target=%s, Prompt=%T", targetVar, promptExpr)
}

// *** ADDED: Handler for Call statement stack ***
func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement: %q", ctx.GetText())
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop value for CALL statement")
		return
	}
	if _, okCast := valueRaw.(*CallableExprNode); !okCast {
		l.addError(ctx, "Internal error: Value popped for CALL statement was not a CallableExprNode (got %T)", valueRaw)
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Call_statement exited with nil currentSteps (problematic if errors occurred)")
	}
	l.logDebugAST("         Consumed CallableExprNode from value stack for standalone CALL statement.")
}

// --- Break/Continue ---
func (l *neuroScriptListenerImpl) ExitBreak_statement(ctx *gen.Break_statementContext) {
	l.logDebugAST("<<< Exit Break_statement: %q", ctx.GetText())
	if !l.isInsideLoop() {
		l.addError(ctx, "'break' statement is not allowed outside of a loop ('while' or 'for each')")
		return
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Break_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "break", Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended BREAK Step")
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(ctx *gen.Continue_statementContext) {
	l.logDebugAST("<<< Exit Continue_statement: %q", ctx.GetText())
	if !l.isInsideLoop() {
		l.addError(ctx, "'continue' statement is not allowed outside of a loop ('while' or 'for each')")
		return
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Continue_statement exited with nil currentSteps")
		return
	}
	step := Step{Pos: tokenToPosition(ctx.GetStart()), Type: "continue", Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CONTINUE Step")
}

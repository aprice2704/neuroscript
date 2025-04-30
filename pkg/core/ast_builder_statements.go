// pkg/core/ast_builder_statements.go
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Simple Statement Exit Handlers ---

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())
	valueNode, ok := l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop value for SET")
		return
	}
	if l.currentSteps == nil {
		l.logger.Warn("Set_statement exited with nil currentSteps")
		return
	}
	varName := ctx.IDENTIFIER().GetText()
	// MODIFIED: Removed nil Args from newStep call
	step := newStep("set", varName, nil, valueNode, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended SET Step: Target=%s, Value=%T", varName, valueNode)
}

// --- REMOVED ExitCall_statement ---
// func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) { ... }

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var returnNodes []interface{} = nil // Default nil if no expression list

	if exprListCtx := ctx.Expression_list(); exprListCtx != nil {
		numExpr := len(exprListCtx.AllExpression())
		if numExpr > 0 { // Only pop if there are expressions
			nodesPopped, ok := l.popNValues(numExpr)
			if !ok {
				l.logger.Error("AST Builder: Failed to pop %d value(s) for RETURN", numExpr)
				returnNodes = nil
			} else {
				returnNodes = nodesPopped
				l.logDebugAST("    Popped %d return nodes", len(returnNodes))
			}
		} else {
			// This case might happen if grammar allows `return ()` vs just `return`
			l.logDebugAST("    RETURN statement has Expression_list context but no expressions (value will be empty list).")
			returnNodes = []interface{}{} // Return empty slice if expr list exists but is empty
		}
	} else {
		l.logDebugAST("    RETURN statement has no expression list (value will be nil)")
	}

	if l.currentSteps == nil {
		l.logger.Warn("Return_statement exited with nil currentSteps")
		return
	}

	// MODIFIED: Removed nil Args from newStep call
	// Store the slice of return nodes (or nil) in the Value field
	step := newStep("return", "", nil, returnNodes, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode interface{} = nil
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Failed to pop value for EMIT")
			// Maybe push nil or some error marker? For now, valueNode remains nil.
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Emit_statement exited with nil currentSteps")
		return
	}
	// MODIFIED: Removed nil Args from newStep call
	step := newStep("emit", "", nil, valueNode, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", ctx.GetText())
	var valueNode interface{}
	var ok bool
	stepType := "must" // Default type
	target := ""       // Target only relevant for mustbe

	valueNode, ok = l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop value/node for MUST/MUSTBE")
		return
	}

	// MODIFIED: Check the type of the node popped from the stack
	if callNode, isCall := valueNode.(CallableExprNode); isCall {
		// If the popped node IS a CallableExprNode, it must be a 'mustbe' statement
		// Note: This assumes the grammar correctly ensures only callable_expr follows KW_MUSTBE
		stepType = "mustbe"
		// Extract target information from the CallableExprNode
		target = callNode.Target.Name
		if callNode.Target.IsTool {
			target = "tool." + target
		}
		l.logDebugAST("    Interpreting as MUSTBE, Target=%s", target)
	} else {
		// If it's not a CallableExprNode, it's a regular 'must' statement
		l.logDebugAST("    Interpreting as MUST")
	}

	if l.currentSteps == nil {
		l.logger.Warn("Must_statement exited with nil currentSteps")
		return
	}

	// MODIFIED: Removed nil Args from newStep call
	step := newStep(stepType, target, nil, valueNode, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended %s Step: Value=%T", stepType, valueNode)
}

func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", ctx.GetText())
	var valueNode interface{} = nil
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Failed to pop value for FAIL")
			// valueNode remains nil
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Fail_statement exited with nil currentSteps")
		return
	}
	// MODIFIED: Removed nil Args from newStep call
	step := newStep("fail", "", nil, valueNode, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Warn("ClearErrorStmt exited with nil currentSteps")
		return
	}
	// MODIFIED: Removed nil Args from newStep call (though it didn't use args before)
	// Step struct directly used before, let's use newStep for consistency
	step := newStep("clear_error", "", nil, nil, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended CLEAR_ERROR Step")
}

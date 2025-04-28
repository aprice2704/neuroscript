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
	step := newStep("set", varName, nil, valueNode, nil, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement: %q", ctx.GetText())
	numArgs := 0
	if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
		if exprList := exprListOpt.Expression_list(); exprList != nil {
			numArgs = len(exprList.AllExpression())
		}
	}

	argNodes, ok := l.popNValues(numArgs)
	if !ok {
		if numArgs > 0 {
			l.logger.Error("AST Builder: Failed to pop %d args for CALL", numArgs)
			return
		}
		argNodes = []interface{}{}
	}

	if l.currentSteps == nil {
		l.logger.Warn("Call_statement exited with nil currentSteps")
		return
	}
	target := ctx.Call_target().GetText()
	step := newStep("call", target, nil, nil, nil, argNodes)
	*l.currentSteps = append(*l.currentSteps, step)
}

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
			l.logger.Warn("RETURN statement has Expression_list context but no expressions found.")
			returnNodes = []interface{}{}
		}
	} else {
		l.logDebugAST("    RETURN statement has no expression list (value will be nil)")
	}

	if l.currentSteps == nil {
		l.logger.Warn("Return_statement exited with nil currentSteps")
		return
	}

	step := newStep("return", "", nil, returnNodes, nil, nil)
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
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Emit_statement exited with nil currentSteps")
		return
	}
	step := newStep("emit", "", nil, valueNode, nil, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", ctx.GetText())
	var valueNode interface{}
	var ok bool
	stepType := "must"

	valueNode, ok = l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop value/node for MUST/MUSTBE")
		return
	}

	target := "" // Target only relevant for mustbe
	// Check if the context has a Function_call node specifically
	if mustBeCtx := ctx.Function_call(); mustBeCtx != nil {
		// If mustBeCtx exists, we know it must be a mustBe statement
		stepType = "mustbe"
		// The value popped should correspond to the FunctionCallNode built from mustBeCtx
		if fnCall, fnOk := valueNode.(FunctionCallNode); fnOk {
			target = fnCall.FunctionName // Retrieve name from the parsed node
			l.logDebugAST("    Interpreting as MUSTBE, Target=%s", target)
		} else {
			l.logger.Error("MUSTBE statement expected FunctionCallNode on stack, got %T", valueNode)
			// Handle error or proceed with incorrect interpretation? Let's log and proceed.
		}
	} else {
		// If no Function_call context, it's a regular must statement
		l.logDebugAST("    Interpreting as MUST")
	}

	if l.currentSteps == nil {
		l.logger.Warn("Must_statement exited with nil currentSteps")
		return
	}

	step := newStep(stepType, target, nil, valueNode, nil, nil)
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
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Fail_statement exited with nil currentSteps")
		return
	}
	step := newStep("fail", "", nil, valueNode, nil, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

// --- ADDED: Exit handler for clear_error statement ---
func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Warn("ClearErrorStmt exited with nil currentSteps")
		return
	}
	// Create a new Step type for clear_error
	step := Step{Type: "clear_error"} // Uses the simpler Step struct directly
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended CLEAR_ERROR Step")
}

// --- REMOVED: Try/Catch/Finally related Exit methods ---
// func (l *neuroScriptListenerImpl) VisitTryStmt(...) ...
// func (l *neuroScriptListenerImpl) VisitCatchClause(...) ...
// func (l *neuroScriptListenerImpl) VisitFinallyClause(...) ...
// Note: Since this builder uses the Listener pattern, these visitor methods
// wouldn't exist. The relevant logic was likely within Enter/Exit Try_statement
// in ast_builder_blocks.go, which will be removed there.

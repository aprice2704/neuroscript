// NeuroScript Version: 0.5.2
// File version: 0.0.3 // Correct Step creation in ExitMust_statement and ExitAsk_stmt.
// Last Modified: 2025-06-02
// filename: pkg/core/ast_builder_statements.go
package core

import (
	"fmt"
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Ensure antlr is imported if tokenToPosition or other antlr types are used directly here
	// "github.com/antlr4-go/antlr/v4"
)

// Helper function (if not already in ast_builder_main.go or a shared util file)
// For standalone use in this file if needed, or ensure it's accessible from where it's defined.
/*
func getRuleText(ctx antlr.RuleContext) string {
	if parser, ok := ctx.GetParser().(antlr.Parser); ok {
		return parser.GetTokenStream().GetTextFromRuleContext(ctx)
	}
	return ctx.GetText()
}
*/

// --- Simple Statement Exit Handlers ---

// ExitSet_statement was already updated in a previous step based on my snippets.
// This version should reflect the use of l.popValue() for LValueNode and RHS Expression.
func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("ExitSet_statement: %s", getRuleText(ctx))

	rhsValueIntf, okRhs := l.popValue()
	if !okRhs {
		l.addErrorf(ctx.GetStart(), "AST Builder: Value stack error when expecting RHS expression for set statement.")
		return
	}
	rhsExpr, castOkRhs := rhsValueIntf.(Expression)
	if !castOkRhs {
		l.addErrorf(ctx.Expression().GetStart(), "AST Builder: Expected Expression for set statement value, got %T", rhsValueIntf)
		l.pushValue(rhsValueIntf)
		return
	}

	lvalueIntf, okLval := l.popValue()
	if !okLval {
		l.addErrorf(ctx.Lvalue().GetStart(), "AST Builder: Value stack error when expecting LValueNode for set statement.")
		l.pushValue(rhsExpr)
		return
	}
	lvalNode, castOkLval := lvalueIntf.(*LValueNode)
	if !castOkLval {
		l.addErrorf(ctx.Lvalue().GetStart(), "AST Builder: Expected *LValueNode on stack for set statement, got %T", lvalueIntf)
		l.pushValue(lvalueIntf)
		l.pushValue(rhsExpr)
		return
	}

	setStep := Step{
		Pos:    tokenToPosition(ctx.KW_SET().GetSymbol()),
		Type:   "set",
		LValue: lvalNode,
		Value:  rhsExpr,
	}

	if l.currentSteps == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: currentSteps is nil, cannot add set statement. Block context issue.")
		var recoverySteps []Step
		l.currentSteps = &recoverySteps
		l.logger.Error("AST Builder: currentSteps was nil in ExitSet_statement. This is a critical issue.")
	}
	*l.currentSteps = append(*l.currentSteps, setStep)
	if lvalNode != nil && rhsExpr != nil { // Added nil checks for logging
		l.logDebugAST("         Appended SET Step: LValue=%s, Value=%s", lvalNode.String(), rhsExpr.String())
	} else {
		l.logDebugAST("         Appended SET Step with nil LValue or Value expression")
	}
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", getRuleText(ctx))
	var returnExprs []Expression

	if exprListCtx := ctx.Expression_list(); exprListCtx != nil {
		numExpr := len(exprListCtx.AllExpression())
		if numExpr > 0 {
			nodesPoppedRaw, ok := l.popNValues(numExpr)
			if !ok {
				l.addError(ctx, "Internal error: Failed to pop %d value(s) for RETURN statement", numExpr)
				return
			}
			returnExprs = make([]Expression, numExpr)
			for i := 0; i < numExpr; i++ {
				idx := i
				nodeExpr, isExpr := nodesPoppedRaw[idx].(Expression)
				if !isExpr {
					actualArgCtx := exprListCtx.Expression(i)
					pos := tokenToPosition(actualArgCtx.GetStart())
					l.addError(actualArgCtx, "RETURN argument %d is not an Expression (got %T)", i+1, nodesPoppedRaw[idx])
					returnExprs[i] = &ErrorNode{Pos: pos, Message: fmt.Sprintf("Return arg %d invalid type %T", i+1, nodesPoppedRaw[idx])}
				} else {
					returnExprs[i] = nodeExpr
				}
			}
			l.logDebugAST("         Popped and asserted %d return nodes", len(returnExprs))
		} else {
			l.logDebugAST("         RETURN statement has empty Expression_list (value will be empty list).")
			returnExprs = []Expression{}
		}
	} else {
		l.logDebugAST("         RETURN statement has no expression list (value will be nil slice of expressions).")
		returnExprs = nil
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Return_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   "return",
		Values: returnExprs,
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended RETURN Step")
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", getRuleText(ctx))
	var valueNode Expression = nil

	if ctx.Expression() != nil {
		valueRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop value for EMIT statement")
			return
		}
		var castOk bool
		valueNode, castOk = valueRaw.(Expression)
		if !castOk {
			l.addError(ctx, "Internal error: Value for EMIT statement is not an Expression (got %T)", valueRaw)
			l.pushValue(valueRaw) // Push back if wrong type
			return
		}
	} else {
		l.addError(ctx, "EMIT statement requires an expression.")
		// Create an error node, valueNode will remain nil. Interpreter should handle nil Value for emit.
		valueNode = &ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "EMIT statement missing expression"}
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Emit_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:   tokenToPosition(ctx.GetStart()),
		Type:  "emit",
		Value: valueNode,
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended EMIT Step")
}

// MODIFIED ExitMust_statement
func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", getRuleText(ctx))
	var valueForStep Expression
	var callForStep *CallableExprNode
	stepType := "must"

	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop node for MUST/MUSTBE statement")
		return
	}

	if ctx.Callable_expr() != nil { // This is 'mustbe callable_expr'
		stepType = "mustbe"
		callNode, isCallable := valueRaw.(*CallableExprNode)
		if !isCallable {
			l.addError(ctx, "Internal error: Expected CallableExprNode for MUSTBE, got %T", valueRaw)
			l.pushValue(valueRaw) // Push back if wrong type
			return
		}
		callForStep = callNode
		// For 'mustbe', the condition IS the callable expression itself.
		// The interpreter will evaluate this CallableExprNode.
		valueForStep = callNode
		l.logDebugAST("         Interpreting as MUSTBE, Call=%s", callNode.String())
	} else if ctx.Expression() != nil { // This is 'must expression'
		stepType = "must"
		exprNode, isExpr := valueRaw.(Expression)
		if !isExpr {
			l.addError(ctx.Expression(), "Internal error: Condition for MUST is not an Expression (got %T)", valueRaw)
			l.pushValue(valueRaw) // Push back if wrong type
			return
		}
		valueForStep = exprNode
		l.logDebugAST("         Interpreting as MUST, Value=%s", exprNode.String())
	} else {
		l.addError(ctx, "Internal error: Invalid structure for Must_statementContext")
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Must_statement exited with nil currentSteps")
		return
	}

	step := Step{
		Pos:   tokenToPosition(ctx.GetStart()),
		Type:  stepType,
		Value: valueForStep, // For 'must', this is the condition. For 'mustbe', this is the CallableExprNode.
		Call:  callForStep,  // Specifically for 'mustbe' to hold the CallableExprNode
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended %s Step", strings.ToUpper(stepType))
}

func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", getRuleText(ctx))
	var valueNode Expression = nil

	if ctx.Expression() != nil {
		valueRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop value for FAIL statement")
			return
		}
		var castOk bool
		valueNode, castOk = valueRaw.(Expression)
		if !castOk {
			l.addError(ctx, "Internal error: Value for FAIL statement is not an Expression (got %T)", valueRaw)
			l.pushValue(valueRaw)
			return
		}
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Fail_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:   tokenToPosition(ctx.GetStart()),
		Type:  "fail",
		Value: valueNode,
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended FAIL Step")
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", getRuleText(ctx))
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: ClearErrorStmt exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "clear_error",
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CLEAR_ERROR Step")
}

// MODIFIED ExitAsk_stmt
func (l *neuroScriptListenerImpl) ExitAsk_stmt(ctx *gen.Ask_stmtContext) {
	l.logDebugAST("<<< Exit Ask_stmt: %q", getRuleText(ctx))
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop prompt expression for ASK statement")
		return
	}
	promptExpr, isExpr := valueRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Internal error: Prompt for ASK statement is not an Expression (got %T)", valueRaw)
		l.pushValue(valueRaw) // Push back if wrong type
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Ask_stmt exited with nil currentSteps")
		return
	}
	targetVarName := ""
	if identNode := ctx.IDENTIFIER(); identNode != nil {
		targetVarName = identNode.GetText()
		l.logDebugAST("         Ask target variable: %s", targetVarName)
	}

	step := Step{
		Pos:        tokenToPosition(ctx.GetStart()),
		Type:       "ask",
		Value:      promptExpr,    // Stores the prompt Expression
		AskIntoVar: targetVarName, // Use the specific field from ast.go's Step struct
	}
	*l.currentSteps = append(*l.currentSteps, step)
	if promptExpr != nil { // Added nil check for logging
		l.logDebugAST("         Appended ASK Step: AskIntoVar=%s, Prompt=%s", targetVarName, promptExpr.String())
	} else {
		l.logDebugAST("         Appended ASK Step: AskIntoVar=%s, Prompt=<nil>", targetVarName)
	}
}

func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement: %q", getRuleText(ctx))
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop CallableExprNode for CALL statement")
		return
	}
	callableNode, isCallable := valueRaw.(*CallableExprNode)
	if !isCallable {
		l.addError(ctx, "Internal error: Value popped for CALL statement was not *CallableExprNode (got %T)", valueRaw)
		l.pushValue(valueRaw) // Push back if wrong type
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Call_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "call",
		Call: callableNode,
	}
	*l.currentSteps = append(*l.currentSteps, step)
	if callableNode != nil && callableNode.Target.Name != "" { // Added nil check for logging
		l.logDebugAST("         Appended CALL Step: Target=%s", callableNode.Target.String())
	} else {
		l.logDebugAST("         Appended CALL Step with nil or unnamed callable")
	}
}

// --- Break/Continue ---
func (l *neuroScriptListenerImpl) ExitBreak_statement(ctx *gen.Break_statementContext) {
	l.logDebugAST("<<< Exit Break_statement: %q", getRuleText(ctx))
	if !l.isInsideLoop() {
		l.addError(ctx, "'break' statement is not allowed outside of a loop ('while' or 'for each')")
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Break_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "break",
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended BREAK Step")
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(ctx *gen.Continue_statementContext) {
	l.logDebugAST("<<< Exit Continue_statement: %q", getRuleText(ctx))
	if !l.isInsideLoop() {
		l.addError(ctx, "'continue' statement is not allowed outside of a loop ('while' or 'for each')")
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Continue_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "continue",
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CONTINUE Step")
}

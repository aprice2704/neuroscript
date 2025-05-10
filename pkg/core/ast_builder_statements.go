// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Align Step creation with revised ast.go
// Last Modified: 2025-05-09 // Updated to reflect new Step struct
package core

import (
	"fmt"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Ensure antlr is imported for tokenToPosition usage with ctx.GetStart()
)

// --- Simple Statement Exit Handlers ---

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop value for SET statement")
		// Ensure a step is added, even if it's an error placeholder or this func returns.
		// For now, addError is sufficient, and a step might not be added if it's a fatal stack issue.
		return
	}
	valueNode, isExpr := valueRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Internal error: Value for SET statement is not an Expression (got %T)", valueRaw)
		// Attempt to push an ErrorNode onto the stack if this statement was part of a larger expression,
		// though for a statement, just logging the error is primary.
		return
	}
	varName := ctx.IDENTIFIER().GetText()
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Set_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   "set",
		Target: varName,
		Value:  valueNode, // Value is Expression, this is correct
		// Metadata: make(map[string]string), // Initialize if steps have metadata
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended SET Step: Target=%s, Value=%T", varName, valueNode)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var returnExprs []Expression // Use specific field Values []Expression

	if exprListCtx := ctx.Expression_list(); exprListCtx != nil {
		numExpr := len(exprListCtx.AllExpression())
		if numExpr > 0 {
			nodesPoppedRaw, ok := l.popNValues(numExpr)
			if !ok {
				l.addError(ctx, "Internal error: Failed to pop %d value(s) for RETURN statement", numExpr)
				return
			}
			returnExprs = make([]Expression, numExpr) // Correctly initialize
			for i := 0; i < numExpr; i++ {
				// popNValues returns in stack order (last pushed = first element). Reverse for parsed order.
				nodeExpr, isExpr := nodesPoppedRaw[numExpr-1-i].(Expression)
				if !isExpr {
					actualArgCtx := exprListCtx.Expression(i)
					pos := tokenToPosition(actualArgCtx.GetStart())
					l.addError(actualArgCtx, "RETURN argument %d is not an Expression (got %T)", i+1, nodesPoppedRaw[numExpr-1-i])
					// Create an ErrorNode to put in the list if desired, or just error out
					returnExprs[i] = &ErrorNode{Pos: pos, Message: fmt.Sprintf("Return arg %d invalid type %T", i+1, nodesPoppedRaw[numExpr-1-i])}
					// Potentially return here or continue with error nodes in the list
				} else {
					returnExprs[i] = nodeExpr
				}
			}
			l.logDebugAST("         Popped and asserted %d return nodes", len(returnExprs))
		} else {
			l.logDebugAST("         RETURN statement has empty Expression_list (value will be empty list).")
			returnExprs = []Expression{} // Explicitly empty, not nil
		}
	} else {
		l.logDebugAST("         RETURN statement has no expression list (value will be nil slice of expressions).")
		returnExprs = nil // Or []Expression{} depending on desired interpreter handling of `return;`
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Return_statement exited with nil currentSteps")
		return
	}
	// Use Step.Values for return expressions
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   "return",
		Values: returnExprs,
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended RETURN Step")
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode Expression = nil // Value is Expression, this is correct

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
	} else { // emit without expression
		l.addError(ctx, "EMIT statement requires an expression.")
		// Create an error node or handle as appropriate. For now, valueNode remains nil.
		// Potentially push an ErrorNode onto stack if EMIT was an expression itself (it's not).
		// The step will be created with nil Value, interpreter should handle.
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Emit_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:   tokenToPosition(ctx.GetStart()),
		Type:  "emit",
		Value: valueNode, // Correctly uses Value Expression
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended EMIT Step")
}

func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", ctx.GetText())
	var valueExpr Expression // For 'must condition' or 'mustbe callable_expr'
	var targetName string    // For 'mustbe callableName'
	stepType := "must"

	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop node for MUST/MUSTBE statement")
		return
	}

	if callableCtx := ctx.Callable_expr(); callableCtx != nil { // mustbe callable_expr
		stepType = "mustbe"
		callNode, isCallable := valueRaw.(*CallableExprNode)
		if !isCallable {
			l.addError(ctx, "Internal error: Expected CallableExprNode for MUSTBE, got %T", valueRaw)
			return
		}
		// For MUSTBE, the 'Value' field of Step will hold the CallableExprNode.
		// The 'Target' field can hold the string name for easier interpreter access if needed,
		// but the full callable is in Value.
		valueExpr = callNode
		targetName = callNode.Target.Name // Store the base name
		if callNode.Target.IsTool {
			targetName = "tool." + callNode.Target.Name // Prepend "tool." if it's a tool
		}
		l.logDebugAST("         Interpreting as MUSTBE, TargetName=%s, Value=%T", targetName, valueExpr)
	} else if exprCtx := ctx.Expression(); exprCtx != nil { // must expression
		stepType = "must"
		exprNode, isExpr := valueRaw.(Expression)
		if !isExpr {
			l.addError(exprCtx, "Internal error: Condition for MUST is not an Expression (got %T)", valueRaw)
			return
		}
		valueExpr = exprNode
		// Target is not used for 'must expression'
		l.logDebugAST("         Interpreting as MUST, Value=%T", valueExpr)
	} else {
		l.addError(ctx, "Internal error: Invalid structure for Must_statementContext")
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Must_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   stepType,
		Target: targetName, // Used for mustbe target name
		Value:  valueExpr,  // Condition for must, or CallableExprNode for mustbe
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended %s Step", stepType)
}

func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", ctx.GetText())
	var valueNode Expression = nil // Value is Expression, this is correct

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
	} // If no expression, valueNode remains nil, which is fine for `fail;`

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Fail_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:   tokenToPosition(ctx.GetStart()),
		Type:  "fail",
		Value: valueNode, // Correctly uses Value Expression
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended FAIL Step")
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: ClearErrorStmt exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "clear_error",
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CLEAR_ERROR Step")
}

func (l *neuroScriptListenerImpl) ExitAsk_stmt(ctx *gen.Ask_stmtContext) {
	l.logDebugAST("<<< Exit Ask_stmt: %q", ctx.GetText())
	valueRaw, ok := l.popValue() // This is the prompt expression
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop prompt expression for ASK statement")
		return
	}
	promptExpr, isExpr := valueRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Internal error: Prompt for ASK statement is not an Expression (got %T)", valueRaw)
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Ask_stmt exited with nil currentSteps")
		return
	}
	targetVar := ""
	if ctx.IDENTIFIER() != nil { // This is the 'into targetVar' part
		targetVar = ctx.IDENTIFIER().GetText()
		l.logDebugAST("         Ask target variable: %s", targetVar)
	}
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   "ask",
		Target: targetVar,  // Stores the 'into' variable name
		Value:  promptExpr, // Stores the prompt Expression
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended ASK Step: Target=%s, Prompt=%T", targetVar, promptExpr)
}

func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement: %q", ctx.GetText())
	// The CallableExprNode was pushed onto the stack by ExitCallable_expr
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop CallableExprNode for CALL statement")
		return
	}
	callableNode, isCallable := valueRaw.(*CallableExprNode)
	if !isCallable {
		l.addError(ctx, "Internal error: Value popped for CALL statement was not *CallableExprNode (got %T)", valueRaw)
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Call_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()), // Position of 'call' keyword
		Type: "call",
		Call: callableNode, // Use the specific 'Call' field
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CALL Step: Target=%s", callableNode.Target.Name)
}

// --- Break/Continue ---
func (l *neuroScriptListenerImpl) ExitBreak_statement(ctx *gen.Break_statementContext) {
	l.logDebugAST("<<< Exit Break_statement: %q", ctx.GetText())
	if !l.isInsideLoop() {
		l.addError(ctx, "'break' statement is not allowed outside of a loop ('while' or 'for each')")
		// Still create a step but interpreter might flag it or it might be benign if error collected.
		// Or, simply return if errors should halt AST step addition for this path.
		// For now, proceed to create the step; interpreter can validate loop context.
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Break_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "break",
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended BREAK Step")
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(ctx *gen.Continue_statementContext) {
	l.logDebugAST("<<< Exit Continue_statement: %q", ctx.GetText())
	if !l.isInsideLoop() {
		l.addError(ctx, "'continue' statement is not allowed outside of a loop ('while' or 'for each')")
		// See comment in ExitBreak_statement
	}
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Continue_statement exited with nil currentSteps")
		return
	}
	step := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "continue",
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("         Appended CONTINUE Step")
}

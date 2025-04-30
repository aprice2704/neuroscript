// filename: pkg/core/ast_builder_statements.go
package core

import (
	"fmt" // Import fmt

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Simple Statement Exit Handlers ---
// *** MODIFIED: Added type assertions for Expression, position setting, and error handling ***

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())

	// 1. Pop the value expression from the stack
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop value for SET statement")
		return // Stop processing this statement
	}

	// 2. Assert the popped value is an Expression
	valueNode, ok := valueRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Value for SET statement is not an Expression (got %T)", valueRaw)
		return // Stop processing this statement
	}

	// 3. Get the variable name
	varName := ctx.IDENTIFIER().GetText()

	// 4. Ensure we have a place to put the step
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Set_statement exited with nil currentSteps")
		return
	}

	// 5. Create and append the step
	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of the 'set' keyword
		Type:      "set",
		Target:    varName,
		Cond:      nil,                     // Not used for set
		Value:     valueNode,               // Assign the asserted Expression
		ElseValue: nil,                     // Not used for set
		Metadata:  make(map[string]string), // Initialize metadata map
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended SET Step: Target=%s, Value=%T", varName, valueNode)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var returnNodes []Expression // Store results as []Expression

	if exprListCtx := ctx.Expression_list(); exprListCtx != nil {
		numExpr := len(exprListCtx.AllExpression())
		if numExpr > 0 {
			nodesPoppedRaw, ok := l.popNValues(numExpr)
			if !ok {
				l.addError(ctx, "Internal error: Failed to pop %d value(s) for RETURN statement", numExpr)
				return // Stop processing
			}

			// Assert each popped value is an Expression
			returnNodes = make([]Expression, 0, numExpr)
			for i, nodeRaw := range nodesPoppedRaw {
				nodeExpr, ok := nodeRaw.(Expression)
				if !ok {
					// Report error for the specific expression that failed assertion
					pos := tokenToPosition(exprListCtx.Expression(i).GetStart()) // Get position of the specific bad expression
					l.errors = append(l.errors, fmt.Errorf("AST build error at %s: RETURN argument %d is not an Expression (got %T)", pos.String(), i+1, nodeRaw))
					// Decide whether to continue or stop entirely
					// For now, let's stop building this step if any assertion fails
					return
				}
				returnNodes = append(returnNodes, nodeExpr)
			}
			l.logDebugAST("    Popped and asserted %d return nodes", len(returnNodes))

		} else {
			// Expression_list exists but is empty (e.g., `return ()`)
			l.logDebugAST("    RETURN statement has empty Expression_list (value will be empty list).")
			returnNodes = []Expression{} // Return empty slice
		}
	} else {
		// No expression list (e.g., `return`)
		l.logDebugAST("    RETURN statement has no expression list (value will be nil)")
		returnNodes = nil // Return nil slice
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Return_statement exited with nil currentSteps")
		return
	}

	// Create and append the step
	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'return' keyword
		Type:      "return",
		Target:    "",          // Not used
		Cond:      nil,         // Not used
		Value:     returnNodes, // Assign the []Expression slice (or nil)
		ElseValue: nil,         // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode Expression // Expecting an Expression or nil

	if ctx.Expression() != nil {
		valueRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop value for EMIT statement")
			return
		}
		// Assert type
		valueNode, ok = valueRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Value for EMIT statement is not an Expression (got %T)", valueRaw)
			return
		}
	} // If no expression, valueNode remains nil

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Emit_statement exited with nil currentSteps")
		return
	}

	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'emit' keyword
		Type:      "emit",
		Target:    "",        // Not used
		Cond:      nil,       // Not used
		Value:     valueNode, // Assign the Expression (or nil)
		ElseValue: nil,       // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", ctx.GetText())
	var valueNode Expression // For 'must', this holds the condition Expression
	var target string        // For 'mustbe', this holds the function name
	stepType := "must"       // Default

	// Pop the single value pushed by visiting the child (either callable_expr for mustbe or expr for must)
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop node for MUST/MUSTBE statement")
		return
	}

	// Check the context type to determine if it was must or mustbe
	if ctx.Callable_expr() != nil {
		// It's a 'mustbe' statement
		stepType = "mustbe"
		// The popped value should be a *CallableExprNode
		callNode, ok := valueRaw.(*CallableExprNode) // Use concrete type here
		if !ok {
			l.addError(ctx, "Internal error: Expected CallableExprNode for MUSTBE statement, got %T", valueRaw)
			return
		}
		target = callNode.Target.Name // Get base name
		if callNode.Target.IsTool {
			// Prepend "tool." if it's a tool call for the target field
			target = "tool." + target
		}
		// The Value field for mustbe holds the *CallableExprNode itself
		valueNode = callNode // Assign the callable node to valueNode (it implements Expression)
		l.logDebugAST("    Interpreting as MUSTBE, Target=%s", target)

	} else if ctx.Expression() != nil {
		// It's a 'must' statement
		stepType = "must"
		// Assert the popped value is an Expression (the condition)
		valueNode, ok = valueRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Condition for MUST statement is not an Expression (got %T)", valueRaw)
			return
		}
		target = "" // Target not used for 'must'
		l.logDebugAST("    Interpreting as MUST")
	} else {
		// Should not happen based on grammar
		l.addError(ctx, "Internal error: Invalid structure for Must_statementContext")
		return
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Must_statement exited with nil currentSteps")
		return
	}

	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'must' or 'mustbe'
		Type:      stepType,
		Target:    target,    // Function name for mustbe, empty for must
		Cond:      nil,       // Not used (condition/call is in Value)
		Value:     valueNode, // Condition Expression for must, CallableExprNode for mustbe
		ElseValue: nil,       // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended %s Step: Value=%T", stepType, valueNode)
}

func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", ctx.GetText())
	var valueNode Expression = nil // Expect Expression or nil

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
	} // valueNode remains nil if no expression

	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Fail_statement exited with nil currentSteps")
		return
	}

	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'fail' keyword
		Type:      "fail",
		Target:    "",        // Not used
		Cond:      nil,       // Not used
		Value:     valueNode, // Assign Expression (or nil)
		ElseValue: nil,       // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(ctx *gen.ClearErrorStmtContext) {
	l.logDebugAST("<<< Exit ClearErrorStmt: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: ClearErrorStmt exited with nil currentSteps")
		return
	}

	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'clear_error'
		Type:      "clear_error",
		Target:    "",  // Not used
		Cond:      nil, // Not used
		Value:     nil, // Not used
		ElseValue: nil, // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended CLEAR_ERROR Step")
}

// *** ADDED: Handler for Ask statement ***
func (l *neuroScriptListenerImpl) ExitAsk_stmt(ctx *gen.Ask_stmtContext) {
	l.logDebugAST("<<< Exit Ask_stmt: %q", ctx.GetText())

	// 1. Pop the prompt expression
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Failed to pop prompt expression for ASK statement")
		return
	}

	// 2. Assert it's an Expression
	promptExpr, ok := valueRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Prompt for ASK statement is not an Expression (got %T)", valueRaw)
		return
	}

	// 3. Check destination
	if l.currentSteps == nil {
		l.addError(ctx, "Internal error: Ask_stmt exited with nil currentSteps")
		return
	}

	// 4. Handle optional 'into' variable
	targetVar := ""
	if ctx.IDENTIFIER() != nil {
		targetVar = ctx.IDENTIFIER().GetText()
		l.logDebugAST("    Ask target variable: %s", targetVar)
	}

	// 5. Create and append step
	step := Step{
		Pos:       tokenToPosition(ctx.GetStart()), // Position of 'ask' keyword
		Type:      "ask",
		Target:    targetVar,  // Store target variable if present
		Cond:      nil,        // Not used
		Value:     promptExpr, // Store the prompt Expression
		ElseValue: nil,        // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended ASK Step: Target=%s, Prompt=%T", targetVar, promptExpr)
}

// --- REMOVED: ExitCall_statement ---
// Call statements are now handled purely as expressions (CallableExprNode)
// The visitor for expressions (`ast_builder_expressions.go` or similar)
// will create the CallableExprNode and push it onto the value stack.
// If a call is used standalone as a statement (e.g., `tool.log("hello")`),
// the grammar rule `statement: call_expr NEWLINE` might exist.
// In the listener, `ExitCall_expr` would push the node, and `ExitStatement`
// might simply pop and discard it if it's not part of a larger structure like `set`.
// --> Let's assume for now that standalone calls used as statements don't need
//     a specific Step type and are handled implicitly by the expression builder.
//     If a dedicated "call_step" is needed later, it can be added.

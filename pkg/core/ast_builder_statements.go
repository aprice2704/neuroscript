// pkg/core/ast_builder_statements.go
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Simple Statement Exit Handlers ---

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())
	valueNode, ok := l.popValue() // Pop the AST node for the RHS expression
	if !ok {
		l.logger.Error("AST Builder: Failed to pop value for SET")
		return
	}
	if l.currentSteps == nil {
		l.logger.Warn("Set_statement exited with nil currentSteps")
		return
	}

	varName := ctx.IDENTIFIER().GetText()
	// Use generic newStep, ElseValue is nil for SET
	step := Step{Type: "set", Target: varName, Value: valueNode}
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

	argNodes, ok := l.popNValues(numArgs) // Pop arg nodes from stack
	if !ok {
		if numArgs > 0 {
			l.logger.Error("AST Builder: Failed to pop %d args for CALL", numArgs)
		}
		if numArgs == 0 {
			argNodes = []interface{}{} // Valid case: no args popped is okay
		} else {
			return // Don't append step if args failed to pop and were expected
		}
	}

	if l.currentSteps == nil {
		l.logger.Warn("Call_statement exited with nil currentSteps")
		return
	}
	target := ctx.Call_target().GetText()
	// Use generic newStep, pass argNodes via Args field
	step := Step{Type: "call", Target: target, Args: argNodes}
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var valueNode interface{} = nil // Default nil if no expression
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue() // Pop the node for the return expression
		if !ok {
			l.logger.Error("AST Builder: Failed to pop value for RETURN")
			// Keep valueNode as nil
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Return_statement exited with nil currentSteps")
		return
	}
	step := Step{Type: "return", Value: valueNode}
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode interface{} = nil // Default nil if no expression
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue() // Pop the node for the emit expression
		if !ok {
			l.logger.Error("AST Builder: Failed to pop value for EMIT")
			// Allow nil value for EMIT
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Emit_statement exited with nil currentSteps")
		return
	}
	step := Step{Type: "emit", Value: valueNode}
	*l.currentSteps = append(*l.currentSteps, step)
}

// --- NEW: Must Statement Handling (v0.2.0) ---
func (l *neuroScriptListenerImpl) ExitMust_statement(ctx *gen.Must_statementContext) {
	l.logDebugAST("<<< Exit Must_statement: %q", ctx.GetText())
	var valueNode interface{} = nil
	var ok bool
	stepType := "must" // Default type

	// Pop the single value from the stack (either the expression or the function call node)
	valueNode, ok = l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop value/node for MUST/MUSTBE")
		return
	}

	// Determine if it was 'must expression' or 'mustbe function_call'
	// We can infer from the type of node popped, assuming the grammar is correct.
	if _, isFuncCall := valueNode.(FunctionCallNode); isFuncCall {
		stepType = "mustbe"
		l.logDebugAST("    Interpreting as MUSTBE")
	} else {
		l.logDebugAST("    Interpreting as MUST")
	}

	if l.currentSteps == nil {
		l.logger.Warn("Must_statement exited with nil currentSteps")
		return
	}

	// Store the expression/FunctionCallNode in the 'Value' field.
	// For 'mustbe', the Target field can optionally store the function name if needed later.
	target := ""
	if stepType == "mustbe" {
		target = valueNode.(FunctionCallNode).FunctionName // Store func name in Target for mustbe
	}
	step := Step{Type: stepType, Target: target, Value: valueNode}
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended %s Step: Value=%T", stepType, valueNode)
}

// --- NEW: Fail Statement Handling (v0.2.0) ---
func (l *neuroScriptListenerImpl) ExitFail_statement(ctx *gen.Fail_statementContext) {
	l.logDebugAST("<<< Exit Fail_statement: %q", ctx.GetText())
	var valueNode interface{} = nil // Default nil if no expression
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue() // Pop the node for the fail message expression
		if !ok {
			l.logger.Error("AST Builder: Failed to pop value for FAIL")
			// Allow nil value for FAIL (will use default message)
		}
	}
	if l.currentSteps == nil {
		l.logger.Warn("Fail_statement exited with nil currentSteps")
		return
	}
	step := Step{Type: "fail", Value: valueNode}
	*l.currentSteps = append(*l.currentSteps, step)
}

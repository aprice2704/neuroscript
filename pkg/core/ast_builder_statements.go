// pkg/core/ast_builder_statements.go
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Statement Exit Handlers (Pop final expression node and store in Step) ---

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement: %q", ctx.GetText())
	valueNode, ok := l.popValue() // Pop the AST node for the RHS expression
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to pop value for SET")
		return
	}
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Set_statement exited with nil currentSteps")
		return
	}

	varName := ctx.IDENTIFIER().GetText()
	step := newStep("SET", varName, nil, valueNode, nil) // Store node in Value
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement: %q", ctx.GetText())
	numArgs := 0
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		numArgs = len(ctx.Expression_list_opt().Expression_list().AllExpression())
	}

	argNodes, ok := l.popNValues(numArgs) // Pop arg nodes from stack
	if !ok {
		l.logger.Printf("[ERROR] AST Builder: Failed to pop %d args for CALL", numArgs)
		return
	} // Error handling

	if l.currentSteps == nil {
		l.logger.Println("[WARN] Call_statement exited with nil currentSteps")
		return
	}
	target := ctx.Call_target().GetText()
	step := newStep("CALL", target, nil, nil, argNodes) // Store nodes in Args
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement: %q", ctx.GetText())
	var valueNode interface{} = nil // Default nil if no expression
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue() // Pop the node for the return expression
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop value for RETURN")
			// Keep valueNode as nil, which is valid for RETURN
		}
	}
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Return_statement exited with nil currentSteps")
		return
	}
	step := newStep("RETURN", "", nil, valueNode, nil) // Store node (or nil) in Value
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement: %q", ctx.GetText())
	var valueNode interface{} = nil // Default nil if no expression
	if ctx.Expression() != nil {
		var ok bool
		valueNode, ok = l.popValue() // Pop the node for the emit expression
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop value for EMIT")
			// Keep valueNode as nil? Or should EMIT require a value? For now, allow nil.
		}
	}
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Emit_statement exited with nil currentSteps")
		return
	}
	step := newStep("EMIT", "", nil, valueNode, nil) // Store node (or nil) in Value
	*l.currentSteps = append(*l.currentSteps, step)
}

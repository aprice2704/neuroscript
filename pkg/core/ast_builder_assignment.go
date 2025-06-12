// ast_builder_assignment.go – builds assignment Step nodes
// file version: 3
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitSet_statement handles the grammar rule:
//
//	set_statement : lvalue '=' expression                 # simple assignment
//	              | lvalue ( ',' lvalue )* '=' expression # multi‑assign  (not yet handled)
//
// It pops RHS then LHS from valueStack, constructs a Step with Type="set",
// and appends it to *currentSteps.
func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	// --- Pop RHS expression ---
	rhsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "AST internal error: RHS missing for assignment")
		return
	}
	rhs, ok := rhsRaw.(Expression)
	if !ok {
		l.addError(ctx, "AST internal error: RHS not Expression (got %T)", rhsRaw)
		return
	}

	// --- Pop LHS lvalue ---
	lhsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "AST internal error: LHS missing for assignment")
		return
	}
	lhs, ok := lhsRaw.(*LValueNode)
	if !ok {
		l.addError(ctx, "AST internal error: LHS not LValueNode (got %T)", lhsRaw)
		return
	}

	// Ensure currentSteps slice exists (for ON_EVENT_BODY first step case)
	if l.currentSteps == nil {
		tmp := make([]Step, 0)
		l.currentSteps = &tmp
	}

	// CORRECTED: The step type should be "set" to match the keyword and test expectations.
	step := Step{
		Pos:    tokenToPosition(ctx.GetStart()),
		Type:   "set",
		LValue: lhs,
		Value:  rhs,
	}
	*l.currentSteps = append(*l.currentSteps, step)
}

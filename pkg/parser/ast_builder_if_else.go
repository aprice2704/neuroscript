// filename: pkg/parser/ast_builder_if_else.go
// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Sets the end position of if/else step nodes using the StopPos field.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) ExitIf_statement(c *gen.If_statementContext) {
	// Pop the 'else' block if it exists.
	var elseBody []ast.Step
	if c.KW_ELSE() != nil {
		val, ok := l.pop()
		if !ok {
			l.addError(c, "internal error in if_statement: could not pop else body")
			return
		}
		elseBody, ok = val.([]ast.Step)
		if !ok {
			l.addError(c, "internal error in if_statement: else body is not a []ast.Step, but %T", val)
			return
		}
	}

	// Pop the 'if' block.
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in if_statement: could not pop if body")
		return
	}
	ifBody, ok := val.([]ast.Step)
	if !ok {
		l.addError(c, "internal error in if_statement: if body is not a []ast.Step, but %T", val)
		return
	}

	// Pop the condition.
	condVal, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in if_statement: could not pop condition")
		return
	}
	cond, ok := condVal.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in if_statement: condition is not an ast.Expression, but %T", condVal)
		return
	}

	pos := tokenToPosition(c.GetStart())
	// Create and add the 'if' step.
	step := ast.Step{
		BaseNode:         ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:             "if",
		Cond:             cond,
		Body:             ifBody,
		ElseBody:         elseBody,
		BlankLinesBefore: l.consumeBlankLines(),
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.KW_ENDIF().GetSymbol())
	l.addStep(step)
}

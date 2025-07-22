// filename: pkg/parser/ast_builder_if_else.go
// NeuroScript Version: 0.6.0
// File version: 5
// Purpose: Removed obsolete blank line counting logic. Association is now handled by the LineInfo algorithm.

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) ExitIf_statement(c *gen.If_statementContext) {
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
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "if",
		Cond:     cond,
		Body:     ifBody,
		ElseBody: elseBody,
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.KW_ENDIF().GetSymbol())
	l.addStep(step)
}

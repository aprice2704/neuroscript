// NeuroScript Version: 0.7.2
// File version: 6
// Purpose: Ensures ElseBody is initialized to a non-nil empty slice for consistency, preventing diffs in canonicalization.
// filename: pkg/parser/ast_builder_if_else.go

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
	} else {
		// FIX: Explicitly create an empty slice for if-statements without an else block.
		elseBody = make([]ast.Step, 0)
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
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, c.KW_ENDIF().GetSymbol())
	l.addStep(step)
}

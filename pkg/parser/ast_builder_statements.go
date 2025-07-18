// NeuroScript Version: 0.6.0
// File version: 16
// Purpose: Removes a redundant newline in a debug print statement that was causing a build failure.
// filename: pkg/parser/ast_builder_statements.go
// nlines: 125
// risk_rating: LOW

package parser

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// addStep is a helper to append a new step to the correct part of the AST.
func (l *neuroScriptListenerImpl) addStep(step ast.Step) {
	if len(l.blockStack) > 0 {
		currentBlock := l.blockStack[len(l.blockStack)-1]
		currentBlock.steps = append(currentBlock.steps, step)
	} else {
		l.addError(nil, "internal AST error: addStep called with no active block context")
	}
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(c *gen.Emit_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in emit_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in emit_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "emit",
		Values:   []ast.Expression{expr},
	})
}

func (l *neuroScriptListenerImpl) ExitReturn_statement(c *gen.Return_statementContext) {
	fmt.Println("\n--- ENTERING ExitReturn_statement ---")
	var returnValues []ast.Expression
	if c.Expression_list() != nil {
		numExpr := len(c.Expression_list().AllExpression())
		fmt.Printf("DEBUG: Expecting %d expressions for return statement.\n", numExpr)
		if numExpr > 0 {
			popped, ok := l.popN(numExpr)
			if !ok {
				l.addError(c, "internal error in return_statement: could not pop values")
				return
			}
			fmt.Printf("DEBUG: Popped %d items from the stack.\n", len(popped))
			for i, val := range popped {
				fmt.Printf("DEBUG: Processing popped value #%d: Type=%T, Value=%#v\n", i, val, val)
				var expr ast.Expression
				isExpr := false

				expr, isExpr = val.(ast.Expression)
				if !isExpr {
					fmt.Printf("DEBUG: Popped value #%d is NOT an ast.Expression. Checking if it's a convertible LValueNode.\n", i)
					if lval, isLval := val.(*ast.LValueNode); isLval && len(lval.Accessors) == 0 {
						fmt.Printf("DEBUG: It IS a simple LValueNode. Converting to VariableNode.\n")
						expr = &ast.VariableNode{
							BaseNode: lval.BaseNode,
							Name:     lval.Identifier,
						}
						isExpr = true
					}
				}

				if isExpr {
					fmt.Printf("DEBUG: Successfully processed value #%d into an expression. Appending to returnValues.\n", i)
					returnValues = append(returnValues, expr)
				} else {
					fmt.Printf("ERROR: Could not process value #%d into an expression.\n", i)
					l.addError(c, "internal error in return_statement: value is not an ast.Expression, but %T", val)
					return
				}
			}
		}
	}
	pos := tokenToPosition(c.GetStart())
	fmt.Printf("DEBUG: Creating return step with %d values.\n", len(returnValues))
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "return",
		Values:   returnValues,
	})
	fmt.Println("--- EXITING ExitReturn_statement ---")
}

func (l *neuroScriptListenerImpl) ExitCall_statement(c *gen.Call_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in call_statement: could not pop value")
		return
	}
	callExpr, ok := val.(*ast.CallableExprNode)
	if !ok {
		l.addError(c, "internal error in call_statement: value is not a *ast.CallableExprNode, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "call",
		Call:     callExpr,
	})
}

func (l *neuroScriptListenerImpl) ExitMust_statement(c *gen.Must_statementContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in must_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in must_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "must",
		Cond:     expr,
	})
}

func (l *neuroScriptListenerImpl) ExitFail_statement(c *gen.Fail_statementContext) {
	var failValue ast.Expression
	if c.Expression() != nil {
		val, ok := l.pop()
		if !ok {
			l.addError(c, "internal error in fail_statement: could not pop value")
			return
		}
		expr, ok := val.(ast.Expression)
		if !ok {
			l.addError(c, "internal error in fail_statement: value is not an ast.Expression, but %T", val)
			return
		}
		failValue = expr
	}

	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "fail",
	}
	if failValue != nil {
		step.Values = []ast.Expression{failValue}
	}
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) ExitAsk_stmt(c *gen.Ask_stmtContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "internal error in ask_statement: could not pop value")
		return
	}
	expr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in ask_statement: value is not an ast.Expression, but %T", val)
		return
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode:   ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position:   pos,
		Type:       "ask",
		Values:     []ast.Expression{expr},
		AskIntoVar: c.IDENTIFIER().GetText(),
	})
}

func (l *neuroScriptListenerImpl) ExitClearErrorStmt(c *gen.ClearErrorStmtContext) {
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "clear_error",
	})
}

func (l *neuroScriptListenerImpl) ExitContinue_statement(c *gen.Continue_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'continue' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "continue",
	})
}

func (l *neuroScriptListenerImpl) ExitBreak_statement(c *gen.Break_statementContext) {
	if !l.isInsideLoop() {
		l.addError(c, "'break' statement found outside of a loop")
	}
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "break",
	})
}

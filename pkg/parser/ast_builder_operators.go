// NeuroScript Version: 0.3.1
// File version: 17
// Purpose: Corrected buildCallTargetFromContext to reconstruct the full tool name by prepending the tool prefix, fixing tool lookup errors.
// filename: pkg/parser/ast_builder_operators.go
// nlines: 606
// risk_rating: MEDIUM

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4" // Using user-specified ANTLR import path
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Helper function to check if a lang.Value is an ast.ErrorNode
func isErrorNode(val interface{}) bool {
	_, ok := val.(*ast.ErrorNode)
	return ok
}

// --- Helper and Operator Exit methods ---

// processBinaryOperators (existing helper)
func (l *neuroScriptListenerImpl) processBinaryOperators(ctx antlr.ParserRuleContext, numOperands int, opGetter func(i int) antlr.TerminalNode) {
	if numOperands <= 1 {
		return
	}

	numOperators := numOperands - 1
	if numOperators < 1 {
		l.addError(ctx, "Internal error: processBinaryOperators with numOperands=%d implies no operators.", numOperands)
		return
	}

	poppedOperands := make([]ast.Expression, numOperands)
	for i := 0; i < numOperands; i++ {
		val, ok := l.pop()
		if !ok {
			pos := tokenToPosition(ctx.GetStart())
			l.addError(ctx, "Stack error popping operand %d for binary op: %s", numOperands-i, ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: fmt.Sprintf("Stack error (binary op operand %d)", numOperands-i)})
			return
		}
		expr, isExpr := val.(ast.Expression)
		if !isExpr {
			errPosToken := ctx.GetStart()
			pos := tokenToPosition(errPosToken)
			l.addError(ctx, "Operand %d is not an ast.Expression (type %T) for binary op: %s", numOperands-i, val, ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: fmt.Sprintf("Type error (binary op operand %d)", numOperands-i)})
			return
		}
		poppedOperands[i] = expr
	}

	currentLHS := poppedOperands[numOperands-1]

	for i := 0; i < numOperators; i++ {
		opToken := opGetter(i)
		if opToken == nil {
			pos := tokenToPosition(ctx.GetStart())
			l.addError(ctx, "Could not find operator token for index %d in: %s", i, ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Missing operator token"})
			return
		}
		opSymbol := opToken.GetSymbol()
		opText := opSymbol.GetText()
		currentRHS := poppedOperands[numOperands-2-i]
		pos := tokenToPosition(opSymbol)
		newNode := &ast.BinaryOpNode{
			Pos:      &pos,
			Left:     currentLHS,
			Operator: opText,
			Right:    currentRHS,
		}
		l.logDebugAST("      Constructed ast.BinaryOpNode: [%T %s %T]", currentLHS, opText, currentRHS)
		currentLHS = newNode
	}
	l.push(currentLHS)
}

// ExitLogical_or_expr
func (l *neuroScriptListenerImpl) ExitLogical_or_expr(ctx *gen.Logical_or_exprContext) {
	l.logDebugAST("--- Exit Logical_or_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllLogical_and_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.KW_OR(i) }
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitLogical_and_expr
func (l *neuroScriptListenerImpl) ExitLogical_and_expr(ctx *gen.Logical_and_exprContext) {
	l.logDebugAST("--- Exit Logical_and_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_or_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.KW_AND(i) }
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_or_expr
func (l *neuroScriptListenerImpl) ExitBitwise_or_expr(ctx *gen.Bitwise_or_exprContext) {
	l.logDebugAST("--- Exit Bitwise_or_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_xor_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.PIPE(i) }
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_xor_expr
func (l *neuroScriptListenerImpl) ExitBitwise_xor_expr(ctx *gen.Bitwise_xor_exprContext) {
	l.logDebugAST("--- Exit Bitwise_xor_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_and_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.CARET(i) }
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_and_expr
func (l *neuroScriptListenerImpl) ExitBitwise_and_expr(ctx *gen.Bitwise_and_exprContext) {
	l.logDebugAST("--- Exit Bitwise_and_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllEquality_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.AMPERSAND(i) }
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitEquality_expr
func (l *neuroScriptListenerImpl) ExitEquality_expr(ctx *gen.Equality_exprContext) {
	l.logDebugAST("--- Exit Equality_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllRelational_expr())
	opGetter := func(i int) antlr.TerminalNode {
		opCount := 0
		for _, child := range ctx.GetChildren() {
			if term, ok := child.(antlr.TerminalNode); ok {
				tokenType := term.GetSymbol().GetTokenType()
				if tokenType == gen.NeuroScriptLexerEQ || tokenType == gen.NeuroScriptLexerNEQ {
					if opCount == i {
						return term
					}
					opCount++
				}
			}
		}
		return nil
	}
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitRelational_expr
func (l *neuroScriptListenerImpl) ExitRelational_expr(ctx *gen.Relational_exprContext) {
	l.logDebugAST("--- Exit Relational_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllAdditive_expr())
	opGetter := func(i int) antlr.TerminalNode {
		opCount := 0
		for _, child := range ctx.GetChildren() {
			if term, ok := child.(antlr.TerminalNode); ok {
				tokenType := term.GetSymbol().GetTokenType()
				if tokenType == gen.NeuroScriptLexerGT || tokenType == gen.NeuroScriptLexerLT ||
					tokenType == gen.NeuroScriptLexerGTE || tokenType == gen.NeuroScriptLexerLTE {
					if opCount == i {
						return term
					}
					opCount++
				}
			}
		}
		return nil
	}
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// EnterAdditive_expr is required to satisfy the listener interface. It does
// not need to perform any logic.
func (l *neuroScriptListenerImpl) EnterAdditive_expr(ctx *gen.Additive_exprContext) {
	l.logDebugAST("--- Enter Additive_expr: %q", ctx.GetText())
}

// ExitAdditive_expr
func (l *neuroScriptListenerImpl) ExitAdditive_expr(ctx *gen.Additive_exprContext) {
	l.logDebugAST("--- Exit Additive_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllMultiplicative_expr())
	opGetter := func(i int) antlr.TerminalNode {
		opCount := 0
		for _, child := range ctx.GetChildren() {
			if term, ok := child.(antlr.TerminalNode); ok {
				tokenType := term.GetSymbol().GetTokenType()
				if tokenType == gen.NeuroScriptLexerPLUS || tokenType == gen.NeuroScriptLexerMINUS {
					if opCount == i {
						return term
					}
					opCount++
				}
			}
		}
		return nil
	}
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitMultiplicative_expr
func (l *neuroScriptListenerImpl) ExitMultiplicative_expr(ctx *gen.Multiplicative_exprContext) {
	l.logDebugAST("--- Exit Multiplicative_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllUnary_expr())
	opGetter := func(i int) antlr.TerminalNode {
		opCount := 0
		for _, child := range ctx.GetChildren() {
			if term, ok := child.(antlr.TerminalNode); ok {
				tokenType := term.GetSymbol().GetTokenType()
				if tokenType == gen.NeuroScriptLexerSTAR || tokenType == gen.NeuroScriptLexerSLASH || tokenType == gen.NeuroScriptLexerPERCENT {
					if opCount == i {
						return term
					}
					opCount++
				}
			}
		}
		return nil
	}
	l.processBinaryOperators(ctx, numOperands, opGetter)
}
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- ExitUnary_expr: %q", ctx.GetText())

	/* ---------- typeof ---------- */

	if ctx.KW_TYPEOF() != nil {
		operandVal, ok := l.pop()
		if !ok {
			startPos := tokenToPosition(ctx.KW_TYPEOF().GetSymbol())
			l.addError(ctx, "Stack error: missing operand for typeof at %s", startPos.String())
			l.push(&ast.ErrorNode{Pos: &startPos, Message: "missing operand for typeof"})
			return
		}
		operandExpr, ok := operandVal.(ast.Expression)
		if !ok {
			startPos := tokenToPosition(ctx.KW_TYPEOF().GetSymbol())
			l.addError(ctx, "typeof operand is not ast.Expression (got %T) at %s", operandVal, startPos.String())
			l.push(&ast.ErrorNode{Pos: &startPos, Message: fmt.Sprintf("typeof operand was %T", operandVal)})
			return
		}
		pos := tokenToPosition(ctx.KW_TYPEOF().GetSymbol())
		l.push(&ast.TypeOfNode{
			Pos:      &pos,
			Argument: operandExpr,
		})
		return
	}

	/* ---------- other unary ops ---------- */

	var tok antlr.TerminalNode
	var op string

	switch {
	case ctx.MINUS() != nil:
		tok, op = ctx.MINUS(), "-"
	case ctx.KW_NOT() != nil:
		tok, op = ctx.KW_NOT(), "not"
	case ctx.KW_NO() != nil:
		tok, op = ctx.KW_NO(), "no"
	case ctx.KW_SOME() != nil:
		tok, op = ctx.KW_SOME(), "some"
	case ctx.TILDE() != nil:
		tok, op = ctx.TILDE(), "~"
	default:
		// pass-through (e.g. power_expr)
		return
	}

	operandRaw, ok := l.pop()
	if !ok {
		pos := tokenToPosition(tok.GetSymbol())
		l.addError(ctx, "Stack error: missing operand for unary %q", op)
		l.push(&ast.ErrorNode{Pos: &pos, Message: "stack underflow (unary)"})
		return
	}
	operandExpr, ok := operandRaw.(ast.Expression)
	if !ok {
		pos := tokenToPosition(tok.GetSymbol())
		l.addError(ctx, "Operand for unary %q is not ast.Expression (got %T)", op, operandRaw)
		l.push(&ast.ErrorNode{Pos: &pos, Message: "type error (unary)"})
		return
	}
	pos := tokenToPosition(tok.GetSymbol())
	l.push(&ast.UnaryOpNode{
		Pos:      &pos,
		Operator: op,
		Operand:  operandExpr,
	})
	l.logDebugAST("      Constructed ast.UnaryOpNode: %s [%T]", op, operandExpr)
}

// ExitPower_expr
func (l *neuroScriptListenerImpl) ExitPower_expr(ctx *gen.Power_exprContext) {
	l.logDebugAST("--- Exit Power_expr: %q", ctx.GetText())
	opTokenNode := ctx.STAR_STAR()
	if opTokenNode == nil {
		// Pass through from accessor_expr
		return
	}
	opSymbol := opTokenNode.GetSymbol()
	opText := opSymbol.GetText()

	exponentRaw, ok := l.pop()
	if !ok {
		pos := tokenToPosition(opSymbol)
		l.addError(ctx, "Stack error popping exponent for POWER")
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (power exponent)"})
		return
	}
	exponentExpr, isExpr := exponentRaw.(ast.Expression)
	if !isExpr {
		pos := tokenToPosition(opSymbol)
		l.addError(ctx, "Exponent for POWER is not an ast.Expression (type %T)", exponentRaw)
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (power exponent)"})
		return
	}
	baseRaw, ok := l.pop()
	if !ok {
		pos := tokenToPosition(opSymbol)
		l.addError(ctx, "Stack error popping base for POWER")
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (power base)"})
		return
	}
	baseExpr, isExpr := baseRaw.(ast.Expression)
	if !isExpr {
		pos := tokenToPosition(opSymbol)
		l.addError(ctx, "Base for POWER is not an ast.Expression (type %T)", baseRaw)
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (power base)"})
		return
	}
	pos := tokenToPosition(opSymbol)
	node := &ast.BinaryOpNode{
		Pos:      &pos,
		Left:     baseExpr,
		Operator: opText,
		Right:    exponentExpr,
	}
	l.push(node)
	l.logDebugAST("      Constructed ast.BinaryOpNode (Power): [%T %s %T]", baseExpr, opText, exponentExpr)
}

// EnterAccessor_expr is required to satisfy the listener interface now that the
// base listener is not embedded. It doesn't need to do anything; the child rules
// will handle pushing their Values onto the stack.
func (l *neuroScriptListenerImpl) EnterAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Enter Accessor_expr: %q", ctx.GetText())
}

// ExitAccessor_expr
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Exit Accessor_expr: %q", ctx.GetText())
	numAccessors := len(ctx.AllLBRACK())
	if numAccessors == 0 {
		// Pass through from primary
		return
	}

	accessorExprs := make([]ast.Expression, numAccessors)
	for i := numAccessors - 1; i >= 0; i-- {
		accessorRaw, ok := l.pop()
		if !ok {
			if i < len(ctx.AllExpression()) {
				l.addError(ctx.Expression(i), "Stack error popping accessor expression %d", i)
			} else {
				l.addError(ctx, "Stack error popping accessor expression %d (index out of bounds)", i)
			}
			pos := tokenToPosition(ctx.LBRACK(i).GetSymbol())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (accessor expr)"})
			return
		}
		accessorExpr, isExpr := accessorRaw.(ast.Expression)
		if !isExpr {
			if i < len(ctx.AllExpression()) {
				l.addError(ctx.Expression(i), "Accessor expression %d is not an ast.Expression (type %T)", i, accessorRaw)
			} else {
				l.addError(ctx, "Accessor expression %d is not an ast.Expression (type %T) (index out of bounds for error pos)", i, accessorRaw)
			}
			pos := tokenToPosition(ctx.LBRACK(i).GetSymbol())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (accessor expr)"})
			return
		}
		accessorExprs[i] = accessorExpr
	}

	collectionRaw, ok := l.pop()
	if !ok {
		pos := tokenToPosition(ctx.Primary().GetStart())
		l.addError(ctx.Primary(), "Stack error popping primary collection")
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (accessor collection)"})
		return
	}
	collectionExpr, isExpr := collectionRaw.(ast.Expression)
	if !isExpr {
		pos := tokenToPosition(ctx.Primary().GetStart())
		l.addError(ctx.Primary(), "Primary collection is not an ast.Expression (type %T)", collectionRaw)
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (accessor collection)"})
		return
	}

	currentCollectionResult := collectionExpr
	for i := 0; i < numAccessors; i++ {
		pos := tokenToPosition(ctx.LBRACK(i).GetSymbol())
		newNode := &ast.ElementAccessNode{
			Pos:        &pos,
			Collection: currentCollectionResult,
			Accessor:   accessorExprs[i],
		}
		l.logDebugAST("      Constructed ast.ElementAccessNode: [Coll: %T Acc: %T]", newNode.Collection, newNode.Accessor)
		currentCollectionResult = newNode
	}
	l.push(currentCollectionResult)
	l.logDebugAST("      Final Accessor_expr result pushed: %T", currentCollectionResult)
}

// buildCallTargetFromContext constructs a ast.CallTarget AST node from an ICall_targetContext.
func (l *neuroScriptListenerImpl) buildCallTargetFromContext(ctx gen.ICall_targetContext) ast.CallTarget {
	l.logDebugAST("      -> buildCallTargetFromContext: %s", ctx.GetText())
	target := ast.CallTarget{}

	if toolKeyword := ctx.KW_TOOL(); toolKeyword != nil {
		target.IsTool = true
		if qiCtx := ctx.Qualified_identifier(); qiCtx != nil {
			idNodes := qiCtx.AllIDENTIFIER()
			var parts []string
			for _, idNode := range idNodes {
				parts = append(parts, idNode.GetText())
			}
			// FIX: Prepend the tool prefix to construct the full, correct tool name.
			groupAndName := strings.Join(parts, types.ToolSep)
			target.Name = types.ToolPrefix + groupAndName

			if len(idNodes) > 0 {
				pos := tokenToPosition(idNodes[0].GetSymbol())
				target.Pos = &pos
			} else {
				pos := tokenToPosition(toolKeyword.GetSymbol())
				target.Pos = &pos
				l.addError(ctx, "Tool call has empty qualified_identifier: %s", ctx.GetText())
			}
		} else {
			l.addError(ctx, "Tool call: Expected Qualified_identifier, but was not found: %s", ctx.GetText())
			target.Name = "<ERROR_NO_QUALIFIED_ID_FOR_TOOL>"
			pos := tokenToPosition(toolKeyword.GetSymbol())
			target.Pos = &pos
		}
		l.logDebugAST("         Tool call identified. Name: '%s', Position: %s", target.Name, target.Pos.String())
	} else if userFuncID := ctx.IDENTIFIER(); userFuncID != nil {
		target.IsTool = false
		target.Name = userFuncID.GetText()
		pos := tokenToPosition(userFuncID.GetSymbol())
		target.Pos = &pos
		l.logDebugAST("         User function call identified. Name: '%s', Position: %s", target.Name, target.Pos.String())
	} else {
		l.addError(ctx, "Unrecognized call_target structure: %s", ctx.GetText())
		target.Name = "<ERROR_INVALID_CALL_TARGET>"
		pos := tokenToPosition(ctx.GetStart())
		target.Pos = &pos
	}
	l.logDebugAST("      <- buildCallTargetFromContext (Name: %s, IsTool: %v)", target.Name, target.IsTool)
	return target
}

// ExitCall_target is called when exiting the call_target rule.
func (l *neuroScriptListenerImpl) ExitCall_target(ctx *gen.Call_targetContext) {
	l.logDebugAST("--- Exit Call_target: %q", ctx.GetText())
	targetNode := l.buildCallTargetFromContext(ctx)
	l.push(&targetNode)
	l.logDebugAST("      Pushed *ast.CallTarget to stack: IsTool=%t, Name=%s", targetNode.IsTool, targetNode.Name)
}

// ExitCallable_expr
func (l *neuroScriptListenerImpl) ExitCallable_expr(ctx *gen.Callable_exprContext) {
	l.logDebugAST("--- Exit Callable_expr: %q", ctx.GetText())

	var args []ast.Expression
	if exprListOptCtx := ctx.Expression_list_opt(); exprListOptCtx != nil {
		if exprListCtx := exprListOptCtx.Expression_list(); exprListCtx != nil {
			numArgs := len(exprListCtx.AllExpression())
			if numArgs > 0 {
				argsRaw, ok := l.popN(numArgs)
				if !ok {
					pos := tokenToPosition(ctx.GetStart())
					l.addError(ctx, "Stack error popping arguments for call %q", ctx.GetText())
					l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (call args)"})
					return
				}
				args = make([]ast.Expression, numArgs)
				for i := 0; i < numArgs; i++ {
					argExpr, isExpr := argsRaw[i].(ast.Expression)
					if !isExpr {
						argSourceCtx := ctx.Expression_list_opt().Expression_list().Expression(i)
						pos := tokenToPosition(argSourceCtx.GetStart())
						l.addError(argSourceCtx, "Argument %d for call %q is not an ast.Expression (type %T)", i+1, ctx.GetText(), argsRaw[i])
						l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (call arg)"})
						return
					}
					args[i] = argExpr
				}
			}
		}
	}

	var finalTargetNode ast.CallTarget
	var callExprPos *lang.Position

	if targetRuleCtx := ctx.Call_target(); targetRuleCtx != nil {
		targetVal, ok := l.pop()
		if !ok {
			pos := tokenToPosition(ctx.GetStart())
			l.addError(ctx, "Stack error popping call target for %q", ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Stack error (call target)"})
			return
		}
		targetPtr, isPtr := targetVal.(*ast.CallTarget)
		if !isPtr {
			pos := tokenToPosition(ctx.GetStart())
			l.addError(ctx, "Popped call target is not *ast.CallTarget (type %T) for %q", targetVal, ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Type error (call target)"})
			return
		}
		finalTargetNode = *targetPtr
		callExprPos = finalTargetNode.Pos
		l.logDebugAST("      Popped *ast.CallTarget from stack: IsTool=%t, Name=%s", finalTargetNode.IsTool, finalTargetNode.Name)
	} else {
		finalTargetNode.IsTool = false
		var keywordToken antlr.TerminalNode
		switch {
		case ctx.KW_LN() != nil:
			keywordToken = ctx.KW_LN()
			finalTargetNode.Name = "ln"
		case ctx.KW_LOG() != nil:
			keywordToken = ctx.KW_LOG()
			finalTargetNode.Name = "log"
		case ctx.KW_SIN() != nil:
			keywordToken = ctx.KW_SIN()
			finalTargetNode.Name = "sin"
		case ctx.KW_COS() != nil:
			keywordToken = ctx.KW_COS()
			finalTargetNode.Name = "cos"
		case ctx.KW_TAN() != nil:
			keywordToken = ctx.KW_TAN()
			finalTargetNode.Name = "tan"
		case ctx.KW_ASIN() != nil:
			keywordToken = ctx.KW_ASIN()
			finalTargetNode.Name = "asin"
		case ctx.KW_ACOS() != nil:
			keywordToken = ctx.KW_ACOS()
			finalTargetNode.Name = "acos"
		case ctx.KW_ATAN() != nil:
			keywordToken = ctx.KW_ATAN()
			finalTargetNode.Name = "atan"
		case ctx.KW_LEN() != nil:
			keywordToken = ctx.KW_LEN()
			finalTargetNode.Name = "len"
		default:
			pos := tokenToPosition(ctx.GetStart())
			l.addError(ctx, "Unhandled built-in or target type in Callable_expr: %q", ctx.GetText())
			l.push(&ast.ErrorNode{Pos: &pos, Message: "Unknown callable target"})
			return
		}
		pos := tokenToPosition(keywordToken.GetSymbol())
		callExprPos = &pos
		finalTargetNode.Pos = callExprPos
		l.logDebugAST("      Identified Built-in function call target: %s", finalTargetNode.Name)
	}

	node := &ast.CallableExprNode{
		Pos:       callExprPos,
		Target:    finalTargetNode,
		Arguments: args,
	}
	l.push(node)
	l.logDebugAST("      Constructed and Pushed ast.CallableExprNode: Target=%s, Args=%d", node.Target.Name, len(node.Arguments))
}

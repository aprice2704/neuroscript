// filename: pkg/parser/ast_builder_operators.go
// NeuroScript Version: 0.5.2
// File version: 23
// Purpose: Corrected tool call parsing to handle fully qualified tool names (e.g., tool.group.Name) instead of just two parts.
// nlines: 215
// risk_rating: HIGH

package parser

import (
	"fmt"
	"strings" // Import strings

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Helper function to check if a lang.Value is an ast.ErrorNode
func isErrorNode(val interface{}) bool {
	_, ok := val.(*ast.ErrorNode)
	return ok
}

// --- Helper and Operator Exit methods ---

// processBinaryOperators
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
			l.addError(ctx, "Stack error popping operand %d for binary op: %s", numOperands-i, ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: fmt.Sprintf("Stack error (binary op operand %d)", numOperands-i)}, ctx.GetStart(), types.KindUnknown))
			return
		}
		expr, isExpr := val.(ast.Expression)
		if !isExpr {
			l.addError(ctx, "Operand %d is not an ast.Expression (type %T) for binary op: %s", numOperands-i, val, ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: fmt.Sprintf("Type error (binary op operand %d)", numOperands-i)}, ctx.GetStart(), types.KindUnknown))
			return
		}
		poppedOperands[i] = expr
	}

	currentLHS := poppedOperands[numOperands-1]

	for i := 0; i < numOperators; i++ {
		opToken := opGetter(i)
		if opToken == nil {
			l.addError(ctx, "Could not find operator token for index %d in: %s", i, ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: "Missing operator token"}, ctx.GetStart(), types.KindUnknown))
			return
		}
		opSymbol := opToken.GetSymbol()
		opText := opSymbol.GetText()
		currentRHS := poppedOperands[numOperands-2-i]
		node := &ast.BinaryOpNode{
			Left:     currentLHS,
			Operator: opText,
			Right:    currentRHS,
		}
		l.logDebugAST("      Constructed ast.BinaryOpNode: [%T %s %T]", currentLHS, opText, currentRHS)
		currentLHS = newNode(node, opSymbol, types.KindBinaryOp)
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

// EnterAdditive_expr
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

	if ctx.KW_TYPEOF() != nil {
		token := ctx.KW_TYPEOF().GetSymbol()
		operandVal, ok := l.pop()
		if !ok {
			l.addError(ctx, "Stack error: missing operand for typeof at %v", tokenToPosition(token))
			l.push(newNode(&ast.ErrorNode{Message: "missing operand for typeof"}, token, types.KindUnknown))
			return
		}
		operandExpr, ok := operandVal.(ast.Expression)
		if !ok {
			l.addError(ctx, "typeof operand is not ast.Expression (got %T) at %v", operandVal, tokenToPosition(token))
			l.push(newNode(&ast.ErrorNode{Message: fmt.Sprintf("typeof operand was %T", operandVal)}, token, types.KindUnknown))
			return
		}
		node := &ast.TypeOfNode{Argument: operandExpr}
		l.push(newNode(node, token, types.KindTypeOfExpr))
		return
	}

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
		return
	}

	token := tok.GetSymbol()
	operandRaw, ok := l.pop()
	if !ok {
		l.addError(ctx, "Stack error: missing operand for unary %q", op)
		l.push(newNode(&ast.ErrorNode{Message: "stack underflow (unary)"}, token, types.KindUnknown))
		return
	}
	operandExpr, ok := operandRaw.(ast.Expression)
	if !ok {
		l.addError(ctx, "Operand for unary %q is not ast.Expression (got %T)", op, operandRaw)
		l.push(newNode(&ast.ErrorNode{Message: "type error (unary)"}, token, types.KindUnknown))
		return
	}
	node := &ast.UnaryOpNode{Operator: op, Operand: operandExpr}
	l.push(newNode(node, token, types.KindUnaryOp))
	l.logDebugAST("      Constructed ast.UnaryOpNode: %s [%T]", op, operandExpr)
}

// ExitPower_expr
func (l *neuroScriptListenerImpl) ExitPower_expr(ctx *gen.Power_exprContext) {
	l.logDebugAST("--- Exit Power_expr: %q", ctx.GetText())
	opTokenNode := ctx.STAR_STAR()
	if opTokenNode == nil {
		return
	}
	opSymbol := opTokenNode.GetSymbol()
	opText := opSymbol.GetText()

	exponentRaw, ok := l.pop()
	if !ok {
		l.addError(ctx, "Stack error popping exponent for POWER")
		l.push(newNode(&ast.ErrorNode{Message: "Stack error (power exponent)"}, opSymbol, types.KindUnknown))
		return
	}
	exponentExpr, isExpr := exponentRaw.(ast.Expression)
	if !isExpr {
		l.addError(ctx, "Exponent for POWER is not an ast.Expression (type %T)", exponentRaw)
		l.push(newNode(&ast.ErrorNode{Message: "Type error (power exponent)"}, opSymbol, types.KindUnknown))
		return
	}
	baseRaw, ok := l.pop()
	if !ok {
		l.addError(ctx, "Stack error popping base for POWER")
		l.push(newNode(&ast.ErrorNode{Message: "Stack error (power base)"}, opSymbol, types.KindUnknown))
		return
	}
	baseExpr, isExpr := baseRaw.(ast.Expression)
	if !isExpr {
		l.addError(ctx, "Base for POWER is not an ast.Expression (type %T)", baseRaw)
		l.push(newNode(&ast.ErrorNode{Message: "Type error (power base)"}, opSymbol, types.KindUnknown))
		return
	}
	node := &ast.BinaryOpNode{Left: baseExpr, Operator: opText, Right: exponentExpr}
	l.push(newNode(node, opSymbol, types.KindBinaryOp))
	l.logDebugAST("      Constructed ast.BinaryOpNode (Power): [%T %s %T]", baseExpr, opText, exponentExpr)
}

// EnterAccessor_expr
func (l *neuroScriptListenerImpl) EnterAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Enter Accessor_expr: %q", ctx.GetText())
}

// ExitAccessor_expr
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Exit Accessor_expr: %q", ctx.GetText())
	numAccessors := len(ctx.AllLBRACK())
	if numAccessors == 0 {
		return
	}

	accessorExprs := make([]ast.Expression, numAccessors)
	for i := numAccessors - 1; i >= 0; i-- {
		accessorRaw, ok := l.pop()
		if !ok {
			l.addError(ctx, "Stack error popping accessor expression %d", i)
			l.push(newNode(&ast.ErrorNode{Message: "Stack error (accessor expr)"}, ctx.LBRACK(i).GetSymbol(), types.KindUnknown))
			return
		}
		accessorExpr, isExpr := accessorRaw.(ast.Expression)
		if !isExpr {
			l.addError(ctx, "Accessor expression %d is not an ast.Expression (type %T)", i, accessorRaw)
			l.push(newNode(&ast.ErrorNode{Message: "Type error (accessor expr)"}, ctx.LBRACK(i).GetSymbol(), types.KindUnknown))
			return
		}
		accessorExprs[i] = accessorExpr
	}

	collectionRaw, ok := l.pop()
	if !ok {
		l.addError(ctx.Primary(), "Stack error popping primary collection")
		l.push(newNode(&ast.ErrorNode{Message: "Stack error (accessor collection)"}, ctx.Primary().GetStart(), types.KindUnknown))
		return
	}
	collectionExpr, isExpr := collectionRaw.(ast.Expression)
	if !isExpr {
		l.addError(ctx.Primary(), "Primary collection is not an ast.Expression (type %T)", collectionRaw)
		l.push(newNode(&ast.ErrorNode{Message: "Type error (accessor collection)"}, ctx.Primary().GetStart(), types.KindUnknown))
		return
	}

	currentCollectionResult := collectionExpr
	for i := 0; i < numAccessors; i++ {
		node := &ast.ElementAccessNode{
			Collection: currentCollectionResult,
			Accessor:   accessorExprs[i],
		}
		l.logDebugAST("      Constructed ast.ElementAccessNode: [Coll: %T Acc: %T]", node.Collection, node.Accessor)
		currentCollectionResult = newNode(node, ctx.LBRACK(i).GetSymbol(), types.KindElementAccess)
	}
	l.push(currentCollectionResult)
	l.logDebugAST("      Final Accessor_expr result pushed: %T", currentCollectionResult)
}

// buildCallTargetFromContext constructs a ast.CallTarget AST node from an ICall_targetContext.
func (l *neuroScriptListenerImpl) buildCallTargetFromContext(ctx gen.ICall_targetContext) *ast.CallTarget {
	l.logDebugAST("      -> buildCallTargetFromContext: %s", ctx.GetText())
	target := &ast.CallTarget{}

	if toolKeyword := ctx.KW_TOOL(); toolKeyword != nil {
		target.IsTool = true
		token := toolKeyword.GetSymbol()
		if qiCtx := ctx.Qualified_identifier(); qiCtx != nil {
			// FIX: Get the full text of the qualified identifier.
			var parts []string
			for _, idNode := range qiCtx.AllIDENTIFIER() {
				parts = append(parts, idNode.GetText())
			}
			target.Name = strings.Join(parts, ".")
			token = qiCtx.GetStart()
		} else {
			l.addError(ctx, "Tool call: Expected Qualified_identifier, but was not found: %s", ctx.GetText())
			target.Name = "<ERROR_NO_QUALIFIED_ID_FOR_TOOL>"
		}
		newNode(target, token, types.KindCallableExpr)
		l.logDebugAST("         Tool call identified. Name: '%s'", target.Name)
	} else if userFuncID := ctx.IDENTIFIER(); userFuncID != nil {
		target.IsTool = false
		target.Name = userFuncID.GetText()
		newNode(target, userFuncID.GetSymbol(), types.KindCallableExpr)
		l.logDebugAST("         User function call identified. Name: '%s'", target.Name)
	} else {
		l.addError(ctx, "Unrecognized call_target structure: %s", ctx.GetText())
		target.Name = "<ERROR_INVALID_CALL_TARGET>"
		newNode(target, ctx.GetStart(), types.KindUnknown)
	}
	l.logDebugAST("      <- buildCallTargetFromContext (Name: %s, IsTool: %v)", target.Name, target.IsTool)
	return target
}

// ExitCall_target is called when exiting the call_target rule.
func (l *neuroScriptListenerImpl) ExitCall_target(ctx *gen.Call_targetContext) {
	l.logDebugAST("--- Exit Call_target: %q", ctx.GetText())
	targetNode := l.buildCallTargetFromContext(ctx)
	l.push(targetNode) // Pushing the pointer
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
					l.addError(ctx, "Stack error popping arguments for call %q", ctx.GetText())
					l.push(newNode(&ast.ErrorNode{Message: "Stack error (call args)"}, ctx.GetStart(), types.KindUnknown))
					return
				}
				args = make([]ast.Expression, numArgs)
				for i := 0; i < numArgs; i++ {
					argExpr, isExpr := argsRaw[i].(ast.Expression)
					if !isExpr {
						l.addError(ctx, "Argument %d for call %q is not an ast.Expression (type %T)", i+1, ctx.GetText(), argsRaw[i])
						l.push(newNode(&ast.ErrorNode{Message: "Type error (call arg)"}, ctx.GetStart(), types.KindUnknown))
						return
					}
					args[i] = argExpr
				}
			}
		}
	}

	var finalTargetNode ast.CallTarget
	var token antlr.Token

	if targetRuleCtx := ctx.Call_target(); targetRuleCtx != nil {
		targetVal, ok := l.pop()
		if !ok {
			l.addError(ctx, "Stack error popping call target for %q", ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: "Stack error (call target)"}, ctx.GetStart(), types.KindUnknown))
			return
		}
		targetPtr, isPtr := targetVal.(*ast.CallTarget)
		if !isPtr {
			l.addError(ctx, "Popped call target is not *ast.CallTarget (type %T) for %q", targetVal, ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: "Type error (call target)"}, ctx.GetStart(), types.KindUnknown))
			return
		}
		finalTargetNode = *targetPtr
		token = ctx.GetStart()
		l.logDebugAST("      Popped *ast.CallTarget from stack: IsTool=%t, Name=%s", finalTargetNode.IsTool, finalTargetNode.Name)
	} else {
		finalTargetNode.IsTool = false
		switch {
		// FIX: Add case for KW_LEN to correctly handle the len() function call.
		case ctx.KW_LEN() != nil:
			token = ctx.KW_LEN().GetSymbol()
			finalTargetNode.Name = "len"
		case ctx.KW_LN() != nil:
			token = ctx.KW_LN().GetSymbol()
			finalTargetNode.Name = "ln"
		case ctx.KW_LOG() != nil:
			token = ctx.KW_LOG().GetSymbol()
			finalTargetNode.Name = "log"
		// ... other cases
		default:
			l.addError(ctx, "Unhandled built-in or target type in Callable_expr: %q", ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: "Unknown callable target"}, ctx.GetStart(), types.KindUnknown))
			return
		}
		l.logDebugAST("      Identified Built-in function call target: %s", finalTargetNode.Name)
	}

	node := &ast.CallableExprNode{
		Target:    finalTargetNode,
		Arguments: args,
	}
	l.push(newNode(node, token, types.KindCallableExpr))
	l.logDebugAST("      Constructed and Pushed ast.CallableExprNode: Target=%s, Args=%d", node.Target.Name, len(node.Arguments))
}

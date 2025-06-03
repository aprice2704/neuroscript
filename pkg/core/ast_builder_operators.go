// NeuroScript Version: 0.3.1
// File version: 0.0.13 // Corrected typeof handling within ExitUnary_expr, removed ExitTypeOfExpression, added isErrorNode.
// Purpose: AST building logic for operators.
// filename: pkg/core/ast_builder_operators.go
// nlines: 372
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4" // Using user-specified ANTLR import path
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// Helper function to check if a value is an ErrorNode
func isErrorNode(val interface{}) bool {
	_, ok := val.(*ErrorNode)
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

	poppedOperands := make([]Expression, numOperands)
	for i := 0; i < numOperands; i++ {
		val, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Stack error popping operand %d for binary op: %s", numOperands-i, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: fmt.Sprintf("Stack error (binary op operand %d)", numOperands-i)})
			return
		}
		expr, isExpr := val.(Expression)
		if !isExpr {
			errPosToken := ctx.GetStart()
			l.addError(ctx, "Operand %d is not an Expression (type %T) for binary op: %s", numOperands-i, val, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(errPosToken), Message: fmt.Sprintf("Type error (binary op operand %d)", numOperands-i)})
			return
		}
		poppedOperands[i] = expr
	}

	currentLHS := poppedOperands[numOperands-1]

	for i := 0; i < numOperators; i++ {
		opToken := opGetter(i)
		if opToken == nil {
			l.addError(ctx, "Could not find operator token for index %d in: %s", i, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Missing operator token"})
			return
		}
		opSymbol := opToken.GetSymbol()
		opText := opSymbol.GetText()
		currentRHS := poppedOperands[numOperands-2-i]

		newNode := &BinaryOpNode{
			Pos:      tokenToPosition(opSymbol),
			Left:     currentLHS,
			Operator: opText,
			Right:    currentRHS,
		}
		l.logDebugAST("    Constructed BinaryOpNode: [%T %s %T]", currentLHS, opText, currentRHS)
		currentLHS = newNode
	}
	l.pushValue(currentLHS)
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

// ExitUnary_expr handles all unary operators, including 'typeof'.
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- ExitUnary_expr: %q", ctx.GetText())

	if ctx.KW_TYPEOF() != nil {
		// Handle 'typeof' operator
		l.logDebugAST("    Unary_expr is a 'typeof' expression.")
		operandVal, ok := l.popValue()
		if !ok {
			startPos := tokenToPosition(ctx.KW_TYPEOF().GetSymbol())
			l.addError(ctx, "Stack error: Could not pop operand for typeof operator at %s", startPos.String())
			l.pushValue(&ErrorNode{Pos: startPos, Message: "Missing operand for typeof"})
			return
		}

		operandExpr, isExpr := operandVal.(Expression)
		if !isExpr {
			startPos := tokenToPosition(ctx.KW_TYPEOF().GetSymbol())
			l.addError(ctx, "Internal AST build error: operand for typeof is not an Expression (got %T) at %s", operandVal, startPos.String())
			l.pushValue(&ErrorNode{Pos: startPos, Message: fmt.Sprintf("typeof operand was %T, expected Expression", operandVal)})
			return
		}

		node := &TypeOfNode{
			Pos:      tokenToPosition(ctx.KW_TYPEOF().GetSymbol()),
			Argument: operandExpr,
		}
		l.logDebugAST("    Constructed TypeOfNode with Argument Type: %T, Pos: %s", operandExpr, node.Pos.String())
		l.pushValue(node)
		return // Processed 'typeof', so return
	}

	// Handle other unary operators (MINUS, NOT, NO, SOME, TILDE)
	var opTokenNode antlr.TerminalNode
	var opText string

	if ctx.MINUS() != nil {
		opTokenNode = ctx.MINUS()
		opText = "-"
	} else if ctx.KW_NOT() != nil {
		opTokenNode = ctx.KW_NOT()
		opText = "not"
	} else if ctx.KW_NO() != nil {
		opTokenNode = ctx.KW_NO()
		opText = "no"
	} else if ctx.KW_SOME() != nil {
		opTokenNode = ctx.KW_SOME()
		opText = "some"
	} else if ctx.TILDE() != nil {
		opTokenNode = ctx.TILDE()
		opText = "~"
	}

	if opTokenNode == nil {
		// This means it's a pass-through from power_expr (or another non-operator alternative within unary_expr's ANTLR rule)
		l.logDebugAST("    Unary_expr is pass-through (no specific operator token found, or was typeof).")
		return
	}

	operandRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping operand for unary op %q", opText)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opTokenNode.GetSymbol()), Message: "Stack error (unary op)"})
		return
	}
	operandExpr, isExpr := operandRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Operand for unary op %q is not an Expression (type %T)", opText, operandRaw)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opTokenNode.GetSymbol()), Message: "Type error (unary op)"})
		return
	}
	node := &UnaryOpNode{
		Pos:      tokenToPosition(opTokenNode.GetSymbol()),
		Operator: opText,
		Operand:  operandExpr,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed UnaryOpNode: %s [%T]", opText, operandExpr)
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

	exponentRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping exponent for POWER")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Stack error (power exponent)"})
		return
	}
	exponentExpr, isExpr := exponentRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Exponent for POWER is not an Expression (type %T)", exponentRaw)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Type error (power exponent)"})
		return
	}
	baseRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping base for POWER")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Stack error (power base)"})
		return
	}
	baseExpr, isExpr := baseRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Base for POWER is not an Expression (type %T)", baseRaw)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Type error (power base)"})
		return
	}
	node := &BinaryOpNode{
		Pos:      tokenToPosition(opSymbol),
		Left:     baseExpr,
		Operator: opText,
		Right:    exponentExpr,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed BinaryOpNode (Power): [%T %s %T]", baseExpr, opText, exponentExpr)
}

// ExitAccessor_expr
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Exit Accessor_expr: %q", ctx.GetText())
	numAccessors := len(ctx.AllLBRACK())
	if numAccessors == 0 {
		// Pass through from primary
		return
	}

	accessorExprs := make([]Expression, numAccessors)
	for i := numAccessors - 1; i >= 0; i-- {
		accessorRaw, ok := l.popValue()
		if !ok {
			if i < len(ctx.AllExpression()) {
				l.addError(ctx.Expression(i), "Stack error popping accessor expression %d", i)
			} else {
				l.addError(ctx, "Stack error popping accessor expression %d (index out of bounds)", i)
			}
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.LBRACK(i).GetSymbol()), Message: "Stack error (accessor expr)"})
			return
		}
		accessorExpr, isExpr := accessorRaw.(Expression)
		if !isExpr {
			if i < len(ctx.AllExpression()) {
				l.addError(ctx.Expression(i), "Accessor expression %d is not an Expression (type %T)", i, accessorRaw)
			} else {
				l.addError(ctx, "Accessor expression %d is not an Expression (type %T) (index out of bounds for error pos)", i, accessorRaw)
			}
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.LBRACK(i).GetSymbol()), Message: "Type error (accessor expr)"})
			return
		}
		accessorExprs[i] = accessorExpr
	}

	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx.Primary(), "Stack error popping primary collection")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.Primary().GetStart()), Message: "Stack error (accessor collection)"})
		return
	}
	collectionExpr, isExpr := collectionRaw.(Expression)
	if !isExpr {
		l.addError(ctx.Primary(), "Primary collection is not an Expression (type %T)", collectionRaw)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.Primary().GetStart()), Message: "Type error (accessor collection)"})
		return
	}

	currentCollectionResult := collectionExpr
	for i := 0; i < numAccessors; i++ {
		newNode := &ElementAccessNode{
			Pos:        tokenToPosition(ctx.LBRACK(i).GetSymbol()),
			Collection: currentCollectionResult,
			Accessor:   accessorExprs[i],
		}
		l.logDebugAST("    Constructed ElementAccessNode: [Coll: %T Acc: %T]", newNode.Collection, newNode.Accessor)
		currentCollectionResult = newNode
	}
	l.pushValue(currentCollectionResult)
	l.logDebugAST("    Final Accessor_expr result pushed: %T", currentCollectionResult)
}

// buildCallTargetFromContext constructs a CallTarget AST node from an ICall_targetContext.
func (l *neuroScriptListenerImpl) buildCallTargetFromContext(ctx gen.ICall_targetContext) CallTarget {
	l.logDebugAST("    -> buildCallTargetFromContext: %s", ctx.GetText())
	target := CallTarget{}

	if toolKeyword := ctx.KW_TOOL(); toolKeyword != nil {
		target.IsTool = true
		if qiCtx := ctx.Qualified_identifier(); qiCtx != nil {
			idNodes := qiCtx.AllIDENTIFIER()
			var parts []string
			for _, idNode := range idNodes {
				parts = append(parts, idNode.GetText())
			}
			target.Name = strings.Join(parts, ".")

			if len(idNodes) > 0 {
				target.Pos = tokenToPosition(idNodes[0].GetSymbol())
			} else {
				target.Pos = tokenToPosition(toolKeyword.GetSymbol())
				l.addError(ctx, "Tool call has empty qualified_identifier: %s", ctx.GetText())
			}
		} else {
			l.addError(ctx, "Tool call: Expected Qualified_identifier, but was not found: %s", ctx.GetText())
			target.Name = "<ERROR_NO_QUALIFIED_ID_FOR_TOOL>"
			target.Pos = tokenToPosition(toolKeyword.GetSymbol())
		}
		l.logDebugAST("       Tool call identified. Name: '%s', Pos: %s", target.Name, target.Pos.String())
	} else if userFuncID := ctx.IDENTIFIER(); userFuncID != nil {
		target.IsTool = false
		target.Name = userFuncID.GetText()
		target.Pos = tokenToPosition(userFuncID.GetSymbol())
		l.logDebugAST("       User function call identified. Name: '%s', Pos: %s", target.Name, target.Pos.String())
	} else {
		l.addError(ctx, "Unrecognized call_target structure: %s", ctx.GetText())
		target.Name = "<ERROR_INVALID_CALL_TARGET>"
		target.Pos = tokenToPosition(ctx.GetStart())
	}
	l.logDebugAST("    <- buildCallTargetFromContext (Name: %s, IsTool: %v)", target.Name, target.IsTool)
	return target
}

// ExitCall_target is called when exiting the call_target rule.
func (l *neuroScriptListenerImpl) ExitCall_target(ctx *gen.Call_targetContext) {
	l.logDebugAST("--- Exit Call_target: %q", ctx.GetText())
	targetNode := l.buildCallTargetFromContext(ctx)
	l.pushValue(&targetNode)
	l.logDebugAST("    Pushed *CallTarget to stack: IsTool=%t, Name=%s", targetNode.IsTool, targetNode.Name)
}

// ExitCallable_expr
func (l *neuroScriptListenerImpl) ExitCallable_expr(ctx *gen.Callable_exprContext) {
	l.logDebugAST("--- Exit Callable_expr: %q", ctx.GetText())

	numArgs := 0
	argExpressions := []Expression{}
	if exprListOptCtx := ctx.Expression_list_opt(); exprListOptCtx != nil {
		if exprListCtx := exprListOptCtx.Expression_list(); exprListCtx != nil {
			numArgs = len(exprListCtx.AllExpression())
		}
	}

	if numArgs > 0 {
		argsRaw, ok := l.popNValues(numArgs)
		if !ok {
			l.addError(ctx, "Stack error popping arguments for call %q", ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Stack error (call args)"})
			return
		}
		argExpressions = make([]Expression, numArgs)
		for i := 0; i < numArgs; i++ {
			argExpr, isExpr := argsRaw[i].(Expression)
			if !isExpr {
				argSourceCtx := ctx.Expression_list_opt().Expression_list().Expression(i)
				l.addError(argSourceCtx, "Argument %d for call %q is not an Expression (type %T)", i+1, ctx.GetText(), argsRaw[i])
				l.pushValue(&ErrorNode{Pos: tokenToPosition(argSourceCtx.GetStart()), Message: "Type error (call arg)"})
				return
			}
			argExpressions[i] = argExpr
		}
	}

	var finalTargetNode CallTarget
	var callExprPos *Position

	if targetRuleCtx := ctx.Call_target(); targetRuleCtx != nil {
		targetVal, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Stack error popping call target for %q", ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Stack error (call target)"})
			return
		}
		targetPtr, isPtr := targetVal.(*CallTarget)
		if !isPtr {
			l.addError(ctx, "Popped call target is not *CallTarget (type %T) for %q", targetVal, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Type error (call target)"})
			return
		}
		finalTargetNode = *targetPtr
		callExprPos = finalTargetNode.Pos
		l.logDebugAST("    Popped *CallTarget from stack: IsTool=%t, Name=%s", finalTargetNode.IsTool, finalTargetNode.Name)
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
		default:
			l.addError(ctx, "Unhandled built-in or target type in Callable_expr: %q", ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Unknown callable target"})
			return
		}
		callExprPos = tokenToPosition(keywordToken.GetSymbol())
		finalTargetNode.Pos = callExprPos
		l.logDebugAST("    Identified Built-in function call target: %s", finalTargetNode.Name)
	}

	node := &CallableExprNode{
		Pos:       callExprPos,
		Target:    finalTargetNode,
		Arguments: argExpressions,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed and Pushed CallableExprNode: Target=%s, Args=%d", node.Target.Name, len(node.Arguments))
}

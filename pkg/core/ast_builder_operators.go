// NeuroScript Version: 0.3.1
// File version: 0.0.12 // Changed TypeOfExpressionContext to Unary_exprContext due to undefined error.
// Purpose: AST building logic for operators.
// filename: pkg/core/ast_builder_operators.go
// nlines: 405
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4" // Using user-specified ANTLR import path
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

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

// ExitUnary_expr handles unary operators other than 'typeof'.
// 'typeof' is handled by ExitTypeOfExpression due to its labeled alternative in the grammar.
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- ExitUnary_expr (General Handler): %q", ctx.GetText())

	// If KW_TYPEOF is found here, it implies ExitTypeOfExpression was not called by ANTLR,
	// which would be unexpected. The primary logic for typeof is in ExitTypeOfExpression.
	if ctx.KW_TYPEOF() != nil {
		// This condition specifically checks if the Unary_exprContext *itself* is the TypeOfExpression alternative.
		// If ExitTypeOfExpression is correctly dispatched by ANTLR's walker, this block in ExitUnary_expr
		// should ideally not be hit for 'typeof' cases *if* ExitTypeOfExpression handles it.
		// However, ANTLR might call ExitUnary_expr for the overall unary_expr rule even if a labeled alternative like ExitTypeOfExpression was also called.
		// If KW_TYPEOF() is non-nil, it means this specific unary_expr *is* a typeof.
		// We rely on ExitTypeOfExpression to have handled pushing the TypeOfNode.
		// So, if we are here and it's a typeof, we should just return to avoid double processing or interfering with stack.
		l.logDebugAST("    Unary_expr is a 'typeof' expression; assuming ExitTypeOfExpression handled it.")
		return
	}

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
		// This means it's a pass-through from power_expr (or another non-operator alternative)
		l.logDebugAST("    Unary_expr is pass-through (no specific operator token found).")
		return
	}

	// If we have an opTokenNode (MINUS, NOT, etc.), pop its operand and build UnaryOpNode
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

// ExitTypeOfExpression is called when exiting the TypeOfExpression alternative of unary_expr
// (grammar: unary_expr: ... | KW_TYPEOF unary_expr #TypeOfExpression)
// Changed ctx type from *gen.TypeOfExpressionContext to *gen.Unary_exprContext
func (l *neuroScriptListenerImpl) ExitTypeOfExpression(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- ExitTypeOfExpression: %q", ctx.GetText())

	kwTypeofToken := ctx.KW_TYPEOF() // Method on *gen.Unary_exprContext (should exist if this is a typeof alternative)
	if kwTypeofToken == nil {
		// This should ideally not happen if ANTLR dispatches to ExitTypeOfExpression only for the correct alternative.
		l.addError(ctx, "Internal error: KW_TYPEOF token missing in ExitTypeOfExpression context (ctx is Unary_exprContext)")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Missing KW_TYPEOF in TypeOfExpression"})
		return
	}

	// The 'unary_expr' child (operand of typeof) would have been visited,
	// and its AST node should be on top of the value stack.
	operandRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping operand for typeof operator")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(kwTypeofToken.GetSymbol()), Message: "Stack error (typeof operand)"})
		return
	}
	operandExpr, isExpr := operandRaw.(Expression)
	if !isExpr {
		errPos := tokenToPosition(kwTypeofToken.GetSymbol()) // Default to typeof keyword position
		// Attempt to get the specific child unary_expr context for better error positioning.
		// Unary_exprContext should provide access to its children.
		// For the alternative 'KW_TYPEOF unary_expr', the children are KW_TYPEOF and a unary_expr.
		// Accessing the child unary_expr context can be done via ctx.Unary_expr(i) or ctx.AllUnary_expr().
		// Assuming the operand unary_expr is the first (or only) such child in this alternative's view.
		// This is heuristic; true child context requires knowing ANTLR's generated accessors for Unary_exprContext.
		// For now, this part of error position refinement remains complex with a generic context.
		// The primary source of position for a type error on the operand should ideally come from the operand itself if possible.
		// The original code had: if childExprRuleCtx := ctx.Expression(); ... which was incorrect as child is unary_expr.
		// If `ctx.Unary_expr(0)` (or similar) is available and refers to the operand:
		// childUnaryExprCtx := ctx.Unary_expr(0) // This is an example, actual accessor may vary
		// if childUnaryExprCtx != nil && childUnaryExprCtx.GetStart() != nil {
		// 	errPos = tokenToPosition(childUnaryExprCtx.GetStart())
		// }
		l.addError(ctx, "Operand for typeof is not an Expression (type %T)", operandRaw)
		l.pushValue(&ErrorNode{Pos: errPos, Message: "Type error (typeof operand)"})
		return
	}

	node := &TypeOfNode{
		Pos:      tokenToPosition(kwTypeofToken.GetSymbol()),
		Argument: operandExpr,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed TypeOfNode for argument: %T", operandExpr)
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

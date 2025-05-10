// filename: pkg/core/ast_builder_operators.go
// NeuroScript Version: 0.3.1 (AST Builder component)
// File version: 0.0.6 // Align with corrected ast.go, use ErrorNode, verified qualified_identifier access
// Last Modified: 2025-05-09

package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Helper and Operator Exit methods ---

// processBinaryOperators (existing helper - adapted for ErrorNode and clearer error pos)
func (l *neuroScriptListenerImpl) processBinaryOperators(ctx antlr.ParserRuleContext, numOperands int, opGetter func(i int) antlr.TerminalNode) {
	if numOperands <= 1 {
		// Single operand, already on stack, pass through.
		return
	}

	numOperators := numOperands - 1
	if numOperators < 1 {
		l.addError(ctx, "Internal error: processBinaryOperators with numOperands=%d implies no operators.", numOperands)
		// If numOperands was 1, it's handled. If >1 but numOperators < 1, it's an issue.
		// We expect one value on stack from the single operand.
		return
	}

	// Operands are pushed L, M, R. Stack top is R, then M, then L for L op M op R.
	// We pop R, then M, then L.
	// Build tree: ( (L op M) op R ) for left-associativity.
	// Pop order: Rightmost operand first.

	// Pop all operands. They will be in reverse parsed order.
	// Example: L op1 M op2 R. Stack has [R_expr, M_expr, L_expr] (top is R_expr)
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

			// Determine a relevant token for error position
			var errPosToken antlr.Token
			// Try to get the token from the specific child that was the operand, if possible
			// This requires knowing which child index corresponds to the operand at (numOperands-1-i)
			// For simplicity and robustness, using ctx.GetStart() is a safe fallback if specific token is hard to get.
			// If you have a more direct way to get the context of the specific operand, use that.
			// For now, let's use the start of the whole binary operation context as a general error position.
			errPosToken = ctx.GetStart() // Use the start of the current binary expression context

			// If you want to try to be more specific (this can be tricky depending on grammar structure):
			// childIndex := numOperands - 1 - i // This is an attempt to map stack order to child order
			// if childIndex < ctx.GetChildCount() {
			// 	if operandCtx, ok := ctx.GetChild(childIndex).(antlr.ParserRuleContext); ok {
			// 		errPosToken = operandCtx.GetStart()
			// 	}
			// }

			l.addError(ctx, "Operand %d is not an Expression (type %T) for binary op: %s", numOperands-i, val, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(errPosToken), Message: fmt.Sprintf("Type error (binary op operand %d)", numOperands-i)})
			return

			// Determine a relevant token for error position
			// var errPosToken antlr.Token
			// if ruleCtx, ok := ctx.(antlr.RuleContext); ok && ruleCtx.GetChildCount() > (numOperands-1-i) { // Simplistic: try to get one of the operand rule contexts
			// 	if termNode, ok := ruleCtx.GetChild(numOperands - 1 - i).(antlr.RuleContext); ok { // This index might not be right
			// 		errPosToken = termNode.GetStart()
			// 	} else {
			// 		errPosToken = ctx.GetStart()
			// 	}
			// } else {
			// 	errPosToken = ctx.GetStart()
			// }
			// l.addError(ctx, "Operand %d is not an Expression (type %T) for binary op: %s", numOperands-i, val, ctx.GetText())
			// l.pushValue(&ErrorNode{Pos: tokenToPosition(errPosToken), Message: fmt.Sprintf("Type error (binary op operand %d)", numOperands-i)})
			// return
		}
		poppedOperands[i] = expr // poppedOperands[0] is R, poppedOperands[1] is M, poppedOperands[numOperands-1] is L
	}

	// Now build left-associative: ((L op M) op R)
	// Leftmost actual operand is at poppedOperands[numOperands-1]
	currentLHS := poppedOperands[numOperands-1]

	for i := 0; i < numOperators; i++ {
		// opSymbols from opGetter are typically in parsed order (0th op is leftmost)
		opToken := opGetter(i)
		if opToken == nil {
			l.addError(ctx, "Could not find operator token for index %d in: %s", i, ctx.GetText())
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Missing operator token"})
			return
		}
		opSymbol := opToken.GetSymbol()
		opText := opSymbol.GetText()

		// Next RHS operand is poppedOperands[numOperands-2-i]
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

// ExitUnary_expr
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- Exit Unary_expr: %q", ctx.GetText())
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
		// Pass through from power_expr
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
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opTokenNode.GetSymbol()), Message: "Type error (unary op)"}) // Use operator pos
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
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Type error (power exponent)"}) // Use operator pos
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
		l.pushValue(&ErrorNode{Pos: tokenToPosition(opSymbol), Message: "Type error (power base)"}) // Use operator pos
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
			l.addError(ctx.Expression(i), "Stack error popping accessor expression %d", i)
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.LBRACK(i).GetSymbol()), Message: "Stack error (accessor expr)"})
			return
		}
		accessorExpr, isExpr := accessorRaw.(Expression)
		if !isExpr {
			l.addError(ctx.Expression(i), "Accessor expression %d is not an Expression (type %T)", i, accessorRaw)
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
// This is the primary location for handling the new qualified_identifier rule.
func (l *neuroScriptListenerImpl) buildCallTargetFromContext(ctx gen.ICall_targetContext) CallTarget {
	l.logDebugAST("    -> buildCallTargetFromContext: %s", ctx.GetText())
	target := CallTarget{}

	if toolKeyword := ctx.KW_TOOL(); toolKeyword != nil {
		target.IsTool = true
		// Use the Qualified_identifier rule from the context.
		// The actual method names (Qualified_identifier, AllIDENTIFIER) come from your *regenerated* parser.
		if qiCtx := ctx.Qualified_identifier(); qiCtx != nil { // This is IQualified_identifierContext
			idNodes := qiCtx.AllIDENTIFIER() // This returns []antlr.TerminalNode
			var parts []string
			for _, idNode := range idNodes {
				parts = append(parts, idNode.GetText())
			}
			target.Name = strings.Join(parts, ".")

			if len(idNodes) > 0 {
				target.Pos = tokenToPosition(idNodes[0].GetSymbol())
			} else { // Should be caught by grammar if qualified_identifier needs at least one ID
				target.Pos = tokenToPosition(toolKeyword.GetSymbol())
				l.addError(ctx, "Tool call has empty qualified_identifier: %s", ctx.GetText())
			}
		} else {
			// This block might be hit if the grammar has an alternative path for `tool.IDENTIFIER`
			// or if there's an issue with parser regeneration making Qualified_identifier optional.
			// Assuming the new grammar `tool DOT qualified_identifier` is strict, this else
			// indicates a problem or an unexpected parse path.
			// The v0.3.7 parser file has `KW_TOOL DOT IDENTIFIER`. If you used *that* to regenerate,
			// this path is expected. If you used the *new* grammar, qiCtx should not be nil.

			// Let's assume the user *has* regenerated with `qualified_identifier`.
			// So, `qiCtx` not being `nil` is the expected path.
			// If `qiCtx` *is* nil here with the new grammar, it implies a parsing issue
			// or that the grammar structure is different from `call_target: KW_TOOL DOT qualified_identifier;`
			l.addError(ctx, "Tool call: Expected Qualified_identifier, but was not found: %s", ctx.GetText())
			target.Name = "<ERROR_NO_QUALIFIED_ID_FOR_TOOL>"
			target.Pos = tokenToPosition(toolKeyword.GetSymbol())
		}
		l.logDebugAST("       Tool call identified. Name: '%s', Pos: %s", target.Name, target.Pos.String())
	} else if userFuncID := ctx.IDENTIFIER(); userFuncID != nil { // For user-defined functions
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
// It constructs a *CallTarget and pushes it to the stack.
func (l *neuroScriptListenerImpl) ExitCall_target(ctx *gen.Call_targetContext) {
	l.logDebugAST("--- Exit Call_target: %q", ctx.GetText())
	targetNode := l.buildCallTargetFromContext(ctx) // ctx is already *gen.Call_targetContext
	l.pushValue(&targetNode)                        // Push a pointer
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
			argExpr, isExpr := argsRaw[numArgs-1-i].(Expression) // Corrected order for popNValues
			if !isExpr {
				argSourceCtx := ctx.Expression_list_opt().Expression_list().Expression(i)
				l.addError(argSourceCtx, "Argument %d for call %q is not an Expression (type %T)", i+1, ctx.GetText(), argsRaw[numArgs-1-i])
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
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Type error (call target)"}) // Use general pos
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

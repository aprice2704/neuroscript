// filename: pkg/core/ast_builder_operators.go
package core

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Helper and Operator Exit methods ---

// Helper to get the specific operator token at a given child index, checking multiple types
func getOperatorToken(ctx antlr.ParserRuleContext, index int, tokenTypes ...int) antlr.TerminalNode {
	if index < 0 || index >= ctx.GetChildCount() {
		return nil
	}
	child := ctx.GetChild(index)
	opNode, ok := child.(antlr.TerminalNode)
	if !ok {
		return nil // Child is not a terminal node
	}
	actualType := opNode.GetSymbol().GetTokenType()
	for _, expectedType := range tokenTypes {
		if actualType == expectedType {
			return opNode
		}
	}
	return nil // Node is terminal, but not one of the expected operator types
}

// processBinaryOperators handles the common logic for left-associative binary operators
func (l *neuroScriptListenerImpl) processBinaryOperators(ctx antlr.ParserRuleContext, numOperands int, opGetter func(i int) antlr.TerminalNode) {
	if numOperands <= 1 {
		// Only one operand, pass through value pushed by child
		return
	}

	// Stack: [Left, Right1, Right2, ..., RightN] where N = numOperands - 1
	numOperators := numOperands - 1

	// Pop initial Right operand
	rightRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping initial right operand")
		l.pushValue(nil)
		return
	}
	rightExpr, ok := rightRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Right operand is not an Expression (got %T)", rightRaw)
		l.pushValue(nil)
		return
	}

	// Loop through remaining operands and operators from right to left
	for i := numOperators - 1; i >= 0; i-- {
		// Pop the corresponding Left operand (which might be result of previous op)
		leftRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping left operand for operator %d", i)
			l.pushValue(nil)
			return
		}
		leftExpr, ok := leftRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Left operand is not an Expression (got %T)", leftRaw)
			l.pushValue(nil)
			return
		}

		// Get the specific operator token for this operation
		opToken := opGetter(i) // Assumes opGetter returns the i-th operator token
		if opToken == nil {
			l.addError(ctx, "Internal error: Could not find operator token for index %d", i)
			l.pushValue(nil)
			return
		}
		opSymbol := opToken.GetSymbol()
		opText := opSymbol.GetText()

		// Create the new BinaryOpNode
		newNode := &BinaryOpNode{
			Pos:      tokenToPosition(opSymbol), // Position of the operator token
			Left:     leftExpr,
			Operator: opText,
			Right:    rightExpr, // Use the right operand popped earlier or result from inner loop
		}
		l.logDebugAST("    Constructed BinaryOpNode: [%T %s %T]", leftExpr, opText, rightExpr)

		// The result becomes the 'rightExpr' for the next iteration (moving left)
		rightExpr = newNode
	}

	// Push the final result (root of the binary expression tree for this level)
	l.pushValue(rightExpr)
}

// ExitLogical_or_expr
func (l *neuroScriptListenerImpl) ExitLogical_or_expr(ctx *gen.Logical_or_exprContext) {
	l.logDebugAST("--- Exit Logical_or_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllLogical_and_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.KW_OR(i) } // Gets the i-th OR token
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitLogical_and_expr
func (l *neuroScriptListenerImpl) ExitLogical_and_expr(ctx *gen.Logical_and_exprContext) {
	l.logDebugAST("--- Exit Logical_and_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_or_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.KW_AND(i) } // Gets the i-th AND token
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_or_expr
func (l *neuroScriptListenerImpl) ExitBitwise_or_expr(ctx *gen.Bitwise_or_exprContext) {
	l.logDebugAST("--- Exit Bitwise_or_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_xor_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.PIPE(i) } // Gets the i-th PIPE token
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_xor_expr
func (l *neuroScriptListenerImpl) ExitBitwise_xor_expr(ctx *gen.Bitwise_xor_exprContext) {
	l.logDebugAST("--- Exit Bitwise_xor_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllBitwise_and_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.CARET(i) } // Gets the i-th CARET token
	l.processBinaryOperators(ctx, numOperands, opGetter)
}

// ExitBitwise_and_expr
func (l *neuroScriptListenerImpl) ExitBitwise_and_expr(ctx *gen.Bitwise_and_exprContext) {
	l.logDebugAST("--- Exit Bitwise_and_expr: %q", ctx.GetText())
	numOperands := len(ctx.AllEquality_expr())
	opGetter := func(i int) antlr.TerminalNode { return ctx.AMPERSAND(i) } // Gets the i-th AMPERSAND token
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

// ExitUnary_expr handles unary minus (-), logical NOT, and no/some.
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST("--- Exit Unary_expr: %q", ctx.GetText())
	var opToken antlr.TerminalNode
	var opSymbol antlr.Token
	var opText string

	// Check which unary operator is present
	if ctx.MINUS() != nil {
		opToken = ctx.MINUS()
	} else if ctx.KW_NOT() != nil {
		opToken = ctx.KW_NOT()
	} else if ctx.KW_NO() != nil {
		opToken = ctx.KW_NO()
	} else if ctx.KW_SOME() != nil {
		opToken = ctx.KW_SOME()
	}

	if opToken == nil {
		// No operator, just pass through value from child (power_expr)
		l.logDebugAST("    Unary is just Power_expr (Pass through)")
		// The result of visiting power_expr is already on the stack.
		return
	}

	// Operator found, pop operand
	opSymbol = opToken.GetSymbol()
	opText = opSymbol.GetText()

	operandRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping operand for unary op %q", opText)
		l.pushValue(nil) // Push error marker
		return
	}
	operandExpr, ok := operandRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Operand for unary op %q is not an Expression (got %T)", opText, operandRaw)
		l.pushValue(nil) // Push error marker
		return
	}

	node := &UnaryOpNode{
		Pos:      tokenToPosition(opSymbol), // Position of the operator
		Operator: opText,
		Operand:  operandExpr,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed UnaryOpNode: %s [%T]", opText, operandExpr)
}

// ExitPower_expr (Handles right-associative power operator)
func (l *neuroScriptListenerImpl) ExitPower_expr(ctx *gen.Power_exprContext) {
	l.logDebugAST("--- Exit Power_expr: %q", ctx.GetText())
	opToken := ctx.STAR_STAR()
	if opToken == nil {
		// No power operator, pass through value from child (accessor_expr)
		l.logDebugAST("    Power is just Accessor_expr (Pass through)")
		// The result of visiting accessor_expr is already on the stack.
		return
	}

	opSymbol := opToken.GetSymbol()
	opText := opSymbol.GetText()

	// Pop exponent (right operand first for right-associativity)
	exponentRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping exponent for POWER")
		l.pushValue(nil)
		return
	}
	exponentExpr, ok := exponentRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Exponent for POWER is not an Expression (got %T)", exponentRaw)
		l.pushValue(nil)
		return
	}

	// Pop base (left operand)
	baseRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping base for POWER")
		l.pushValue(nil)
		return
	}
	baseExpr, ok := baseRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Base for POWER is not an Expression (got %T)", baseRaw)
		l.pushValue(nil)
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

// --- ADDED ---
// ExitAccessor_expr handles list/map element access like list[index] or map["key"]
// Grammar: accessor_expr: primary ( LBRACK expression RBRACK )* ;
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST("--- Exit Accessor_expr: %q", ctx.GetText())
	// Get results for all accessor expressions (inside brackets)
	numAccessors := len(ctx.AllExpression())

	if numAccessors == 0 {
		// No brackets, just pass through the primary expression result.
		// The result of visiting Primary() is already on the stack.
		l.logDebugAST("    Accessor is just Primary (Pass through)")
		return
	}

	// Stack should contain: [PrimaryResult, Accessor1Result, Accessor2Result, ...]
	// Pop results in reverse order (last accessor first)
	accessorExprs := make([]Expression, numAccessors)
	for i := numAccessors - 1; i >= 0; i-- {
		accessorRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx.Expression(i), "Internal error: Stack error popping accessor %d", i)
			l.pushValue(nil) // Push error marker
			return
		}
		accessorExpr, ok := accessorRaw.(Expression)
		if !ok {
			l.addError(ctx.Expression(i), "Internal error: Accessor %d is not an Expression (got %T)", i, accessorRaw)
			l.pushValue(nil) // Push error marker
			return
		}
		accessorExprs[i] = accessorExpr
	}

	// Pop the primary collection expression result
	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx.Primary(), "Internal error: Stack error popping primary collection")
		l.pushValue(nil) // Push error marker
		return
	}
	collectionExpr, ok := collectionRaw.(Expression)
	if !ok {
		l.addError(ctx.Primary(), "Internal error: Primary collection is not an Expression (got %T)", collectionRaw)
		l.pushValue(nil) // Push error marker
		return
	}

	// Build nested ElementAccessNodes from left to right
	currentCollectionExpr := collectionExpr
	for i := 0; i < numAccessors; i++ {
		lBracketToken := ctx.LBRACK(i) // Get the '[' token for position
		if lBracketToken == nil {
			l.addError(ctx, "Internal error: Missing LBRACK token for accessor %d", i)
			l.pushValue(nil) // Push error marker
			return
		}

		newNode := &ElementAccessNode{
			Pos:        tokenToPosition(lBracketToken.GetSymbol()), // Position of the '['
			Collection: currentCollectionExpr,
			Accessor:   accessorExprs[i],
		}
		l.logDebugAST("    Constructed ElementAccessNode: [Coll: %T Acc: %T]", newNode.Collection, newNode.Accessor)
		currentCollectionExpr = newNode // The new node becomes the collection for the next access
	}

	// Push the final result (the outermost ElementAccessNode)
	l.pushValue(currentCollectionExpr)
	l.logDebugAST("    Final Accessor_expr result: %T", currentCollectionExpr)
}

// ExitCall_target -- *** CORRECTED PREVIOUSLY ***
func (l *neuroScriptListenerImpl) ExitCall_target(ctx *gen.Call_targetContext) {
	l.logDebugAST("--- Exit Call_target: %q", ctx.GetText())
	var node CallTarget
	var idTokenNode antlr.TerminalNode // The IDENTIFIER node

	idTokenNode = ctx.IDENTIFIER() // Get the single IDENTIFIER token
	if idTokenNode == nil {
		// Grammar should prevent this, but check defensively
		l.addError(ctx, "Internal error: Missing IDENTIFIER in call target")
		l.pushValue(nil)
		return
	}

	// Set name and determine position token
	node.Name = idTokenNode.GetText()
	posToken := idTokenNode.GetSymbol() // Default position is the identifier itself

	// Check if it's a tool call
	if ctx.KW_TOOL() != nil {
		node.IsTool = true
		posToken = ctx.KW_TOOL().GetSymbol() // Position starts at 'tool' keyword if present
		l.logDebugAST("    Identified Tool call target name: %s", node.Name)
	} else {
		node.IsTool = false
		l.logDebugAST("    Identified User Function call target name: %s", node.Name)
	}

	node.Pos = tokenToPosition(posToken)
	l.pushValue(&node) // Push pointer to the CallTarget struct
	l.logDebugAST("    Constructed CallTarget: IsTool=%t, Name=%s", node.IsTool, node.Name)
}

// ExitCallable_expr
func (l *neuroScriptListenerImpl) ExitCallable_expr(ctx *gen.Callable_exprContext) {
	l.logDebugAST("--- Exit Callable_expr: %q", ctx.GetText())

	// 1. Pop Arguments
	numArgs := 0
	argExpressions := []Expression{} // Initialize empty slice

	// Check if optional expression list exists
	if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
		// Check if the expression list itself exists within the optional rule
		if exprList := exprListOpt.Expression_list(); exprList != nil {
			numArgs = len(exprList.AllExpression())
		}
	}

	if numArgs > 0 {
		argsRaw, ok := l.popNValues(numArgs)
		if !ok {
			// Error already added by popNValues
			l.pushValue(nil)
			return
		}
		// Assert each argument
		argExpressions = make([]Expression, numArgs) // Allocate correct size
		for i, argRaw := range argsRaw {
			argExpr, ok := argRaw.(Expression)
			if !ok {
				l.addError(ctx, "Internal error: Argument %d for call %q is not an Expression (got %T)", i+1, ctx.GetText(), argRaw)
				l.pushValue(nil)
				return
			}
			argExpressions[i] = argExpr
		}
	} // If numArgs is 0, argExpressions remains an empty slice

	// 2. Pop Call Target (should be *CallTarget pushed by ExitCall_target or built-in keyword)
	var targetNode CallTarget // Store the actual target info
	var nodePos *Position     // Position of the start of the call expression

	if targetCtx := ctx.Call_target(); targetCtx != nil {
		// Target was pushed by ExitCall_target
		targetRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping call target node for %q", ctx.GetText())
			l.pushValue(nil)
			return
		}
		targetPtr, ok := targetRaw.(*CallTarget) // Assert it's a pointer
		if !ok {
			l.addError(ctx, "Internal error: Popped call target for %q is not *CallTarget (got %T)", ctx.GetText(), targetRaw)
			l.pushValue(nil)
			return
		}
		targetNode = *targetPtr       // Dereference to get the value
		nodePos = targetNode.GetPos() // Position comes from the target node
		l.logDebugAST("    Popped CallTarget: IsTool=%t, Name=%s", targetNode.IsTool, targetNode.Name)
	} else {
		// Built-in function call (target info determined directly from keyword token)
		targetNode.IsTool = false
		var keywordToken antlr.TerminalNode
		switch {
		case ctx.KW_LN() != nil:
			keywordToken = ctx.KW_LN()
			targetNode.Name = "ln"
		case ctx.KW_LOG() != nil:
			keywordToken = ctx.KW_LOG()
			targetNode.Name = "log"
		case ctx.KW_SIN() != nil:
			keywordToken = ctx.KW_SIN()
			targetNode.Name = "sin"
		case ctx.KW_COS() != nil:
			keywordToken = ctx.KW_COS()
			targetNode.Name = "cos"
		case ctx.KW_TAN() != nil:
			keywordToken = ctx.KW_TAN()
			targetNode.Name = "tan"
		case ctx.KW_ASIN() != nil:
			keywordToken = ctx.KW_ASIN()
			targetNode.Name = "asin"
		case ctx.KW_ACOS() != nil:
			keywordToken = ctx.KW_ACOS()
			targetNode.Name = "acos"
		case ctx.KW_ATAN() != nil:
			keywordToken = ctx.KW_ATAN()
			targetNode.Name = "atan"
		default:
			l.addError(ctx, "Internal error: Unhandled target type in Callable_expr: %q", ctx.GetText())
			l.pushValue(nil)
			return
		}
		nodePos = tokenToPosition(keywordToken.GetSymbol()) // Position of the built-in keyword
		targetNode.Pos = nodePos                            // Also set Pos in the constructed target for consistency
		l.logDebugAST("    Identified Built-in function call target: %s", targetNode.Name)
	}

	// 3. Create and Push Node
	node := &CallableExprNode{
		Pos:       nodePos, // Use position determined above
		Target:    targetNode,
		Arguments: argExpressions, // Assign the []Expression slice
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed CallableExprNode: Target=%+v, Args=%d", node.Target, len(node.Arguments))
}

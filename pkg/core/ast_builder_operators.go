// pkg/core/ast_builder_operators.go
package core

import (
	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// Helper function to create BinaryOpNode from context
// Builds a BinaryOpNode by popping two values (left, right) from the stack
// and using the provided operator token. Pushes the resulting node back.
func (l *neuroScriptListenerImpl) buildBinaryOpNode(ctx antlr.ParserRuleContext, opToken antlr.TerminalNode) {
	right, okR := l.popValue()
	if !okR {
		l.logger.Printf("[ERROR] AST Builder: Stack error popping right operand for op %q in rule %T", opToken.GetText(), ctx)
		l.pushValue(nil) // Push error marker
		return
	}
	left, okL := l.popValue()
	if !okL {
		l.logger.Printf("[ERROR] AST Builder: Stack error popping left operand for op %q in rule %T", opToken.GetText(), ctx)
		l.pushValue(nil) // Push error marker
		return
	}
	op := opToken.GetText()
	l.pushValue(BinaryOpNode{Left: left, Operator: op, Right: right})
	l.logDebugAST("    Constructed BinaryOpNode: %T %s %T", left, op, right)
}

// --- Exit methods for expression precedence rules (Operators) ---

// ExitLogical_or_expr handles the OR operator (lowest precedence).
// It iterates through all OR operators found in the context and builds
// BinaryOpNodes left-associatively.
func (l *neuroScriptListenerImpl) ExitLogical_or_expr(ctx *gen.Logical_or_exprContext) {
	l.logDebugAST(">>> Exit Logical_or_expr: %q", ctx.GetText())
	operators := ctx.AllKW_OR() // Get all OR tokens
	if len(operators) > 0 {
		// Build nodes left-associatively by processing operators as found
		for i := 1; i < len(ctx.GetChildren()); i += 2 { // Step by 2: operand, operator, operand...
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerKW_OR {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Logical_or_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitLogical_and_expr handles the AND operator.
func (l *neuroScriptListenerImpl) ExitLogical_and_expr(ctx *gen.Logical_and_exprContext) {
	l.logDebugAST(">>> Exit Logical_and_expr: %q", ctx.GetText())
	operators := ctx.AllKW_AND()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerKW_AND {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Logical_and_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitBitwise_or_expr handles the bitwise OR (|) operator.
func (l *neuroScriptListenerImpl) ExitBitwise_or_expr(ctx *gen.Bitwise_or_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_or_expr: %q", ctx.GetText())
	operators := ctx.AllPIPE()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerPIPE {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Bitwise_or_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitBitwise_xor_expr handles the bitwise XOR (^) operator.
func (l *neuroScriptListenerImpl) ExitBitwise_xor_expr(ctx *gen.Bitwise_xor_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_xor_expr: %q", ctx.GetText())
	operators := ctx.AllCARET()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerCARET {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Bitwise_xor_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitBitwise_and_expr handles the bitwise AND (&) operator.
func (l *neuroScriptListenerImpl) ExitBitwise_and_expr(ctx *gen.Bitwise_and_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_and_expr: %q", ctx.GetText())
	operators := ctx.AllAMPERSAND()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerAMPERSAND {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Bitwise_and_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitEquality_expr handles equality (==) and inequality (!=) operators.
func (l *neuroScriptListenerImpl) ExitEquality_expr(ctx *gen.Equality_exprContext) {
	l.logDebugAST(">>> Exit Equality_expr: %q", ctx.GetText())
	operators := append(ctx.AllEQ(), ctx.AllNEQ()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok && (opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerEQ || opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerNEQ) {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Equality_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitRelational_expr handles relational (<, >, <=, >=) operators.
func (l *neuroScriptListenerImpl) ExitRelational_expr(ctx *gen.Relational_exprContext) {
	l.logDebugAST(">>> Exit Relational_expr: %q", ctx.GetText())
	operators := append(append(append(ctx.AllLT(), ctx.AllGT()...), ctx.AllLTE()...), ctx.AllGTE()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok {
				atype := opNode.GetSymbol().GetTokenType()
				if atype == gen.NeuroScriptLexerLT || atype == gen.NeuroScriptLexerGT ||
					atype == gen.NeuroScriptLexerLTE || atype == gen.NeuroScriptLexerGTE {
					l.buildBinaryOpNode(ctx, opNode)
				} else {
					l.logger.Printf("[WARN] Unexpected token type %d at operator position in Relational_expr", atype)
				}
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T at operator position in Relational_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitAdditive_expr handles addition (+) and subtraction (-) operators.
// Note: '+' also handles string concatenation in the evaluation phase.
func (l *neuroScriptListenerImpl) ExitAdditive_expr(ctx *gen.Additive_exprContext) {
	l.logDebugAST(">>> Exit Additive_expr: %q", ctx.GetText())
	operators := append(ctx.AllPLUS(), ctx.AllMINUS()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok && (opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerPLUS || opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMINUS) {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Additive_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitMultiplicative_expr handles multiplication (*), division (/), and modulo (%) operators.
func (l *neuroScriptListenerImpl) ExitMultiplicative_expr(ctx *gen.Multiplicative_exprContext) {
	l.logDebugAST(">>> Exit Multiplicative_expr: %q", ctx.GetText())
	operators := append(append(ctx.AllSTAR(), ctx.AllSLASH()...), ctx.AllPERCENT()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok && (opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerSTAR ||
				opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerSLASH ||
				opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerPERCENT) {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Printf("[WARN] Unexpected child type %T or token type at operator position in Multiplicative_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitUnary_expr handles unary minus (-) and logical NOT operators.
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST(">>> Exit Unary_expr: %q", ctx.GetText())
	var opToken antlr.TerminalNode
	// Check which unary operator is present in this specific context instance
	if ctx.MINUS() != nil {
		opToken = ctx.MINUS()
	} else if ctx.KW_NOT() != nil {
		opToken = ctx.KW_NOT()
	}

	// If an operator was found, pop the operand and build the UnaryOpNode
	if opToken != nil {
		operand, ok := l.popValue()
		if !ok {
			l.logger.Printf("[ERROR] AST Builder: Stack error popping operand for unary op %q", opToken.GetText())
			l.pushValue(nil) // Push error marker
			return
		}
		op := opToken.GetText()
		l.pushValue(UnaryOpNode{Operator: op, Operand: operand})
		l.logDebugAST("    Constructed UnaryOpNode: %s %T", op, operand)
	}
	// If no operator token (e.g., just a power_expr), the value from Power_expr just passes through.
}

// ExitPower_expr handles the exponentiation (**) operator.
// Note: This is right-associative in the grammar, handled by recursive structure.
func (l *neuroScriptListenerImpl) ExitPower_expr(ctx *gen.Power_exprContext) {
	l.logDebugAST(">>> Exit Power_expr: %q", ctx.GetText())
	opToken := ctx.STAR_STAR() // Check if the '**' token exists in this context instance
	if opToken != nil {
		// Pop the right operand (exponent) first due to right-associativity processing
		exponent, okE := l.popValue()
		if !okE {
			l.logger.Println("[ERROR] AST Builder: Stack error popping exponent for POWER")
			l.pushValue(nil)
			return
		}
		// Pop the left operand (base)
		base, okB := l.popValue()
		if !okB {
			l.logger.Println("[ERROR] AST Builder: Stack error popping base for POWER")
			l.pushValue(nil)
			return
		}

		l.pushValue(BinaryOpNode{Left: base, Operator: "**", Right: exponent})
		l.logDebugAST("    Constructed BinaryOpNode: %T ** %T", base, exponent)
	}
	// If no '**' token, the value from Primary just passes through the stack.
}

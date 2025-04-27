// pkg/core/ast_builder_operators.go
package core

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// Helper function to create BinaryOpNode from context
// (Unchanged)
func (l *neuroScriptListenerImpl) buildBinaryOpNode(ctx antlr.ParserRuleContext, opToken antlr.TerminalNode) {
	right, okR := l.popValue()
	if !okR {
		l.logger.Error("AST Builder: Stack error popping right operand for op %q in rule %T", opToken.GetText(), ctx)
		l.pushValue(nil)
		return
	}
	left, okL := l.popValue()
	if !okL {
		l.logger.Error("AST Builder: Stack error popping left operand for op %q in rule %T", opToken.GetText(), ctx)
		l.pushValue(nil)
		return
	}
	op := opToken.GetText()
	l.pushValue(BinaryOpNode{Left: left, Operator: op, Right: right})
	l.logDebugAST("    Constructed BinaryOpNode: %T %s %T", left, op, right)
}

// --- Exit methods for expression precedence rules (Operators) ---

// ExitLogical_or_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitLogical_or_expr(ctx *gen.Logical_or_exprContext) {
	l.logDebugAST(">>> Exit Logical_or_expr: %q", ctx.GetText())
	operators := ctx.AllKW_OR()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerKW_OR {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Logical_or_expr")
			}
		}
	}
}

// ExitLogical_and_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitLogical_and_expr(ctx *gen.Logical_and_exprContext) {
	l.logDebugAST(">>> Exit Logical_and_expr: %q", ctx.GetText())
	operators := ctx.AllKW_AND()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerKW_AND {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Logical_and_expr")
			}
		}
	}
}

// ExitBitwise_or_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitBitwise_or_expr(ctx *gen.Bitwise_or_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_or_expr: %q", ctx.GetText())
	operators := ctx.AllPIPE()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerPIPE {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Bitwise_or_expr")
			}
		}
	}
}

// ExitBitwise_xor_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitBitwise_xor_expr(ctx *gen.Bitwise_xor_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_xor_expr: %q", ctx.GetText())
	operators := ctx.AllCARET()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerCARET {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Bitwise_xor_expr")
			}
		}
	}
}

// ExitBitwise_and_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitBitwise_and_expr(ctx *gen.Bitwise_and_exprContext) {
	l.logDebugAST(">>> Exit Bitwise_and_expr: %q", ctx.GetText())
	operators := ctx.AllAMPERSAND()
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			if opNode, ok := ctx.GetChild(i).(antlr.TerminalNode); ok && opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerAMPERSAND {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Bitwise_and_expr")
			}
		}
	}
}

// ExitEquality_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitEquality_expr(ctx *gen.Equality_exprContext) {
	l.logDebugAST(">>> Exit Equality_expr: %q", ctx.GetText())
	operators := append(ctx.AllEQ(), ctx.AllNEQ()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok && (opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerEQ || opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerNEQ) {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Equality_expr")
			}
		}
	}
}

// ExitRelational_expr (Unchanged)
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
					l.logger.Warn("Unexpected token type %d at operator position in Relational_expr", atype)
				}
			} else {
				l.logger.Warn("Unexpected child type %T at operator position in Relational_expr", ctx.GetChild(i))
			}
		}
	}
}

// ExitAdditive_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitAdditive_expr(ctx *gen.Additive_exprContext) {
	l.logDebugAST(">>> Exit Additive_expr: %q", ctx.GetText())
	operators := append(ctx.AllPLUS(), ctx.AllMINUS()...)
	if len(operators) > 0 {
		for i := 1; i < len(ctx.GetChildren()); i += 2 {
			opNode, ok := ctx.GetChild(i).(antlr.TerminalNode)
			if ok && (opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerPLUS || opNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMINUS) {
				l.buildBinaryOpNode(ctx, opNode)
			} else {
				l.logger.Warn("Unexpected child type/token at operator position in Additive_expr")
			}
		}
	}
}

// ExitMultiplicative_expr (Unchanged)
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
				l.logger.Warn("Unexpected child type/token at operator position in Multiplicative_expr")
			}
		}
	}
}

// ExitUnary_expr handles unary minus (-), logical NOT, and NEW: no/some.
func (l *neuroScriptListenerImpl) ExitUnary_expr(ctx *gen.Unary_exprContext) {
	l.logDebugAST(">>> Exit Unary_expr: %q", ctx.GetText())
	var opToken antlr.TerminalNode
	// Check which unary operator is present
	if ctx.MINUS() != nil {
		opToken = ctx.MINUS()
	} else if ctx.KW_NOT() != nil {
		opToken = ctx.KW_NOT()
	} else if ctx.KW_NO() != nil { // NEW
		opToken = ctx.KW_NO()
	} else if ctx.KW_SOME() != nil { // NEW
		opToken = ctx.KW_SOME()
	}

	if opToken != nil {
		operand, ok := l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Stack error popping operand for unary op %q", opToken.GetText())
			l.pushValue(nil) // Push error marker
			return
		}
		op := opToken.GetText() // Get operator as string ("-", "not", "no", "some")
		l.pushValue(UnaryOpNode{Operator: op, Operand: operand})
		l.logDebugAST("    Constructed UnaryOpNode: %s %T", op, operand)
	}
	// If no operator token (just power_expr), the value from Power_expr passes through.
}

// ExitPower_expr (Unchanged)
func (l *neuroScriptListenerImpl) ExitPower_expr(ctx *gen.Power_exprContext) {
	l.logDebugAST(">>> Exit Power_expr: %q", ctx.GetText())
	opToken := ctx.STAR_STAR()
	if opToken != nil {
		exponent, okE := l.popValue()
		if !okE {
			l.logger.Error("AST Builder: Stack error popping exponent for POWER")
			l.pushValue(nil)
			return
		}
		base, okB := l.popValue()
		if !okB {
			l.logger.Error("AST Builder: Stack error popping base for POWER")
			l.pushValue(nil)
			return
		}
		l.pushValue(BinaryOpNode{Left: base, Operator: "**", Right: exponent})
		l.logDebugAST("    Constructed BinaryOpNode: %T ** %T", base, exponent)
	}
}

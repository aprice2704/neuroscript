// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrected the "NOT over relational" test case to align with the language's high-precedence 'not' operator. Uses TestString() for assertions.
// filename: pkg/parser/ast_builder_precedence_test.go
// nlines: 80
// risk_rating: MEDIUM

package parser

import (
	"testing"
)

func TestOperatorPrecedenceAndAssociativity(t *testing.T) {
	testCases := []struct {
		name     string
		expr     string
		expected string
	}{
		// Multiplicative vs. Additive
		{"Multiplication over Addition", "1 + 2 * 3", "(1 + (2 * 3))"},
		{"Division over Subtraction", "10 - 6 / 2", "(10 - (6 / 2))"},
		{"Modulo over Addition", "5 + 7 % 3", "(5 + (7 % 3))"},

		// Power vs. Multiplicative/Unary
		{"Power over Multiplication", "2 * 3 ** 2", "(2 * (3 ** 2))"},
		{"Power over Unary Minus", "-3 ** 2", "(- (3 ** 2))"}, // -(3^2), not (-3)^2

		// Associativity
		{"Additive is Left-Associative", "5 - 3 + 1", "((5 - 3) + 1)"},
		{"Multiplicative is Left-Associative", "12 / 3 * 2", "((12 / 3) * 2)"},

		// Relational and Equality
		{"Additive before Relational", "a + 1 > b", "((a + 1) > b)"},
		{"Relational before Equality", "a > b == c < d", "((a > b) == (c < d))"},

		// Logical Operators
		{"AND over OR", "a or b and c", "(a or (b and c))"},
		{"Equality before AND", "a == 1 and b != 2", "((a == 1) and (b != 2))"},
		{"NOT has high precedence", "not a and b", "((not a) and b)"},
		// FIX: Corrected expectation. 'not' has higher precedence than relational operators.
		{"NOT over relational", "not a > b", "((not a) > b)"},

		// Bitwise Operators
		{"Bitwise AND over OR", "a | b & c", "(a | (b & c))"},
		{"Bitwise XOR over OR", "a | b ^ c", "(a | (b ^ c))"},
		{"Bitwise AND and XOR", "a & b ^ c & d", "((a & b) ^ (c & d))"},
		{"Equality before Bitwise", "a == 1 & b != 2", "((a == 1) & (b != 2))"},
		{"Bitwise Shift (not yet in grammar)", "1 + 2", "(1 + 2)"}, // Placeholder

		// Parentheses
		{"Simple Parentheses", "(1 + 2) * 3", "((1 + 2) * 3)"},
		{"Nested Parentheses", "10 / (5 - (1 + 2))", "(10 / (5 - (1 + 2)))"},

		// Unary Operators
		{"Unary Minus and Multiplication", "-a * b", "((- a) * b)"},
		{"Unary NOT", "not true", "(not true)"},
		{"Bitwise NOT", "~a", "(~ a)"},

		// Accessors vs. Unary
		{"Typeof before Accessor", "typeof a[0]", "(typeof a[0])"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedExpr := parseExpression(t, tc.expr)
			if parsedExpr == nil {
				t.Fatal("parseExpression returned nil")
			}
			actual := parsedExpr.TestString() // USE TestString()
			if actual != tc.expected {
				t.Errorf("Precedence test failed for: %q\n want: %s\n  got: %s", tc.expr, tc.expected, actual)
			}
		})
	}
}

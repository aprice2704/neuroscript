// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 16
// :: description: Updated formatMapLiteral to sort by key.TestString() instead of .Value.
// :: latestChange: Fix sorting of map entries now that Key is an Expression.
// :: filename: pkg/nsfmt/format_expr.go
// :: serialization: go

package nsfmt

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// getPrecedence returns a numeric precedence for binary operators.
// Higher numbers bind tighter.
func getPrecedence(op string) int {
	switch op {
	case "*", "/", "%":
		return 5
	case "+", "-":
		return 4
	case "==", "!=", "<", "<=", ">", ">=":
		return 3
	case "and":
		return 2
	case "or":
		return 1
	default:
		return 0
	}
}

// formatExpression recursively formats an AST expression node into a string.
// It now takes a prefixLen to determine if wrapping is needed.
func (f *formatter) formatExpression(expr ast.Expression, prefixLen int) string {
	if expr == nil {
		return "<nil_expr>"
	}

	switch n := expr.(type) {
	// Simple literals and identifiers
	case *ast.NumberLiteralNode,
		*ast.NilLiteralNode,
		*ast.BooleanLiteralNode,
		*ast.VariableNode,
		*ast.LValueNode,
		*ast.CallTarget:
		return n.TestString()

	case *ast.StringLiteralNode:
		// Logic:
		// 1. If the string contains newlines, we generally want Triple Backticks (```) for readability.
		// 2. EXCEPTION: If the string is PURE WHITESPACE (e.g. "\n", "\n\n"), backticks are visually heavy.
		//    Keep these as standard quoted strings (e.g. "\n").
		// 3. EXCEPTION: If the content contains "```", we cannot use backticks.

		hasNewline := strings.Contains(n.Value, "\n")
		hasTripleBacktick := strings.Contains(n.Value, "```")
		isPureWhitespace := len(strings.TrimSpace(n.Value)) == 0

		if hasNewline && !hasTripleBacktick && !isPureWhitespace {
			return fmt.Sprintf("```%s```", n.Value)
		}

		// Fallback to standard quoted string (escapes newlines if present)
		return n.TestString()

	// Collections: These are the complex multi-line cases
	case *ast.ListLiteralNode:
		return f.formatListLiteral(n, prefixLen) // Pass prefix
	case *ast.MapLiteralNode:
		return f.formatMapLiteral(n, prefixLen) // Pass prefix

	// Recursive expression types
	case *ast.BinaryOpNode:
		// Calculate precedence
		currentPrec := getPrecedence(n.Operator)

		// Format Left
		leftStr := f.formatExpression(n.Left, 0)
		if leftOp, ok := n.Left.(*ast.BinaryOpNode); ok {
			leftPrec := getPrecedence(leftOp.Operator)
			// Wrap if child has lower precedence
			if leftPrec < currentPrec {
				leftStr = fmt.Sprintf("(%s)", leftStr)
			}
		}

		// Format Right
		rightStr := f.formatExpression(n.Right, 0)
		if rightOp, ok := n.Right.(*ast.BinaryOpNode); ok {
			rightPrec := getPrecedence(rightOp.Operator)
			// Wrap if child has lower precedence OR same precedence (for left-associativity)
			// e.g. (a - b) - c -> left is same, no parens.
			// a - (b - c) -> right is same, needs parens.
			if rightPrec <= currentPrec {
				rightStr = fmt.Sprintf("(%s)", rightStr)
			}
		}

		return fmt.Sprintf("%s %s %s", leftStr, n.Operator, rightStr)

	case *ast.UnaryOpNode:
		var operatorLen int
		var operatorStr string

		// Add space for word-based operators
		switch n.Operator {
		case "not", "some", "no":
			operatorStr = n.Operator + " "
			operatorLen = len(operatorStr)
		default:
			operatorStr = n.Operator
			operatorLen = len(operatorStr)
		}

		// Calculate the prefix for the operand
		operandPrefixLen := prefixLen + operatorLen
		operand := f.formatExpression(n.Operand, operandPrefixLen)

		if binOp, ok := n.Operand.(*ast.BinaryOpNode); ok {
			// Always wrap binary ops inside unary ops to avoid ambiguity
			// e.g. not (a and b)
			_ = binOp
			operand = fmt.Sprintf("(%s)", operand)
		}

		return fmt.Sprintf("%s%s", operatorStr, operand)

	case *ast.ElementAccessNode:
		// The object gets the prefix, as it's the start of the expression
		obj := f.formatExpression(n.Collection, prefixLen)
		// The index is "inside" the brackets, pass 0
		index := f.formatExpression(n.Accessor, 0)
		return fmt.Sprintf("%s[%s]", obj, index)

	case *ast.CallableExprNode:
		return f.formatCallable(n, prefixLen) // Pass prefix

	default:
		// Fallback for any other expression types
		return expr.TestString()
	}
}

// formatListLiteral formats a list. It tries to fit on one line,
// but switches to multi-line if it's too long.
func (f *formatter) formatListLiteral(list *ast.ListLiteralNode, prefixLen int) string {
	if len(list.Elements) == 0 {
		return "[]"
	}

	// 1. Try to build the single-line version
	var bSingle strings.Builder
	bSingle.WriteString("[")
	for i, el := range list.Elements {
		bSingle.WriteString(f.formatExpression(el, 0)) // Pass 0 for sub-elements
		if i < len(list.Elements)-1 {
			bSingle.WriteString(", ")
		}
	}
	bSingle.WriteString("]")
	singleLine := bSingle.String()

	// 2. Decide if we must switch to multi-line
	fullLineLen := prefixLen + len(singleLine)
	if fullLineLen > maxLineLength {
		// Fallthrough to multi-line logic
	} else {
		return singleLine // Single line is fine
	}

	// 3. Build multi-line ("vertical") version
	var bMulti strings.Builder
	itemIndentStr := strings.Repeat(indentString, f.indent+1)
	closingIndentStr := strings.Repeat(indentString, f.indent)

	bMulti.WriteString("[ \\\n")
	for i, el := range list.Elements {
		bMulti.WriteString(itemIndentStr)
		bMulti.WriteString(f.formatExpression(el, 0)) // Pass 0 for sub-elements
		// FIX: Only add comma if it's not the last element
		if i < len(list.Elements)-1 {
			bMulti.WriteString(", \\\n")
		} else {
			bMulti.WriteString(" \\\n") // No trailing comma
		}
	}
	bMulti.WriteString(closingIndentStr)
	bMulti.WriteString("]")
	return bMulti.String()
}

// formatMapLiteral formats a map. It tries to fit on one line,
// but switches to multi-line if it's too long.
func (f *formatter) formatMapLiteral(m *ast.MapLiteralNode, prefixLen int) string {
	if len(m.Entries) == 0 {
		return "{}"
	}

	// 1. Try to build the single-line version
	var bSingle strings.Builder
	bSingle.WriteString("{")

	// Sort entries by key for stable output
	sort.Slice(m.Entries, func(i, j int) bool {
		// FIX: Use TestString() for sorting expressions
		return m.Entries[i].Key.TestString() < m.Entries[j].Key.TestString()
	})

	for i, entry := range m.Entries {
		val := f.formatExpression(entry.Value, 0) // Pass 0 for sub-elements
		bSingle.WriteString(fmt.Sprintf("%s: %s", entry.Key.TestString(), val))
		if i < len(m.Entries)-1 {
			bSingle.WriteString(", ")
		}
	}
	bSingle.WriteString("}")
	singleLine := bSingle.String()

	// 2. Decide if we must switch to multi-line
	fullLineLen := prefixLen + len(singleLine)
	if fullLineLen > maxLineLength {
		// Fallthrough to multi-line logic
	} else {
		return singleLine // Single line is fine
	}

	// 3. Build multi-line ("vertical") version
	var bMulti strings.Builder
	itemIndentStr := strings.Repeat(indentString, f.indent+1)
	closingIndentStr := strings.Repeat(indentString, f.indent)

	bMulti.WriteString("{ \\\n")
	for i, entry := range m.Entries {
		bMulti.WriteString(itemIndentStr)
		val := f.formatExpression(entry.Value, 0) // Pass 0 for sub-elements
		// FIX: Only add comma if's not the last element
		if i < len(m.Entries)-1 {
			bMulti.WriteString(fmt.Sprintf("%s: %s, \\\n", entry.Key.TestString(), val))
		} else {
			bMulti.WriteString(fmt.Sprintf("%s: %s \\\n", entry.Key.TestString(), val)) // No trailing comma
		}
	}
	bMulti.WriteString(closingIndentStr)
	bMulti.WriteString("}")
	return bMulti.String()
}

// formatCallable formats a function call. It tries to fit on one line,
// but switches to multi-line if it's too long.
func (f *formatter) formatCallable(n *ast.CallableExprNode, prefixLen int) string {
	// The callee target gets the prefix, as it's the start
	callee := f.formatExpression(&n.Target, prefixLen)
	if len(n.Arguments) == 0 {
		return fmt.Sprintf("%s()", callee)
	}

	// 1. Try to build the single-line version
	var bSingle strings.Builder
	bSingle.WriteString(callee)
	bSingle.WriteString("(")

	argStrings := make([]string, len(n.Arguments))
	for i, arg := range n.Arguments {
		// Sub-expressions pass 0 for prefix
		argStrings[i] = f.formatExpression(arg, 0)
	}
	bSingle.WriteString(strings.Join(argStrings, ", "))

	bSingle.WriteString(")")
	singleLine := bSingle.String()

	// 2. Decide if we must switch to multi-line
	fullLineLen := prefixLen + len(singleLine)
	if fullLineLen > maxLineLength {
		// Fallthrough to multi-line logic
	} else {
		return singleLine // Single line is fine
	}

	// 3. Build multi-line ("vertical") version
	var bMulti strings.Builder
	itemIndentStr := strings.Repeat(indentString, f.indent+1)
	closingIndentStr := strings.Repeat(indentString, f.indent)

	bMulti.WriteString(callee)
	bMulti.WriteString("( \\\n")
	for i, arg := range n.Arguments {
		bMulti.WriteString(itemIndentStr)
		// Pass prefix for args in multi-line mode?
		// No, they are on their own lines, pass 0.
		bMulti.WriteString(f.formatExpression(arg, 0))
		// FIX: Only add comma if it's not the last element
		if i < len(n.Arguments)-1 {
			bMulti.WriteString(", \\\n")
		} else {
			bMulti.WriteString(" \\\n") // No trailing comma
		}
	}
	bMulti.WriteString(closingIndentStr)
	bMulti.WriteString(")")
	return bMulti.String()
}

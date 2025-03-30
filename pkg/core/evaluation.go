package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// --- Evaluation Logic Helpers ---

// evaluateCondition - Handles ==, !=, >, <, >=, <= for strings, plus true/false checks
func (i *Interpreter) evaluateCondition(conditionStr string) (bool, error) {
	trimmedCond := strings.TrimSpace(conditionStr)

	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	comparisonFuncs := map[string]func(string, string) bool{
		"==": func(a, b string) bool { return a == b }, "!=": func(a, b string) bool { return a != b },
		">": func(a, b string) bool { return a > b }, "<": func(a, b string) bool { return a < b },
		">=": func(a, b string) bool { return a >= b }, "<=": func(a, b string) bool { return a <= b },
	}

	for _, op := range operators {
		parts := strings.SplitN(trimmedCond, op, 2)
		if len(parts) == 2 {
			lhs := strings.TrimSpace(parts[0])
			rhs := strings.TrimSpace(parts[1])
			resolvedLhsVal := i.evaluateExpression(lhs)
			resolvedRhsVal := i.evaluateExpression(rhs)
			resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
			resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
			return comparisonFuncs[op](resolvedLhsStr, resolvedRhsStr), nil
		}
	}

	resolvedValue := i.evaluateExpression(trimmedCond)
	resolvedValueStr := fmt.Sprintf("%v", resolvedValue)
	lowerResolved := strings.ToLower(resolvedValueStr)

	if lowerResolved == "true" {
		return true, nil
	}
	if lowerResolved == "false" {
		return false, nil
	}

	return false, fmt.Errorf("unsupported condition format or non-boolean result: %s -> %q", conditionStr, resolvedValueStr)
}

// resolvePlaceholders - Performs recursive placeholder lookup and replacement.
func (i *Interpreter) resolvePlaceholders(input string) string {
	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`)
	const maxDepth = 10
	originalInput := input

	var resolve func(s string, depth int) string
	resolve = func(s string, depth int) string {
		if depth > maxDepth {
			return originalInput
		} // Return original on depth limit

		changedInPass := false
		result := reVar.ReplaceAllStringFunc(s, func(match string) string {
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			}
			varName := strings.TrimSpace(varNameSubmatch[1])
			var lookupVal interface{}
			var found bool

			if varName == "__last_call_result" {
				lookupVal = i.lastCallResult
				found = true
			} else if isValidIdentifier(varName) {
				lookupVal, found = i.variables[varName]
			} else {
				return match
			}

			if found {
				replacement := fmt.Sprintf("%v", lookupVal)
				if replacement != match {
					changedInPass = true
					if strings.Contains(replacement, "{{") {
						return resolve(replacement, depth+1)
					}
				}
				return replacement
			} else {
				return match
			}
		})

		if !changedInPass {
			return result
		}
		if strings.Contains(result, "{{") && strings.Contains(result, "}}") {
			if result != s {
				return resolve(result, depth+1)
			} // Recurse only if changed and possible loop
		}
		return result
	}
	return resolve(input, 0)
}

// resolveValue - Resolves ONLY direct variable names & __last_call_result. Returns raw value.
func (i *Interpreter) resolveValue(input string) (value interface{}, found bool) {
	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "__last_call_result" {
		if i.lastCallResult != nil {
			return i.lastCallResult, true
		}
		return "", true
	}
	if isValidIdentifier(trimmedInput) {
		val, exists := i.variables[trimmedInput]
		if exists {
			return val, true
		}
	}
	return trimmedInput, false
}

// splitExpression splits an expression string by the top-level '+' operator.
func splitExpression(expr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune
	escapeNext := false
	placeholderLevel := 0
	parenLevel := 0
	flushCurrentPart := func() {
		trimmedPart := strings.TrimSpace(current.String())
		if trimmedPart != "" {
			parts = append(parts, trimmedPart)
		}
		current.Reset()
	}
	for i := 0; i < len(expr); i++ {
		ch := rune(expr[i])
		if escapeNext {
			current.WriteRune(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' {
			current.WriteRune(ch)
			if i+1 < len(expr) && (expr[i+1] == '"' || expr[i+1] == '\'' || expr[i+1] == '\\') {
				escapeNext = true
			}
			continue
		}
		if ch == '{' && i+1 < len(expr) && expr[i+1] == '{' {
			current.WriteString("{{")
			if !inQuotes && parenLevel == 0 {
				placeholderLevel++
			}
			i++
			continue
		}
		if ch == '}' && i > 0 && expr[i-1] == '}' {
			current.WriteRune('}')
			if !inQuotes && parenLevel == 0 && placeholderLevel > 0 {
				placeholderLevel--
			}
			continue
		}
		if (ch == '"' || ch == '\'') && placeholderLevel == 0 && parenLevel == 0 {
			current.WriteRune(ch)
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
			}
			continue
		}
		if ch == '(' && !inQuotes && placeholderLevel == 0 {
			parenLevel++
			current.WriteRune(ch)
			continue
		}
		if ch == ')' && !inQuotes && placeholderLevel == 0 {
			if parenLevel > 0 {
				parenLevel--
			}
			current.WriteRune(ch)
			continue
		}
		if ch == '+' && !inQuotes && placeholderLevel == 0 && parenLevel == 0 {
			flushCurrentPart()
			parts = append(parts, "+")
		} else {
			current.WriteRune(ch)
		}
	}
	flushCurrentPart()
	finalParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			finalParts = append(finalParts, p)
		}
	}
	return finalParts
}

// isValidIdentifier checks if a string is a valid NeuroScript identifier.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	upperName := strings.ToUpper(name)
	keywords := map[string]bool{"DEFINE": true, "PROCEDURE": true, "COMMENT": true, "END": true, "SET": true, "CALL": true, "RETURN": true, "IF": true, "THEN": true, "ELSE": true, "WHILE": true, "DO": true, "FOR": true, "EACH": true, "IN": true, "TOOL": true, "LLM": true}
	if keywords[upperName] {
		return false
	}
	if name == "__last_call_result" {
		return true
	}
	return true
}

// evaluateExpression - Central Evaluator. Returns final value (interface{}).
// ** FIX: Reworked Concat logic for better recursion/evaluation **
func (i *Interpreter) evaluateExpression(expr string) interface{} {
	trimmedExpr := strings.TrimSpace(expr)

	// --- Check 0: Parenthesized Expression ---
	if len(trimmedExpr) >= 2 && trimmedExpr[0] == '(' && trimmedExpr[len(trimmedExpr)-1] == ')' {
		innerExpr := trimmedExpr[1 : len(trimmedExpr)-1]
		return i.evaluateExpression(innerExpr) // Recursively evaluate inner content
	}

	// --- Check 1: Direct variable/keyword lookup ---
	rawValue, found := i.resolveValue(trimmedExpr)
	if found {
		return rawValue // Return raw value (could be string, int, slice, etc.)
	}
	// If not found, 'rawValue' holds the trimmed input string.

	// --- Check 2: Concatenation ---
	// Split the original expression *before* placeholder resolution for accurate structure
	parts := splitExpression(trimmedExpr)

	if len(parts) > 1 { // Contains '+' operator at top level
		var builder strings.Builder
		isValidConcat := true
		if len(parts) == 0 {
			isValidConcat = false
		}

		for idx, part := range parts {
			isOperatorPart := (idx%2 == 1) // Expects operand, +, operand...
			if isOperatorPart {
				if part != "+" {
					isValidConcat = false
					break
				}
				continue // Skip '+'
			}

			// *** Evaluate the operand part recursively ***
			// This is the crucial change: evaluate each operand *within* the loop.
			// This handles loop variables correctly.
			evaluatedPart := i.evaluateExpression(part) // Evaluate the part (e.g., `{{output}}` or `{{char}}` or `"-"`)

			// Stringify the result of the evaluation for concatenation
			partStr := fmt.Sprintf("%v", evaluatedPart)

			// Unquote ONLY if the *original part string* looked like a literal string
			trimmedOriginalPart := strings.TrimSpace(part)
			isOriginalLiteral := len(trimmedOriginalPart) >= 2 && ((trimmedOriginalPart[0] == '"' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '"') || (trimmedOriginalPart[0] == '\'' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '\''))

			if isOriginalLiteral {
				// Unquote the stringified result if it still looks quoted
				if len(partStr) >= 2 && ((partStr[0] == '"' && partStr[len(partStr)-1] == '"') || (partStr[0] == '\'' && partStr[len(partStr)-1] == '\'')) {
					unquoted, err := strconv.Unquote(partStr)
					if err == nil {
						partStr = unquoted
					}
				}
			}
			builder.WriteString(partStr)
		}

		if isValidConcat {
			return builder.String() // Return concatenated result
		}
		// If not valid concat (e.g., "a" + + "b"), fall through to treat original expression as single unit
	}

	// --- Treat as Single Unit (Literal or Identifier/Placeholder String) ---
	// If not direct var, not parenthesized, and not valid concatenation.
	// Resolve placeholders *now* on the original expression.
	resolvedExprStr := i.resolvePlaceholders(trimmedExpr)

	// Unquote if the *original* expression looked like a quoted literal.
	if len(trimmedExpr) >= 2 &&
		((trimmedExpr[0] == '"' && trimmedExpr[len(trimmedExpr)-1] == '"') ||
			(trimmedExpr[0] == '\'' && trimmedExpr[len(trimmedExpr)-1] == '\'')) {
		// Use the placeholder-resolved string for unquoting
		unquoted, err := strconv.Unquote(resolvedExprStr)
		if err == nil {
			return unquoted
		}
		// If unquote fails, return the placeholder-resolved string as is.
		return resolvedExprStr
	}

	// If not originally quoted, return the placeholder-resolved string
	// (could be an unresolved placeholder, an unknown identifier, etc.)
	return resolvedExprStr
}

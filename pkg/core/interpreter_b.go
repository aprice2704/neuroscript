package core

import (
	"fmt"
	"regexp"
	"strings"
	// NOTE: No os, path/filepath, json needed here
)

// --- Evaluation Logic Helpers ---

// evaluateCondition evaluates simple conditions (LHS == RHS, LHS != RHS, or boolean variable/literal)
func (i *Interpreter) evaluateCondition(conditionStr string) (bool, error) {
	trimmedCond := strings.TrimSpace(conditionStr)
	// 1. Check for equality (==)
	partsEq := strings.SplitN(trimmedCond, "==", 2)
	if len(partsEq) == 2 {
		lhs := strings.TrimSpace(partsEq[0])
		rhs := strings.TrimSpace(partsEq[1])
		resolvedLhsVal := i.evaluateExpression(lhs)
		resolvedRhsVal := i.evaluateExpression(rhs) // evaluateExpression is in _c.go
		resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
		resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
		fmt.Printf("      [Cond Eval ==] Comparing: %q == %q\n", resolvedLhsStr, resolvedRhsStr)
		return resolvedLhsStr == resolvedRhsStr, nil
	}
	// 2. Check for inequality (!=)
	partsNeq := strings.SplitN(trimmedCond, "!=", 2)
	if len(partsNeq) == 2 {
		lhs := strings.TrimSpace(partsNeq[0])
		rhs := strings.TrimSpace(partsNeq[1])
		resolvedLhsVal := i.evaluateExpression(lhs)
		resolvedRhsVal := i.evaluateExpression(rhs) // evaluateExpression is in _c.go
		resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
		resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
		fmt.Printf("      [Cond Eval !=] Comparing: %q != %q\n", resolvedLhsStr, resolvedRhsStr)
		return resolvedLhsStr != resolvedRhsStr, nil
	}
	// TODO: Add numeric comparisons >, <, >=, <=

	// 3. Not equality/inequality? Evaluate the whole string and check if it's "true" or "false"
	resolvedValue := i.evaluateExpression(trimmedCond) // evaluateExpression is in _c.go
	resolvedValueStr := fmt.Sprintf("%v", resolvedValue)
	lowerResolved := strings.ToLower(resolvedValueStr)
	if lowerResolved == "true" {
		return true, nil
	}
	if lowerResolved == "false" {
		return false, nil
	}

	// 4. Invalid condition format
	return false, fmt.Errorf("unsupported condition format or non-boolean result: %s -> %q", conditionStr, resolvedValueStr)
}

// resolvePlaceholders - Recursively substitutes {{var}} or __last_call_result
func (i *Interpreter) resolvePlaceholders(input string) string {
	resolved := input
	// Use package-level compiled regex for efficiency if possible
	reVar := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	reLastResult := regexp.MustCompile(`__last_call_result`)
	const maxDepth = 10
	var resolveRecursive func(s string, depth int) string
	resolveRecursive = func(s string, depth int) string {
		if depth > maxDepth {
			fmt.Printf("  [Warn] Max placeholder recursion depth (%d) exceeded for: %q\n", maxDepth, input)
			return s
		}
		madeChange := false
		// Resolve __last_call_result first
		resolvedLast := reLastResult.ReplaceAllStringFunc(s, func(match string) string {
			if i.lastCallResult != nil {
				valueStr := fmt.Sprintf("%v", i.lastCallResult)
				madeChange = true
				return resolveRecursive(valueStr, depth+1)
			} else {
				fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
				madeChange = true
				return ""
			}
		})
		// Resolve {{var}} placeholders
		resolvedVars := reVar.ReplaceAllStringFunc(resolvedLast, func(match string) string {
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			} // Should not happen
			varName := strings.TrimSpace(varNameSubmatch[1])
			if value, exists := i.variables[varName]; exists {
				valueStr := fmt.Sprintf("%v", value)
				madeChange = true
				return resolveRecursive(valueStr, depth+1)
			} else {
				fmt.Printf("  [Warn] Variable '%s' not found for placeholder %q\n", varName, match)
				return match
			} // Keep placeholder if var not found
		})
		// Recurse if changes were made
		if madeChange && resolvedVars != s {
			return resolveRecursive(resolvedVars, depth+1)
		}
		return resolvedVars // No changes or fully resolved
	}
	resolved = resolveRecursive(input, 0)
	return resolved
}

// resolveValue - Handles simple variable lookup OR resolving placeholders in a literal.
// Returns the *value* (interface{}), not necessarily a string representation.
func (i *Interpreter) resolveValue(input string) interface{} {
	trimmedInput := strings.TrimSpace(input)
	// 1. Check for __last_call_result keyword explicitly
	if trimmedInput == "__last_call_result" {
		if i.lastCallResult != nil {
			return i.lastCallResult
		} // Return stored result (any type)
		fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
		return "" // Return empty string if not set
	}
	// 2. Check if input is a plain variable name
	isPlainVarName := false
	if isValidIdentifier(trimmedInput) { // isValidIdentifier defined in parser_c.go utils
		if !strings.Contains(trimmedInput, "{{") {
			isPlainVarName = true
		}
	}
	if isPlainVarName {
		if val, exists := i.variables[trimmedInput]; exists {
			return val
		} // Variable found, return its value (interface{})
		// If plain var name but not found, fall through to treat as literal.
		fmt.Printf("  [Warn] Variable '%s' not found, treating as literal.\n", trimmedInput)
		return trimmedInput // Return the name itself as a string
	}
	// 3. Not a plain variable lookup (or var not found). Treat input as literal string.
	// Resolve placeholders *within the input literal itself*. Returns a string.
	return i.resolvePlaceholders(trimmedInput) // Returns string
}

// splitExpression - Splits an expression by the '+' operator, respecting quotes and {{placeholders}}.
// Returns a list of parts, including the '+' operators themselves.
func splitExpression(expr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	placeholderLevel := 0 // Simple nesting check

	for i, ch := range expr {
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
		// Placeholders
		if ch == '{' && i+1 < len(expr) && expr[i+1] == '{' {
			if !inQuotes {
				placeholderLevel++
			}
			current.WriteRune(ch)
			continue
		}
		if ch == '}' && i > 0 && expr[i-1] == '}' {
			current.WriteRune(ch)
			if !inQuotes && placeholderLevel > 0 {
				placeholderLevel--
			}
			continue
		}
		// Quotes (only outside placeholders)
		if (ch == '"' || ch == '\'') && placeholderLevel == 0 {
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
			}
			current.WriteRune(ch)
			continue
		}
		// Split on '+' outside quotes and placeholders
		if ch == '+' && !inQuotes && placeholderLevel == 0 {
			trimmedPart := strings.TrimSpace(current.String())
			if trimmedPart != "" {
				parts = append(parts, trimmedPart)
			}
			parts = append(parts, "+") // Add operator
			current.Reset()
		} else {
			current.WriteRune(ch)
		} // Add regular char
	}
	// Add final part
	trimmedPart := strings.TrimSpace(current.String())
	if trimmedPart != "" {
		parts = append(parts, trimmedPart)
	}
	// Filter empty strings (should be minimal with this logic)
	finalParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			finalParts = append(finalParts, p)
		}
	}
	return finalParts
}

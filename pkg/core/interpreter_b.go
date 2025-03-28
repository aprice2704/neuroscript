package core

import (
	"fmt"
	"regexp"
	"strings"
)

// --- Evaluation Logic ---

// evaluateCondition evaluates simple conditions (LHS == RHS, LHS != RHS, or boolean variable/literal)
func (i *Interpreter) evaluateCondition(conditionStr string) (bool, error) {
	// Trim whitespace for reliable splitting
	trimmedCond := strings.TrimSpace(conditionStr)

	// 1. Check for equality (==)
	partsEq := strings.SplitN(trimmedCond, "==", 2)
	if len(partsEq) == 2 {
		lhs := strings.TrimSpace(partsEq[0])
		rhs := strings.TrimSpace(partsEq[1])
		// Evaluate both sides using the main expression evaluator
		resolvedLhsVal := i.evaluateExpression(lhs)
		resolvedRhsVal := i.evaluateExpression(rhs)
		// Compare the evaluated results (likely strings at this point)
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
		resolvedRhsVal := i.evaluateExpression(rhs)
		resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
		resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
		fmt.Printf("      [Cond Eval !=] Comparing: %q != %q\n", resolvedLhsStr, resolvedRhsStr)
		return resolvedLhsStr != resolvedRhsStr, nil
	}

	// 3. Not equality/inequality? Evaluate the whole string and check if it's "true" or "false"
	resolvedValue := i.evaluateExpression(trimmedCond)   // Evaluate the condition string itself
	resolvedValueStr := fmt.Sprintf("%v", resolvedValue) // Ensure it's a string
	lowerResolved := strings.ToLower(resolvedValueStr)
	if lowerResolved == "true" {
		return true, nil
	}
	if lowerResolved == "false" {
		return false, nil
	}

	// 4. Not "true" or "false" - invalid condition format for now
	// Consider adding numeric comparisons (>, <, >=, <=) here later if needed.
	return false, fmt.Errorf("unsupported condition format or non-boolean result: %s -> %q", conditionStr, resolvedValueStr)
}

// resolvePlaceholders - Recursively substitutes {{var}} or __last_call_result
func (i *Interpreter) resolvePlaceholders(input string) string {
	resolved := input
	reVar := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	reLastResult := regexp.MustCompile(`__last_call_result`)

	// Limit recursion depth to prevent infinite loops with circular references
	const maxDepth = 10
	var resolveRecursive func(s string, depth int) string
	resolveRecursive = func(s string, depth int) string {
		if depth > maxDepth {
			fmt.Printf("  [Warn] Max placeholder recursion depth (%d) exceeded for: %q\n", maxDepth, input)
			return s // Return unresolved string if depth limit hit
		}
		madeChange := false

		// Resolve __last_call_result first
		resolvedLast := reLastResult.ReplaceAllStringFunc(s, func(match string) string {
			if i.lastCallResult != nil {
				valueStr := fmt.Sprintf("%v", i.lastCallResult)
				madeChange = true
				// Recursively resolve placeholders within the substituted value
				return resolveRecursive(valueStr, depth+1)
			} else {
				fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
				madeChange = true // Treat as resolved to empty string
				return ""
			}
		})

		// Resolve {{var}} placeholders
		resolvedVars := reVar.ReplaceAllStringFunc(resolvedLast, func(match string) string {
			// Extract variable name robustly, handle potential spaces if spec allows later
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			} // Should not happen with regex
			varName := strings.TrimSpace(varNameSubmatch[1]) // Trim space if regex allows {{ name }}

			if value, exists := i.variables[varName]; exists {
				valueStr := fmt.Sprintf("%v", value)
				// Basic self-reference check removed as nested resolution handles it implicitly
				madeChange = true
				// Recursively resolve placeholders within the substituted value
				return resolveRecursive(valueStr, depth+1)
			} else {
				fmt.Printf("  [Warn] Variable '%s' not found for placeholder %q\n", varName, match)
				return match // Keep placeholder if var not found
			}
		})

		// If any change was made in this iteration, recurse again on the whole string
		// This handles cases like "{{a}} {{b}}" where resolving 'a' might reveal placeholders for 'b'
		if madeChange && resolvedVars != s { // Check if string actually changed to prevent infinite loop on no-op replaces
			return resolveRecursive(resolvedVars, depth+1)
		}
		return resolvedVars // No changes or already fully resolved, return result
	}

	resolved = resolveRecursive(input, 0)
	return resolved
}

// resolveValue - Handles simple variable lookup OR resolving placeholders in a literal.
// Returns the *value* (interface{}), not necessarily a string representation.
// *** Does NOT resolve placeholders within a returned variable's value here. ***
func (i *Interpreter) resolveValue(input string) interface{} {
	trimmedInput := strings.TrimSpace(input)

	// 1. Check for __last_call_result keyword explicitly
	if trimmedInput == "__last_call_result" {
		if i.lastCallResult != nil {
			// Return the stored result directly (could be any type)
			return i.lastCallResult
		}
		fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
		return "" // Return empty string if not set
	}

	// 2. Check if input is a plain variable name
	isPlainVarName := false
	if isValidIdentifier(trimmedInput) { // Use parser's definition of valid identifier (checks keywords too)
		// Ensure it doesn't look like a placeholder handled elsewhere
		if !strings.Contains(trimmedInput, "{{") {
			isPlainVarName = true
		}
	}

	if isPlainVarName {
		if val, exists := i.variables[trimmedInput]; exists {
			// Variable found. Return its stored value (interface{}) directly.
			// Do NOT resolve placeholders within 'val' here.
			return val
		}
		// If plain var name but not found, fall through to treat as literal string.
		fmt.Printf("  [Warn] Variable '%s' not found, treating as literal.\n", trimmedInput)
		// Treat unresolved variable name as a literal string of that name
		return trimmedInput // Return the name itself as a string
	}

	// 3. Not a plain variable lookup (or var not found). Treat input as literal string.
	// Resolve placeholders *within the input literal itself*. Returns a string.
	// Example: Input `"Hello {{name}}"`, if name="World", returns `"Hello World"`.
	// Example: Input `my var` (not identifier), returns `"my var"`.
	// Example: Input `{{unknown}}`, returns `"{{unknown}}"`.
	return i.resolvePlaceholders(trimmedInput) // Returns string
}

// splitExpression - Splits an expression for concatenation, respecting quotes and {{placeholders}}
// (No changes needed here based on recent test failures)
// splitExpression - Replaces the previous version with a single-pass state machine.
// Splits an expression by the '+' operator, respecting quotes and {{placeholders}}.
// Returns a list of parts, including the '+' operators themselves.
func splitExpression(expr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	placeholderLevel := 0 // Use level for nested {{ {{...}} }} if needed, though likely not

	for i, ch := range expr {
		if escapeNext {
			current.WriteRune(ch)
			escapeNext = false
			continue
		}

		if ch == '\\' {
			current.WriteRune(ch)
			// Check if next char is one that can be escaped inside quotes
			if i+1 < len(expr) && (expr[i+1] == '"' || expr[i+1] == '\'' || expr[i+1] == '\\') {
				escapeNext = true
			}
			continue
		}

		// Track placeholders {{ }} - Simple non-nested version
		if ch == '{' && i+1 < len(expr) && expr[i+1] == '{' {
			if !inQuotes {
				placeholderLevel++
			}
			current.WriteRune(ch)
			continue // Don't check for '+' inside placeholder open
		}
		if ch == '}' && i > 0 && expr[i-1] == '}' {
			current.WriteRune(ch)
			if !inQuotes && placeholderLevel > 0 {
				placeholderLevel--
			}
			continue // Don't check for '+' inside placeholder close
		}

		// Track quotes "" ''
		if (ch == '"' || ch == '\'') && placeholderLevel == 0 { // Only handle quotes outside placeholders
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
			}
			current.WriteRune(ch)
			continue // Don't check for '+' right after quote char
		}

		// Check for '+' operator outside quotes and placeholders
		if ch == '+' && !inQuotes && placeholderLevel == 0 {
			// Found a delimiter. Add the preceding part (if any) and the '+'.
			trimmedPart := strings.TrimSpace(current.String())
			if trimmedPart != "" {
				parts = append(parts, trimmedPart)
			}
			parts = append(parts, "+") // Add the operator itself as a part
			current.Reset()            // Start collecting the next part
		} else {
			// Regular character, add to current part
			current.WriteRune(ch)
		}
	}

	// Add the final part after the loop finishes
	trimmedPart := strings.TrimSpace(current.String())
	if trimmedPart != "" {
		parts = append(parts, trimmedPart)
	}

	// Clean up empty strings potentially caused by multiple '+' or leading/trailing '+'
	// (Though the logic above should minimize this)
	finalParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" { // Ensure '+' operator isn't accidentally removed if it was valid
			finalParts = append(finalParts, p)
		}
	}

	return finalParts
}

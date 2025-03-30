package core

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// --- Evaluation Logic Helpers ---

// evaluateCondition - Handles ==, !=, >, <, >=, <= for strings, plus true/false checks
func (i *Interpreter) evaluateCondition(conditionStr string) (bool, error) {
	trimmedCond := strings.TrimSpace(conditionStr)

	// Define operators and corresponding comparison functions
	operators := []string{">=", "<=", "==", "!=", ">", "<"} // Order matters: check >= before >
	comparisonFuncs := map[string]func(string, string) bool{
		"==": func(a, b string) bool { return a == b },
		"!=": func(a, b string) bool { return a != b },
		">":  func(a, b string) bool { return a > b },
		"<":  func(a, b string) bool { return a < b },
		">=": func(a, b string) bool { return a >= b },
		"<=": func(a, b string) bool { return a <= b },
	}

	// Iterate through operators to find the first match
	for _, op := range operators {
		parts := strings.SplitN(trimmedCond, op, 2)
		if len(parts) == 2 {
			lhs := strings.TrimSpace(parts[0])
			rhs := strings.TrimSpace(parts[1])

			// Evaluate LHS and RHS expressions
			resolvedLhsVal := i.evaluateExpression(lhs) // evaluateExpression from interpreter_c.go
			resolvedRhsVal := i.evaluateExpression(rhs) // evaluateExpression from interpreter_c.go

			// Convert evaluated results to strings for comparison
			resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
			resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)

			// Perform the comparison using the corresponding function
			fmt.Printf("      [Eval Cond %s] LHS: %q (%v), RHS: %q (%v)\n", op, resolvedLhsStr, resolvedLhsVal, resolvedRhsStr, resolvedRhsVal)
			result := comparisonFuncs[op](resolvedLhsStr, resolvedRhsStr)
			return result, nil
		}
	}

	// If no binary operator was found, evaluate the condition as a single expression
	// Expected to resolve to "true" or "false" (case-insensitive)
	resolvedValue := i.evaluateExpression(trimmedCond) // evaluateExpression from interpreter_c.go
	resolvedValueStr := fmt.Sprintf("%v", resolvedValue)
	lowerResolved := strings.ToLower(resolvedValueStr)

	fmt.Printf("      [Eval Cond Bool] Expr: %q -> %q\n", trimmedCond, resolvedValueStr)

	if lowerResolved == "true" {
		return true, nil
	}
	if lowerResolved == "false" {
		return false, nil
	}

	// If it's not a recognized operator expression and doesn't resolve to true/false
	return false, fmt.Errorf("unsupported condition format or non-boolean result: %s -> %q", conditionStr, resolvedValueStr)
}

// resolvePlaceholders - ** CORRECTED to perform recursive lookup and replacement **
func (i *Interpreter) resolvePlaceholders(input string) string {
	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`) // Regex for {{...}}
	const maxDepth = 10                          // Prevent infinite recursion

	var resolve func(s string, depth int) string
	resolve = func(s string, depth int) string {
		if depth > maxDepth {
			fmt.Printf("  [Warn] Max placeholder recursion depth (%d) exceeded for: %q\n", maxDepth, input)
			return s // Return original string if depth exceeded
		}

		changed := false
		result := reVar.ReplaceAllStringFunc(s, func(match string) string {
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match // Should not happen with valid regex match
			}
			varName := strings.TrimSpace(varNameSubmatch[1])

			var replacement string
			var lookupVal interface{}
			var found bool

			if varName == "__last_call_result" {
				lookupVal = i.lastCallResult
				found = true // Consider it found even if nil
			} else if isValidIdentifier(varName) {
				lookupVal, found = i.variables[varName]
			} else {
				fmt.Printf("  [Warn] Invalid identifier '%s' inside placeholder: %q\n", varName, match)
				return match // Return original match if identifier invalid
			}

			if found {
				// Convert the found value (could be string, slice, etc.) to its string representation
				replacement = fmt.Sprintf("%v", lookupVal)
				// Mark that a change occurred in this pass
				changed = true
				// Recursively resolve placeholders *within* the replacement itself
				return resolve(replacement, depth+1)
			} else {
				fmt.Printf("  [Warn] Variable '%s' not found for placeholder: %q\n", varName, match)
				return match // Return original placeholder if variable not found
			}
		})

		// If any replacement happened in this pass, we might need another pass
		// if the replacement itself contained more placeholders.
		if changed {
			// Recurse on the entire result string *if* it still contains placeholders
			// This check prevents infinite loops if a var resolves to itself like {{var}}
			if strings.Contains(result, "{{") && strings.Contains(result, "}}") {
				return resolve(result, depth+1)
			}
		}

		return result // Return the final string after this pass
	}

	finalResult := resolve(input, 0)
	// Special handling for literal __last_call_result outside of {{}} - should be minimal now
	// This was causing issues before, let's rely on {{__last_call_result}} primarily
	// finalResult = strings.ReplaceAll(finalResult, "__last_call_result", fmt.Sprintf("%v", i.lastCallResult))

	return finalResult
}

// resolveValue - Resolves ONLY direct variable names & __last_call_result. Returns raw value.
// Does NOT handle literals or placeholders directly.
func (i *Interpreter) resolveValue(input string) (value interface{}, found bool) {
	trimmedInput := strings.TrimSpace(input)

	if trimmedInput == "__last_call_result" {
		// Return last result (or "" if nil), indicate found=true
		if i.lastCallResult != nil {
			return i.lastCallResult, true
		}
		return "", true // Treat as found, value is ""
	}

	if isValidIdentifier(trimmedInput) {
		val, exists := i.variables[trimmedInput]
		if exists {
			// fmt.Printf("      [ResolveValue] Direct lookup var '%s' returning raw type %T\n", trimmedInput, val)
			return val, true // Return raw value and found=true
		}
	}

	// If not a keyword or known variable, indicate not found
	// The input itself might be returned by the caller if needed
	// fmt.Printf("      [ResolveValue] '%s' not keyword or known var.\n", trimmedInput)
	return nil, false
}

// splitExpression (Unchanged)
func splitExpression(expr string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	placeholderLevel := 0
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
		if ch == '+' && !inQuotes && placeholderLevel == 0 {
			trimmedPart := strings.TrimSpace(current.String())
			if trimmedPart != "" {
				parts = append(parts, trimmedPart)
			}
			parts = append(parts, "+")
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}
	trimmedPart := strings.TrimSpace(current.String())
	if trimmedPart != "" {
		parts = append(parts, trimmedPart)
	}
	finalParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			finalParts = append(finalParts, p)
		}
	}
	return finalParts
}

// isValidIdentifier (Make sure this helper exists, e.g., from parser_c.go)
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
	return true
}

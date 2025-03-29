package core

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// --- Evaluation Logic Helpers ---

// evaluateCondition (Unchanged)
func (i *Interpreter) evaluateCondition(conditionStr string) (bool, error) {
	trimmedCond := strings.TrimSpace(conditionStr)
	partsEq := strings.SplitN(trimmedCond, "==", 2)
	if len(partsEq) == 2 {
		lhs := strings.TrimSpace(partsEq[0])
		rhs := strings.TrimSpace(partsEq[1])
		resolvedLhsVal := i.evaluateExpression(lhs)
		resolvedRhsVal := i.evaluateExpression(rhs)
		resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
		resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
		return resolvedLhsStr == resolvedRhsStr, nil
	}
	partsNeq := strings.SplitN(trimmedCond, "!=", 2)
	if len(partsNeq) == 2 {
		lhs := strings.TrimSpace(partsNeq[0])
		rhs := strings.TrimSpace(partsNeq[1])
		resolvedLhsVal := i.evaluateExpression(lhs)
		resolvedRhsVal := i.evaluateExpression(rhs)
		resolvedLhsStr := fmt.Sprintf("%v", resolvedLhsVal)
		resolvedRhsStr := fmt.Sprintf("%v", resolvedRhsVal)
		return resolvedLhsStr != resolvedRhsStr, nil
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

// resolvePlaceholders - ** CORRECTED literal __last_call_result handling **
func (i *Interpreter) resolvePlaceholders(input string) string {
	// Optimization: if no placeholders likely, return early
	if !strings.Contains(input, "{{") && !strings.Contains(input, "__last_call_result") {
		return input
	}

	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`) // Regex for {{...}}
	literalLastCall := "__last_call_result"      // Literal string to replace

	const maxDepth = 10
	var resolveRecursive func(s string, depth int) string

	resolveRecursive = func(s string, depth int) string {
		if depth > maxDepth {
			fmt.Printf("  [Warn] Max placeholder recursion depth (%d) exceeded for: %q\n", maxDepth, input)
			return s
		}

		madeChangeThisPass := false
		current := s

		// --- Stage 1: Replace {{...}} placeholders ---
		next := reVar.ReplaceAllStringFunc(current, func(match string) string {
			// ... (Inner logic for {{...}} replacement remains the same as previous correct version) ...
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			}
			varName := strings.TrimSpace(varNameSubmatch[1])

			var replacement string
			found := false

			if varName == "__last_call_result" {
				if i.lastCallResult != nil {
					replacement = fmt.Sprintf("%v", i.lastCallResult)
					found = true
				} else {
					fmt.Printf("  [Warn] Evaluating {{__last_call_result}} before CALL\n")
					replacement = ""
					found = true
				}
			} else if isValidIdentifier(varName) {
				if value, exists := i.variables[varName]; exists {
					replacement = fmt.Sprintf("%v", value)
					found = true
				} else {
					fmt.Printf("  [Warn] Variable '%s' not found for %q\n", varName, match)
					replacement = match
					found = false
				}
			} else {
				fmt.Printf("  [Warn] Invalid content '%s' in %q\n", varName, match)
				replacement = match
				found = false
			}

			if found {
				madeChangeThisPass = true
				return resolveRecursive(replacement, depth+1) // Recurse on replacement
			} else {
				return replacement
			} // Return original match
		})

		// Update current string after {{...}} replacements
		current = next

		// --- Stage 2: Replace literal __last_call_result ---
		// This needs to handle cases where the literal appears *after* {{...}} subs
		if strings.Contains(current, literalLastCall) {
			var replacementValue string
			if i.lastCallResult != nil {
				// Recursively resolve placeholders within the replacement value *once*
				replacementValue = resolveRecursive(fmt.Sprintf("%v", i.lastCallResult), depth+1)
			} else {
				fmt.Printf("  [Warn] Evaluating literal __last_call_result before CALL\n")
				replacementValue = ""
			}
			// Use simple ReplaceAll for the literal string
			next = strings.ReplaceAll(current, literalLastCall, replacementValue)
			if next != current { // Check if ReplaceAll actually changed something
				madeChangeThisPass = true
				current = next
			}
		}

		// If any changes were made in *either stage*, recurse on the final result of this pass
		if madeChangeThisPass && current != s {
			return resolveRecursive(current, depth+1)
		}

		return current // No changes in this full pass, return final result
	}

	return resolveRecursive(input, 0)
}

// resolveValue (Unchanged)
func (i *Interpreter) resolveValue(input string) interface{} {
	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "__last_call_result" {
		if i.lastCallResult != nil {
			return i.lastCallResult
		}
		fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
		return ""
	}
	isPlainVarName := false
	if isValidIdentifier(trimmedInput) {
		if !strings.Contains(trimmedInput, "{{") {
			isPlainVarName = true
		}
	}
	if isPlainVarName {
		if val, exists := i.variables[trimmedInput]; exists {
			return val
		}
		fmt.Printf("  [Warn] Variable '%s' not found, treating as literal.\n", trimmedInput)
		return trimmedInput
	}
	return i.resolvePlaceholders(trimmedInput)
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

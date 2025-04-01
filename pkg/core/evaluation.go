package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// --- Evaluation Logic Helpers ---

// tryParseFloat attempts to parse a string as float64.
func tryParseFloat(s string) (float64, bool) {
	val, err := strconv.ParseFloat(s, 64)
	return val, err == nil
}

// evaluateCondition evaluates an AST node intended to be a condition.
// Returns true if the node evaluates to true (bool), non-zero (numeric), or "true" (string).
// Returns false otherwise. Logs warnings for non-boolean/numeric types.
func (i *Interpreter) evaluateCondition(condNode interface{}) (bool, error) {
	// First, evaluate the expression represented by the node
	// Note: This doesn't handle comparison operators yet (e.g., x == y).
	// It evaluates the primary expression (like 'x' in 'x == y')
	evaluatedValue := i.evaluateExpression(condNode)

	switch v := evaluatedValue.(type) {
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0.0, nil // Explicitly compare float to 0.0
	case string:
		lowerV := strings.ToLower(v)
		if lowerV == "true" {
			return true, nil
		}
		if lowerV == "false" {
			return false, nil
		}
		// Fallthrough: Non-boolean string is treated as false, but log warning
	}

	// Log warning for types that aren't directly true/false/zero/non-zero
	// This includes non-"true"/"false" strings, slices, maps, nil etc.
	logMsg := fmt.Sprintf("condition evaluated to non-boolean/numeric value: %T (%v)", evaluatedValue, evaluatedValue)
	if i.logger != nil {
		i.logger.Printf("[WARN] %s", logMsg)
	}
	// Treat non-interpretable conditions as false for control flow purposes
	return false, nil // Return false, but no error that stops execution
}

// resolvePlaceholders - REFINED to handle nested resolution more explicitly.
func (i *Interpreter) resolvePlaceholders(input string) string {
	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`)
	const maxDepth = 10
	originalInput := input

	var resolve func(s string, depth int) string
	resolve = func(s string, depth int) string {
		if depth > maxDepth {
			if i.logger != nil {
				i.logger.Printf("[WARN] Placeholder resolution exceeded max depth (%d) for: %q", maxDepth, originalInput)
			}
			return s // Return current string at max depth
		}

		madeChangeInPass := false
		resolvedString := reVar.ReplaceAllStringFunc(s, func(match string) string {
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			} // Invalid format

			varName := strings.TrimSpace(varNameSubmatch[1])
			var nodeToEval interface{}
			var found bool

			if varName == "__last_call_result" {
				nodeToEval = LastCallResultNode{}
				found = true // Assume conceptually found
			} else if isValidIdentifier(varName) {
				nodeToEval = VariableNode{Name: varName}
				_, found = i.variables[varName]
			} else {
				if i.logger != nil {
					i.logger.Printf("[WARN] Invalid identifier '%s' inside placeholder: %s", varName, match)
				}
				return match // Return literal match
			}

			if found {
				evaluatedValue := i.evaluateExpression(nodeToEval)
				replacement := fmt.Sprintf("%v", evaluatedValue) // Convert evaluated value to string

				// --- Check if the replacement itself needs further resolution ---
				if replacement != match && strings.Contains(replacement, "{{") {
					recursiveReplacement := resolve(replacement, depth+1) // Recurse NOW
					if recursiveReplacement != replacement {
						madeChangeInPass = true
					}
					return recursiveReplacement
				} else if replacement != match {
					madeChangeInPass = true
					return replacement
				} else {
					// No change or value stringifies to the match itself
					return match
				}
			} else {
				// Variable not found for placeholder
				if i.logger != nil {
					i.logger.Printf("[INFO] Placeholder variable '{{%s}}' not found.", varName)
				}
				return match // Leave placeholder as is
			}
		})

		// Re-run resolve on the *entire* string *if* changes were made in the pass,
		// but limit recursion.
		if madeChangeInPass && strings.Contains(resolvedString, "{{") && resolvedString != s {
			return resolve(resolvedString, depth+1) // Only recurse if string actually changed
		}

		return resolvedString
	}

	return resolve(input, 0)
}

// isValidIdentifier checks if a string is a valid NeuroScript identifier (and not a keyword).
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
	// Check against keywords (case-insensitive)
	upperName := strings.ToUpper(name)
	// Define keywords explicitly or read from lexer symbols if possible
	// Note: Adding all keywords from the grammar here
	keywords := map[string]bool{
		"DEFINE": true, "SPLAT": true, "PROCEDURE": true, "COMMENT": true, "END": true, "ENDCOMMENT": true, "ENDBLOCK": true,
		"SET": true, "CALL": true, "RETURN": true, "EMIT": true,
		"IF": true, "THEN": true, "ELSE": true,
		"WHILE": true, "DO": true,
		"FOR": true, "EACH": true, "IN": true,
		"TOOL": true, "LLM": true,
		"__LAST_CALL_RESULT": true, // Treat as keyword for identifier checks
	}
	if keywords[upperName] {
		return false // It's a keyword, not a valid *variable* identifier
	}
	// Allow __last_call_result specifically as it's handled like a variable lookup
	// This logic might seem contradictory, but it prevents users defining a variable named 'IF',
	// while still allowing lookup of the special result variable.
	if name == "__last_call_result" {
		return true
	}
	return true
}

// evaluateExpression evaluates an AST node representing an expression.
// It handles literals, variables, placeholders, concatenations, lists, maps.
func (i *Interpreter) evaluateExpression(node interface{}) interface{} {

	switch n := node.(type) {
	// --- Handle AST Node Types ---
	case StringLiteralNode:
		return i.resolvePlaceholders(n.Value)
	case NumberLiteralNode:
		return n.Value // Return the stored int64 or float64
	case BooleanLiteralNode:
		return n.Value
	case VariableNode:
		val, exists := i.variables[n.Name]
		if exists {
			// If the retrieved value is a string, resolve placeholders within it *now*.
			if strVal, ok := val.(string); ok {
				return i.resolvePlaceholders(strVal)
			}
			return val // Return raw value (could be slice, map, number, bool etc.)
		}
		if i.logger != nil {
			i.logger.Printf("[WARN] Variable '%s' not found during evaluation.", n.Name)
		}
		return nil
	case PlaceholderNode: // Placeholders are primarily resolved within strings, but handle direct eval too
		var refValue interface{}
		if n.Name == "__last_call_result" {
			refValue = i.lastCallResult
		} else {
			val, exists := i.variables[n.Name]
			if !exists {
				if i.logger != nil {
					i.logger.Printf("[WARN] Variable '{{%s}}' referenced in placeholder not found.", n.Name)
				}
				return nil // Return nil if underlying variable not found
			}
			refValue = val
		}
		// If the referenced value IS a string, resolve placeholders inside it as well
		if strVal, ok := refValue.(string); ok {
			return i.resolvePlaceholders(strVal)
		}
		// If not a string, return the raw value (number, bool, slice, map...)
		return refValue
	case LastCallResultNode: // Similar to PlaceholderNode for __last_call_result
		if strVal, ok := i.lastCallResult.(string); ok {
			return i.resolvePlaceholders(strVal)
		}
		return i.lastCallResult

	case ConcatenationNode:
		var builder strings.Builder
		for _, operandNode := range n.Operands {
			evaluatedOperand := i.evaluateExpression(operandNode)
			// --- FIX: Convert operand to string before appending ---
			builder.WriteString(fmt.Sprintf("%v", evaluatedOperand))
			// --- END FIX ---
		}
		return builder.String()

	case ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx] = i.evaluateExpression(elemNode)
		}
		// Return the evaluated slice
		return evaluatedElements

	case MapLiteralNode:
		evaluatedMap := make(map[string]interface{})
		for _, entry := range n.Entries {
			// Key is already a string literal node, use its value
			mapKey := entry.Key.Value                     // Already unquoted during AST build
			mapValue := i.evaluateExpression(entry.Value) // Evaluate the value node
			evaluatedMap[mapKey] = mapValue
		}
		// Return the evaluated map
		return evaluatedMap

	// --- Handle Raw Go Types (if passed directly, e.g., from tool result) ---
	case string:
		// If a raw string is passed, assume it might need placeholder resolution too
		return i.resolvePlaceholders(n)
	case int64, float64, bool, nil:
		// Pass through basic types without modification
		return n
	case []interface{}: // Pass through slices
		return n
	case map[string]interface{}: // Pass through maps
		return n

	default:
		// Log an error for unexpected types
		if i.logger != nil {
			i.logger.Printf("[ERROR] evaluateExpression encountered unexpected type: %T (%+v)", node, node)
		}
		return nil // Return nil for unhandled types
	}
}

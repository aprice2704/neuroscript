// pkg/core/evaluation_resolve.go
package core

import (
	"fmt"
	"regexp"
	"strings"
)

// resolvePlaceholdersWithError recursively substitutes {{variable}} placeholders within a string.
// This version returns an error if a referenced variable is not found or max depth is exceeded.
func (i *Interpreter) resolvePlaceholdersWithError(input string) (string, error) {
	var firstError error
	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`)
	const maxDepth = 10
	originalInput := input // Keep original for logging/errors

	var resolve func(s string, depth int) string
	resolve = func(s string, depth int) string {
		// --- Priority 1: Check depth BEFORE any processing ---
		if depth > maxDepth {
			if firstError == nil {
				firstError = fmt.Errorf("placeholder resolution exceeded max depth (%d)", maxDepth)
				if i.logger != nil {
					i.logger.Printf("[WARN] %v for input starting with: %q", firstError, originalInput[:min(len(originalInput), 50)])
				}
			}
			return s // Return original string immediately
		}
		// --- Priority 2: Check if error already set ---
		if firstError != nil {
			return s
		}

		madeChangeThisPass := false

		resolvedString := reVar.ReplaceAllStringFunc(s, func(match string) string {
			// --- Check error inside closure ---
			if firstError != nil {
				return match
			}

			// Extract variable name
			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			}
			varName := strings.TrimSpace(varNameSubmatch[1])

			var nodeToEval interface{}
			var found bool

			if varName == "__last_call_result" {
				nodeToEval = LastCallResultNode{}
				found = true
			} else if isValidIdentifier(varName) {
				if _, exists := i.variables[varName]; exists {
					nodeToEval = VariableNode{Name: varName}
					found = true
				}
			} else { // Invalid identifier
				if firstError == nil {
					firstError = fmt.Errorf("invalid identifier '%s' inside placeholder '%s'", varName, match)
					if i.logger != nil {
						i.logger.Printf("[WARN] %v", firstError)
					}
				}
				return match
			}

			if found {
				evaluatedValue, evalErr := i.evaluateExpression(nodeToEval)
				if evalErr != nil {
					if firstError == nil {
						firstError = fmt.Errorf("evaluating placeholder '{{%s}}': %w", varName, evalErr)
					}
					return match
				}

				var replacement string
				if evaluatedValue == nil {
					replacement = ""
				} else {
					replacement = fmt.Sprintf("%v", evaluatedValue)
				}

				if replacement != match {
					madeChangeThisPass = true // Mark change happened
					if strings.Contains(replacement, "{{") {
						// No predictive check here, rely on start-of-call check
						recursiveReplacement := resolve(replacement, depth+1)
						if firstError != nil { // Check if recursion itself failed
							return match
						}
						return recursiveReplacement
					}
					return replacement // No nested placeholders
				}
				return match // No change

			} else { // Variable valid but not found
				if firstError == nil {
					firstError = fmt.Errorf("placeholder variable '{{%s}}' not found", varName)
					if i.logger != nil {
						i.logger.Printf("[INFO] %v during resolution.", firstError)
					}
				}
				return match
			}
		}) // End ReplaceAllStringFunc

		// --- Priority 4: Check error status AFTER ReplaceAllStringFunc ---
		if firstError != nil {
			return s // Return original string if any error occurred
		}

		// --- Simplified Re-run Logic ---
		// If a change was made AND placeholders remain, recurse one more level.
		if madeChangeThisPass && strings.Contains(resolvedString, "{{") {
			// The depth check at the START of the next resolve call handles termination
			return resolve(resolvedString, depth+1)
		}
		// --- End Simplified Re-run Logic ---

		return resolvedString // Return fully resolved string for this level
	}

	finalResult := resolve(input, 0)
	// If an error occurred, return the original input string.
	if firstError != nil {
		return originalInput, firstError
	}
	return finalResult, nil // Return final result and nil error
}

// Helper for logging
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

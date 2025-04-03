// pkg/core/evaluation_resolve.go
package core

import (
	"fmt"
	"regexp"
	"strings"
)

// resolvePlaceholdersWithError iteratively substitutes {{variable}} placeholders within a string.
// Returns an error if a referenced variable is not found, max iterations exceeded, or a cycle is detected.
func (i *Interpreter) resolvePlaceholdersWithError(input string) (string, error) {
	const maxIterations = 10 // Limit iterations
	currentString := input
	originalInput := input

	reVar := regexp.MustCompile(`\{\{(.*?)\}\}`)

	for iteration := 0; iteration < maxIterations; iteration++ {
		var firstErrorThisPass error = nil
		madeChangeThisPass := false
		visitedInThisPass := make(map[string]bool) // Cycle detection for this pass

		nextString := reVar.ReplaceAllStringFunc(currentString, func(match string) string {
			if firstErrorThisPass != nil {
				return match
			} // Stop if error occurred in this pass

			varNameSubmatch := reVar.FindStringSubmatch(match)
			if len(varNameSubmatch) < 2 {
				return match
			}
			varName := strings.TrimSpace(varNameSubmatch[1])

			if visitedInThisPass[varName] { // Cycle detected this pass
				firstErrorThisPass = fmt.Errorf("detected cycle during placeholder resolution pass involving '{{%s}}' in string starting with %q", varName, originalInput[:min(len(originalInput), 50)])
				return match
			}
			visitedInThisPass[varName] = true

			var nodeToEval interface{}
			var found bool
			isLast := false
			if varName == "LAST" {
				nodeToEval = LastNode{}
				found = true
				isLast = true
			} else if isValidIdentifier(varName) {
				_, exists := i.variables[varName]
				if exists {
					nodeToEval = VariableNode{Name: varName}
					found = true
				}
			} else {
				firstErrorThisPass = fmt.Errorf("invalid identifier '%s' inside placeholder '%s'", varName, match)
				return match
			}

			if found {
				evaluatedValue, evalErr := i.evaluateExpression(nodeToEval) // Gets RAW value
				if evalErr != nil {
					placeholderContext := varName
					if isLast {
						placeholderContext = "LAST"
					}
					firstErrorThisPass = fmt.Errorf("evaluating placeholder '{{%s}}': %w", placeholderContext, evalErr)
					return match
				}
				replacement := ""
				if evaluatedValue != nil {
					replacement = fmt.Sprintf("%v", evaluatedValue)
				}
				if replacement != match {
					madeChangeThisPass = true
				}
				return replacement // Return value from this level, iteration handles recursion
			} else { // Variable not found
				firstErrorThisPass = fmt.Errorf("placeholder variable '{{%s}}' not found", varName)
				return match
			}
		}) // End ReplaceAllStringFunc

		if firstErrorThisPass != nil {
			return originalInput, firstErrorThisPass
		} // Return original on error
		if !madeChangeThisPass {
			return nextString, nil
		} // Success: Done if no changes this pass

		currentString = nextString
		if !strings.Contains(currentString, "{{") {
			return currentString, nil
		} // Success: Done if no placeholders left
	} // End for loop

	// If loop finishes, max iterations were exceeded
	return originalInput, fmt.Errorf("placeholder resolution exceeded max iterations (%d) for input starting with: %q", maxIterations, originalInput[:min(len(originalInput), 50)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

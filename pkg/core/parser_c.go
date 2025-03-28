// pkg/core/parser_c.go
package core

import (
	"fmt"
	"strings"
	"unicode"
)

// parseForHeader parses only "FOR EACH var IN collection DO", returns Step{Value: nil}
func parseForHeader(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	upperLine := strings.ToUpper(line)

	forEachPrefixLen := len("FOR EACH ")
	if !strings.HasPrefix(upperLine, "FOR EACH ") {
		return Step{}, fmt.Errorf("malformed FOR EACH statement (must start with 'FOR EACH '): %q", originalLine)
	}

	// Find ' IN '
	inIndex := findKeywordIndex(line, "IN", forEachPrefixLen)
	if inIndex == -1 {
		return Step{}, fmt.Errorf("missing ' IN ' keyword (with spaces) in FOR EACH statement: %q", originalLine)
	}
	if inIndex < forEachPrefixLen {
		return Step{}, fmt.Errorf("missing or invalid loop variable name after FOR EACH: %q", originalLine)
	}

	// Extract loop variable
	loopVar := strings.TrimSpace(line[forEachPrefixLen:inIndex])
	if !isValidIdentifier(loopVar) {
		return Step{}, fmt.Errorf("missing or invalid loop variable name '%s' after FOR EACH: %q", loopVar, originalLine)
	}

	// Find ' DO' - must be end of line now
	inKeywordLen := len(" IN ")
	collectionStartIndex := inIndex + inKeywordLen
	doIndex := findKeywordIndex(line, "DO", collectionStartIndex)
	if doIndex == -1 {
		return Step{}, fmt.Errorf("missing ' DO' keyword at the end of FOR EACH line: %q", originalLine)
	}

	// Check if anything follows DO on the same line
	if strings.TrimSpace(line[doIndex+len(" DO"):]) != "" {
		return Step{}, fmt.Errorf("unexpected content after ' DO' on FOR EACH line (body must be on subsequent lines): %q", originalLine)
	}

	// Check indices for collection slicing
	if doIndex < collectionStartIndex {
		return Step{}, fmt.Errorf("missing collection expression after IN: %q", originalLine)
	}

	// Extract collection expression
	collectionExpr := strings.TrimSpace(line[collectionStartIndex:doIndex])
	// Collection being empty should be allowed, check happens at runtime

	// Return Step with Type "FOR", Target=loopVar, Cond=collectionExpr, Value=nil
	return Step{Type: "FOR", Target: loopVar, Cond: collectionExpr, Value: nil}, nil
}

// --- Utility Functions ---
// (isValidIdentifier, splitParams, isValidCallTarget, parseCallArgs, trimComments, findCharOutsideQuotes, findKeywordIndex, findMatchingParen, isInsideQuotes unchanged)
// ... KEEP ALL EXISTING UTILITY FUNCTIONS FROM parser_c.go HERE ...
// isValidIdentifier checks if a string is a valid NeuroScript identifier.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			// First char must be letter or underscore
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			// Subsequent chars can be letter, digit, or underscore
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	// Check if it's a reserved keyword
	upperName := strings.ToUpper(name)
	keywords := map[string]bool{
		"DEFINE": true, "PROCEDURE": true, "COMMENT": true, "END": true,
		"SET": true, "CALL": true, "RETURN": true, "IF": true, "THEN": true,
		"ELSE": true, "WHILE": true, "DO": true, "FOR": true, "EACH": true,
		"IN": true, "TOOL": true, "LLM": true,
	}
	if keywords[upperName] {
		return false // Cannot use keywords as identifiers
	}
	return true
}

// splitParams splits a comma-separated parameter list string.
func splitParams(paramStr string) []string {
	trimmedParamStr := strings.TrimSpace(paramStr)
	if trimmedParamStr == "" {
		return []string{}
	}
	parts := strings.Split(trimmedParamStr, ",")
	params := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmedParam := strings.TrimSpace(p)
		if trimmedParam != "" {
			params = append(params, trimmedParam)
		} else {
			fmt.Printf("[Warn] Ignoring empty parameter name resulting from '%s'\n", paramStr)
		}
	}
	return params
}

// isValidCallTarget checks if a string is a valid target for CALL (Name or TOOL.Name or LLM).
func isValidCallTarget(name string) bool {
	if name == "" {
		return false
	}
	if name == "LLM" {
		return true
	}
	if strings.HasPrefix(name, "TOOL.") {
		toolFuncName := name[len("TOOL."):]
		return isValidIdentifier(toolFuncName)
	}
	if isValidIdentifier(name) {
		return true
	}
	return false
}

// parseCallArgs parses the arguments string from a CALL statement.
func parseCallArgs(argsStr string) ([]string, error) {
	trimmedArgsStr := strings.TrimSpace(argsStr)
	if trimmedArgsStr == "" {
		return []string{}, nil // No arguments
	}

	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	parenLevel := 0 // Track parentheses within args if needed later

	for i, ch := range argsStr {
		if escapeNext {
			current.WriteRune(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' {
			current.WriteRune(ch)
			if i+1 < len(argsStr) && (argsStr[i+1] == '"' || argsStr[i+1] == '\'' || argsStr[i+1] == '\\') {
				escapeNext = true
			}
			continue
		}

		switch {
		case (ch == '"' || ch == '\'') && !inQuotes: // Start quote
			inQuotes = true
			quoteChar = ch
			current.WriteRune(ch)
		case ch == quoteChar && inQuotes: // End quote
			inQuotes = false
			current.WriteRune(ch)
		case ch == ',' && !inQuotes && parenLevel == 0: // Separator outside quotes/parens
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset() // Reset for next argument
		case ch == '(' && !inQuotes: // Track parentheses if needed
			parenLevel++
			current.WriteRune(ch)
		case ch == ')' && !inQuotes: // Track parentheses if needed
			if parenLevel > 0 {
				parenLevel--
			}
			current.WriteRune(ch)
		default: // Any other character
			if current.Len() == 0 && len(args) > 0 && unicode.IsSpace(ch) {
				continue
			}
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 || len(args) > 0 || argsStr != "" {
		args = append(args, strings.TrimSpace(current.String()))
	}

	if inQuotes {
		return nil, fmt.Errorf("mismatched quotes in argument list: %s", argsStr)
	}
	// Ignore parenLevel != 0 for now

	return args, nil
}

// trimComments removes NeuroScript comments (# or --) from a line, respecting quotes.
func trimComments(line string) string {
	commentIdx := -1
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false

	for i, ch := range line {
		if escapeNext {
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			continue
		}

		if (ch == '"' || ch == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = ch
		} else if ch == quoteChar && inQuotes {
			inQuotes = false
		} else if !inQuotes {
			if ch == '#' {
				commentIdx = i
				break
			}
			if ch == '-' && i+1 < len(line) && line[i+1] == '-' {
				commentIdx = i
				break
			}
		}
	}

	if commentIdx != -1 {
		return line[:commentIdx]
	}
	return line
}

// findCharOutsideQuotes finds the first index of a character outside quotes.
func findCharOutsideQuotes(line string, char rune) int {
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	for i, r := range line {
		if escapeNext {
			escapeNext = false
			continue
		}
		if r == '\\' {
			escapeNext = true
			continue
		}
		if (r == '"' || r == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = r
		} else if r == quoteChar && inQuotes {
			inQuotes = false
		} else if r == char && !inQuotes {
			return i
		}
	}
	return -1
}

// findKeywordIndex finds the starting index of the space *before* a keyword
// (case-insensitive, surrounded by spaces or line ends)
// within a line *after* a given start position, skipping matches inside quotes. Returns -1 if not found.
// **MODIFIED: Looks for keyword potentially at END of line**
func findKeywordIndex(line string, keyword string, searchStartOffset int) int {
	upperLine := strings.ToUpper(line)
	upperKeyword := strings.ToUpper(keyword)
	searchStart := searchStartOffset

	for searchStart < len(upperLine) {
		relativeIndex := strings.Index(upperLine[searchStart:], upperKeyword)
		if relativeIndex == -1 {
			return -1
		}

		absKeywordStartIndex := searchStart + relativeIndex

		// Check boundaries: Must be preceded by space (or start of line after offset)
		precededBySpace := (absKeywordStartIndex == searchStartOffset) || (absKeywordStartIndex > 0 && unicode.IsSpace(rune(line[absKeywordStartIndex-1])))

		// Modified Check: followed by space OR end of line
		isEndOfLine := absKeywordStartIndex+len(keyword) == len(line)
		followedBySpaceOrEnd := isEndOfLine || (absKeywordStartIndex+len(keyword) < len(line) && unicode.IsSpace(rune(line[absKeywordStartIndex+len(keyword)])))

		if precededBySpace && followedBySpaceOrEnd {
			if !isInsideQuotes(line, absKeywordStartIndex) {
				// Return index of the space *before* the keyword (if keyword isn't at start)
				if absKeywordStartIndex > 0 { // Need space before unless it's start of search offset?
					if unicode.IsSpace(rune(line[absKeywordStartIndex-1])) {
						return absKeywordStartIndex - 1
					}
					// Keyword found but not preceded by required space, continue search
				} else if absKeywordStartIndex == 0 && searchStartOffset == 0 {
					// Keyword at very beginning, no preceding space possible
					// This might be valid for some keywords but not THEN/DO/IN
					// Let's treat it as not found for THEN/DO/IN context
					// For other potential uses, might return 0. For now, return -1.
					return -1
				}
			}
		}
		searchStart = absKeywordStartIndex + 1
	}
	return -1
}

// findMatchingParen finds the index of the matching closing parenthesis for an opening one at startIndex.
func findMatchingParen(line string, startIndex int) int {
	if startIndex < 0 || startIndex >= len(line) || line[startIndex] != '(' {
		return -1
	}

	level := 0
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false

	for i := startIndex; i < len(line); i++ {
		r := rune(line[i])
		if escapeNext {
			escapeNext = false
			continue
		}
		if r == '\\' {
			escapeNext = true
			continue
		}

		if (r == '"' || r == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = r
		} else if r == quoteChar && inQuotes {
			inQuotes = false
		} else if r == '(' && !inQuotes {
			level++
		} else if r == ')' && !inQuotes {
			level--
			if level == 0 {
				return i
			}
			if level < 0 {
				return -1
			}
		}
	}
	return -1
}

// isInsideQuotes checks if a given index in a line is inside quotes.
func isInsideQuotes(line string, index int) bool {
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	for i, r := range line {
		if i >= index {
			break
		}

		if escapeNext {
			escapeNext = false
			continue
		}
		if r == '\\' {
			escapeNext = true
			continue
		}
		if (r == '"' || r == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = r
		} else if r == quoteChar && inQuotes {
			inQuotes = false
		}
	}
	return inQuotes
}

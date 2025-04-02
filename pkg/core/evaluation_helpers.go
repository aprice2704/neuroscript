// pkg/core/evaluation_helpers.go
package core

import (
	"strconv"
	"strings"
	"unicode"
)

// tryParseFloat attempts to parse a string as float64.
func tryParseFloat(s string) (float64, bool) {
	val, err := strconv.ParseFloat(s, 64)
	return val, err == nil
}

// isValidIdentifier checks if a string is a valid NeuroScript identifier (and not a keyword).
// Moved from evaluation.go during refactoring.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for idx, r := range name {
		if idx == 0 {
			// Must start with a letter or underscore
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			// Subsequent characters can be letters, digits, or underscores
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	upperName := strings.ToUpper(name)
	// Using map for efficient keyword lookup
	keywords := map[string]bool{
		"DEFINE":             true, // Deprecated? Replaced by SPLAT? Keep for compatibility?
		"SPLAT":              true,
		"PROCEDURE":          true,
		"COMMENT":            true, // Part of COMMENT_BLOCK
		"END":                true,
		"ENDCOMMENT":         true, // Part of COMMENT_BLOCK
		"ENDBLOCK":           true,
		"SET":                true,
		"CALL":               true,
		"RETURN":             true,
		"EMIT":               true,
		"IF":                 true,
		"THEN":               true,
		"ELSE":               true,
		"WHILE":              true,
		"DO":                 true,
		"FOR":                true,
		"EACH":               true,
		"IN":                 true,
		"TOOL":               true,
		"LLM":                true,
		"TRUE":               true, // Treat true/false as keywords
		"FALSE":              true,
		"__LAST_CALL_RESULT": true, // Treat special var as keyword-like
	}
	if keywords[upperName] {
		// Allow __last_call_result specifically even though it's keyword-like
		return name == "__last_call_result"
	}
	// If not a keyword, it's a valid identifier structure
	return true
}

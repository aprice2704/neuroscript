// filename: pkg/core/utils.go
package core

import (
	"strings"
)

// --- Utility Helpers ---

// FIX: Added parseMetadataLine helper
// parseMetadataLine extracts the key and value from a metadata line content.
// Assumes input is the raw text *after* the initial ":: ".
// Returns key, value, and ok=true if successful.
func parseMetadataLine(lineContent string) (key string, value string, ok bool) {
	lineContent = strings.TrimSpace(lineContent)
	parts := strings.SplitN(lineContent, ":", 2)
	if len(parts) != 2 {
		// Malformed line, missing colon or key/value
		return "", "", false
	}
	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])
	if key == "" {
		// Key cannot be empty
		return "", "", false
	}
	ok = true
	return
}

// trimCodeFences removes code fences (```) from the beginning and end of a code string.
// Handles optional language identifiers and whitespace.
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return trimmed
	}
	startFenceFound := false
	endFenceFound := false
	firstLineTrimmed := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLineTrimmed, "```") {
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
		isValidLangID := true
		if len(restOfLine) > 0 {
			for _, r := range restOfLine {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
					isValidLangID = false
					break
				}
			}
		}
		if isValidLangID {
			startFenceFound = true
			lines = lines[1:]
		}
	}
	if len(lines) > 0 {
		lastLineTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastLineTrimmed == "```" {
			endFenceFound = true
			lines = lines[:len(lines)-1]
		}
	}
	if startFenceFound || endFenceFound {
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return trimmed
}

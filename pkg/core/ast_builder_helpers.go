// filename: pkg/core/ast_builder_helpers.go
package core

import (
	"strings"
)

// ParseMetadataLine attempts to parse a line potentially containing metadata (e.g., ":: key: value").
// It returns the extracted key, value, and a boolean indicating if the line was a valid metadata line.
// Key and value are trimmed of whitespace.
func ParseMetadataLine(line string) (key string, value string, ok bool) {
	trimmedLine := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmedLine, "::") {
		return "", "", false // Not a metadata line
	}

	// Remove "::" prefix and trim surrounding space
	content := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "::"))

	// Find the first colon
	colonIndex := strings.Index(content, ":")
	if colonIndex == -1 {
		// Treat as a key-only metadata line (value is empty)
		key = strings.TrimSpace(content)
		value = ""
		return key, value, true
		// Alternatively, consider this invalid: return "", "", false
	}

	// Extract key and value based on the first colon
	key = strings.TrimSpace(content[:colonIndex])
	value = strings.TrimSpace(content[colonIndex+1:])

	// Basic validation: key cannot be empty
	if key == "" {
		return "", "", false
	}

	return key, value, true
}

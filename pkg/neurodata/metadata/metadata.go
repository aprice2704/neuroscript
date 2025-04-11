// Package metadata provides functions for extracting structured metadata
// (formatted as ':: key: value') from the beginning of text content,
// typically used for files or embedded code/data blocks.
package metadata

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// Unexported regex patterns
var (
	metadataPattern          = regexp.MustCompile(`^\s*::\s+([a-zA-Z0-9_.-]+)\s*:\s*(.*)`)
	commentOrBlankPattern    = regexp.MustCompile(`^\s*($|#|--)`)
	startsWithMetadataPrefix = regexp.MustCompile(`^\s*::`) // Checks for potential start
)

// --- Exported Helper Functions ---

// IsMetadataLine checks if a line *could* be a metadata line (starts with `:: `).
// Note: It doesn't validate the full key:value format. Use ExtractKeyValue for that.
func IsMetadataLine(line string) bool {
	// We check for '::' followed by at least one space, allowing leading whitespace.
	// metadataPattern requires the full structure, StartsWithMetadataPrefix is too loose.
	// Let's refine this check slightly.
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, ":: ")
}

// IsCommentOrBlank checks if a line is a comment (#, --) or blank.
func IsCommentOrBlank(line string) bool {
	// Uses the unexported pattern
	return commentOrBlankPattern.MatchString(line)
}

// ExtractKeyValue attempts to parse a line as a ':: key: value' entry.
// Returns the key, value, and true if successful, otherwise empty strings and false.
func ExtractKeyValue(line string) (key, value string, ok bool) {
	matches := metadataPattern.FindStringSubmatch(line)
	if len(matches) == 3 {
		key = strings.TrimSpace(matches[1])
		value = strings.TrimSpace(matches[2])
		ok = true
		return key, value, ok
	}
	return "", "", false
}

// --- Main Extraction Function ---

// Extract scans the beginning of content for metadata lines using the helper functions.
// It stops at the first line that is not valid metadata, a comment, or blank.
func Extract(content string) (map[string]string, error) {
	metadata := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Use the exported helper to check if it *might* be metadata
		if IsMetadataLine(line) {
			// Try to extract the key/value pair using the stricter pattern
			key, value, ok := ExtractKeyValue(line)
			if ok {
				// Only add if key doesn't exist yet (first wins)
				if _, exists := metadata[key]; !exists {
					metadata[key] = value
				}
				continue // Successfully processed metadata line
			} else {
				// Line started like metadata (`:: `) but was malformed (e.g., no colon)
				// Stop processing metadata here.
				break
			}
		}

		// If not potentially metadata, check if it's a comment or blank line
		if IsCommentOrBlank(line) {
			continue // Skip comments and blank lines within the metadata section
		}

		// If it's not metadata, not a comment, and not blank, then metadata section ends.
		break
	}

	if err := scanner.Err(); err != nil {
		return metadata, fmt.Errorf("error scanning content for metadata: %w", err)
	}

	return metadata, nil
}

// --- Unexported helpers used by the exported Extract function ---
// These are kept unexported as they are implementation details of Extract.

// StartsWithMetadataPrefix (unexported now) checks if a line starts with the potential metadata pattern `::`.
// Note: This is less strict than IsMetadataLine, used internally by Extract if needed,
// but external callers should use IsMetadataLine or ExtractKeyValue.
func startsWithMetadataPrefixFunc(line string) bool {
	return startsWithMetadataPrefix.MatchString(line)
}

// CommentOrBlankPattern (unexported now) checks if a line matches a comment or is blank.
// External callers should use IsCommentOrBlank.
func commentOrBlankPatternFunc(line string) bool {
	return commentOrBlankPattern.MatchString(line)
}

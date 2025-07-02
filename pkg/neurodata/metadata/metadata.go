// filename: pkg/neurodata/metadata/metadata.go
// Package metadata provides functions for extracting structured metadata
// (formatted as ':: key: value') from the beginning of text content,
// typically used for files or embedded code/data blocks.
package metadata

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Define Error within this package
var (
	// ErrMalformedMetadata indicates a line started like metadata (::) but was malformed.
	ErrMalformedMetadata = errors.New("malformed metadata line")
)

// Unexported regex patterns
var (
	// Stricter pattern for extracting key/value, requiring space after ::
	metadataPattern		= regexp.MustCompile(`^\s*::\s+([a-zA-Z0-9_.-]+)\s*:\s*(.*)`)
	commentOrBlankPattern	= regexp.MustCompile(`^\s*($|#|--)`)
	// More lenient pattern just to check if a line *might* be metadata
	startsWithMetadataPrefix	= regexp.MustCompile(`^\s*::`)
)

// --- Exported Helper Functions ---

// IsMetadataLine checks if a line *could* be a metadata line (starts with `::`).
// Allows leading whitespace, does NOT require whitespace after :: for this initial check.
func IsMetadataLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	// CHANGE: Only check for the "::" prefix now
	return strings.HasPrefix(trimmed, "::")
}

// IsCommentOrBlank checks if a line is a comment (#, --) or blank.
func IsCommentOrBlank(line string) bool {
	return commentOrBlankPattern.MatchString(line)
}

// ExtractKeyValue attempts to parse a line as a ':: key: value' entry using the strict pattern.
// Returns the key, value, and true if successful, otherwise empty strings and false.
func ExtractKeyValue(line string) (key, value string, ok bool) {
	matches := metadataPattern.FindStringSubmatch(line)	// Uses the strict regex
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
// It stops and returns an error if a line starts with '::' but is malformed according to ExtractKeyValue.
// It stops normally at the first line that is not potentially metadata, a comment, or blank.
func Extract(content string) (map[string]string, error) {
	metadata := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Use the updated, less strict IsMetadataLine to identify potential metadata lines
		if IsMetadataLine(line) {	// Check for "::" prefix only
			// Now try to extract key/value using the strict pattern
			key, value, ok := ExtractKeyValue(line)
			if ok {
				// Only add if key doesn't exist yet (first wins)
				if _, exists := metadata[key]; !exists {
					metadata[key] = value
				}
				continue	// Successfully processed valid metadata line
			} else {
				// Line started "::" but did not match ":: key: value". Return ERROR.
				err := fmt.Errorf("%w: detected on line %d: %s", ErrMalformedMetadata, lineNumber, line)
				return metadata, err	// Return partially collected metadata AND the error
			}
		}

		// If not potentially metadata, check if it's a comment or blank line
		if IsCommentOrBlank(line) {
			continue	// Skip comments and blank lines within the metadata section
		}

		// If it's not potentially metadata, not comment, not blank, then metadata section ends normally.
		break
	}

	if err := scanner.Err(); err != nil {
		// Return any previously collected metadata along with the scanner error
		return metadata, fmt.Errorf("error scanning content for metadata: %w", err)
	}

	// Return successfully collected metadata and nil error if loop finished normally
	return metadata, nil
}

// --- Unexported helpers ---
// (Keep these unexported as they are implementation details)

func startsWithMetadataPrefixFunc(line string) bool {
	return startsWithMetadataPrefix.MatchString(line)
}

func commentOrBlankPatternFunc(line string) bool {
	return commentOrBlankPattern.MatchString(line)
}
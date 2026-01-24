// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 5
// :: description: Provides utilities for extracting and validating metadata, including the SINGLE SOURCE OF TRUTH for parsing logic.
// :: latestChange: Centralized MetaRegex, ReadLines, ParseHeaderBlock, and ParseFooterBlock to eliminate parser drift.
// :: filename: pkg/metadata/utility.go
// :: serialization: go
package metadata

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// MetaRegex captures the key and value from a metadata line.
// It is the Single Source of Truth for what constitutes a metadata line.
// We use a lenient regex here to allow the parser to read slightly malformed files,
// though the spec enforces strict spacing for validation/linting.
var MetaRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_.-]+)\s*:\s*(.*?)\s*$`)

// Pre-defined sets of required metadata keys for different schemas.
var (
	// RequiredSourceFileKeys are the essential keys for any source file.
	RequiredSourceFileKeys = []string{"schema", "serialization", "fileversion", "description"}
	// RequiredCapsuleKeys are the essential keys for a capsule markdown file.
	RequiredCapsuleKeys = []string{"schema", "serialization", "id", "version", "description"}
)

// keyNormalizeRegex is used to remove characters that are ignored during key matching.
var keyNormalizeRegex = regexp.MustCompile(`[._-]+`)

// NormalizeKey implements the key matching rule from the spec:
// "the case of the letters, and the characters underscore, dot and dash (_.-) are ignored"
func NormalizeKey(key string) string {
	lower := strings.ToLower(key)
	return keyNormalizeRegex.ReplaceAllString(lower, "")
}

// Extractor provides a safe and convenient way to access values from a metadata Store.
// It automatically normalizes keys for lookups.
type Extractor struct {
	store Store
}

// NewExtractor creates a new extractor for a given metadata store.
func NewExtractor(s Store) *Extractor {
	// We create a new store with normalized keys for efficient lookups.
	normalizedStore := make(Store)
	for k, v := range s {
		normalizedStore[NormalizeKey(k)] = v
	}
	return &Extractor{store: normalizedStore}
}

// Get retrieves a value by key. Returns the value and true if the key exists.
func (e *Extractor) Get(key string) (string, bool) {
	val, ok := e.store[NormalizeKey(key)]
	return val, ok
}

// GetOr retrieves a value by key, returning the provided default value if the key is not found.
func (e *Extractor) GetOr(key string, defaultValue string) string {
	if val, ok := e.Get(key); ok {
		return val
	}
	return defaultValue
}

// MustGet retrieves a value by key. It returns the value, or an empty string if not found.
func (e *Extractor) MustGet(key string) string {
	return e.store[NormalizeKey(key)]
}

// GetInt retrieves a value by key and attempts to parse it as an integer.
func (e *Extractor) GetInt(key string) (int, bool, error) {
	val, ok := e.Get(key)
	if !ok {
		return 0, false, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, true, fmt.Errorf("metadata key %q is not a valid integer: %w", key, err)
	}
	return i, true, nil
}

// GetIntOr retrieves a value by key and parses it as an integer, returning the
// provided default value if the key is not found. If the key is found but the value
// is not a valid integer, it returns an error.
func (e *Extractor) GetIntOr(key string, defaultValue int) (int, error) {
	i, ok, err := e.GetInt(key)
	if err != nil {
		return 0, err // Found the key, but it failed to parse.
	}
	if !ok {
		return defaultValue, nil // Key not found, return default.
	}
	return i, nil // Key found and parsed successfully.
}

// CheckRequired verifies that the underlying store contains all the specified required keys.
// It uses the same normalization logic for checking.
func (e *Extractor) CheckRequired(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if _, ok := e.store[NormalizeKey(k)]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required metadata keys: %v", missing)
	}
	return nil
}

// --- Unified Parsing Logic (Single Source of Truth) ---

// ReadLines reads all lines from a reader using a scanner to handle line endings reliably.
// It automatically handles \n and \r\n line endings.
func ReadLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// ParseHeaderBlock extracts a NeuroScript-style metadata block from the beginning of the lines.
// Returns the store and the line index where the content begins (inclusive).
// Returns (nil, 0) if no valid header block is found.
func ParseHeaderBlock(lines []string) (Store, int) {
	store := make(Store)
	metaEndLine := 0
	pastLeadingWhitespace := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if !pastLeadingWhitespace && trimmedLine == "" {
			metaEndLine = i + 1
			continue
		}
		pastLeadingWhitespace = true

		if !MetaRegex.MatchString(trimmedLine) {
			break // End of the contiguous block
		}

		// It matches, so extract it
		matches := MetaRegex.FindStringSubmatch(trimmedLine)
		if len(matches) == 3 {
			key := strings.ToLower(matches[1])
			val := strings.TrimSpace(matches[2])
			store[key] = val
		}

		// This line is metadata, so content starts after it
		metaEndLine = i + 1
	}

	if len(store) == 0 {
		return nil, 0
	}

	return store, metaEndLine
}

// ParseFooterBlock extracts a Markdown-style metadata block from the end of the lines.
// Returns the store and the line index where the content ends (exclusive).
// Returns (nil, len(lines)) if no valid footer block is found.
func ParseFooterBlock(lines []string) (Store, int) {
	store := make(Store)
	metaStartLine := len(lines)
	inBlock := false

	// Scan backwards
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			if inBlock {
				// A blank line breaks the contiguity of the block.
				break
			}
			continue // Skip trailing blank lines
		}

		if MetaRegex.MatchString(trimmedLine) {
			inBlock = true
			metaStartLine = i // Potentially moves up
		} else {
			// Content line found, block ends here.
			break
		}
	}

	// Now parse forward from the identified start
	if inBlock && metaStartLine < len(lines) {
		for i := metaStartLine; i < len(lines); i++ {
			line := lines[i]
			trimmedLine := strings.TrimSpace(line)
			matches := MetaRegex.FindStringSubmatch(trimmedLine)
			if len(matches) == 3 {
				key := strings.ToLower(matches[1])
				val := strings.TrimSpace(matches[2])
				store[key] = val
			}
		}
		return store, metaStartLine
	}

	return nil, len(lines)
}

// NeuroScript Version: 0.3.0
// File version: 4
// Purpose: Implements a metadata parser for Markdown files.
// filename: pkg/metadata/markdown.go
// nlines: 67
// risk_rating: LOW
package metadata

import (
	"io"
	"regexp"
	"strings"
)

// metaRegex captures the key and value, trimming trailing whitespace from the value group.
var metaRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_.-]+)\s*:\s*(.*?)\s*$`)

// MarkdownParser implements the Parser interface for markdown files.
// It expects metadata to be in a block at the very end of the file.
type MarkdownParser struct{}

// NewMarkdownParser creates a new parser for Markdown files.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Parse extracts metadata from the end of a reader's content.
func (p *MarkdownParser) Parse(r io.Reader) (Store, []byte, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	store := make(Store)
	lines := strings.Split(string(content), "\n")

	// Find the start of the metadata block by scanning from the end.
	metaStartLine := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // Skip trailing blank lines
		}
		// A line is part of the metadata block only if it matches the full regex.
		if !metaRegex.MatchString(line) {
			break // Found the last line of content before metadata
		}
		metaStartLine = i
	}

	if metaStartLine == -1 {
		return store, content, nil // No metadata block found
	}

	// Parse the metadata lines
	for i := metaStartLine; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		matches := metaRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := strings.ToLower(matches[1])
			// The regex now handles trailing space, but we still trim the value
			// to handle leading space and to be fully compliant with the spec.
			val := strings.TrimSpace(matches[2])
			store[key] = val
		}
	}

	// Join the content lines before the metadata block
	contentBytes := []byte(strings.Join(lines[:metaStartLine], "\n"))

	return store, contentBytes, nil
}

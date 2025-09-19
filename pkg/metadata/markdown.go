// NeuroScript Version: 0.3.0
// File version: 18
// Purpose: Implements a metadata parser for Markdown files, enforcing that metadata blocks must be contiguous and cannot contain blank lines.
// filename: pkg/metadata/markdown.go
// nlines: 83
// risk_rating: LOW
package metadata

import (
	"io"
	"regexp"
	"strings"
)

// MetaRegex captures the key and value from a metadata line. It is exported for use in other packages.
var MetaRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_.-]+)\s*:\s*(.*?)\s*$`)

func init() {
	RegisterParser("md", func() Parser { return NewMarkdownParser() })
}

// MarkdownParser implements the Parser interface for markdown files.
// It expects metadata to be in a block at the very end of the file.
type MarkdownParser struct{}

// NewMarkdownParser creates a new parser for Markdown files.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Parse extracts metadata from the end of a reader's content.
func (p *MarkdownParser) Parse(r io.Reader) (Store, []byte, error) {
	contentBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	store := make(Store)
	content := strings.ReplaceAll(string(contentBytes), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	//log.Printf("[DEBUG] markdown.Parse: Total lines received: %d", len(lines))

	// Find the start of the last contiguous block of metadata lines.
	// Blank lines are not allowed in the metadata block.
	metaStartLine := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
		isMeta := MetaRegex.MatchString(line)
		//	log.Printf("[DEBUG] line[%d]: q=%q | isMeta=%v", i, line, isMeta)

		if trimmedLine == "" {
			// If we haven't found the block yet, this is just trailing whitespace.
			// If we HAVE found the block, this terminates it.
			if metaStartLine != len(lines) {
				break // Terminate the block if a blank line is found above it.
			}
			continue
		}

		if isMeta {
			metaStartLine = i // This line is part of the block.
		} else {
			// This is a content line, so the block (if any) starts on the next line.
			metaStartLine = i + 1
			break
		}
	}
	//log.Printf("[DEBUG] markdown.Parse: Final metaStartLine: %d", metaStartLine)

	// Parse the metadata lines
	for i := metaStartLine; i < len(lines); i++ {
		line := lines[i]
		// We can now safely assume no blank lines are in the metadata block itself.
		matches := MetaRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := strings.ToLower(matches[1])
			val := strings.TrimSpace(matches[2])
			store[key] = val
			//	log.Printf("[DEBUG] markdown.Parse: Stored meta: {%q: %q}", key, val)
		}
	}

	contentBody := strings.Join(lines[:metaStartLine], "\n")
	contentBody = strings.TrimRight(contentBody, " \t")

	//	log.Printf("[DEBUG] markdown.Parse: Final content returned: %q", contentBody)
	return store, []byte(contentBody), nil
}

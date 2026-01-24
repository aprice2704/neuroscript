// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 19
// :: description: Implements a metadata parser for Markdown files using Unified Parsing Logic.
// :: latestChange: Refactored to use utility.ReadLines and ParseFooterBlock.
// :: filename: pkg/metadata/markdown.go
// :: serialization: go
package metadata

import (
	"io"
	"strings"
)

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
	lines, err := ReadLines(r)
	if err != nil {
		return nil, nil, err
	}

	store, metaStartLine := ParseFooterBlock(lines)
	if store == nil {
		// If no block found, everything is content
		metaStartLine = len(lines)
		store = make(Store)
	}

	contentBody := strings.Join(lines[:metaStartLine], "\n")
	contentBody = strings.TrimRight(contentBody, " \t")

	return store, []byte(contentBody), nil
}

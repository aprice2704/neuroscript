// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 10
// :: description: Implements a metadata parser for NeuroScript (.ns) files using Unified Parsing Logic.
// :: latestChange: Refactored to use utility.ReadLines and ParseHeaderBlock.
// :: filename: pkg/metadata/nsparser.go
// :: serialization: go
package metadata

import (
	"io"
	"strings"
)

func init() {
	RegisterParser("ns", func() Parser { return NewNeuroScriptParser() })
}

// NeuroScriptParser implements the Parser interface for .ns files.
// It expects metadata to be in a block at the very beginning of the file.
type NeuroScriptParser struct{}

// NewNeuroScriptParser creates a new parser for NeuroScript files.
func NewNeuroScriptParser() *NeuroScriptParser {
	return &NeuroScriptParser{}
}

// Parse extracts metadata from the start of a reader's content.
func (p *NeuroScriptParser) Parse(r io.Reader) (Store, []byte, error) {
	lines, err := ReadLines(r)
	if err != nil {
		return nil, nil, err
	}

	store, metaEndLine := ParseHeaderBlock(lines)
	if store == nil {
		// If no block found, nothing is metadata, everything is content (or empty)
		metaEndLine = 0
		store = make(Store)
	}

	var contentLines []string
	if metaEndLine < len(lines) {
		contentLines = lines[metaEndLine:]
	}

	contentBody := strings.Join(contentLines, "\n")
	return store, []byte(contentBody), nil
}

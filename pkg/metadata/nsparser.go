// NeuroScript Version: 0.3.0
// File version: 9
// Purpose: Implements a metadata parser for NeuroScript (.ns) files, correcting a compiler error in the constructor.
// filename: pkg/metadata/nsparser.go
// nlines: 70
// risk_rating: LOW
package metadata

import (
	"bufio"
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
	store := make(Store)
	var allLines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	//	log.Printf("[DEBUG] nsparser.Parse: Total lines received: %d", len(allLines))

	metaEndLine := 0
	pastLeadingWhitespace := false
	for i, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		//		log.Printf("[DEBUG] nsparser.Parse: line[%d] q=%q", i, line)

		if !pastLeadingWhitespace && trimmedLine == "" {
			metaEndLine = i + 1
			continue
		}
		pastLeadingWhitespace = true

		if !MetaRegex.MatchString(trimmedLine) {
			break // End of the contiguous block
		}
		// This line is metadata, so the content starts after it.
		metaEndLine = i + 1
	}
	//	log.Printf("[DEBUG] nsparser.Parse: Final metaEndLine: %d", metaEndLine)

	// Parse the metadata lines
	for i := 0; i < metaEndLine; i++ {
		line := allLines[i]
		trimmedLine := strings.TrimSpace(line)
		if MetaRegex.MatchString(trimmedLine) {
			matches := MetaRegex.FindStringSubmatch(trimmedLine)
			if len(matches) == 3 {
				key := strings.ToLower(matches[1])
				val := strings.TrimSpace(matches[2])
				store[key] = val
				//				log.Printf("[DEBUG] nsparser.Parse: Stored meta: {%q: %q}", key, val)
			}
		}
	}

	var contentLines []string
	if metaEndLine < len(allLines) {
		contentLines = allLines[metaEndLine:]
	}

	contentBody := strings.Join(contentLines, "\n")
	//	log.Printf("[DEBUG] nsparser.Parse: Final content returned: %q", contentBody)
	return store, []byte(contentBody), nil
}

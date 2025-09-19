// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Implements an extensible, auto-detecting parser for different serialization formats.
// filename: pkg/metadata/autodetect.go
// nlines: 115
// risk_rating: LOW
package metadata

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ParserFactory is a function that returns a new instance of a Parser.
type ParserFactory func() Parser

// parserRegistry holds the mapping from serialization format names to their parser factories.
var parserRegistry = make(map[string]ParserFactory)

// RegisterParser makes a Parser available by a given serialization name.
// If RegisterParser is called twice with the same name, it panics.
func RegisterParser(serialization string, factory ParserFactory) {
	if _, dup := parserRegistry[serialization]; dup {
		panic("metadata: RegisterParser called twice for " + serialization)
	}
	parserRegistry[serialization] = factory
}

// NewParserForSerialization returns a new parser for the given format.
func NewParserForSerialization(serialization string) (Parser, error) {
	factory, ok := parserRegistry[serialization]
	if !ok {
		return nil, fmt.Errorf("no metadata parser registered for serialization format %q", serialization)
	}
	return factory(), nil
}

// DetectSerialization reads just enough of a reader to find the '::serialization:' key
// and determine the file's format. It is designed to be efficient by reading
// only the necessary parts of the file.
//
// It gives precedence to a key found in a valid end-of-file block (for formats
// like Markdown), then checks for a key in a valid start-of-file block (for
// formats like NeuroScript). A block is considered valid if it is a contiguous
// set of metadata lines, optionally separated from content by blank lines.
func DetectSerialization(r io.ReadSeeker) (string, error) {
	contentBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(contentBytes), "\n")

	// Check for a markdown-style block at the very end of the file.
	var lastFoundSer string
	inBlock := false
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			if inBlock {
				break
			} // A blank line breaks contiguity.
			continue // Skip trailing blank lines.
		}

		if MetaRegex.MatchString(trimmedLine) {
			inBlock = true
			matches := MetaRegex.FindStringSubmatch(trimmedLine)
			if len(matches) == 3 && NormalizeKey(matches[1]) == "serialization" {
				lastFoundSer = strings.TrimSpace(matches[2])
			}
		} else {
			break // As soon as we hit a content line, the potential block is over.
		}
	}
	if lastFoundSer != "" {
		return lastFoundSer, nil
	}

	// Check for a neuroscript-style block at the very beginning of the file.
	pastWhitespace := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !pastWhitespace && trimmedLine == "" {
			continue // Skip leading blank lines
		}
		pastWhitespace = true

		if !MetaRegex.MatchString(trimmedLine) {
			break // End of contiguous block
		}
		matches := MetaRegex.FindStringSubmatch(trimmedLine)
		if len(matches) == 3 && NormalizeKey(matches[1]) == "serialization" {
			return strings.TrimSpace(matches[2]), nil
		}
	}

	return "", fmt.Errorf("key '::serialization:' not found in a valid start or end block")
}

// ParseWithAutoDetect uses DetectSerialization to determine the file format and then
// parses the content using the appropriate registered parser.
func ParseWithAutoDetect(r io.ReadSeeker) (Store, []byte, string, error) {
	// Read the content once
	contentBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, "", err
	}

	// Detect serialization from the byte slice
	serialization, err := DetectSerialization(bytes.NewReader(contentBytes))
	if err != nil {
		return nil, nil, "", err
	}

	// Get the correct parser
	parser, err := NewParserForSerialization(serialization)
	if err != nil {
		return nil, nil, "", err
	}

	// Parse using the selected parser
	meta, content, err := parser.Parse(bytes.NewReader(contentBytes))
	if err != nil {
		return nil, nil, "", err
	}

	return meta, content, serialization, nil
}

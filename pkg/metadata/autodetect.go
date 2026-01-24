// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 5
// :: description: Implements an extensible, auto-detecting parser for different serialization formats. Refactored to use Unified Parsing Logic and fixed error messages.
// :: latestChange: Corrected error message to avoid misleading spacing in key citation.
// :: filename: pkg/metadata/autodetect.go
// :: serialization: go
package metadata

import (
	"bytes"
	"fmt"
	"io"
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

// DetectSerialization reads the content and uses the shared parsing logic
// to determine the serialization format.
func DetectSerialization(r io.ReadSeeker) (string, error) {
	lines, err := ReadLines(r)
	if err != nil {
		return "", err
	}

	// 1. Check for Footer Block (Markdown style)
	if store, _ := ParseFooterBlock(lines); store != nil {
		if ser, ok := store["serialization"]; ok {
			return ser, nil
		}
	}

	// 2. Check for Header Block (NeuroScript style)
	if store, _ := ParseHeaderBlock(lines); store != nil {
		if ser, ok := store["serialization"]; ok {
			return ser, nil
		}
	}

	// Corrected error message: explicitly state "metadata key 'serialization'"
	// to avoid the confusing "::serialization:" artifact which violates the spacing spec.
	return "", fmt.Errorf("metadata key 'serialization' not found in a valid start or end block")
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

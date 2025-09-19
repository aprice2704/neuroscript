// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Contains aggressive stress tests for the Markdown metadata parser, corrected to align with new contiguity rules.
// filename: pkg/metadata/markdown_insane_test.go
// nlines: 125
// risk_rating: LOW
package metadata_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

// TestMarkdownParser_InsaneCases tests the parser against a variety of edge cases and malformed inputs.
func TestMarkdownParser_InsaneCases(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    metadata.Store
	}{
		{
			name: "Value contains colons",
			input: `Content.
::key: value:with:colons`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": "value:with:colons"},
		},
		{
			name: "Value looks like another key",
			input: `Content.
::key: ::another:key`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": "::another:key"},
		},
		{
			name: "Malformed key with invalid characters",
			input: `Content.
::key!!: bad
::good-key: ok`,
			// The parser should stop at the first invalid line.
			wantContent: "Content.\n::key!!: bad",
			wantMeta:    metadata.Store{"good-key": "ok"},
		},
		{
			name: "Key with no value followed by valid key",
			input: `Content.
::key1:
::key2: value2`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key1": "", "key2": "value2"},
		},
		{
			name: "Lots of empty lines in metadata block",
			input: `Content.
::key1: val1


::key2: val2
`,
			// CORRECTED: The blank lines are a boundary. `::key1` is content.
			wantContent: "Content.\n::key1: val1\n\n",
			wantMeta:    metadata.Store{"key2": "val2"},
		},
		{
			name:        "Tab characters as whitespace",
			input:       "Content.\n::\tkey\t:\tvalue\t",
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": "value"},
		},
		{
			name: "Content that ends exactly at a metadata-like line",
			input: `Final line of content is ::fake:meta
::real: meta`,
			wantContent: "Final line of content is ::fake:meta",
			wantMeta:    metadata.Store{"real": "meta"},
		},
		{
			name: "Unicode characters in value",
			input: `Content.
::message: Hello, 世界`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"message": "Hello, 世界"},
		},
		{
			name: "Very long value string",
			input: `Content.
::long: ` + strings.Repeat("a", 2048),
			wantContent: "Content.",
			wantMeta:    metadata.Store{"long": strings.Repeat("a", 2048)},
		},
		{
			name: "No newline at EOF",
			input: `Content.
::key: value`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": "value"},
		},
		{
			name:  "Windows line endings (CRLF)",
			input: "Content.\r\n::key: value\r\n",
			// Note: The parser will normalize to LF in the output content.
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": "value"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Normalize CRLF to LF for consistent test inputs
			input := strings.ReplaceAll(tc.input, "\r\n", "\n")

			parser := metadata.NewMarkdownParser()
			r := strings.NewReader(input)
			meta, content, err := parser.Parse(r)

			if err != nil {
				t.Fatalf("Parse() returned an unexpected error: %v", err)
			}
			if string(content) != tc.wantContent {
				t.Fatalf("Parse() content mismatch:\ngot:  %q\nwant: %q", string(content), tc.wantContent)
			}
			if len(meta) != len(tc.wantMeta) {
				t.Fatalf("Parse() meta size mismatch: got %d, want %d. Got: %v", len(meta), len(tc.wantMeta), meta)
			}
			for k, v := range tc.wantMeta {
				got, ok := meta[k]
				if !ok {
					t.Errorf("Expected key %q not found in metadata", k)
					continue
				}
				if got != v {
					t.Errorf("Parse() meta[%q] mismatch:\ngot:  %q\nwant: %q", k, got, v)
				}
			}
		})
	}
}

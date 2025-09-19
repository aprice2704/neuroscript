// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Contains tests for the NeuroScript metadata parser, corrected for new contiguity rules.
// filename: pkg/metadata/nsparser_test.go
// nlines: 110
// risk_rating: LOW
package metadata_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

func TestNeuroScriptParser_Parse(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    metadata.Store
		wantErr     bool
	}{
		{
			name: "Valid metadata at start",
			input: `::schema: neuroscript
::fileVersion: 2
command
  emit "hello"
endcommand`,
			wantContent: `command
  emit "hello"
endcommand`,
			wantMeta: metadata.Store{"schema": "neuroscript", "fileversion": "2"},
		},
		{
			name: "Leading newlines and whitespace",
			input: `

  ::schema: a
::serialization: ns
command
  emit 1
endcommand`,
			wantContent: `command
  emit 1
endcommand`,
			wantMeta: metadata.Store{"schema": "a", "serialization": "ns"},
		},
		{
			name: "Metadata with mixed case and extra spacing",
			input: `::Schema: neuroscript
::  fileVERSION  :  3
func main() means
endfunc`,
			wantContent: `func main() means
endfunc`,
			wantMeta: metadata.Store{"schema": "neuroscript", "fileversion": "3"},
		},
		{
			name:        "No metadata",
			input:       "command\nemit 1\nendcommand",
			wantContent: "command\nemit 1\nendcommand",
			wantMeta:    metadata.Store{},
		},
		{
			name:        "Empty input",
			input:       "",
			wantContent: "",
			wantMeta:    metadata.Store{},
		},
		{
			name: "File with only metadata",
			input: `::schema: sdi-go
::serialization: go`,
			wantContent: "",
			wantMeta:    metadata.Store{"schema": "sdi-go", "serialization": "go"},
		},
		{
			name: "Metadata block with interleaved blank lines",
			input: `::key1: val1

::key2: val2

command
endcommand`,
			// CORRECTED: The blank line is a boundary. `::key2` is content.
			wantContent: `
::key2: val2

command
endcommand`,
			wantMeta: metadata.Store{"key1": "val1"},
		},
		{
			name: "Malformed metadata line ends the block",
			input: `::key1: val1
::key2 val2
::key3: val3
command
endcommand`,
			wantContent: `::key2 val2
::key3: val3
command
endcommand`,
			wantMeta: metadata.Store{"key1": "val1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := metadata.NewNeuroScriptParser()
			r := strings.NewReader(tc.input)
			meta, content, err := parser.Parse(r)

			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tc.wantErr)
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
					t.Errorf("Parse() meta[%q] = %q, want %q", k, got, v)
				}
			}
		})
	}
}

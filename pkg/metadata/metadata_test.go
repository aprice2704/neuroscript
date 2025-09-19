// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Contains robust tests for the Markdown metadata parser, corrected to align with new contiguity rules.
// filename: pkg/metadata/markdown_test.go
// nlines: 154
// risk_rating: LOW

package metadata_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

func TestMarkdownParser_Parse_Robust(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		wantContent string
		wantMeta    metadata.Store
		wantErr     bool
	}{
		{
			name: "Valid metadata at end",
			input: `This is the content.
More content.
::schema: spec
::fileVersion: 1`,
			wantContent: "This is the content.\nMore content.",
			wantMeta:    metadata.Store{"schema": "spec", "fileversion": "1"},
		},
		{
			name: "With trailing newlines and whitespace",
			input: `Content here.

::schema: a
::serialization: b

  `,
			wantContent: "Content here.\n",
			wantMeta:    metadata.Store{"schema": "a", "serialization": "b"},
		},
		{
			name: "Metadata with mixed case keys",
			input: `Content.
::Schema: spec
::fileVERSION: 1`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"schema": "spec", "fileversion": "1"},
		},
		{
			name: "Metadata with extra spacing",
			input: `Content.
::  key-one  :  value one  `,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key-one": "value one"},
		},
		{
			name:        "No metadata",
			input:       "Just content.",
			wantContent: "Just content.",
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
			name: "File with only metadata and trailing newlines",
			input: `::schema: sdi-go

`,
			wantContent: "",
			wantMeta:    metadata.Store{"schema": "sdi-go"},
		},
		{
			name: "Metadata block with interleaved blank lines",
			input: `Content.
::key1: val1

::key2: val2
`,
			// CORRECTED: The blank line is a boundary. `::key1` is content.
			wantContent: "Content.\n::key1: val1\n",
			wantMeta:    metadata.Store{"key2": "val2"},
		},
		{
			name: "Malformed metadata line is ignored",
			input: `Content.
::key1: val1
::key2 val2
::key3: val3`,
			wantContent: "Content.\n::key1: val1\n::key2 val2",
			wantMeta:    metadata.Store{"key3": "val3"},
		},
		{
			name: "No value for key",
			input: `Content.
::key:`,
			wantContent: "Content.",
			wantMeta:    metadata.Store{"key": ""},
		},
		{
			name: "Metadata-like content not at the end",
			input: `This content has ::fake: metadata.
It should be preserved.
::real: meta`,
			wantContent: "This content has ::fake: metadata.\nIt should be preserved.",
			wantMeta:    metadata.Store{"real": "meta"},
		},
		{
			name:        "Content ending with a colon",
			input:       "this is a test:",
			wantContent: "this is a test:",
			wantMeta:    metadata.Store{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := metadata.NewMarkdownParser()
			r := strings.NewReader(tc.input)
			meta, content, err := parser.Parse(r)

			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tc.wantErr)
			}
			if string(content) != tc.wantContent {
				t.Fatalf("Parse() content = %q, want %q", string(content), tc.wantContent)
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

func TestStore_CheckRequired(t *testing.T) {
	store := metadata.Store{
		"schema":        "spec",
		"serialization": "md",
	}

	t.Run("All keys present", func(t *testing.T) {
		if err := store.CheckRequired("schema", "serialization"); err != nil {
			t.Errorf("CheckRequired() returned an unexpected error: %v", err)
		}
	})

	t.Run("One key missing", func(t *testing.T) {
		if err := store.CheckRequired("schema", "version"); err == nil {
			t.Errorf("CheckRequired() expected an error for missing key 'version', but got nil")
		}
	})

	t.Run("No keys required", func(t *testing.T) {
		if err := store.CheckRequired(); err != nil {
			t.Errorf("CheckRequired() with no args should not return an error, but got: %v", err)
		}
	})
}

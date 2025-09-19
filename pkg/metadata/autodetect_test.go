// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Corrects the test to rely on automatic parser registration via init() instead of manual registration, fixing a panic.
// filename: pkg/metadata/autodetect_test.go
// nlines: 85
// risk_rating: LOW
package metadata_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

func TestDetectSerialization(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		want      string
		expectErr bool
	}{
		{
			name: "Serialization at the beginning (ns style)",
			input: `::serialization: ns
::id: capsule/some-id
command
endcommand`,
			want: "ns",
		},
		{
			name: "Serialization at the end (md style)",
			input: `This is the capsule content.
::id: capsule/some-id
::serialization: md`,
			want: "md",
		},
		{
			name: "Serialization in the middle only (should fail)",
			input: `Content line 1
::serialization: md
Content line 2`,
			expectErr: true,
		},
		{
			name:      "Empty file",
			input:     "",
			expectErr: true,
		},
		{
			name:      "No serialization key",
			input:     "::id: capsule/test\n::version: 1",
			expectErr: true,
		},
		{
			name:  "Key only present at end of a large file",
			input: strings.Repeat("a", 2048) + "\n::serialization: md",
			want:  "md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tc.input))
			var readSeeker io.ReadSeeker = reader

			got, err := metadata.DetectSerialization(readSeeker)

			if (err != nil) != tc.expectErr {
				t.Fatalf("DetectSerialization() error = %v, expectErr %v", err, tc.expectErr)
			}
			if !tc.expectErr && got != tc.want {
				t.Errorf("DetectSerialization() got = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseWithAutoDetect(t *testing.T) {
	// NOTE: The parsers are now automatically registered via their init() functions.
	// No manual registration is needed here.

	nsContent := `::serialization: ns
::id: test-ns
command
endcommand`

	mdContent := `This is markdown.
::serialization: md
::id: test-md`

	t.Run("Detects and parses ns", func(t *testing.T) {
		meta, _, ser, err := metadata.ParseWithAutoDetect(strings.NewReader(nsContent))
		if err != nil {
			t.Fatalf("ParseWithAutoDetect failed for ns: %v", err)
		}
		if ser != "ns" {
			t.Errorf("Expected serialization 'ns', got %q", ser)
		}
		if meta["id"] != "test-ns" {
			t.Errorf("Expected id 'test-ns', got %q", meta["id"])
		}
	})

	t.Run("Detects and parses md", func(t *testing.T) {
		meta, _, ser, err := metadata.ParseWithAutoDetect(strings.NewReader(mdContent))
		if err != nil {
			t.Fatalf("ParseWithAutoDetect failed for md: %v", err)
		}
		if ser != "md" {
			t.Errorf("Expected serialization 'md', got %q", ser)
		}
		if meta["id"] != "test-md" {
			t.Errorf("Expected id 'test-md', got %q", meta["id"])
		}
	})
}

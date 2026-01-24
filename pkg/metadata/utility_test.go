// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 3
// :: description: Contains tests for the metadata utility, extractor, and UNIFIED PARSING LOGIC with tricky line ending cases.
// :: latestChange: Added extensive test cases for CRLF, mixed line endings, and trailing whitespace to verify unified parser robustness.
// :: filename: pkg/metadata/utility_test.go
// :: serialization: go
package metadata_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

func TestExtractor(t *testing.T) {
	store := metadata.Store{
		"schema":       "spec",
		"file-Version": "123",
		"description":  "A test file.",
		"camelCaseKey": "value",
		"not_an_int":   "abc",
		"optional.key": "present",
	}
	extractor := metadata.NewExtractor(store)

	t.Run("GetFound", func(t *testing.T) {
		// Test different variations of the same key
		keysToTest := []string{"file-Version", "file_version", "file.version", "FileVersion"}
		for _, key := range keysToTest {
			val, ok := extractor.Get(key)
			if !ok {
				t.Errorf("Get(%q) expected to find key", key)
			}
			if val != "123" {
				t.Errorf("Get(%q) got %q, want %q", key, val, "123")
			}
		}

		// Test camelCase
		val, ok := extractor.Get("Camel.Case_Key")
		if !ok || val != "value" {
			t.Errorf("Get(Camel.Case_Key) failed, got: %q, %v", val, ok)
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		_, ok := extractor.Get("nonexistent")
		if ok {
			t.Error("Get(\"nonexistent\") expected not to find key")
		}
	})

	t.Run("GetWithDefaults", func(t *testing.T) {
		// Test GetOr
		if val := extractor.GetOr("schema", "default"); val != "spec" {
			t.Errorf("GetOr on existing key failed, got %q", val)
		}
		if val := extractor.GetOr("nonexistent", "default"); val != "default" {
			t.Errorf("GetOr on missing key failed, got %q", val)
		}

		// Test GetIntOr
		i, err := extractor.GetIntOr("file.version", 999)
		if err != nil || i != 123 {
			t.Errorf("GetIntOr on existing key failed, got %d, %v", i, err)
		}
		i, err = extractor.GetIntOr("nonexistent", 999)
		if err != nil || i != 999 {
			t.Errorf("GetIntOr on missing key failed, got %d, %v", i, err)
		}
		_, err = extractor.GetIntOr("not_an_int", 999)
		if err == nil {
			t.Error("GetIntOr on non-int value expected an error")
		}
	})

	t.Run("MustGet", func(t *testing.T) {
		if val := extractor.MustGet("schema"); val != "spec" {
			t.Errorf("MustGet(\"schema\") got %q, want %q", val, "spec")
		}
		if val := extractor.MustGet("nonexistent"); val != "" {
			t.Errorf("MustGet(\"nonexistent\") got %q, want \"\"", val)
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		i, ok, err := extractor.GetInt("file.version")
		if !ok || err != nil || i != 123 {
			t.Errorf("GetInt(\"file.version\") got %d, %v, %v; want 123, true, nil", i, ok, err)
		}

		_, ok, _ = extractor.GetInt("nonexistent")
		if ok {
			t.Error("GetInt(\"nonexistent\") expected ok=false")
		}

		_, _, err = extractor.GetInt("not_an_int")
		if err == nil {
			t.Error("GetInt(\"not_an_int\") expected a parse error")
		}
	})

	t.Run("CheckRequired", func(t *testing.T) {
		// Test with keys that will be normalized
		err := extractor.CheckRequired("schema", "file-version")
		if err != nil {
			t.Errorf("CheckRequired() returned unexpected error: %v", err)
		}

		err = extractor.CheckRequired("schema", "missing-key")
		if err == nil {
			t.Error("CheckRequired() expected an error for missing key")
		}
	})
}

// --- Tricky Line Ending & Parsing Tests ---

func TestReadLines_TrickyEndings(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Unix Newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Windows CRLF",
			input:    "Line 1\r\nLine 2\r\nLine 3",
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Mixed Line Endings",
			input:    "Line 1\nLine 2\r\nLine 3",
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Trailing Newline",
			input:    "Line 1\n",
			expected: []string{"Line 1"},
		},
		{
			name:     "No Trailing Newline",
			input:    "Line 1",
			expected: []string{"Line 1"},
		},
		{
			name:     "Empty File",
			input:    "",
			expected: nil, // Scanner returns nothing for empty input
		},
		{
			name:     "Only Newlines",
			input:    "\n\n",
			expected: []string{"", ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			lines, err := metadata.ReadLines(r)
			if err != nil {
				t.Fatalf("ReadLines failed: %v", err)
			}
			if len(lines) != len(tc.expected) {
				t.Fatalf("Line count mismatch. Got %d, want %d. Got: %q", len(lines), len(tc.expected), lines)
			}
			for i, line := range lines {
				if line != tc.expected[i] {
					t.Errorf("Line %d mismatch. Got %q, want %q", i, line, tc.expected[i])
				}
			}
		})
	}
}

func TestParseHeaderBlock_Tricky(t *testing.T) {
	input := []string{
		"",                // Leading blank line (ignored)
		":: schema: ns",   // Valid
		":: key: val   ",  // Trailing spaces on line (should be trimmed)
		"   ",             // Blank line with spaces (ends block)
		":: content: yes", // Should NOT be in metadata
	}

	store, endLine := metadata.ParseHeaderBlock(input)
	if store == nil {
		t.Fatal("ParseHeaderBlock returned nil store")
	}

	if val := store["schema"]; val != "ns" {
		t.Errorf("Expected schema=ns, got %q", val)
	}
	if val := store["key"]; val != "val" {
		t.Errorf("Expected key=val, got %q (check whitespace trimming)", val)
	}
	if _, ok := store["content"]; ok {
		t.Error("ParseHeaderBlock consumed past the blank line")
	}

	// 0: "", 1: meta, 2: meta, 3: blank(break).
	// endLine should point to index 4 (content start)?
	// Loop:
	// i=0: blank, continue.
	// i=1: match.
	// i=2: match.
	// i=3: no match (trim="" not regex). Break.
	// metaEndLine was last set at i+1 during i=2 -> 3.
	// So lines[3] is the start of non-meta. Correct.
	if endLine != 3 {
		t.Errorf("Expected endLine=3, got %d", endLine)
	}
}

func TestParseFooterBlock_Tricky(t *testing.T) {
	input := []string{
		"Content",
		"",
		":: schema: md  ", // Trailing whitespace
		":: id: test",
		"", // Trailing blank line (should be skipped)
	}

	store, startLine := metadata.ParseFooterBlock(input)
	if store == nil {
		t.Fatal("ParseFooterBlock returned nil store")
	}

	if val := store["schema"]; val != "md" {
		t.Errorf("Expected schema=md, got %q", val)
	}
	if val := store["id"]; val != "test" {
		t.Errorf("Expected id=test, got %q", val)
	}

	// Index 2 is the start of metadata.
	if startLine != 2 {
		t.Errorf("Expected startLine=2, got %d", startLine)
	}
}

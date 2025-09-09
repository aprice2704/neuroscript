// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Fuzz and sanitization tests for the path-lite parser.
// filename: pkg/json-lite/path_fuzz_test.go
// nlines: 91
// risk_rating: LOW

package json_lite

import (
	"math/rand"
	"testing"
	"time"
)

// TestParsePath_Sanitization ensures that inputs that could create
// empty or invalid segments are rejected.
func TestParsePath_Sanitization(t *testing.T) {
	testCases := []string{
		".",
		"..",
		"a..b",
		".a",
		"a.",
		"[]",
		"a[]",
		"a[ ]",
		"a[?]",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			_, err := ParsePath(tc)
			if err == nil {
				t.Fatalf("expected an error for invalid path '%s', but got nil", tc)
			}
		})
	}
}

// TestParsePath_Fuzz sprinkles random junk into otherwise valid paths
// and ensures the parser always fails gracefully.
func TestParsePath_Fuzz(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	validPaths := []string{
		"a",
		"meta.version",
		"items[0].id",
		"items[1]",
		"a.b.c.d.e.f",
	}

	// Characters that should break the parser
	junk := []rune{' ', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '=', '+', ',', '<', '>', '/', ':', ';', '{', '}', '\\', '|'}

	for _, path := range validPaths {
		t.Run("fuzzing_"+path, func(t *testing.T) {
			for i := 0; i < 20; i++ { // Run 20 fuzz attempts per path
				// Insert a random junk character at a random position
				insertPos := rng.Intn(len(path))
				junkChar := string(junk[rng.Intn(len(junk))])
				fuzzedPath := path[:insertPos] + junkChar + path[insertPos:]

				_, err := ParsePath(fuzzedPath)
				if err == nil {
					t.Fatalf("expected error for fuzzed path '%s', but it parsed successfully", fuzzedPath)
				}
			}
		})
	}
}

// TestSelectWithWeirdKeys ensures that keys with non-alphanumeric characters
// (which are valid in JSON) don't crash the selector, although they cannot be
// accessed by the string parser. They should be silently ignored by the parser
// as part of a larger key name.
func TestSelectWithWeirdKeys(t *testing.T) {
	data := map[string]any{
		"a.b":  1, // A key containing a dot
		"c[0]": 2, // A key containing brackets
	}

	// The string parser will interpret "a.b" as two segments, 'a' and 'b'.
	// This will correctly result in a "key not found" error because there is no 'a' key.
	path, err := ParsePath("a.b")
	if err != nil {
		t.Fatalf("parsing 'a.b' failed: %v", err)
	}
	_, err = Select(data, path, nil)
	if err == nil {
		t.Error("expected error when selecting path 'a.b' but got nil")
	}
}

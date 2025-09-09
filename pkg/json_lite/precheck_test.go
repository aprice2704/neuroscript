// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Tests for regex and sanitization pre-checks in parsers.
// filename: pkg/json-lite/precheck_test.go
// nlines: 66
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

func TestParsePath_RegexPreCheck(t *testing.T) {
	testCases := []struct {
		name      string
		pathStr   string
		shouldErr bool
	}{
		{"valid simple", "a", false},
		{"valid nested", "meta.version", false},
		{"valid indexed", "items[0].id", false},
		{"invalid double dot", "a..b", true},
		{"invalid leading dot", ".a", true},
		{"invalid trailing dot", "a.", true},
		{"invalid empty index", "a[]", true},
		{"invalid char in index", "a[b]", true},
		{"invalid unterminated index", "a[0", true},
		{"valid key with hyphen", "a[0].my-key", false},
		{"invalid double dot after index", "a[0]..mykey", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePath(tc.pathStr)
			if tc.shouldErr {
				if !errors.Is(err, ErrInvalidPath) {
					t.Fatalf("expected ErrInvalidPath for path '%s', but got: %v", tc.pathStr, err)
				}
			} else if err != nil {
				t.Fatalf("path '%s' should have passed but failed: %v", tc.pathStr, err)
			}
		})
	}
}

func TestParseShape_KeySanitization(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		isInvalid bool
	}{
		{"valid", "name", false},
		{"valid optional", "name?", false},
		{"valid list", "items[]", false},
		{"valid list optional", "items[]?", false},
		{"invalid empty", "", true},
		{"invalid just optional", "?", true},
		{"invalid just list", "[]", true},
		{"invalid just list optional", "[]?", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shapeDef := map[string]any{tc.key: "string"}
			_, err := ParseShape(shapeDef)
			if tc.isInvalid {
				if !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument for key '%s', got: %v", tc.key, err)
				}
			} else if err != nil {
				t.Fatalf("key '%s' should have been valid but failed: %v", tc.key, err)
			}
		})
	}
}

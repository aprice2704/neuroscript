// NeuroScript Version: 0.7.0
// File version: 9
// Purpose: Corrects the usage of errors.As to fix a compiler error.
// filename: aeiou/canonical_json_test.go
// nlines: 155
// risk_rating: LOW

package aeiou

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
)

func TestCanonicalize_Comprehensive(t *testing.T) {
	testCases := []struct {
		name        string
		input       interface{}
		expected    string
		expectError error
	}{
		// --- Passing Cases ---
		{
			name:     "Simple key reordering",
			input:    json.RawMessage(`{"c":1,"a":2,"b":3}`),
			expected: `{"a":2,"b":3,"c":1}`,
		},
		{
			name:     "Unicode keys and values",
			input:    json.RawMessage(`{"ü":"a","a":"b"}`),
			expected: `{"a":"b","ü":"a"}`,
		},
		{
			name:     "Escaped characters",
			input:    json.RawMessage(`{"a":"\n","b":"\""}`),
			expected: `{"a":"\n","b":"\""}`,
		},
		{
			name:     "Nested object keys are sorted",
			input:    json.RawMessage(`{"z":true,"a":{"y":1,"x":2}}`),
			expected: `{"a":{"x":2,"y":1},"z":true}`,
		},
		{
			name: "Large integer representation",
			input: struct {
				A int64 `json:"a"`
			}{
				A: math.MaxInt64,
			},
			expected: `{"a":9223372036854775807}`,
		},
		{
			name:     "Deeply nested structure",
			input:    json.RawMessage(`{"d": [1, {"c": 3, "b": [9, 8]}]}`),
			expected: `{"d":[1,{"b":[9,8],"c":3}]}`,
		},
		{
			name:     "Array with mixed types",
			input:    json.RawMessage(`[1, "two", null, true, {"a":1}]`),
			expected: `[1,"two",null,true,{"a":1}]`, // Array order is preserved
		},
		{
			name:     "Empty object and array",
			input:    json.RawMessage(`{"b": [], "a": {}}`),
			expected: `{"a":{},"b":[]}`,
		},
		{
			name:     "Scalar string",
			input:    `"hello"`,
			expected: `"hello"`,
		},
		{
			name:     "Scalar number",
			input:    json.RawMessage(`123`),
			expected: `123`,
		},
		{
			name:     "Scalar null",
			input:    json.RawMessage(`null`),
			expected: `null`,
		},

		// --- Failing Cases (Invalid JSON) ---
		{
			name:        "Invalid JSON - trailing comma in object",
			input:       `{"a":1,"b":2,}`,
			expectError: &json.SyntaxError{},
		},
		{
			name:        "Invalid JSON - trailing comma in array",
			input:       `[1,2,3,]`,
			expectError: &json.SyntaxError{},
		},
		{
			name:        "Invalid JSON - mismatched brackets",
			input:       `{"a":1]`,
			expectError: &json.SyntaxError{},
		},
		{
			name:        "Invalid JSON - unquoted key",
			input:       `{a:1}`,
			expectError: &json.SyntaxError{},
		},
		// --- Failing Case (DoS Attack) ---
		{
			name: "Recursion depth bomb",
			input: (func() string {
				// Build a JSON string like {"a":{"a":{"a":...}}}
				// that is deeper than our recursion limit.
				depth := maxRecursionDepth + 5
				var sb strings.Builder
				for i := 0; i < depth; i++ {
					sb.WriteString(`{"a":`)
				}
				sb.WriteString("null")
				for i := 0; i < depth; i++ {
					sb.WriteString(`}`)
				}
				return sb.String()
			})(),
			expectError: ErrMaxRecursionDepth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			canonicalBytes, err := Canonicalize(tc.input)

			if tc.expectError != nil {
				if err == nil {
					t.Fatal("Expected an error, but got nil")
				}
				// Check if the error is of the expected type or a specific sentinel
				if !errors.Is(err, tc.expectError) {
					var syntaxErr *json.SyntaxError
					if _, ok := tc.expectError.(*json.SyntaxError); !ok || !errors.As(err, &syntaxErr) {
						t.Fatalf("Expected error target %T, got %T (%v)", tc.expectError, err, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Canonicalize() failed unexpectedly: %v", err)
			}

			if !bytes.Equal(canonicalBytes, []byte(tc.expected)) {
				t.Errorf("Canonicalize() mismatch:\n- want: %s\n- got:  %s", tc.expected, string(canonicalBytes))
			}
		})
	}
}

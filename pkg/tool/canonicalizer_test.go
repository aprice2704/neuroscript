// NeuroScript Version: 0.1.0
// File version: 1.0.0
// Purpose: Provides tests for the tool name canonicalizer.
// filename: pkg/tool/canonicalizer_test.go

package tool

import "testing"

func TestCanonicalizeToolName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no prefix",
			input:    "math.add",
			expected: "tool.math.add",
		},
		{
			name:     "correct prefix",
			input:    "tool.math.add",
			expected: "tool.math.add",
		},
		{
			name:     "double prefix",
			input:    "tool.tool.math.add",
			expected: "tool.math.add",
		},
		{
			name:     "triple prefix",
			input:    "tool.tool.tool.math.add",
			expected: "tool.math.add",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "tool.",
		},
		{
			name:     "just the prefix",
			input:    "tool.",
			expected: "tool.",
		},
		{
			name:     "just a double prefix",
			input:    "tool.tool.",
			expected: "tool.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := CanonicalizeToolName(tc.input)
			if actual != tc.expected {
				t.Errorf("for input '%s', expected '%s' but got '%s'", tc.input, tc.expected, actual)
			}
		})
	}
}

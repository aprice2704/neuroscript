package metadata

import (
	"errors" // Import errors for error checking
	"reflect"
	"testing"
	// REMOVED: "github.com/aprice2704/neuroscript/pkg/core"
)

func TestExtractMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		expected    map[string]string // Expected map only if no error expected
		wantErr     bool
		wantErrType error // Specific error type to check using errors.Is
	}{
		{
			name: "Basic Metadata",
			content: `:: version: 1.0
:: id: test-123
:: author:  Gemini
- [ ] First real item`, // Content stops metadata scan
			expected: map[string]string{
				"version": "1.0",
				"id":      "test-123",
				"author":  "Gemini",
			},
			wantErr: false,
		},
		{
			name: "Metadata with Comments and Blank Lines",
			content: `
# Standard Comment
:: version: 0.5
  :: type: Checklist

-- Another comment style
:: status : draft

Actual content starts here.
- [x] Done item`,
			expected: map[string]string{
				"version": "0.5",
				"type":    "Checklist",
				"status":  "draft",
			},
			wantErr: false,
		},
		{
			name: "Metadata Stops at First Content (DEFINE PROCEDURE)",
			content: `:: key1: value1
DEFINE PROCEDURE Test()
:: key2: value2`, // This won't be extracted
			expected: map[string]string{
				"key1": "value1",
			},
			wantErr: false,
		},
		{
			name: "Metadata Stops at First Content (FILE_VERSION)",
			content: `:: key1: value1
FILE_VERSION "1.0"
:: key2: value2`, // This won't be extracted
			expected: map[string]string{
				"key1": "value1",
			},
			wantErr: false,
		},
		{
			name: "Metadata Before FILE_VERSION and DEFINE PROCEDURE",
			content: `:: meta1: valueA
:: meta2: valueB

FILE_VERSION "1.1.0"

DEFINE PROCEDURE ActualCode()
COMMENT: ... ENDCOMMENT
END
`,
			expected: map[string]string{
				"meta1": "valueA",
				"meta2": "valueB",
			},
			wantErr: false,
		},
		{
			name: "No Metadata",
			content: `DEFINE PROCEDURE Test()
COMMENT: ... ENDCOMMENT
SET x = 1
END`,
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "Empty Input",
			content:  ``,
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name: "Only Comments and Whitespace",
			content: `

# Comment 1
  -- Comment 2

`,
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name: "Duplicate Keys (First Wins)",
			content: `:: version: 1.0
:: id: first-id
:: version: 2.0
- Content`,
			expected: map[string]string{
				"version": "1.0",
				"id":      "first-id",
			},
			wantErr: false,
		},
		{
			// --- UPDATED Test Case: Expect Local Error ---
			name:    "Invalid_Metadata_Format_(No_Space_after_::)",
			content: `::version: 1.0\n:: id: test\n- Content`, // Malformed line first
			// Function should error on line 1, return whatever metadata was collected *before* it (none).
			expected:    map[string]string{},
			wantErr:     true,
			wantErrType: ErrMalformedMetadata, // Expect specific local error
		},
		{
			// --- UPDATED Test Case: Expect Local Error ---
			name:    "Invalid_Metadata_Format_(No_Colon)",
			content: `:: version 1.0\n:: id: test\n- Content`, // Malformed line first
			// Function should error on line 1, return whatever metadata was collected *before* it (none).
			expected:    map[string]string{},
			wantErr:     true,
			wantErrType: ErrMalformedMetadata, // Expect specific local error
		},
		{
			name: "Value with Colons",
			content: `:: url: https://example.com
:: description: This value : has a colon.
- item`,
			expected: map[string]string{
				"url":         "https://example.com",
				"description": "This value : has a colon.",
			},
			wantErr: false,
		},
		{
			name: "Key with Hyphen and Dot",
			content: `:: neuro-version.major: 1
- item`,
			expected: map[string]string{
				"neuro-version.major": "1",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Extract(tc.content) // Use the Extract function directly

			// Check error expectation
			if tc.wantErr {
				if err == nil {
					t.Errorf("Extract() expected an error but got nil")
				} else if tc.wantErrType != nil {
					// Check if the error IS or WRAPS the expected type
					if !errors.Is(err, tc.wantErrType) {
						t.Errorf("Extract() error type mismatch:\n Got error: %v (%T)\nWant error type: %v", err, err, tc.wantErrType)
					}
				}
				// Check the returned map on expected error
				if !reflect.DeepEqual(got, tc.expected) {
					t.Errorf("Extract() map result on expected error mismatch:\ngot = %#v\nwant %#v", got, tc.expected)
				}

			} else { // No error expected
				if err != nil {
					t.Errorf("Extract() unexpected error: %v", err)
				}
				// Compare map results only if no error was expected
				if !reflect.DeepEqual(got, tc.expected) {
					t.Errorf("Extract() map result mismatch:\ngot = %#v\nwant %#v", got, tc.expected)
				}
			}
		})
	}
}

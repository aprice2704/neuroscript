package metadata

import (
	"reflect"
	"testing"
)

func TestExtractMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
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
			name: "Invalid Metadata Format (No Space after ::)",
			content: `::version: 1.0
:: id: test
- Content`,
			expected: map[string]string{
				// "version" is missed due to format requirement
				"id": "test",
			},
			wantErr: false,
		},
		{
			name: "Invalid Metadata Format (No Colon)",
			content: `:: version 1.0
:: id: test
- Content`,
			expected: map[string]string{
				// "version" is missed
				"id": "test",
			},
			wantErr: false,
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

			if (err != nil) != tc.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			// Use DeepEqual for map comparison
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Extract() got = %#v, want %#v", got, tc.expected)
			}
		})
	}
}

// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Tests for path-lite parsing and selection logic.
// filename: pkg/json-lite/path_test.go
// nlines: 122
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

var pathTestData = map[string]any{
	"name": "test",
	"meta": map[string]any{
		"version": 2,
	},
	"items": []any{
		map[string]any{"id": 100},
		map[string]any{"id": 200},
	},
}

func TestParsePath(t *testing.T) {
	testCases := []struct {
		name        string
		pathStr     string
		expectedLen int
		expectErr   bool
	}{
		{"simple", "a", 1, false},
		{"dotted", "a.b", 2, false},
		{"indexed", "a[0]", 2, false},
		{"mixed", "a.b[0].c", 4, false},
		{"empty", "", 0, true},
		{"just dots", "..", 0, true},
		{"invalid index", "a[x]", 0, true},
		{"unterminated index", "a[0", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := ParsePath(tc.pathStr)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("did not expect an error but got: %v", err)
			}
			if len(path) != tc.expectedLen {
				t.Errorf("expected path length %d, got %d", tc.expectedLen, len(path))
			}
		})
	}
}

func TestSelect(t *testing.T) {
	testCases := []struct {
		name        string
		pathStr     string
		data        any
		expected    any
		expectedErr error
	}{
		{
			name:     "simple",
			pathStr:  "name",
			data:     pathTestData,
			expected: "test",
		},
		{
			name:     "nested",
			pathStr:  "meta.version",
			data:     pathTestData,
			expected: 2,
		},
		{
			name:     "list access",
			pathStr:  "items[1].id",
			data:     pathTestData,
			expected: 200,
		},
		{
			name:        "key not found",
			pathStr:     "meta.foo",
			data:        pathTestData,
			expectedErr: ErrMapKeyNotFound,
		},
		{
			name:        "index out of bounds",
			pathStr:     "items[5]",
			data:        pathTestData,
			expectedErr: ErrListIndexOutOfBounds,
		},
		{
			name:        "accessing key on list",
			pathStr:     "items.key",
			data:        pathTestData,
			expectedErr: ErrCannotAccessType,
		},
		{
			name:        "nil data",
			pathStr:     "a.b",
			data:        nil,
			expectedErr: ErrCollectionIsNil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := ParsePath(tc.pathStr)
			if err != nil && tc.expectedErr == nil {
				t.Fatalf("path parsing failed: %v", err)
			}

			result, err := Select(tc.data, path, nil)

			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error '%v', but got nil", tc.expectedErr)
				}
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error to be '%v', but got '%v'", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("did not expect error, but got: %v", err)
			}

			if result != tc.expected {
				t.Errorf("expected result '%v', but got '%v'", tc.expected, result)
			}
		})
	}
}

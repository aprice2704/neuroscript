// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Tests for case-insensitive options in Select and Validate.
// filename: pkg/json-lite/options_test.go
// nlines: 125
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

var caseTestData = map[string]any{
	"NAME": "test",
	"Meta": map[string]any{
		"VERSION": 2,
	},
	"ITEMS": []any{
		map[string]any{"ID": 100},
		map[string]any{"Id": 200},
	},
}

func TestSelect_WithOptions_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name        string
		pathStr     string
		options     *SelectOptions
		expected    any
		expectedErr error
	}{
		{
			name:     "case-insensitive simple",
			pathStr:  "name",
			options:  &SelectOptions{CaseInsensitive: true},
			expected: "test",
		},
		{
			name:     "case-insensitive nested",
			pathStr:  "meta.version",
			options:  &SelectOptions{CaseInsensitive: true},
			expected: 2,
		},
		{
			name:     "case-insensitive list access",
			pathStr:  "items[1].id",
			options:  &SelectOptions{CaseInsensitive: true},
			expected: 200,
		},
		{
			name:        "case-sensitive fail",
			pathStr:     "name",
			options:     nil, // Default is case-sensitive
			expectedErr: ErrMapKeyNotFound,
		},
		{
			name:        "case-insensitive key not found",
			pathStr:     "meta.foo",
			options:     &SelectOptions{CaseInsensitive: true},
			expectedErr: ErrMapKeyNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := ParsePath(tc.pathStr)
			if err != nil {
				t.Fatalf("path parsing failed: %v", err)
			}
			result, err := Select(caseTestData, path, tc.options)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error '%v', but got '%v'", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}

			if result != tc.expected {
				t.Errorf("expected result '%v', but got '%v'", tc.expected, result)
			}
		})
	}
}

func TestShapeValidate_WithOptions(t *testing.T) {
	shapeDef := map[string]any{
		"name": "string",
		"meta": map[string]any{
			"version": "int",
		},
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// This data matches the shape, just with different casing.
	validCaseData := map[string]any{
		"NAME": "test",
		"Meta": map[string]any{"VERSION": 2},
	}

	// This data has an extra key.
	extraKeyData := map[string]any{
		"NAME":  "test",
		"Meta":  map[string]any{"VERSION": 2},
		"EXTRA": "field",
	}

	testCases := []struct {
		name    string
		data    map[string]any
		options *ValidateOptions
		wantErr bool
	}{
		{"case-insensitive pass", validCaseData, &ValidateOptions{CaseInsensitive: true, AllowExtra: false}, false},
		{"case-sensitive fail", validCaseData, &ValidateOptions{CaseInsensitive: false, AllowExtra: false}, true},
		{"case-insensitive with extra fail", extraKeyData, &ValidateOptions{CaseInsensitive: true, AllowExtra: false}, true},
		{"case-insensitive with extra pass", extraKeyData, &ValidateOptions{CaseInsensitive: true, AllowExtra: true}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Validate(tc.data, tc.options)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
			} else if err != nil {
				t.Fatalf("did not expect an error but got: %v", err)
			}
		})
	}
}

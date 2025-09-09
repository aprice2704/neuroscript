// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Fuzz and sanitization tests for the shape-lite parser and validator.
// filename: pkg/json-lite/shape_fuzz_test.go
// nlines: 137
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

func TestParseShape_Sanitization(t *testing.T) {
	testCases := []struct {
		name      string
		shapeDef  map[string]any
		expectErr bool
	}{
		{"empty key", map[string]any{"": "string"}, true},
		{"just ?", map[string]any{"?": "string"}, true},
		{"just []", map[string]any{"[]": "int"}, true},
		{"mixed junk", map[string]any{"?[]": "any"}, true},
		{"mixed junk reversed", map[string]any{"[]?": "any"}, true},
		{"valid complex suffix", map[string]any{"key[]?": "string"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseShape(tc.shapeDef)
			if tc.expectErr {
				if !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got: %v", err)
				}
			} else if err != nil {
				t.Fatalf("key should have been valid but failed: %v", err)
			}
		})
	}
}

func TestShapeValidate_NilValues(t *testing.T) {
	shapeDef := map[string]any{
		"required_field":  "string",
		"optional_field?": "string",
		"any_field?":      "any",
	}
	shape, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	testCases := []struct {
		name        string
		data        map[string]any
		expectedErr error
	}{
		{
			"nil for required field",
			map[string]any{"required_field": nil, "any_field": "ok"},
			ErrValidationTypeMismatch,
		},
		{
			"nil for optional field",
			map[string]any{"required_field": "ok", "optional_field": nil},
			ErrValidationTypeMismatch,
		},
		{
			"nil for any field",
			// CORRECTED: Added the required field to make the test valid
			map[string]any{"required_field": "ok", "any_field": nil},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := shape.Validate(tc.data, nil)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error '%v', got '%v'", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}
		})
	}
}

func TestShapeValidate_FuzzMutate(t *testing.T) {
	shapeDef := map[string]any{
		"name": "string",
		"age":  "int",
	}
	shape, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}
	validData := map[string]any{"name": "John", "age": 42}
	if err := shape.Validate(validData, nil); err != nil {
		t.Fatalf("base valid data failed validation: %v", err)
	}

	mutations := []struct {
		name   string
		mutate func(d map[string]any)
	}{
		{"change type", func(d map[string]any) { d["age"] = "not-an-int" }},
		{"remove required", func(d map[string]any) { delete(d, "name") }},
		{"add extra", func(d map[string]any) { d["extra"] = true }},
	}

	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			mutatedData := make(map[string]any)
			for k, v := range validData {
				mutatedData[k] = v
			}
			m.mutate(mutatedData)

			err := shape.Validate(mutatedData, nil)
			if err == nil {
				t.Fatalf("expected validation to fail for mutation '%s', but it passed", m.name)
			}
		})
	}
}

// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Tests for the core shape parser.
// filename: pkg/json-lite/shape_test.go
// nlines: 43
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

func TestParseShape(t *testing.T) {
	validShape := map[string]any{
		"name": "string",
		"meta?": map[string]any{
			"version": "int",
		},
		"tags[]": "string",
	}
	invalidShapeType := map[string]any{"key": 123}
	invalidNestedShape := map[string]any{"meta": map[string]any{"version": 123}}

	t.Run("valid shape", func(t *testing.T) {
		s, err := ParseShape(validShape)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(s.Fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(s.Fields))
		}
		// CORRECTED: Check for the normalized key 'meta' which the parser now uses for storage.
		if spec, ok := s.Fields["meta"]; !ok || spec.NestedShape == nil {
			t.Error("expected nested shape for 'meta' field")
		}
	})
	t.Run("invalid shape type", func(t *testing.T) {
		_, err := ParseShape(invalidShapeType)
		if !errors.Is(err, ErrValidationTypeMismatch) {
			t.Fatalf("expected type mismatch error, got %v", err)
		}
	})
	t.Run("invalid nested shape type", func(t *testing.T) {
		_, err := ParseShape(invalidNestedShape)
		if !errors.Is(err, ErrValidationTypeMismatch) {
			t.Fatalf("expected type mismatch error in nested shape, got %v", err)
		}
	})
}

// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Adds extra test coverage for edge cases and complex interactions.
// filename: pkg/json-lite/extra_coverage_test.go
// nlines: 249
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

func TestShapeValidate_AnyType(t *testing.T) {
	shapeDef := map[string]any{
		"anything?":  "any",
		"any_list[]": "any",
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	testCases := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name:    "any accepts nil",
			data:    map[string]any{"anything": nil, "any_list": []any{}},
			wantErr: false,
		},
		{
			name:    "any accepts string",
			data:    map[string]any{"anything": "hello", "any_list": []any{}},
			wantErr: false,
		},
		{
			name:    "any accepts number",
			data:    map[string]any{"anything": 123, "any_list": []any{}},
			wantErr: false,
		},
		{
			name:    "any accepts map",
			data:    map[string]any{"anything": map[string]any{"a": 1}, "any_list": []any{}},
			wantErr: false,
		},
		{
			name:    "any list accepts mixed types",
			data:    map[string]any{"any_list": []any{"a", 1, true, nil}},
			wantErr: false,
		},
		{
			name:    "any list fails on non-list",
			data:    map[string]any{"any_list": "not a list"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Validate(tc.data, nil)
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

func TestShapeValidate_EmptyLists(t *testing.T) {
	shapeDef := map[string]any{
		"optional_list[]?": "string",
		"required_list[]":  "string",
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	testCases := []struct {
		name        string
		data        map[string]any
		expectedErr error
	}{
		{
			name:        "empty list for required field is valid",
			data:        map[string]any{"required_list": []any{}},
			expectedErr: nil,
		},
		{
			name:        "empty list for optional field is valid",
			data:        map[string]any{"required_list": []any{}, "optional_list": []any{}},
			expectedErr: nil,
		},
		{
			name:        "missing optional list is valid",
			data:        map[string]any{"required_list": []any{}},
			expectedErr: nil,
		},
		{
			name:        "missing required list is invalid",
			data:        map[string]any{},
			expectedErr: ErrValidationRequiredArgMissing,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Validate(tc.data, nil)
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

func TestNilOptions(t *testing.T) {
	// Ensures that passing a nil options struct is handled gracefully and defaults to the strictest settings.
	t.Run("Validate with nil options", func(t *testing.T) {
		s, _ := ParseShape(map[string]any{"name": "string"})
		// Data with an extra key should fail with nil options (defaults to AllowExtra: false)
		data := map[string]any{"name": "test", "extra": "field"}
		if err := s.Validate(data, nil); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for extra field with nil options, got: %v", err)
		}
	})

	t.Run("Select with nil options", func(t *testing.T) {
		p, _ := ParsePath("NAME")
		// Data with different casing should fail with nil options (defaults to CaseInsensitive: false)
		data := map[string]any{"name": "test"}
		if _, err := Select(data, p, nil); !errors.Is(err, ErrMapKeyNotFound) {
			t.Fatalf("expected ErrMapKeyNotFound for case-sensitive mismatch with nil options, got: %v", err)
		}
	})
}

func TestShapeValidate_ComplexNesting(t *testing.T) {
	// A more complex, realistic shape to test recursion.
	shapeDef := map[string]any{
		"id":          "string",
		"eventType":   "string",
		"actor?":      map[string]any{"id": "int", "name": "string"},
		"permissions": map[string]any{"read": "bool", "write": "bool"},
		"assets[]?": map[string]any{
			"assetId":   "string",
			"assetType": "string",
			"meta[]?":   map[string]any{"key": "string", "value": "any"},
		},
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Data that perfectly matches the complex shape.
	validData := map[string]any{
		"id":          "evt-123",
		"eventType":   "asset.created",
		"actor":       map[string]any{"id": 1, "name": "test-user"},
		"permissions": map[string]any{"read": true, "write": false},
		"assets": []any{
			map[string]any{
				"assetId":   "ast-456",
				"assetType": "image",
				"meta": []any{
					map[string]any{"key": "size", "value": 1024},
					map[string]any{"key": "format", "value": "png"},
				},
			},
		},
	}

	// Data with a type error deep inside the structure.
	invalidData := map[string]any{
		"id":          "evt-123",
		"eventType":   "asset.created",
		"permissions": map[string]any{"read": true, "write": false},
		"assets": []any{
			map[string]any{
				"assetId":   "ast-456",
				"assetType": "image",
				"meta": []any{
					map[string]any{"key": "size", "value": "this-should-be-any-so-its-ok"},
					map[string]any{"key": 123, "value": "but-this-key-should-be-string"},
				},
			},
		},
	}

	t.Run("complex valid data", func(t *testing.T) {
		if err := s.Validate(validData, nil); err != nil {
			t.Fatalf("expected complex data to be valid, but got error: %v", err)
		}
	})

	t.Run("complex invalid data", func(t *testing.T) {
		err := s.Validate(invalidData, nil)
		if !errors.Is(err, ErrValidationTypeMismatch) {
			t.Fatalf("expected ErrValidationTypeMismatch for deep type error, got: %v", err)
		}
	})
}

func TestSelect_MismatchedAccess(t *testing.T) {
	data := map[string]any{
		"items": []any{"a", "b"},
		"user":  map[string]any{"name": "test"},
	}

	t.Run("access key on list", func(t *testing.T) {
		p, _ := ParsePath("items.name")
		_, err := Select(data, p, nil)
		if !errors.Is(err, ErrCannotAccessType) {
			t.Fatalf("expected ErrCannotAccessType, got %v", err)
		}
	})

	t.Run("access index on map", func(t *testing.T) {
		p, _ := ParsePath("user[0]")
		_, err := Select(data, p, nil)
		if !errors.Is(err, ErrCannotAccessType) {
			t.Fatalf("expected ErrCannotAccessType, got %v", err)
		}
	})
}

func TestParseShape_InvalidDefinitions(t *testing.T) {
	t.Run("list as type definition", func(t *testing.T) {
		invalidShape := map[string]any{
			"key": []any{"string"},
		}
		_, err := ParseShape(invalidShape)
		if !errors.Is(err, ErrValidationTypeMismatch) {
			t.Fatalf("expected ErrValidationTypeMismatch for invalid shape def, got %v", err)
		}
	})
}

func TestShapeValidate_DuplicateCaseInsensitiveKeys(t *testing.T) {
	s, _ := ParseShape(map[string]any{"name": "string"})
	// The validator will find the first match ('name') and validate it.
	// It will then see 'NAME' as an unvalidated, extra key.
	data := map[string]any{"name": "a", "NAME": "b"}

	t.Run("fail on duplicates when extra not allowed", func(t *testing.T) {
		err := s.Validate(data, &ValidateOptions{CaseInsensitive: true, AllowExtra: false})
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for duplicate key, got: %v", err)
		}
	})

	t.Run("pass on duplicates when extra allowed", func(t *testing.T) {
		err := s.Validate(data, &ValidateOptions{CaseInsensitive: true, AllowExtra: true})
		if err != nil {
			t.Fatalf("did not expect error for duplicate keys when extra is allowed, got: %v", err)
		}
	})
}

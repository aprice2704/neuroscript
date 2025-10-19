// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Additional spec-compliance tests for path-lite and shape-lite (v0.2) — updated with float-int coercion test.
// filename: pkg/json-lite/spec_compliance_additional_test.go
// nlines: 213
// risk_rating: LOW

package json_lite

import (
	"errors"
	"strings"
	"testing"
)

func TestParsePath_InvalidNegativeIndex(t *testing.T) {
	_, err := ParsePath("a[-1]")
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath for negative index, got: %v", err)
	}
}

func TestParsePath_InvalidAlphaNumIndex(t *testing.T) {
	_, err := ParsePath("a[1a]")
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath for alphanumeric index, got: %v", err)
	}
}

func TestParsePath_IndexOverflow(t *testing.T) {
	huge := strings.Repeat("9", 40)
	_, err := ParsePath("a[" + huge + "]")
	if !errors.Is(err, ErrListInvalidIndexType) {
		t.Fatalf("expected ErrListInvalidIndexType on overflow index, got: %v", err)
	}
}

func TestParsePath_IndexSegmentTooLong(t *testing.T) {
	longIdx := strings.Repeat("1", maxPathSegmentLen+1)
	_, err := ParsePath("a[" + longIdx + "]")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for overlong index segment, got: %v", err)
	}
}

func TestSelect_IndexZeroOnEmptyList(t *testing.T) {
	p, err := ParsePath("items[0]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	data := map[string]any{"items": []any{}}
	_, err = Select(data, p, nil)
	if !errors.Is(err, ErrListIndexOutOfBounds) {
		t.Fatalf("expected ErrListIndexOutOfBounds on empty list, got: %v", err)
	}
}

func TestParseShape_SuffixOrderVariants(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		wantField string
	}{
		{"list_then_optional", "items[]?", "items"},
		{"optional_then_list", "maybe?[]", "maybe"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := ParseShape(map[string]any{tc.key: "string"})
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			spec, ok := s.Fields[tc.wantField]
			if !ok {
				t.Fatalf("expected field %q present; got keys: %v", tc.wantField, func() []string {
					keys := make([]string, 0, len(s.Fields))
					for k := range s.Fields {
						keys = append(keys, k)
					}
					return keys
				}())
			}
			if !spec.IsList {
				t.Fatalf("expected IsList=true for key %q", tc.key)
			}
			if !spec.IsOptional {
				t.Fatalf("expected IsOptional=true for key %q", tc.key)
			}
			if spec.PrimitiveType != "string" {
				t.Fatalf("expected primitive string for key %q, got %q", tc.key, spec.PrimitiveType)
			}
		})
	}
}

func TestShapeValidate_ListOfPrimitives(t *testing.T) {
	shapeDef := map[string]any{
		"tags[]": "string",
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	data := map[string]any{"tags": []any{"a", "b", "c"}}
	if err := s.Validate(data, nil); err != nil {
		t.Fatalf("validation should pass, got: %v", err)
	}

	bad := map[string]any{"tags": []any{"a", 42}}
	if err := s.Validate(bad, nil); !errors.Is(err, ErrValidationTypeMismatch) {
		t.Fatalf("expected type mismatch for list element, got: %v", err)
	}
}

func TestShapeValidate_ListOfMaps(t *testing.T) {
	shapeDef := map[string]any{
		"items[]": map[string]any{
			"sku": "string",
			"qty": "int",
		},
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ok := map[string]any{
		"items": []any{
			map[string]any{"sku": "A", "qty": 1},
			map[string]any{"sku": "B", "qty": 2},
		},
	}
	if err := s.Validate(ok, nil); err != nil {
		t.Fatalf("validation should pass, got: %v", err)
	}

	miss := map[string]any{
		"items": []any{
			map[string]any{"sku": "A", "qty": 1},
			map[string]any{"sku": "B"},
		},
	}
	if err := s.Validate(miss, nil); !errors.Is(err, ErrValidationRequiredArgMissing) {
		t.Fatalf("expected required-missing error, got: %v", err)
	}
}

func TestShapeValidate_AllowExtraToggle(t *testing.T) {
	shapeDef := map[string]any{
		"name": "string",
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	data := map[string]any{"name": "ok", "extra": true}

	if err := s.Validate(data, &ValidateOptions{AllowExtra: false}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid-argument for extra key, got: %v", err)
	}

	if err := s.Validate(data, &ValidateOptions{AllowExtra: true}); err != nil {
		t.Fatalf("validation should pass with allow_extra=true, got: %v", err)
	}
}

func TestShapeValidate_SpecialTypesAcceptance(t *testing.T) {
	shapeDef := map[string]any{
		"e": "email",
		"u": "url",
		"d": "isoDatetime",
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	ok := map[string]any{"e": "a@b.com", "u": "http://example.com", "d": "2025-01-01T00:00:00Z"}
	if err := s.Validate(ok, nil); err != nil {
		t.Fatalf("validation should pass for special-string types, got: %v", err)
	}

	bad := map[string]any{"e": 123, "u": "http://example.com", "d": "2025-01-01T00:00:00Z"}
	if err := s.Validate(bad, nil); !errors.Is(err, ErrValidationTypeMismatch) {
		t.Fatalf("expected type mismatch for non-string special type, got: %v", err)
	}
}

func TestShapeValidate_MissingRequiredNested(t *testing.T) {
	shapeDef := map[string]any{
		"user": map[string]any{
			"name":  "string",
			"email": "string",
		},
	}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	data := map[string]any{
		"user": map[string]any{
			"name": "A",
		},
	}
	err = s.Validate(data, nil)
	if !errors.Is(err, ErrValidationRequiredArgMissing) {
		t.Fatalf("expected required missing error, got: %v", err)
	}
}

// This test specifically verifies the fix for float64 values (like unwrapped
// timestamps) being validated against an "int" shape type.
func TestShapeValidate_IntFloatCoercion(t *testing.T) {
	shapeDef := map[string]any{"ts": "int"}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	cases := []struct {
		name    string
		value   any
		wantErr error
	}{
		{"native int", int(12345), nil},
		{"native int64", int64(1234567890), nil},
		{"float64 whole number", float64(12345.0), nil},
		{"float32 whole number", float32(123.0), nil},
		{"float64 with fraction", float64(123.45), ErrValidationTypeMismatch},
		{"float32 with fraction", float32(123.5), ErrValidationTypeMismatch},
		{"string value", "12345", ErrValidationTypeMismatch},
		{"bool value", true, ErrValidationTypeMismatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"ts": tc.value}
			err := s.Validate(data, nil)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestParseShape_ExactDepthLimitOK(t *testing.T) {
	shape := make(map[string]any)
	cur := shape
	for i := 0; i < maxShapeDepth; i++ {
		next := make(map[string]any)
		cur["next"] = next
		cur = next
	}
	cur["name"] = "string"

	if _, err := ParseShape(shape); err != nil {
		t.Fatalf("shape at exact depth limit should parse, got: %v", err)
	}
}

func TestSelect_IndexOutOfBoundsLarge(t *testing.T) {
	p, err := ParsePath("items[" + strings.Repeat("9", 5) + "]")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	data := map[string]any{"items": []any{1, 2}}
	_, err = Select(data, p, nil)
	if !errors.Is(err, ErrListIndexOutOfBounds) {
		t.Fatalf("expected ErrListIndexOutOfBounds, got: %v", err)
	}
}

func TestParsePath_MaxSegmentsBoundary(t *testing.T) {
	ok := strings.Repeat("a.", maxPathSegments-1) + "a"
	if _, err := ParsePath(ok); err != nil {
		t.Fatalf("path with exactly max segments should pass, got: %v", err)
	}
	bad := ok + ".b"
	_, err := ParsePath(bad)
	if !errors.Is(err, ErrNestingDepthExceeded) {
		t.Fatalf("expected ErrNestingDepthExceeded, got: %v", err)
	}
}

func TestParsePath_MaxSegmentLenBoundary(t *testing.T) {
	seg := strings.Repeat("x", maxPathSegmentLen)
	if _, err := ParsePath(seg); err != nil {
		t.Fatalf("segment at limit should pass, got: %v", err)
	}
	over := seg + "x"
	_, err := ParsePath(over)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for overlong key segment, got: %v", err)
	}
}

func TestShapeValidate_ListNestingDepthLimit(t *testing.T) {
	// Build a self-referential shape so validation can exceed max depth
	shape := &Shape{Fields: make(map[string]*FieldSpec)}
	fieldSpec := &FieldSpec{Name: "items", IsList: true}
	fieldSpec.NestedShape = shape // self-reference
	shape.Fields["items"] = fieldSpec

	// Data deeper than max depth
	data := make(map[string]any)
	cur := data
	for i := 0; i < maxShapeDepth+1; i++ {
		next := map[string]any{}
		cur["items"] = []any{next}
		cur = next
	}

	err := shape.Validate(data, nil)
	if !errors.Is(err, ErrNestingDepthExceeded) {
		t.Fatalf("expected ErrNestingDepthExceeded for deep list/map nesting, got: %v", err)
	}
}

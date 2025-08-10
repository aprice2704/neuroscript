// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Consolidated tests for path and shape depth/length limits.
// filename: pkg/json-lite/limits_test.go
// nlines: 98
// risk_rating: LOW

package json_lite

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- Path Limits ---

func TestPath_StringForm_MaxSegmentsBoundary(t *testing.T) {
	t.Run("should pass at exact segment limit", func(t *testing.T) {
		pathStr := strings.Repeat("a.", maxPathSegments-1) + "a"
		if _, err := ParsePath(pathStr); err != nil {
			t.Fatalf("path with exactly max segments should pass, got: %v", err)
		}
	})

	t.Run("should fail when exceeding segment limit", func(t *testing.T) {
		pathStr := strings.Repeat("a.", maxPathSegments) + "a"
		_, err := ParsePath(pathStr)
		if !errors.Is(err, lang.ErrNestingDepthExceeded) {
			t.Fatalf("expected ErrNestingDepthExceeded for path over segment limit, got: %v", err)
		}
	})
}

func TestPath_StringForm_MaxSegmentLenBoundary(t *testing.T) {
	t.Run("should pass for key at exact length limit", func(t *testing.T) {
		pathStr := strings.Repeat("x", maxPathSegmentLen)
		if _, err := ParsePath(pathStr); err != nil {
			t.Fatalf("segment at length limit should pass, got: %v", err)
		}
	})

	t.Run("should fail for key exceeding length limit", func(t *testing.T) {
		pathStr := strings.Repeat("x", maxPathSegmentLen+1)
		_, err := ParsePath(pathStr)
		if !errors.Is(err, lang.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for overlong key segment, got: %v", err)
		}
	})

	t.Run("should fail for index exceeding length limit", func(t *testing.T) {
		longIdx := strings.Repeat("1", maxPathSegmentLen+1)
		pathStr := "a[" + longIdx + "]"
		_, err := ParsePath(pathStr)
		if !errors.Is(err, lang.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for overlong index segment, got: %v", err)
		}
	})
}

func TestPath_ArrayForm_MaxSegmentsBoundary(t *testing.T) {
	t.Run("should fail when exceeding segment limit", func(t *testing.T) {
		longPathArr := make([]any, maxPathSegments+1)
		for i := 0; i < len(longPathArr); i++ {
			longPathArr[i] = "a"
		}
		_, err := buildPathFromArray(longPathArr)
		if !errors.Is(err, lang.ErrNestingDepthExceeded) {
			t.Fatalf("expected ErrNestingDepthExceeded for array-form path over limit, got: %v", err)
		}
	})
}

// --- Shape Limits ---

func TestShape_Parse_MaxDepthBoundary(t *testing.T) {
	t.Run("should pass at exact depth limit", func(t *testing.T) {
		shape := make(map[string]any)
		current := shape
		// Build a shape exactly maxShapeDepth deep
		for i := 0; i < maxShapeDepth; i++ {
			next := make(map[string]any)
			current["next"] = next
			current = next
		}
		current["leaf"] = "string" // Add a field at the deepest level

		if _, err := ParseShape(shape); err != nil {
			t.Fatalf("shape at exact depth limit should parse, got: %v", err)
		}
	})

	t.Run("should fail when exceeding depth limit", func(t *testing.T) {
		shape := make(map[string]any)
		current := shape
		// Build a shape deeper than maxShapeDepth
		for i := 0; i < maxShapeDepth+1; i++ {
			next := make(map[string]any)
			current["next"] = next
			current = next
		}
		current["leaf"] = "string"

		_, err := ParseShape(shape)
		if !errors.Is(err, lang.ErrNestingDepthExceeded) {
			t.Fatalf("expected ErrNestingDepthExceeded, got: %v", err)
		}
	})
}

func TestShape_Validate_MaxDepthBoundary(t *testing.T) {
	// We build a self-referential shape to test validation depth on the data.
	shape := &Shape{Fields: make(map[string]*FieldSpec)}
	fieldSpec := &FieldSpec{Name: "items", IsList: true}
	fieldSpec.NestedShape = shape // The field 'items' contains a list of shapes like itself
	shape.Fields["items"] = fieldSpec

	// Now build a data structure that is deeper than the validation limit
	data := make(map[string]any)
	current := data
	for i := 0; i < maxShapeDepth+1; i++ {
		next := map[string]any{}
		current["items"] = []any{next}
		current = next
	}

	err := shape.Validate(data, false)
	if !errors.Is(err, lang.ErrNestingDepthExceeded) {
		t.Fatalf("expected ErrNestingDepthExceeded for deeply nested data, got: %v", err)
	}
}

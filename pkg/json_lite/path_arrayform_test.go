// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Array-form path tests for path-lite (build Path from []any and use Select)
// filename: pkg/json-lite/path_arrayform_test.go
// nlines: 184
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

// buildPathFromArray converts an array-form path (e.g., []any{"items", 1, "id"})
// into a json_lite.Path. Strings become key segments; ints become index segments.
// Any other type -> ErrInvalidArgument. Nil -> ErrInvalidArgument.
// NOTE: This is test-local glue; production code still uses ParsePath for strings.
func buildPathFromArray(arr []any) (Path, error) {
	if arr == nil {
		return nil, ErrInvalidArgument
	}
	if len(arr) == 0 {
		return nil, ErrInvalidPath
	}
	if len(arr) > maxPathSegments {
		return nil, ErrNestingDepthExceeded
	}
	p := make(Path, 0, len(arr))
	for _, el := range arr {
		switch v := el.(type) {
		case string:
			if v == "" {
				return nil, ErrInvalidArgument
			}
			p = append(p, PathSegment{Key: v, IsKey: true})
		case int:
			p = append(p, PathSegment{Index: v, IsKey: false})
		default:
			return nil, ErrInvalidArgument
		}
	}
	return p, nil
}

func TestArrayForm_Basics(t *testing.T) {
	data := map[string]any{
		"name": "test",
		"meta": map[string]any{"version": 2},
		"items": []any{
			map[string]any{"id": 100},
			map[string]any{"id": 200},
		},
	}

	cases := []struct {
		name    string
		pathArr []any
		want    any
		wantErr error
	}{
		{
			name:    "simple key",
			pathArr: []any{"name"},
			want:    "test",
		},
		{
			name:    "nested key",
			pathArr: []any{"meta", "version"},
			want:    2,
		},
		{
			name:    "list index then key",
			pathArr: []any{"items", 1, "id"},
			want:    200,
		},
		{
			name:    "map key not found",
			pathArr: []any{"meta", "nope"},
			wantErr: ErrMapKeyNotFound,
		},
		{
			name:    "list index OOB",
			pathArr: []any{"items", 5},
			wantErr: ErrListIndexOutOfBounds,
		},
		{
			name:    "cannot access key on list",
			pathArr: []any{"items", "id"},
			wantErr: ErrCannotAccessType,
		},
		{
			name:    "nil arr",
			pathArr: nil,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "empty arr",
			pathArr: []any{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "bad element type",
			pathArr: []any{"items", int64(1)}, // unsupported type
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "negative index",
			pathArr: []any{"items", -1},
			wantErr: ErrListIndexOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := buildPathFromArray(tc.pathArr)
			if tc.wantErr != nil && (tc.pathArr == nil || len(tc.pathArr) == 0 || (len(tc.pathArr) > 0 && (tc.pathArr[0] == nil || (len(tc.pathArr) == 2 && (tc.pathArr[1] == int64(1)))))) {
				// Construction-time errors for invalid inputs
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected constructor error %v, got %v", tc.wantErr, err)
				}
				return
			}
			// If constructor failed unexpectedly
			if err != nil {
				t.Fatalf("constructor failed: %v", err)
			}
			got, selErr := Select(data, path, nil)
			if tc.wantErr != nil {
				if !errors.Is(selErr, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, selErr)
				}
				return
			}
			if selErr != nil {
				t.Fatalf("unexpected select error: %v", selErr)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestArrayForm_WeirdKeys(t *testing.T) {
	// Keys that the string parser treats as structure can be addressed literally via array-form.
	data := map[string]any{
		"a.b":  1,
		"c[0]": 2,
		"nest": map[string]any{
			"x.y":  3,
			"z[1]": 4,
		},
		"list": []any{
			map[string]any{"dot.key": 9},
		},
	}

	cases := []struct {
		name    string
		pathArr []any
		want    any
	}{
		{"top-level dotted key", []any{"a.b"}, 1},
		{"top-level bracket key", []any{"c[0]"}, 2},
		{"nested dotted key", []any{"nest", "x.y"}, 3},
		{"nested bracket key", []any{"nest", "z[1]"}, 4},
		{"list then dotted key", []any{"list", 0, "dot.key"}, 9},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := buildPathFromArray(tc.pathArr)
			if err != nil {
				t.Fatalf("constructor failed: %v", err)
			}
			got, err := Select(data, p, nil)
			if err != nil {
				t.Fatalf("select failed: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestArrayForm_Limits(t *testing.T) {
	// Build a path longer than maxPathSegments to ensure constructor enforces depth.
	long := make([]any, maxPathSegments+1)
	for i := 0; i < len(long); i++ {
		long[i] = "a"
	}
	if _, err := buildPathFromArray(long); !errors.Is(err, ErrNestingDepthExceeded) {
		t.Fatalf("expected ErrNestingDepthExceeded for array-form path over limit, got: %v", err)
	}
}

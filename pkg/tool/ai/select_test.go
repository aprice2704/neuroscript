// NeuroScript Version: 0.3.0
// File version: 7
// Purpose: Provides unit tests for the Select tool. Corrected expected error types.
// filename: pkg/tool/ai/select_test.go
// nlines: 100
// risk_rating: LOW

package ai

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestSelect(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Ada Lovelace",
		},
		"items": []interface{}{
			map[string]interface{}{"id": float64(100)},
			map[string]interface{}{"id": float64(200)},
		},
		"a.b": "literal dot",
	}

	testCases := []struct {
		name      string
		args      []interface{}
		want      interface{}
		wantErrIs error
	}{
		{
			name: "select nested key with string path",
			args: []interface{}{data, "user.name"},
			want: "Ada Lovelace",
		},
		{
			name: "select from list with string path",
			args: []interface{}{data, "items[1].id"},
			want: float64(200),
		},
		{
			name: "select with literal dot key using slice path",
			args: []interface{}{data, []interface{}{"a.b"}},
			want: "literal dot",
		},
		{
			name:      "missing key not ok",
			args:      []interface{}{data, "user.email"},
			wantErrIs: json_lite.ErrMapKeyNotFound,
		},
		{
			name: "missing key ok",
			args: []interface{}{data, "user.email", true},
			want: nil,
		},
		{
			name:      "invalid path string",
			args:      []interface{}{data, "items[1]name"},
			wantErrIs: json_lite.ErrInvalidPath,
		},
		{
			name:      "out of bounds list access",
			args:      []interface{}{data, "items[2]"},
			wantErrIs: json_lite.ErrListIndexOutOfBounds,
		},
		{
			name: "out of bounds list access ok",
			args: []interface{}{data, "items[2]", true},
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Select(interp, tc.args)
			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Select() error = %v, wantErrIs %v", err, tc.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Errorf("Select() unexpected error = %v", err)
			}
			if got != tc.want {
				t.Errorf("Select() = %v (%T), want %v (%T)", got, got, tc.want, tc.want)
			}
		})
	}
}

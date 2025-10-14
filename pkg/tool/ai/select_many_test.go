// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Provides unit tests for the SelectMany tool. Corrected expected error types.
// filename: pkg/tool/ai/select_many_test.go
// nlines: 90
// risk_rating: LOW

package ai

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestSelectMany(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":    "Ada Lovelace",
			"address": map[string]interface{}{"city": "London"},
		},
		"items": []interface{}{
			map[string]interface{}{"id": float64(100)},
			map[string]interface{}{"id": float64(200)},
		},
	}

	testCases := []struct {
		name      string
		args      []interface{}
		want      map[string]interface{}
		wantErrIs error
	}{
		{
			name: "extract multiple keys",
			args: []interface{}{data, map[string]interface{}{
				"username":  "user.name",
				"firstItem": "items[0].id",
			}},
			want: map[string]interface{}{
				"username":  "Ada Lovelace",
				"firstItem": float64(100),
			},
		},
		{
			name: "missing key not ok",
			args: []interface{}{data, map[string]interface{}{
				"email": "user.email",
			}},
			wantErrIs: json_lite.ErrMapKeyNotFound,
		},
		{
			name: "missing key ok",
			args: []interface{}{data, map[string]interface{}{
				"username": "user.name",
				"email":    "user.email",
			}, true},
			want: map[string]interface{}{
				"username": "Ada Lovelace",
			},
		},
		{
			name: "invalid path in extracts",
			args: []interface{}{data, map[string]interface{}{
				"bad": "user..name",
			}},
			wantErrIs: json_lite.ErrInvalidPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectMany(interp, tc.args)
			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("SelectMany() error = %v, wantErrIs %v", err, tc.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Errorf("SelectMany() unexpected error = %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SelectMany() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

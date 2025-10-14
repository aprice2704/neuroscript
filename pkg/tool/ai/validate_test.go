// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Provides unit tests for the Validate tool. Corrected expected error types.
// filename: pkg/tool/ai/validate_test.go
// nlines: 85
// risk_rating: LOW

package ai

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestValidate(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}

	shape := map[string]interface{}{
		"name":     "string",
		"email":    "email",
		"company?": "string",
	}

	testCases := []struct {
		name      string
		args      []interface{}
		want      bool
		wantErrIs error
	}{
		{
			name: "valid data",
			args: []interface{}{
				map[string]interface{}{
					"name":  "Ada Lovelace",
					"email": "ada@example.com",
				},
				shape,
				false,
			},
			want: true,
		},
		{
			name: "missing required field",
			args: []interface{}{
				map[string]interface{}{
					"name": "Ada Lovelace",
				},
				shape,
				false,
			},
			wantErrIs: json_lite.ErrValidationRequiredArgMissing,
		},
		{
			name: "invalid email format",
			args: []interface{}{
				map[string]interface{}{
					"name":  "Ada Lovelace",
					"email": "ada",
				},
				shape,
				false,
			},
			wantErrIs: json_lite.ErrValidationFailed,
		},
		{
			name: "extra field not allowed",
			args: []interface{}{
				map[string]interface{}{
					"name":  "Ada Lovelace",
					"email": "ada@example.com",
					"extra": "field",
				},
				shape,
				false,
			},
			wantErrIs: json_lite.ErrInvalidArgument,
		},
		{
			name: "extra field allowed",
			args: []interface{}{
				map[string]interface{}{
					"name":  "Ada Lovelace",
					"email": "ada@example.com",
					"extra": "field",
				},
				shape,
				true,
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Validate(interp, tc.args)
			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Validate() error = %v, wantErrIs %v", err, tc.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
			if got != tc.want {
				t.Errorf("Validate() = %v, want %v", got, tc.want)
			}
		})
	}
}

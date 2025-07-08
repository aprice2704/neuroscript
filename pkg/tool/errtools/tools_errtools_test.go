// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Provides unit tests for the 'Error.New' tool implementation.
// filename: pkg/tool/errtools/tools_error_test.go
// nlines: 83
// risk_rating: LOW

package errtools

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolErrorNew(t *testing.T) {
	// The tool.Runtime argument is not used by this function, so we can pass nil.
	var mockInterpreter tool.Runtime = nil

	testCases := []struct {
		name    string
		args    []interface{}
		wantVal interface{}
		wantErr bool
	}{
		{
			name: "happy path with string code",
			args: []interface{}{"ERR_NOT_FOUND", "File could not be located."},
			wantVal: lang.ErrorValue{Value: map[string]lang.Value{
				"code":    lang.StringValue{Value: "ERR_NOT_FOUND"},
				"message": lang.StringValue{Value: "File could not be located."},
			}},
			wantErr: false,
		},
		{
			name: "happy path with integer code",
			args: []interface{}{404, "Resource not found."},
			wantVal: lang.ErrorValue{Value: map[string]lang.Value{
				"code":    lang.NumberValue{Value: 404.0},
				"message": lang.StringValue{Value: "Resource not found."},
			}},
			wantErr: false,
		},
		{
			name: "happy path with float code",
			args: []interface{}{500.1, "Internal server error variant."},
			wantVal: lang.ErrorValue{Value: map[string]lang.Value{
				"code":    lang.NumberValue{Value: 500.1},
				"message": lang.StringValue{Value: "Internal server error variant."},
			}},
			wantErr: false,
		},
		{
			name:    "unhappy path - too few arguments",
			args:    []interface{}{"ERR_ALONE"},
			wantErr: true,
		},
		{
			name:    "unhappy path - too many arguments",
			args:    []interface{}{"ERR_CODE", "message", "extra"},
			wantErr: true,
		},
		{
			name:    "unhappy path - message is not a string",
			args:    []interface{}{"ERR_CODE", 12345},
			wantErr: true,
		},
		{
			name:    "unhappy path - code is not string or number",
			args:    []interface{}{true, "A boolean is not a valid code."},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toolErrorNew(mockInterpreter, tc.args)

			if (err != nil) != tc.wantErr {
				t.Fatalf("toolErrorNew() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr && !reflect.DeepEqual(got, tc.wantVal) {
				t.Errorf("toolErrorNew() got = %v, want %v", got, tc.wantVal)
			}
		})
	}
}

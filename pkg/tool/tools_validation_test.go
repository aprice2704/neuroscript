// :: product: NS
// :: majorVersion: 1
// :: fileVersion: 4
// :: description: Removes tests for variadic arguments and adds identity type validation.
// :: latestChange: Added validation test cases for NodeID, EntityID, and Handle types.
// :: filename: pkg/tool/tools_validation_test.go
// :: serialization: go

package tool

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestValidateAndCoerceArgs(t *testing.T) {
	// --- Mock Tool Specs ---
	specSimple := ToolSpec{
		FullName: "test.simple",
		Args: []ArgSpec{
			{Name: "arg1", Type: ArgTypeString, Required: true},
		},
		ReturnType: ArgTypeBool,
	}
	specIdentity := ToolSpec{
		FullName: "test.identity",
		Args: []ArgSpec{
			{Name: "node", Type: ArgTypeNodeID, Required: true},
			{Name: "entity", Type: ArgTypeEntityID, Required: false},
			{Name: "hdl", Type: ArgTypeHandle, Required: false},
		},
		ReturnType: ArgTypeNil,
	}

	// --- Test Cases ---
	testCases := []struct {
		name       string
		spec       ToolSpec
		rawArgs    []any
		expected   []any
		wantErr    bool
		wantErrIs  error  // Expected sentinel error type
		wantErrMsg string // Substring to check in error message
		skipDeepEq bool   // Skip DeepEqual check if error is expected
	}{
		// === Simple Spec ===
		{"Simple OK", specSimple, []any{"hello"}, []any{"hello"}, false, nil, "", false},
		{"Simple Wrong Type", specSimple, []any{123}, nil, true, lang.ErrInvalidArgument, "argument 'arg1': expected string, got int", true},

		// === Identity Spec ===
		{"Identity OK Min", specIdentity, []any{"N_123"}, []any{"N_123", nil, nil}, false, nil, "", false},
		{"Identity OK Max", specIdentity, []any{"N_123", "E_456", "sys.user"}, []any{"N_123", "E_456", "sys.user"}, false, nil, "", false},
		{"Identity Invalid NodeID", specIdentity, []any{"E_123"}, nil, true, lang.ErrInvalidArgument, "argument 'node': invalid NodeID", true},
		{"Identity Invalid EntityID", specIdentity, []any{"N_123", "N_456"}, nil, true, lang.ErrInvalidArgument, "argument 'entity': invalid EntityID", true},
		{"Identity Invalid Handle", specIdentity, []any{"N_123", "E_456", "bad handle!"}, nil, true, lang.ErrInvalidArgument, "argument 'hdl': invalid handle format", true},
		{"Identity Missing Required", specIdentity, []any{nil}, nil, true, lang.ErrInvalidArgument, "argument 'node' is required", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coerced, err := validateAndCoerceArgs(tc.spec.FullName, tc.rawArgs, tc.spec)

			// Check error presence
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateAndCoerceArgs() error = %v, wantErr %v", err, tc.wantErr)
			}

			// If error expected, check type and message substring
			if tc.wantErr {
				if tc.wantErrIs != nil {
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("validateAndCoerceArgs() error type = %T, want error type %T", err, tc.wantErrIs)
					}
				}
				if tc.wantErrMsg != "" {
					re, ok := err.(*lang.RuntimeError)
					if !ok {
						t.Fatalf("Expected error to be *lang.RuntimeError, got %T", err)
					}
					if !strings.Contains(re.Message, tc.wantErrMsg) {
						t.Errorf("validateAndCoerceArgs() error message = %q, want message containing %q", re.Message, tc.wantErrMsg)
					}
				}
				return // Skip value check
			}

			// If no error expected, check coerced values
			if !tc.skipDeepEq {
				expectedLen := len(tc.spec.Args)
				if len(coerced) != expectedLen {
					t.Fatalf("validateAndCoerceArgs() length mismatch: got %d elements, want %d", len(coerced), expectedLen)
				}

				if !reflect.DeepEqual(coerced, tc.expected) {
					t.Errorf("validateAndCoerceArgs() value mismatch:\ngot = %#v\nwant= %#v", coerced, tc.expected)
				}
			}
		})
	}
}

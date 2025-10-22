// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Removes tests for variadic arguments.
// filename: pkg/tool/tools_validation_test.go
// nlines: 168
// risk_rating: LOW

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
	specOptional := ToolSpec{
		FullName: "test.optional",
		Args: []ArgSpec{
			{Name: "req", Type: ArgTypeInt, Required: true},
			{Name: "optStr", Type: ArgTypeString, Required: false},
			{Name: "optBool", Type: ArgTypeBool, Required: false},
		},
		ReturnType: ArgTypeNil,
	}
	specSlice := ToolSpec{
		FullName: "test.slice",
		Args: []ArgSpec{
			{Name: "reqSlice", Type: ArgTypeSliceString, Required: true},
			{Name: "optSlice", Type: ArgTypeSliceInt, Required: false},
		},
		ReturnType: ArgTypeNil,
	}
	// REMOVED specVariadic

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
		{"Simple Too Few Args", specSimple, []any{}, nil, true, lang.ErrArgumentMismatch, "expected 1 arguments, got 0", true},
		{"Simple Too Many Args", specSimple, []any{"hello", 123}, nil, true, lang.ErrArgumentMismatch, "expected 1 arguments, got 2", true},
		{"Simple Wrong Type", specSimple, []any{123}, nil, true, lang.ErrInvalidArgument, "argument 'arg1': expected string, got int", true},
		{"Simple Missing Required", specSimple, []any{nil}, nil, true, lang.ErrInvalidArgument, "argument 'arg1' is required", true},

		// === Optional Spec ===
		{"Optional OK Min", specOptional, []any{int64(10)}, []any{int64(10), nil, nil}, false, nil, "", false},
		{"Optional OK Mid", specOptional, []any{int64(10), "opt"}, []any{int64(10), "opt", nil}, false, nil, "", false},
		{"Optional OK Max", specOptional, []any{int64(10), "opt", true}, []any{int64(10), "opt", true}, false, nil, "", false},
		{"Optional Nil Allowed", specOptional, []any{int64(10), nil, false}, []any{int64(10), nil, false}, false, nil, "", false},
		{"Optional Too Few", specOptional, []any{}, nil, true, lang.ErrArgumentMismatch, "expected 1 to 3 arguments, got 0", true},
		{"Optional Too Many", specOptional, []any{1, "a", true, 4}, nil, true, lang.ErrArgumentMismatch, "expected 1 to 3 arguments, got 4", true},
		{"Optional Wrong Type Required", specOptional, []any{"wrong"}, nil, true, lang.ErrInvalidArgument, "argument 'req': expected integer", true},
		{"Optional Wrong Type Optional", specOptional, []any{1, 123, true}, nil, true, lang.ErrInvalidArgument, "argument 'optStr': expected string", true},
		{"Optional Missing Required", specOptional, []any{nil, "opt"}, nil, true, lang.ErrInvalidArgument, "argument 'req' is required", true},

		// === Slice Spec ===
		{"Slice OK Min", specSlice, []any{[]string{"a", "b"}}, []any{[]string{"a", "b"}, nil}, false, nil, "", false},
		{"Slice OK Max", specSlice, []any{[]string{"a", "b"}, []int64{1, 2}}, []any{[]string{"a", "b"}, []int64{1, 2}}, false, nil, "", false},
		{"Slice OK Nil Optional", specSlice, []any{[]string{"a", "b"}, nil}, []any{[]string{"a", "b"}, nil}, false, nil, "", false},
		{"Slice Missing Required", specSlice, []any{nil}, nil, true, lang.ErrInvalidArgument, "argument 'reqSlice' is required", true},
		{"Slice Wrong Type Required", specSlice, []any{123}, nil, true, lang.ErrInvalidArgument, "argument 'reqSlice': expected slice of strings, got int", true},
		{"Slice Wrong Type Optional", specSlice, []any{[]string{"a"}, 123}, nil, true, lang.ErrInvalidArgument, "argument 'optSlice': expected a slice (list), got int", true},
		{"Slice Wrong Elem Type Required", specSlice, []any{[]any{"a", 1}}, nil, true, lang.ErrInvalidArgument, "argument 'reqSlice': expected slice of strings, but element 1 has incompatible type int", true},
		{"Slice Wrong Elem Type Optional", specSlice, []any{[]string{"a"}, []any{1, "b"}}, nil, true, lang.ErrInvalidArgument, "argument 'optSlice': element 1 (string) could not be converted to int64", true},

		// === Variadic Spec (REMOVED) ===
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
				// Coerced args should now always have length equal to numSpecArgs
				expectedLen := len(tc.spec.Args)
				if len(coerced) != expectedLen {
					t.Fatalf("validateAndCoerceArgs() length mismatch: got %d elements, want %d.\nGot: %#v\nWant: %#v", len(coerced), expectedLen, coerced, tc.expected)
				}

				if !reflect.DeepEqual(coerced, tc.expected) {
					t.Errorf("validateAndCoerceArgs() value mismatch:\ngot = %#v\nwant= %#v", coerced, tc.expected)
				}
			}
		})
	}
}

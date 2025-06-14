// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to test the primitive-based Input tool implementation directly.
// filename: pkg/core/tools_io_test.go
// nlines: 48
// risk_rating: LOW

package core

import (
	"errors"
	"testing"
)

func TestToolIOInputValidation(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	toolImpl, _ := interp.ToolRegistry().GetTool("Input")

	testCases := []struct {
		name      string
		args      []interface{}
		wantErrIs error
	}{
		{name: "Valid prompt (string)", args: MakeArgs("Enter name: "), wantErrIs: nil},
		{name: "No arguments (optional prompt)", args: MakeArgs(), wantErrIs: nil},
		{name: "Valid argument type (nil for optional)", args: MakeArgs(nil), wantErrIs: nil},
		{name: "Incorrect argument type (number)", args: MakeArgs(123), wantErrIs: ErrInvalidArgument},
		{name: "Incorrect argument type (bool)", args: MakeArgs(true), wantErrIs: ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't test the stdin reading part automatically, so we just check
			// if the function handles the arguments without returning an unexpected error.
			// We expect an EOF error from trying to read stdin in a non-interactive test.
			_, err := toolImpl.Func(interp, tc.args)

			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, err)
				}
			} else {
				// In a non-interactive environment, reading from stdin should fail.
				// We expect an I/O error here, NOT a nil error.
				if err == nil {
					t.Errorf("Expected an I/O error from reading stdin, but got nil")
				} else if !errors.Is(err, ErrIOFailed) {
					t.Logf("Got an error as expected, but it wasn't the specific I/O error: %v", err)
				}
			}
		})
	}
}

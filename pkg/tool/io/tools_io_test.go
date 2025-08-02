// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Corrected tool lookup to use the fully qualified name, fixing the nil pointer panic.
// filename: pkg/tool/io/tools_io_test.go
// nlines: 48
// risk_rating: LOW

package io

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// MakeArgs is a convenience function to create a slice of interfaces, useful for constructing tool arguments programmatically.
func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

func TestToolIOInputValidation(t *testing.T) {
	interp := interpreter.NewInterpreter()
	for _, toolImpl := range ioToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}

	testCases := []struct {
		name      string
		args      []interface{}
		wantErrIs error
	}{
		{name: "Valid prompt (string)", args: MakeArgs("Enter name: "), wantErrIs: nil},
		{name: "No arguments (optional prompt)", args: MakeArgs(), wantErrIs: nil},
		{name: "Valid argument type (nil for optional)", args: MakeArgs(nil), wantErrIs: nil},
		{name: "Incorrect argument type (number)", args: MakeArgs(123), wantErrIs: lang.ErrInvalidArgument},
		{name: "Incorrect argument type (bool)", args: MakeArgs(true), wantErrIs: lang.ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't test the stdin reading part automatically, so we just check
			// if the function handles the arguments without returning an unexpected error.
			// We expect an EOF error from trying to read stdin in a non-interactive test.
			fullName := types.MakeFullName(group, "Input")
			toolImpl, found := interp.ToolRegistry().GetTool(fullName)
			if !found {
				t.Fatalf("Tool %q not found in registry", fullName)
			}
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
				} else if !errors.Is(err, lang.ErrIOFailed) {
					// It's okay if we get a different I/O error, just log it.
					t.Logf("Got an error as expected, but it wasn't the specific I/O error: %v", err)
				}
			}
		})
	}
}

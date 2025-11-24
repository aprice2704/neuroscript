// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Tests for the 'handle' toolset.
// filename: pkg/tool/handle/tools_test.go
// nlines: 110
// risk_rating: LOW

package handle

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// newHandleTestInterpreter creates a self-contained interpreter for testing.
func newHandleTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}

	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(policy.AllowAll()),
	)
	return interp
}

func TestToolHandle(t *testing.T) {
	interp := newHandleTestInterpreter(t)

	// Setup: Create a valid handle in the interpreter's registry
	reg := interp.HandleRegistry()
	if reg == nil {
		t.Fatal("Interpreter HandleRegistry is nil")
	}

	// Create a "live" handle
	validHandle, err := reg.NewHandle("some payload", "test_kind")
	if err != nil {
		t.Fatalf("Failed to create test handle: %v", err)
	}

	// Create a handle value that points to nothing (simulate stale/invalid)
	// We construct a HandleValue manually if possible, or create and delete.
	staleHandle, _ := reg.NewHandle("to delete", "test_kind")
	reg.DeleteHandle(staleHandle.HandleID())

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// --- Type ---
		{
			name:       "Type Valid",
			toolName:   "Type",
			args:       []interface{}{validHandle},
			wantResult: "test_kind",
		},
		{
			name:       "Type Stale (Still has type info locally)",
			toolName:   "Type",
			args:       []interface{}{staleHandle},
			wantResult: "test_kind",
		},
		{
			name:      "Type Invalid Arg",
			toolName:  "Type",
			args:      []interface{}{"not a handle"},
			wantErrIs: lang.ErrArgumentMismatch,
		},

		// --- IsValid ---
		{
			name:       "IsValid True",
			toolName:   "IsValid",
			args:       []interface{}{validHandle},
			wantResult: true,
		},
		{
			name:       "IsValid False (Deleted)",
			toolName:   "IsValid",
			args:       []interface{}{staleHandle},
			wantResult: false,
		},
		{
			name:      "IsValid Invalid Arg",
			toolName:  "IsValid",
			args:      []interface{}{123},
			wantErrIs: lang.ErrArgumentMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullname := types.MakeFullName(group, tt.toolName)
			toolImpl, found := interp.ToolRegistry().GetTool(fullname)
			if !found {
				t.Fatalf("Tool %q not found", fullname)
			}

			got, err := toolImpl.Func(interp, tt.args)

			if tt.wantErrIs != nil {
				if err == nil {
					t.Errorf("Expected an error wrapping [%v], but got nil", tt.wantErrIs)
				} else if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Expected error to wrap [%v], but got: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("Result mismatch:\n  Got:  %#v\n  Want: %#v", got, tt.wantResult)
			}
		})
	}
}

// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to test the primitive-based shell tool implementation directly.
// filename: pkg/core/tools_shell_test.go
// nlines: 106
// risk_rating: HIGH

package core

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

// testShellToolHelper tests the toolExecuteCommand implementation directly.
func testShellToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name         string
	args         []interface{}
	wantSuccess  bool
	wantExitCode int64
	wantStdout   string
	wantStderr   string // Check if stderr *contains* this string
	wantErrIs    error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		toolImpl, _ := interp.ToolRegistry().GetTool("Shell.Execute")
		gotResult, toolErr := toolImpl.Func(interp, tc.args)

		if tc.wantErrIs != nil {
			if !errors.Is(toolErr, tc.wantErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, toolErr)
			}
			return
		}
		if toolErr != nil {
			t.Fatalf("Unexpected error: %v", toolErr)
		}

		resultMap, ok := gotResult.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map[string]interface{}, but got %T", gotResult)
		}

		if got, want := resultMap["success"].(bool), tc.wantSuccess; got != want {
			t.Errorf("Mismatch in 'success' field. Got: %v, Want: %v", got, want)
		}
		if got, want := resultMap["exit_code"].(int64), tc.wantExitCode; got != want {
			t.Errorf("Mismatch in 'exit_code' field. Got: %v, Want: %v", got, want)
		}
		if got, want := resultMap["stdout"].(string), tc.wantStdout; got != want {
			t.Errorf("Mismatch in 'stdout' field.\nGot:\n%s\nWant:\n%s", got, want)
		}
		if stderr, ok := resultMap["stderr"].(string); ok {
			if tc.wantStderr != "" && !strings.Contains(stderr, tc.wantStderr) {
				t.Errorf("Mismatch in 'stderr' field.\nGot:\n%s\nDid not contain:\n%s", stderr, tc.wantStderr)
			}
		} else {
			t.Errorf("stderr field was not a string")
		}
	})
}

func TestToolExecuteCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell command tests on Windows")
	}
	interp, _ := NewDefaultTestInterpreter(t)
	sandboxDir := t.TempDir()
	if err := interp.SetSandboxDir(sandboxDir); err != nil {
		t.Fatalf("Failed to set sandbox dir: %v", err)
	}

	tests := []struct {
		name         string
		args         []interface{}
		wantSuccess  bool
		wantExitCode int64
		wantStdout   string
		wantStderr   string
		wantErrIs    error
	}{
		{name: "Simple Echo", args: MakeArgs("echo", []string{"hello"}), wantSuccess: true, wantExitCode: 0, wantStdout: "hello\n"},
		{name: "Command False Failure", args: MakeArgs("false"), wantSuccess: false, wantExitCode: 1},
		{name: "Command Not Found", args: MakeArgs("nonexistent_command_ajsdflk"), wantSuccess: false, wantExitCode: -1, wantStderr: "executable file not found"},
		{name: "Run in specified dir (pwd)", args: MakeArgs("pwd", nil, "."), wantSuccess: true, wantExitCode: 0, wantStdout: sandboxDir + "\n"},
		{name: "Directory outside sandbox", args: MakeArgs("pwd", nil, "../escaped"), wantErrIs: ErrPathViolation},
		{name: "Invalid Command Arg Type", args: MakeArgs(123), wantErrIs: ErrInvalidArgument},
		{name: "Invalid Args_list Type", args: MakeArgs("echo", "not-a-list"), wantErrIs: ErrInvalidArgument},
		{name: "Invalid Dir Type", args: MakeArgs("echo", nil, 123), wantErrIs: ErrInvalidArgument},
	}

	for _, tt := range tests {
		testShellToolHelper(t, interp, tt)
	}
}

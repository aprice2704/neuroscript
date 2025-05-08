// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 20:13:11 PDT // Updated timestamp
// filename: pkg/core/tools_shell_test.go
package core

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// Assume NewTestInterpreter and MakeArgs are defined elsewhere

// --- Test ToolExecuteCommand ---
func TestToolExecuteCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell command tests on Windows")
	}
	dummyInterp, _ := NewDefaultTestInterpreter(t)
	sandboxDir := t.TempDir()
	dummyInterp.SetSandboxDir(sandboxDir)

	tests := []struct {
		name         string
		command      string
		cmdArgs      interface{} // Holds intended args for command, or malformed rawArgs for validation tests
		dirArg       interface{}
		wantStdout   string
		wantStderr   string
		wantExitCode int64
		wantSuccess  bool
		valWantErrIs error
	}{
		// Test cases ...
		{name: "Simple Echo Success", command: "echo", cmdArgs: []string{"hello", "world"}, dirArg: nil, wantStdout: "hello world\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command True Success", command: "true", cmdArgs: nil, dirArg: nil, wantStdout: "", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command False Failure", command: "false", cmdArgs: []string{}, dirArg: nil, wantStdout: "", wantStderr: "", wantExitCode: 1, wantSuccess: false},
		{name: "Command Writes to Stderr", command: "sh", cmdArgs: []string{"-c", "echo 'error output' >&2"}, dirArg: nil, wantStdout: "", wantStderr: "error output\n", wantExitCode: 0, wantSuccess: true},
		{name: "Command Not Found", command: "nonexistent_command_ajsdflk", cmdArgs: nil, dirArg: nil, wantStdout: "", wantStderr: "executable file not found", wantExitCode: -1, wantSuccess: false},
		{name: "Run in specified dir (pwd)", command: "pwd", cmdArgs: nil, dirArg: ".", wantStdout: sandboxDir + "\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Validation Valid Arg Count (Only Command)", command: "echo", cmdArgs: nil, dirArg: nil, wantStdout: "\n", wantStderr: "", wantExitCode: 0, wantSuccess: true, valWantErrIs: nil}, // Corrected expectation
		{name: "Validation Wrong Arg Type (Command not string)", command: "", cmdArgs: MakeArgs(123, []string{}, nil), valWantErrIs: ErrValidationTypeMismatch},                                  // cmdArgs holds the bad rawArgs from MakeArgs
		{name: "Validation Wrong Arg Type (Args not string slice)", command: "echo", cmdArgs: MakeArgs("echo", "not_a_slice", nil), valWantErrIs: ErrValidationTypeMismatch},                     // cmdArgs holds the bad rawArgs from MakeArgs
		{name: "Validation Wrong Arg Type (Dir not string)", command: "echo", cmdArgs: MakeArgs("echo", []string{}, 123), valWantErrIs: ErrValidationTypeMismatch},                               // cmdArgs holds the bad rawArgs from MakeArgs
		{name: "Directory outside sandbox", command: "pwd", cmdArgs: nil, dirArg: "../escaped", wantStderr: "ExecuteCommand path validation failed", wantExitCode: -1, wantSuccess: false, valWantErrIs: nil},
	}

	toolImpl, found := dummyInterp.ToolRegistry().GetTool("Shell.Execute")
	if !found {
		t.Fatalf("Tool %q not found in registry", "Shell.Execute")
	}
	spec := toolImpl.Spec

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			// Construct rawArgs based on test case structure
			if strings.Contains(tt.name, "Validation Wrong Arg Type") {
				// For these specific tests, tt.cmdArgs holds the pre-constructed bad []interface{} from MakeArgs
				var ok bool
				// --- Apply Type Assertion ---
				rawArgs, ok = tt.cmdArgs.([]interface{}) // Assert tt.cmdArgs holds the []interface{} from MakeArgs
				if !ok {
					t.Fatalf("Test setup error: Expected tt.cmdArgs to hold []interface{} for validation test %q, but got %T", tt.name, tt.cmdArgs)
				}
				// --- End Assertion ---
			} else {
				// Construct args for execution or other validation tests: command, args_list, directory
				var argsListForTool []string
				// Handle conversion from test case's cmdArgs (which holds intended command args)
				if tt.cmdArgs != nil {
					switch v := tt.cmdArgs.(type) {
					case []string:
						argsListForTool = v
					case []interface{}:
						// Convert []interface{} to []string
						argsListForTool = make([]string, len(v))
						for i, item := range v {
							argsListForTool[i] = fmt.Sprintf("%v", item)
						}
					default:
						// This case should ideally not be hit if test setup is correct for execution tests
						t.Fatalf("Test setup error: cmdArgs has unexpected type %T for execution test %q", tt.cmdArgs, tt.name)
					}
				} // If tt.cmdArgs is nil, argsListForTool remains nil/empty, which is fine

				rawArgs = MakeArgs(tt.command, argsListForTool, tt.dirArg)
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)

			if tt.valWantErrIs != nil {
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil (Raw Args: %#v)", tt.valWantErrIs, rawArgs)
				} else if !errors.Is(valErr, tt.valWantErrIs) {
					t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tt.valWantErrIs, valErr, valErr)
				}
				return // Stop if validation failed as expected
			}
			if valErr != nil && tt.valWantErrIs == nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// --- Execution Checks (Only if Validation Passed) ---
			gotInterface, toolErr := toolExecuteCommand(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolExecuteCommand() returned unexpected Go error: %v", toolErr)
			}
			gotMap, ok := gotInterface.(map[string]interface{})
			if !ok {
				t.Fatalf("toolExecuteCommand() did not return a map, got %T", gotInterface)
			}
			gotSuccess, _ := gotMap["success"].(bool)
			if gotSuccess != tt.wantSuccess {
				t.Errorf("success field: got %v, want %v", gotSuccess, tt.wantSuccess)
			}
			gotExitCode, _ := gotMap["exit_code"].(int64)

			if tt.wantExitCode == -1 {
				if gotSuccess {
					t.Errorf("Expected failure (success=false), but got success=true")
				}
			} else if gotExitCode != tt.wantExitCode {
				t.Errorf("exit_code field: got %d, want %d", gotExitCode, tt.wantExitCode)
			}

			gotStdout, _ := gotMap["stdout"].(string)
			if strings.TrimSpace(gotStdout) != strings.TrimSpace(tt.wantStdout) {
				t.Errorf("stdout field:\ngot:  %q\nwant: %q", gotStdout, tt.wantStdout)
			}

			gotStderr, _ := gotMap["stderr"].(string)
			if tt.wantStderr != "" {
				if !strings.Contains(gotStderr, tt.wantStderr) {
					t.Errorf("stderr field:\ngot:  %q\ndoes not contain: %q", gotStderr, tt.wantStderr)
				}
			} else if tt.wantStderr == "" && gotStderr != "" {
				if tt.wantSuccess {
					t.Errorf("stderr field: got %q, want empty", gotStderr)
				} else {
					t.Logf("stderr field (on intended failure): got %q", gotStderr)
				}
			}
		})
	}
}

// NeuroScript Version: 0.3.0
// File version: 0.1.7
// Add execution expectations to the valid validation test case.
// nlines: 145
// risk_rating: LOW
// filename: pkg/core/tools_shell_test.go
package core

import (
	"errors"
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
	err := dummyInterp.SetSandboxDir(sandboxDir) // Ensure sandbox is set
	if err != nil {
		t.Fatalf("Failed to set sandbox dir: %v", err)
	}

	tests := []struct {
		name          string
		command       string
		cmdArgs       interface{} // Holds intended args for command, or malformed rawArgs for validation tests
		dirArg        interface{}
		wantStdout    string
		wantStderr    string
		wantExitCode  int64
		wantSuccess   bool
		valWantErrIs  error // Expected validation error
		toolWantErrIs error // Expected tool execution error
	}{
		{name: "Simple Echo Success", command: "echo", cmdArgs: []string{"hello", "world"}, dirArg: nil, wantStdout: "hello world\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command True Success", command: "true", cmdArgs: nil, dirArg: nil, wantStdout: "", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command False Failure", command: "false", cmdArgs: []string{}, dirArg: nil, wantStdout: "", wantStderr: "", wantExitCode: 1, wantSuccess: false},
		{name: "Command Writes to Stderr", command: "sh", cmdArgs: []string{"-c", "echo 'error output' >&2"}, dirArg: nil, wantStdout: "", wantStderr: "error output\n", wantExitCode: 0, wantSuccess: true},
		{name: "Command Not Found", command: "nonexistent_command_ajsdflk", cmdArgs: nil, dirArg: nil, wantStdout: "", wantStderr: "executable file not found", wantExitCode: -1, wantSuccess: false}, // Specific check for stderr content
		{name: "Run in specified dir (pwd)", command: "pwd", cmdArgs: nil, dirArg: ".", wantStdout: sandboxDir + "\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		// Corrected: Added wantSuccess etc. since validation passes and execution occurs
		{name: "Validation_Valid_Arg_Count_(Only_Command)", command: "echo", cmdArgs: MakeArgs("echo", nil, nil), valWantErrIs: nil, wantStdout: "\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Validation Wrong Arg Type (Command not string)", command: "", cmdArgs: MakeArgs(123, []string{}, nil), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Wrong Arg Type (Args not string slice)", command: "echo", cmdArgs: MakeArgs("echo", "not_a_slice", nil), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Wrong Arg Type (Dir not string)", command: "echo", cmdArgs: MakeArgs("echo", []string{}, 123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Directory_outside_sandbox", command: "pwd", cmdArgs: nil, dirArg: "../escaped", toolWantErrIs: ErrPathViolation},
		{name: "Validation Missing Command Arg", command: "", cmdArgs: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
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
			// For validation tests, tt.cmdArgs now holds the *entire* raw arg list from MakeArgs
			if strings.HasPrefix(tt.name, "Validation") {
				var ok bool
				rawArgs, ok = tt.cmdArgs.([]interface{}) // Assert tt.cmdArgs holds the []interface{} from MakeArgs
				if !ok {
					// Handle the case where cmdArgs might be nil for the missing arg test
					if tt.name == "Validation Missing Command Arg" && tt.cmdArgs == nil {
						rawArgs = MakeArgs() // Ensure rawArgs is an empty slice
					} else {
						t.Fatalf("Test setup error: Expected tt.cmdArgs to hold []interface{} for validation test %q, but got %T", tt.name, tt.cmdArgs)
					}
				}
			} else {
				// Construct args for execution tests: command, args_list, directory
				var argsListForTool []string
				if tt.cmdArgs != nil {
					switch v := tt.cmdArgs.(type) {
					case []string:
						argsListForTool = v
					case []interface{}: // Allow []interface{} for flexibility if needed
						argsListForTool = make([]string, 0, len(v))
						for _, item := range v {
							if strItem, ok := item.(string); ok {
								argsListForTool = append(argsListForTool, strItem)
							} else {
								// Handle non-string items if necessary, or fail the test setup
								t.Fatalf("Test setup error: cmdArgs contains non-string element %T for execution test %q", item, tt.name)
							}
						}
					default:
						t.Fatalf("Test setup error: cmdArgs has unexpected type %T for execution test %q", tt.cmdArgs, tt.cmdArgs)
					}
				} // If tt.cmdArgs is nil, argsListForTool remains nil/empty

				rawArgs = MakeArgs(tt.command, argsListForTool, tt.dirArg)
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)

			// --- Check Validation Error ---
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

			// --- Check Tool Execution Error ---
			if tt.toolWantErrIs != nil {
				if toolErr == nil {
					t.Errorf("toolExecuteCommand() expected error [%v], but got nil. Result: %+v", tt.toolWantErrIs, gotInterface)
				} else if !errors.Is(toolErr, tt.toolWantErrIs) {
					t.Errorf("toolExecuteCommand() expected error type [%v], but got type [%T] with value: %v", tt.toolWantErrIs, toolErr, toolErr)
				}
				// Don't check result map content if a tool error was expected
				return
			}
			if toolErr != nil && tt.toolWantErrIs == nil {
				t.Fatalf("toolExecuteCommand() returned unexpected Go error: %v", toolErr)
			}

			// --- Check Result Map Content (Only if No Tool Error Expected/Occurred) ---
			gotMap, ok := gotInterface.(map[string]interface{})
			if !ok {
				t.Fatalf("toolExecuteCommand() did not return a map, got %T", gotInterface)
			}
			gotSuccess, _ := gotMap["success"].(bool)
			if gotSuccess != tt.wantSuccess {
				t.Errorf("success field: got %v, want %v", gotSuccess, tt.wantSuccess)
			}
			gotExitCode, _ := gotMap["exit_code"].(int64)

			// Allow -1 exit code check for specific failures like command not found
			if tt.wantExitCode == -1 {
				if gotSuccess {
					t.Errorf("Expected failure (success=false), but got success=true")
				}
				// For -1, don't strictly check the exit code value, just that it failed
				// Optionally check stderr contains expected message
				gotStderr, _ := gotMap["stderr"].(string)
				if tt.wantStderr != "" && !strings.Contains(gotStderr, tt.wantStderr) {
					t.Errorf("stderr field on expected failure:\ngot:  %q\ndoes not contain: %q", gotStderr, tt.wantStderr)
				}
			} else if gotExitCode != tt.wantExitCode {
				t.Errorf("exit_code field: got %d, want %d", gotExitCode, tt.wantExitCode)
			}

			gotStdout, _ := gotMap["stdout"].(string)
			if strings.TrimSpace(gotStdout) != strings.TrimSpace(tt.wantStdout) {
				t.Errorf("stdout field:\ngot:  %q\nwant: %q", gotStdout, tt.wantStdout)
			}

			gotStderr, _ := gotMap["stderr"].(string)
			if tt.wantStderr != "" && tt.wantExitCode != -1 { // Don't double-check stderr if already checked for -1 exit code
				if !strings.Contains(gotStderr, tt.wantStderr) {
					t.Errorf("stderr field:\ngot:  %q\ndoes not contain: %q", gotStderr, tt.wantStderr)
				}
			} else if tt.wantStderr == "" && gotStderr != "" {
				// Only fail on unexpected stderr if the command was expected to succeed
				if tt.wantSuccess {
					t.Errorf("stderr field: got %q, want empty", gotStderr)
				} else {
					t.Logf("stderr field (on intended failure): got %q", gotStderr)
				}
			}
		})
	}
}

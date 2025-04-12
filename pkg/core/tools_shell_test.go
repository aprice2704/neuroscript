// pkg/core/tools_shell_test.go
package core

import (
	"errors" // Import errors
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined elsewhere (e.g., testing_helpers.go)

// --- Test ToolExecuteCommand ---
func TestToolExecuteCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell command tests on Windows")
	}
	dummyInterp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name         string
		command      string
		cmdArgs      []interface{} // Args for the *command* itself
		wantStdout   string
		wantStderr   string
		wantExitCode int64
		wantSuccess  bool
		valWantErrIs error // Expect specific validation error?
		// Removed errContains field
	}{
		// Success cases
		{name: "Simple Echo Success", command: "echo", cmdArgs: []interface{}{"hello", "world"}, wantStdout: "hello world\n", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command True Success", command: "true", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 0, wantSuccess: true},
		{name: "Command False Failure", command: "false", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 1, wantSuccess: false},
		{name: "Command Writes to Stderr", command: "sh", cmdArgs: []interface{}{"-c", "echo 'error output' >&2"}, wantStdout: "", wantStderr: "error output\n", wantExitCode: 0, wantSuccess: true},
		{name: "Command Not Found", command: "nonexistent_command_ajsdflk", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "executable file not found", wantExitCode: -1, wantSuccess: false},

		// Validation Error Tests
		{name: "Validation Wrong Arg Count (1)", command: "echo", cmdArgs: nil, valWantErrIs: ErrValidationArgCount},                                             // cmdArgs is nil, but makeArgs below passes only command
		{name: "Validation Wrong Arg Type (Command not string)", command: "", cmdArgs: makeArgs(123, []interface{}{}), valWantErrIs: ErrValidationTypeMismatch},  // Setup provides wrong type
		{name: "Validation Wrong Arg Type (Args not slice)", command: "echo", cmdArgs: makeArgs("echo", "not_a_slice"), valWantErrIs: ErrValidationTypeMismatch}, // Setup provides wrong type
	}
	spec := ToolSpec{Name: "ExecuteCommand", Args: []ArgSpec{{Name: "command", Type: ArgTypeString, Required: true}, {Name: "args_list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count (1)" {
				rawArgs = makeArgs(tt.command) // Only pass command
			} else if tt.name == "Validation Wrong Arg Type (Command not string)" {
				rawArgs = tt.cmdArgs // Use predefined setup with wrong type
			} else if tt.name == "Validation Wrong Arg Type (Args not slice)" {
				rawArgs = tt.cmdArgs // Use predefined setup with wrong type
			} else {
				// Correct structure for execution tests: command string, then slice of command args
				rawArgs = makeArgs(tt.command, tt.cmdArgs) // Pass the command args as the second element
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)

			// *** UPDATED: Use errors.Is for validation check ***
			if tt.valWantErrIs != nil {
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil (Raw Args: %#v)", tt.valWantErrIs, rawArgs)
				} else if !errors.Is(valErr, tt.valWantErrIs) {
					t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tt.valWantErrIs, valErr, valErr)
				}
				return // Stop if validation failed as expected
			}
			// *** END UPDATED VALIDATION CHECK ***

			// Check for unexpected validation error
			if valErr != nil && tt.valWantErrIs == nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// Tool execution checks (only run if validation passed and was expected to pass)
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
			if gotStdout != tt.wantStdout {
				t.Errorf("stdout field:\ngot:  %q\nwant: %q", gotStdout, tt.wantStdout)
			}
			gotStderr, _ := gotMap["stderr"].(string)
			if tt.wantStderr != "" {
				if !strings.Contains(gotStderr, tt.wantStderr) {
					t.Errorf("stderr field:\ngot:  %q\nwant contains: %q", gotStderr, tt.wantStderr)
				}
			} else if tt.wantStderr == "" && gotStderr != "" {
				// Allow stderr on failure even if wantStderr is empty, unless it's specifically checking for command not found
				if tt.wantSuccess || tt.command == "nonexistent_command_ajsdflk" {
					t.Errorf("stderr field: got %q, want empty", gotStderr)
				} else {
					t.Logf("stderr field (on intended failure): got %q", gotStderr)
				}
			}
		})
	}
}

// --- Test Go Mod Tidy ---
func TestToolGoModTidy(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Skipping GoModTidy test: 'go' command not found")
	}

	dummyInterp, _ := newDefaultTestInterpreter(t)
	testDir := "gomodtidy_test_run_cwd_dir_5" // Use different dir name
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	goModPath := filepath.Join(testDir, "go.mod")
	goModContent := []byte("module neuroscript_test_gomod_5\n\ngo 1.20\nrequire rsc.io/quote v1.5.2\n")
	err = os.WriteFile(goModPath, goModContent, 0644)
	if err != nil {
		t.Fatalf("Write go.mod failed: %v", err)
	}
	goFilePath := filepath.Join(testDir, "dummy.go")
	goFileContent := []byte("package main\nimport (\n\t\"fmt\"\n\t\"rsc.io/quote\"\n)\nfunc main(){ fmt.Println(quote.Hello())}\n")
	err = os.WriteFile(goFilePath, goFileContent, 0644)
	if err != nil {
		t.Fatalf("Write .go failed: %v", err)
	}

	originalWd, _ := os.Getwd()
	err = os.Chdir(testDir)
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer os.Chdir(originalWd)

	rawArgs := makeArgs() // GoModTidy takes no arguments
	spec := ToolSpec{Name: "GoModTidy", Args: []ArgSpec{}, ReturnType: ArgTypeAny}
	convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs) // Use convertedArgs
	if valErr != nil {
		t.Fatalf("Validate unexpected error: %v", valErr)
	}

	gotResult, toolErr := toolGoModTidy(dummyInterp, convertedArgs) // Use convertedArgs
	if toolErr != nil {
		t.Fatalf("toolGoModTidy unexpected Go error: %v", toolErr)
	}

	gotMap, ok := gotResult.(map[string]interface{})
	if !ok {
		t.Fatalf("toolGoModTidy did not return map, got %T", gotResult)
	}
	if success, ok := gotMap["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got success=%v, stderr=%q", gotMap["success"], gotMap["stderr"])
	}
	if exitCode, ok := gotMap["exit_code"].(int64); !ok || exitCode != 0 {
		t.Errorf("Expected exit_code=0, got %v", gotMap["exit_code"])
	}
	t.Logf("GoModTidy stdout: %q", gotMap["stdout"])
	t.Logf("GoModTidy stderr: %q", gotMap["stderr"])
}

// pkg/core/tools_shell_test.go
package core

import (
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
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name         string
		command      string
		cmdArgs      []interface{} // Raw args before validation
		wantStdout   string
		wantStderr   string
		wantExitCode int64
		wantSuccess  bool
		valWantErr   bool   // Expect validation error?
		errContains  string // Validation error substring
	}{
		{name: "Simple Echo Success", command: "echo", cmdArgs: []interface{}{"hello", "world"}, wantStdout: "hello world\n", wantStderr: "", wantExitCode: 0, wantSuccess: true, valWantErr: false, errContains: ""},
		{name: "Command True Success", command: "true", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 0, wantSuccess: true, valWantErr: false, errContains: ""},
		{name: "Command False Failure", command: "false", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 1, wantSuccess: false, valWantErr: false, errContains: ""},
		{name: "Command Writes to Stderr", command: "sh", cmdArgs: []interface{}{"-c", "echo 'error output' >&2"}, wantStdout: "", wantStderr: "error output\n", wantExitCode: 0, wantSuccess: true, valWantErr: false, errContains: ""},
		{name: "Command Not Found", command: "nonexistent_command_ajsdflk", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "executable file not found", wantExitCode: -1, wantSuccess: false, valWantErr: false, errContains: ""},
		// Validation Error Tests
		{name: "Validation Wrong Arg Count (1)", command: "echo", cmdArgs: nil, valWantErr: true, errContains: "tool 'ExecuteCommand' expected exactly 2 arguments, but received 1"}, // Use cmdArgs nil to pass makeArgs(cmd)
		{name: "Validation Wrong Arg Type (Command not string)", command: "", cmdArgs: makeArgs(123, []interface{}{}), valWantErr: true, errContains: "tool 'ExecuteCommand' argument 'command' (index 0): expected string, but received type int"},
		// ** FIX: Update expected error substring for slice_any mismatch **
		{name: "Validation Wrong Arg Type (Args not slice)", command: "echo", cmdArgs: makeArgs("echo", "not_a_slice"), valWantErr: true, errContains: "tool 'ExecuteCommand' argument 'args_list' (index 1): expected slice_any (e.g., list literal), but received incompatible type string"},
	}
	spec := ToolSpec{Name: "ExecuteCommand", Args: []ArgSpec{{Name: "command", Type: ArgTypeString, Required: true}, {Name: "args_list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			// Special handling for arg count test where cmdArgs is nil
			if tt.name == "Validation Wrong Arg Count (1)" {
				rawArgs = makeArgs(tt.command) // Only pass command
			} else {
				rawArgs = tt.cmdArgs // Use the predefined args (which might be wrong type for validation tests)
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, wantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.errContains, valErr)
				}
				return // Stop if validation failed as expected
			}
			// Proceed only if validation passed
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
			// Allow -1 or any non-zero for failure check when wantExitCode is -1
			if tt.wantExitCode == -1 {
				if gotSuccess {
					t.Errorf("Expected failure (success=false), but got success=true")
				}
				if gotExitCode == 0 {
					t.Errorf("Expected non-zero exit code on failure, got 0")
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
				// Allow stderr content if command succeeded, might be warnings
				if !gotSuccess {
					t.Errorf("stderr field: got %q, want empty on failure (unless specific msg expected)", gotStderr)
				} else {
					t.Logf("stderr field: got %q, want empty (allowing warnings on success)", gotStderr)
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

	dummyInterp := newDefaultTestInterpreter()
	testDir := "gomodtidy_test_run_cwd_dir_2" // Use different dir name
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	goModPath := filepath.Join(testDir, "go.mod")
	// Intentionally use a slightly different module path
	goModContent := []byte("module neuroscript_test_gomod_2\n\ngo 1.20\nrequire rsc.io/quote v1.5.2\n")
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

	rawArgs := makeArgs()
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

// Helper funcs assumed to be in testing_helpers.go
// func newDefaultTestInterpreter() *Interpreter
// func makeArgs(vals ...interface{}) []interface{}

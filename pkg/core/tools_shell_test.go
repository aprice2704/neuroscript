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

// Assume newDummyInterpreter and makeArgs are defined elsewhere (e.g., interpreter_test.go)

// --- Test ToolExecuteCommand ---
func TestToolExecuteCommand(t *testing.T) { /* ... no changes ... */
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell command tests on Windows")
	}
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name         string
		command      string
		cmdArgs      []interface{}
		wantStdout   string
		wantStderr   string
		wantExitCode int64
		wantSuccess  bool
		wantErr      bool
		errContains  string
	}{
		{name: "Simple Echo Success", command: "echo", cmdArgs: []interface{}{"hello", "world"}, wantStdout: "hello world\n", wantStderr: "", wantExitCode: 0, wantSuccess: true, wantErr: false, errContains: ""},
		{name: "Command True Success", command: "true", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 0, wantSuccess: true, wantErr: false, errContains: ""},
		{name: "Command False Failure", command: "false", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "", wantExitCode: 1, wantSuccess: false, wantErr: false, errContains: ""},
		{name: "Command Writes to Stderr", command: "sh", cmdArgs: []interface{}{"-c", "echo 'error output' >&2"}, wantStdout: "", wantStderr: "error output\n", wantExitCode: 0, wantSuccess: true, wantErr: false, errContains: ""},
		{name: "Command Not Found", command: "nonexistent_command_ajsdflk", cmdArgs: []interface{}{}, wantStdout: "", wantStderr: "executable file not found", wantExitCode: -1, wantSuccess: false, wantErr: false, errContains: ""},
		{name: "Wrong Arg Count (1)", command: "echo", cmdArgs: nil, wantErr: true, errContains: "tool 'ExecuteCommand' expected exactly 2 arguments, but received 1", wantSuccess: false},
		{name: "Wrong Arg Type (Command not string)", command: "", cmdArgs: makeArgs(123, []interface{}{}), wantErr: true, errContains: "tool 'ExecuteCommand' argument 'command' (index 0): expected string, but received type int", wantSuccess: false},
		{name: "Wrong Arg Type (Args not slice)", command: "echo", cmdArgs: makeArgs("echo", "not_a_slice"), wantErr: true, errContains: "tool 'ExecuteCommand' argument 'args_list' (index 1): expected slice_any, but received incompatible type string", wantSuccess: false},
	}
	spec := ToolSpec{Name: "ExecuteCommand", Args: []ArgSpec{{Name: "command", Type: ArgTypeString, Required: true}, {Name: "args_list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.wantErr {
				if tt.name == "Wrong Arg Count (1)" {
					rawArgs = makeArgs(tt.command)
				} else {
					rawArgs = tt.cmdArgs
				}
			} else {
				rawArgs = makeArgs(tt.command, tt.cmdArgs)
			}
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.wantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, wantErr %v", valErr, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.errContains, valErr)
				}
				return
			}
			gotInterface, toolErr := toolExecuteCommand(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolExecuteCommand() returned unexpected error: %v", toolErr)
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
			if gotExitCode != tt.wantExitCode {
				if !(tt.wantExitCode == -1 && gotExitCode != 0 && !gotSuccess) {
					t.Errorf("exit_code field: got %d, want %d", gotExitCode, tt.wantExitCode)
				}
			}
			gotStdout, _ := gotMap["stdout"].(string)
			if gotStdout != tt.wantStdout {
				t.Errorf("stdout field: got %q, want %q", gotStdout, tt.wantStdout)
			}
			gotStderr, _ := gotMap["stderr"].(string)
			if tt.wantStderr != "" {
				if !strings.Contains(gotStderr, tt.wantStderr) {
					t.Errorf("stderr field: got %q, want contains %q", gotStderr, tt.wantStderr)
				}
			} else if gotStderr != "" && tt.wantStderr == "" {
				if tt.name != "Command Not Found" {
					t.Logf("stderr field: got %q, want empty (but allowing content)", gotStderr)
				} else if !strings.Contains(gotStderr, "Execution Error:") {
					t.Errorf("stderr field for Command Not Found: got %q, want contains 'Execution Error:'", gotStderr)
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

	dummyInterp := newDummyInterpreter()
	testDir := "gomodtidy_test_run_cwd_dir"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	goModPath := filepath.Join(testDir, "go.mod")
	goModContent := []byte("module neuroscript_test_gomod\n\ngo 1.20\nrequire rsc.io/quote v1.5.2\n")
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
	_, valErr := ValidateAndConvertArgs(spec, rawArgs)
	if valErr != nil {
		t.Fatalf("Validate unexpected error: %v", valErr)
	}

	gotResult, toolErr := toolGoModTidy(dummyInterp, rawArgs)
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

// --- Helper Functions ---
// Removed - Assume defined elsewhere
// func newDummyInterpreter() *Interpreter          { return NewInterpreter(nil) }
// func makeArgs(vals ...interface{}) []interface{} { return vals }

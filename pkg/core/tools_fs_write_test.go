// filename: pkg/core/tools_fs_write_test.go
package core

import (
	// Keep errors
	"fmt" // Keep fmt
	"os"
	"path/filepath" // Keep filepath

	// Keep strings for error message check
	"testing"
)

// Remove the duplicate testFsToolHelper function definition here
/*
func testFsToolHelper(...) {
    ... // REMOVED
}
*/

func TestToolWriteFile(t *testing.T) {
	interp, sandboxDir := newDefaultTestInterpreter(t) // Get interpreter and sandbox

	// Use the unified fsTestCase struct from testing_helpers_test.go
	tests := []fsTestCase{
		{
			name:        "Write New File",
			toolName:    "WriteFile",
			args:        makeArgs("newWrite.txt", "content here"),
			wantResult:  "OK",
			wantContent: "content here",
			// cleanupFunc: func() error { return os.Remove("newWrite.txt") }, // Handled by TempDir
		},
		{
			name:        "Overwrite Existing File",
			toolName:    "WriteFile",
			args:        makeArgs("overwrite.txt", "new data"),
			wantResult:  "OK",
			wantContent: "new data",
			setupFunc: func() error {
				return os.WriteFile("overwrite.txt", []byte("old data"), 0644) // Relative path
			},
			// cleanupFunc: func() error { return os.Remove("overwrite.txt") },
		},
		{
			name:        "Write Empty Content",
			toolName:    "WriteFile",
			args:        makeArgs("emptyWrite.txt", ""),
			wantResult:  "OK",
			wantContent: "",
			// cleanupFunc: func() error { return os.Remove("emptyWrite.txt") },
		},
		{
			name:        "Create Subdirectory",
			toolName:    "WriteFile",
			args:        makeArgs("newdir/nestedfile.txt", "nested content"),
			wantResult:  "OK",
			wantContent: "nested content",
			// cleanupFunc: func() error { return os.RemoveAll("newdir") },
		},
		{
			name:         "Validation_Wrong_Path_Type",
			toolName:     "WriteFile",
			args:         makeArgs(123, "content"),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Wrong_Content_Type",
			toolName:     "WriteFile",
			args:         makeArgs("path.txt", true),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Missing_Content",
			toolName:     "WriteFile",
			args:         makeArgs("path.txt"),
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:     "Path_Outside_Sandbox",
			toolName: "WriteFile",
			args:     makeArgs("../outsideWrite.txt", "data"),
			// Construct expected error message based on ErrPathViolation
			wantResult:    fmt.Sprintf("WriteFile path error for '../outsideWrite.txt': %s: relative path '../outsideWrite.txt' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(sandboxDir, "../outsideWrite.txt")), sandboxDir),
			wantToolErrIs: ErrPathViolation, // Expect the specific Go error
		},
		{
			name:     "Write_To_Directory_Path",
			toolName: "WriteFile",
			args:     makeArgs("subdir_write", "content"),
			// Adjust expected error message based on OS behavior
			wantResult:    fmt.Sprintf("WriteFile failed for 'subdir_write': open %s: is a directory", filepath.Join(sandboxDir, "subdir_write")),
			wantToolErrIs: ErrInternalTool,                                        // Expect wrapped OS error
			setupFunc:     func() error { return os.Mkdir("subdir_write", 0755) }, // Relative path
			// cleanupFunc: func() error { return os.Remove("subdir_write") },
		},
		{
			name:         "Validation_Nil_Path",
			toolName:     "WriteFile",
			args:         makeArgs(nil, "content"),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
		{
			name:         "Validation_Nil_Content",
			toolName:     "WriteFile",
			args:         makeArgs("path.txt", nil),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		// Pass interp and tt to the helper in testing_helpers_test.go
		testFsToolHelper(t, interp, "../temp", tt)
	}
}

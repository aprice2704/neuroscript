// filename: pkg/core/tools_fs_list_test.go
package core

import (
	// ADDED errors import
	"os"
	"path/filepath"
	"testing"
)

// Assuming testFsToolHelper is defined in testing_helpers_test.go

func TestToolListDirectory(t *testing.T) {
	// *** MODIFIED: Ignore the unused sandboxDir return value ***
	interp, _ := newDefaultTestInterpreter(t) // Use blank identifier '_'

	// Setup test directory structure within the sandbox
	baseDir := "listTest"
	subDir := filepath.Join(baseDir, "sub")
	file1 := filepath.Join(baseDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	os.MkdirAll(subDir, 0755)                              // Create dirs
	os.WriteFile(file1, []byte("hello"), 0644)             // Create file in base
	os.WriteFile(file2, []byte("world"), 0644)             // Create file in sub
	os.WriteFile("file_at_root.txt", []byte("root"), 0644) // Create file at sandbox root

	// Get size info for expectations
	file1Info, _ := os.Stat(file1)
	file2Info, _ := os.Stat(file2)
	subDirInfo, _ := os.Stat(subDir)
	rootFileInfo, _ := os.Stat("file_at_root.txt")

	// Define fsTestCase struct locally or ensure it's accessible
	// type fsTestCase struct { ... } // Assuming it's defined elsewhere

	tests := []fsTestCase{
		{
			name:     "List root of test dir",
			toolName: "ListDirectory",
			args:     makeArgs(baseDir),
			// *** MODIFIED: Added "size" to expectations ***
			wantResult: []interface{}{
				map[string]interface{}{"name": "file1.txt", "is_dir": false, "size": file1Info.Size()},
				map[string]interface{}{"name": "sub", "is_dir": true, "size": subDirInfo.Size()},
			},
		},
		{
			name:     "List sub directory",
			toolName: "ListDirectory",
			args:     makeArgs(subDir),
			// *** MODIFIED: Added "size" to expectations ***
			wantResult: []interface{}{
				map[string]interface{}{"name": "file2.txt", "is_dir": false, "size": file2Info.Size()},
			},
		},
		{
			name:     "List sandbox root", // Add test for sandbox root
			toolName: "ListDirectory",
			args:     makeArgs("."), // Use "." for current (sandbox root)
			// *** MODIFIED: Added "size" to expectations ***
			wantResult: []interface{}{
				// Note: Order matters for DeepEqual if not sorted by helper
				// Adding baseDir first as it usually comes first alphabetically
				map[string]interface{}{"name": baseDir, "is_dir": true, "size": int64(4096)}, // Assuming baseDir size is standard block size
				map[string]interface{}{"name": "file_at_root.txt", "is_dir": false, "size": rootFileInfo.Size()},
			},
		},
		{
			name:     "List non-existent dir",
			toolName: "ListDirectory",
			args:     makeArgs(filepath.Join(baseDir, "nonexistent")),
			// *** MODIFIED: Expect Go error, not string result ***
			wantResult:    nil,             // No specific result expected on error
			wantToolErrIs: ErrInternalTool, // Expect wrapped OS error
			valWantErrIs:  nil,             // Validation passes
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "ListDirectory",
			args:         makeArgs(123),
			wantResult:   nil,                       // No result expected on validation error
			valWantErrIs: ErrValidationTypeMismatch, // Expect validation error
		},
		{
			name:     "Path_Outside_Sandbox",
			toolName: "ListDirectory",
			args:     makeArgs("../outside"),
			// *** MODIFIED: Expect Go error (ErrPathViolation) ***
			wantResult:    nil,
			wantToolErrIs: ErrPathViolation, // Expect path violation error from tool
			valWantErrIs:  nil,              // Validation passes
		},
		{
			name:     "Path_Is_File",
			toolName: "ListDirectory",
			args:     makeArgs(file1), // Try to list a file
			// *** MODIFIED: Expect Go error, not string result ***
			wantResult:    nil,             // No specific result expected on error
			wantToolErrIs: ErrInternalTool, // Expect wrapped OS error (e.g., "not a directory")
			valWantErrIs:  nil,             // Validation passes
		},
	}

	for _, tt := range tests {
		// Ensure testFsToolHelper is called correctly
		testFsToolHelper(t, interp, "../temp", tt)
	}
}

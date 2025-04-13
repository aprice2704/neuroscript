// filename: pkg/core/tools_fs_dirs_test.go
package core

import (
	// Import errors for checking
	// Import fmt for error message construction (can potentially be removed if not used elsewhere)
	"os"            // Import os for file operations in setup/verification
	"path/filepath" // Import filepath for joining paths
	"testing"
)

// Assume newTestInterpreter, makeArgs, and testFsToolHelper are defined in testing_helpers_test.go

func TestToolMkdir(t *testing.T) {
	interp, sandboxDir := newDefaultTestInterpreter(t) // Get interpreter and sandbox

	// Define test cases using the fsTestCase struct from helpers
	tests := []fsTestCase{
		// Success Cases
		{
			name:       "Create Single Dir",
			toolName:   "Mkdir",
			args:       makeArgs("newdir"),
			wantResult: "OK", // Expect "OK" string on success
		},
		{
			name:       "Create Nested Dirs",
			toolName:   "Mkdir",
			args:       makeArgs("a/b/c"),
			wantResult: "OK",
		},
		{
			name:     "Create Existing Dir",
			toolName: "Mkdir",
			args:     makeArgs("existing_dir"),
			setupFunc: func() error {
				return os.Mkdir(filepath.Join(sandboxDir, "existing_dir"), 0755)
			},
			wantResult: "OK", // Should succeed without error
		},

		// Validation Error Cases (wantResult is implicitly nil here as helper returns on valErr)
		{
			name:         "Validation Missing Arg",
			toolName:     "Mkdir",
			args:         makeArgs(),
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:         "Validation Wrong Arg Type",
			toolName:     "Mkdir",
			args:         makeArgs(123),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation Nil Arg",
			toolName:     "Mkdir",
			args:         makeArgs(nil),
			valWantErrIs: ErrValidationRequiredArgNil,
		},

		// Execution Error Cases (wantResult is nil because helper ignores it when wantToolErrIs is set)
		{
			name:          "Path Outside Sandbox",
			toolName:      "Mkdir",
			args:          makeArgs("../outside_dir"),
			wantResult:    nil, // Ignored by helper when wantToolErrIs is set
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:     "Path Is Existing File",
			toolName: "Mkdir",
			args:     makeArgs("existing_file.txt"),
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(sandboxDir, "existing_file.txt"), []byte("hello"), 0644)
			},
			wantResult:    nil, // Ignored by helper when wantToolErrIs is set
			wantToolErrIs: ErrCannotCreateDir,
		},
		{
			name:          "Empty Path String",
			toolName:      "Mkdir",
			args:          makeArgs(""),
			wantResult:    nil, // Ignored by helper when wantToolErrIs is set
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:          "Path Contains Null Byte",
			toolName:      "Mkdir",
			args:          makeArgs("dir\x00name"),
			wantResult:    nil, // Ignored by helper when wantToolErrIs is set
			wantToolErrIs: ErrPathViolation,
		},
	}

	for _, tt := range tests {
		// Pass interp and tt to the helper defined in testing_helpers_test.go
		testFsToolHelper(t, interp, "../temp", tt)

		// Verification for successful mkdir cases (remains the same)
		if tt.valWantErrIs == nil && tt.wantToolErrIs == nil && tt.wantResult == "OK" {
			t.Run(tt.name+"_VerifyExists", func(t *testing.T) {
				pathArg := tt.args[0].(string)
				verifyPath := filepath.Join(sandboxDir, pathArg)
				info, err := os.Stat(verifyPath)
				if err != nil {
					if tt.name != "Create Existing Dir" {
						t.Errorf("os.Stat failed for expected directory '%s': %v", verifyPath, err)
					}
				} else if !info.IsDir() {
					t.Errorf("Path '%s' exists but is not a directory", verifyPath)
				}
			})
		}
	}
}

// Add TestToolDeleteFile here later...

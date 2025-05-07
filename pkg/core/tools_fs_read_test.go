// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Correct test cases to use existing fsTestCase fields (wantToolErrIs, wantResult)
// and align with updated error handling in tools_fs_read.go.
// filename: pkg/core/tools_fs_read_test.go
package core

import (
	"fmt"
	"os"
	"path/filepath" // For checking error message substrings
	"testing"
)

// Assume testFsToolHelper is defined in testing_helpers_test.go or universal_test_helpers.go

func TestToolReadFile(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Get interpreter and sandbox path

	// --- Test Setup Data ---
	testFilePathRel := "readTest.txt"
	testDirPathRel := "readTestDir"         // This will be created as a directory
	nonExistentPathRel := "nonexistent.txt" // This file will not be created
	testContent := "Hello\nWorld!"

	// --- Setup Function ---
	setupReadFileTest := func(sandboxRoot string) error {
		// Construct absolute paths *within* the sandbox for setup
		testFileAbs := filepath.Join(sandboxRoot, testFilePathRel)
		testDirAbs := filepath.Join(sandboxRoot, testDirPathRel)

		// Create file using absolute path
		if err := os.WriteFile(testFileAbs, []byte(testContent), 0644); err != nil {
			if !os.IsExist(err) { // Allow already exists for idempotent tests
				return fmt.Errorf("setup WriteFile failed for %s: %w", testFileAbs, err)
			}
		}
		// Create directory using absolute path
		if err := os.Mkdir(testDirAbs, 0755); err != nil {
			if !os.IsExist(err) { // Allow already exists
				return fmt.Errorf("setup Mkdir failed for %s: %w", testDirAbs, err)
			}
		}
		// Ensure nonExistentPathRel does NOT exist
		_ = os.Remove(filepath.Join(sandboxRoot, nonExistentPathRel))

		return nil
	}

	tests := []fsTestCase{
		{
			name:       "Read Existing File",
			toolName:   "ReadFile",
			args:       MakeArgs(testFilePathRel),
			setupFunc:  setupReadFileTest,
			wantResult: testContent,
			// No error expected
		},
		{
			name:          "Read Non-Existent File",
			toolName:      "ReadFile",
			args:          MakeArgs(nonExistentPathRel),
			setupFunc:     setupReadFileTest,
			wantResult:    fmt.Sprintf("file not found at path '%s'", nonExistentPathRel), // Check for this substring in the error
			wantToolErrIs: ErrFileNotFound,                                                // Check for the wrapped sentinel error
		},
		{
			name:          "Read Directory",
			toolName:      "ReadFile",
			args:          MakeArgs(testDirPathRel),
			setupFunc:     setupReadFileTest,
			wantResult:    fmt.Sprintf("path '%s' is a directory, not a file", testDirPathRel), // Check for this substring
			wantToolErrIs: ErrInvalidArgument,                                                  // Check for the wrapped sentinel error
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "ReadFile",
			args:         MakeArgs(123),
			valWantErrIs: ErrValidationTypeMismatch, // This is a validation error
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "ReadFile",
			args:         MakeArgs(),
			valWantErrIs: ErrValidationArgCount, // Validation error
		},
		{
			name:      "Path_Outside_Sandbox",
			toolName:  "ReadFile",
			args:      MakeArgs("../outside.txt"),
			setupFunc: setupReadFileTest,
			// The ResolvePath function is expected to create the specific error message.
			// We check for the sentinel and a key part of the message.
			wantResult:    "path resolves outside allowed directory", // Check for this substring
			wantToolErrIs: ErrPathViolation,                          // ResolvePath wraps ErrPathViolation
		},
		{
			name:         "Validation_Nil_Arg",
			toolName:     "ReadFile",
			args:         MakeArgs(nil),
			valWantErrIs: ErrValidationRequiredArgNil, // Validation error
		},
		{
			name:          "Validation_Empty_Filepath_Arg",
			toolName:      "ReadFile",
			args:          MakeArgs(""),
			wantResult:    "ReadFile filepath cannot be empty", // Check for this substring
			wantToolErrIs: ErrInvalidArgument,                  // toolReadFile now returns this for empty filepath
		},
	}

	for _, tt := range tests {
		// The testFsToolHelper should:
		// 1. If tt.valWantErrIs is set, check for that validation error.
		// 2. Otherwise, run the tool.
		// 3. If tt.wantToolErrIs is set:
		//    a. Check if an error was returned.
		//    b. Check if errors.Is(actualError, tt.wantToolErrIs).
		//    c. If tt.wantResult is a string, check if strings.Contains(actualError.Error(), tt.wantResult.(string)).
		// 4. If tt.wantToolErrIs is nil, check if no error was returned and actualResult matches tt.wantResult.
		testFsToolHelper(t, interp, tt)
	}
}

// Mock implementation of testFsToolHelper for clarity on how it might work
// This is NOT part of the file to be changed, just for understanding.
// The actual helper is in universal_test_helpers.go or testing_helpers.go
/*
type fsTestCase struct {
	name          string
	toolName      string
	args          []interface{}
	setupFunc     func(sandboxRoot string) error
	wantResult    interface{}
	wantToolErrIs error // For checking errors.Is(err, wantToolErrIs)
	valWantErrIs  error // For checking validation errors before tool execution
}

func testFsToolHelper(t *testing.T, interp *Interpreter, tc fsTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		t.Helper()
		if tc.setupFunc != nil {
			if err := tc.setupFunc(interp.SandboxDirAbs()); err != nil { // Assuming SandboxDirAbs() gives the absolute path
				t.Fatalf("Setup function failed: %v", err)
			}
		}

		// Simplified: Tool execution and validation logic would be here.
		// This part simulates how the helper might use the fsTestCase fields.

		tool, toolExists := interp.Tools[tc.toolName]
		if !toolExists {
			t.Fatalf("Tool %s not found", tc.toolName)
		}

		// Simulate validation phase (this is often part of a generic tool runner)
		if tc.valWantErrIs != nil {
			// In a real scenario, validation would be called here.
			// For this example, assume a placeholder validation error if args are nil and valWantErrIs is set.
			var validationErr error
			if tc.args == nil && tc.valWantErrIs != nil { // Highly simplified validation check
				validationErr = tc.valWantErrIs
			}

			if validationErr == nil {
				if tc.valWantErrIs != nil {
					t.Errorf("Expected validation error %q, but got nil", tc.valWantErrIs)
				}
			} else if !errors.Is(validationErr, tc.valWantErrIs) {
				t.Errorf("Expected validation error %q, got %q", tc.valWantErrIs, validationErr)
			}
			return // Validation tests usually stop here.
		}


		// Actual tool execution
		gotResult, gotErr := tool.Func(interp, tc.args)

		// Check error expectations
		if tc.wantToolErrIs != nil {
			if gotErr == nil {
				t.Errorf("Expected error type %T, but got nil", tc.wantToolErrIs)
				return
			}
			if !errors.Is(gotErr, tc.wantToolErrIs) {
				t.Errorf("Expected error to wrap %T, but it did not. Got error: %v", tc.wantToolErrIs, gotErr)
			}
			// If wantResult is a string, check if the error message contains it
			if expectedMsgPart, ok := tc.wantResult.(string); ok && expectedMsgPart != "" {
				if !strings.Contains(gotErr.Error(), expectedMsgPart) {
					t.Errorf("Error message %q does not contain expected substring %q", gotErr.Error(), expectedMsgPart)
				}
			}
		} else { // No error expected
			if gotErr != nil {
				t.Errorf("Expected no error, but got: %v", gotErr)
				return
			}
			// Check result if no error was expected
			if !reflect.DeepEqual(gotResult, tc.wantResult) {
				t.Errorf("Result mismatch:\nGot:  %#v (%T)\nWant: %#v (%T)", gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}
*/

// filename: core/tools_git_test.go
package core

import (
	"errors"
	// NOTE: strings package no longer needed here
	"testing"
)

// Helper to run validation checks (similar to other test files)
// Uses only errors.Is for checking expected errors.
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t) // Basic interpreter for validation context
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found {
		t.Fatalf("Tool %s not found in registry", toolName)
	}
	spec := toolImpl.Spec

	for _, tc := range testCases {
		// Use tc.Name from the struct field for the subtest name
		t.Run(tc.Name, func(t *testing.T) {
			// Use tc.InputArgs from the struct field
			_, err := ValidateAndConvertArgs(spec, tc.InputArgs)

			// Use tc.ExpectedError from the struct field
			if tc.ExpectedError != nil {
				if err == nil {
					// Use tc.ExpectedError in the error message
					t.Errorf("Expected error [%v], got nil", tc.ExpectedError)
				} else if !errors.Is(err, tc.ExpectedError) {
					// Use tc.ExpectedError in the error message
					// Check only errors.Is - if this fails, the wrapping or expected error is wrong.
					t.Errorf("Expected error wrapping [%v], but errors.Is is false. Got error: [%T] %v", tc.ExpectedError, err, err)
				} else {
					t.Logf("Got expected error type via errors.Is: %v", err) // Log success for clarity
				}
			} else if err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			// We don't call the actual tool function here as we're not mocking toolExec yet
		})
	}
}

// --- GitNewBranch Validation Tests ---
func TestToolGitNewBranchValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: MakeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: MakeArgs("new-feature"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitNewBranch", testCases)
	// Note: Actual branch name validation happens inside the tool function currently
}

// --- GitCheckout Validation Tests ---
func TestToolGitCheckoutValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: MakeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: MakeArgs("main"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitCheckout", testCases)
}

// --- GitRm Validation Tests ---
func TestToolGitRmValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: MakeArgs("file1", "file2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: MakeArgs("path/to/file.txt"), ExpectedError: nil}, // Validation should pass
		// Note: SecureFilePath validation happens inside the tool function
	}
	runValidationTestCases(t, "GitRm", testCases)
}

// --- GitMerge Validation Tests ---
func TestToolGitMergeValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: MakeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: MakeArgs("develop"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitMerge", testCases)
}

// --- GitPull Validation Tests ---
func TestToolGitPullValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: MakeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitPull", testCases)
}

// --- GitPush Validation Tests (NEW) ---
func TestToolGitPushValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: MakeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitPush", testCases)
}

// --- GitDiff Validation Tests (NEW) ---
func TestToolGitDiffValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: MakeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitDiff", testCases)
}

// TODO: Add functional tests for Git tools, likely requiring mocking of toolExec or a test repo setup.

// Ensure required error variables are defined
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
)

// Ensure MakeArgs is available (implicitly via testing_helpers_test.go)
var _ = MakeArgs

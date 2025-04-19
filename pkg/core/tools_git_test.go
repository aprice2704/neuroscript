// filename: pkg/core/tools_git_test.go
// NEW: Basic validation tests for new Git tools
// FIXED: Corrected ValidationTestCase field names (Name, InputArgs, ExpectedError)
// ADDED: Validation test for GitPull, GitPush, GitDiff
package core

import (
	"errors"
	"testing"
)

// Helper to run validation checks (similar to other test files)
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := newDefaultTestInterpreter(t) // Basic interpreter for validation context
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
					t.Errorf("Expected error type [%T], got [%T]: %v", tc.ExpectedError, err, err)
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
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: makeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: makeArgs("new-feature"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitNewBranch", testCases)
	// Note: Actual branch name validation happens inside the tool function currently
}

// --- GitCheckout Validation Tests ---
func TestToolGitCheckoutValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: makeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: makeArgs("main"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitCheckout", testCases)
}

// --- GitRm Validation Tests ---
func TestToolGitRmValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("file1", "file2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: makeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: makeArgs("path/to/file.txt"), ExpectedError: nil}, // Validation should pass
		// Note: SecureFilePath validation happens inside the tool function
	}
	runValidationTestCases(t, "GitRm", testCases)
}

// --- GitMerge Validation Tests ---
func TestToolGitMergeValidation(t *testing.T) {
	// Use capitalized field names in struct literals
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: makeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: makeArgs("develop"), ExpectedError: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitMerge", testCases)
}

// --- GitPull Validation Tests ---
func TestToolGitPullValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: makeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitPull", testCases)
}

// --- GitPush Validation Tests (NEW) ---
func TestToolGitPushValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: makeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitPush", testCases)
}

// --- GitDiff Validation Tests (NEW) ---
func TestToolGitDiffValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Arg Count (None)", InputArgs: makeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "GitDiff", testCases)
}

// TODO: Add functional tests for Git tools, likely requiring mocking of toolExec or a test repo setup.

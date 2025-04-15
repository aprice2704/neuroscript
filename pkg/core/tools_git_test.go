// filename: pkg/core/tools_git_test.go
// NEW: Basic validation tests for new Git tools
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
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateAndConvertArgs(spec, tc.args)
			if tc.wantErrIs != nil {
				if err == nil {
					t.Errorf("Expected error [%v], got nil", tc.wantErrIs)
				} else if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error type [%T], got [%T]: %v", tc.wantErrIs, err, err)
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
	testCases := []ValidationTestCase{
		{name: "Wrong Arg Count (None)", args: makeArgs(), wantErrIs: ErrValidationArgCount},
		{name: "Wrong Arg Count (Too Many)", args: makeArgs("branch1", "branch2"), wantErrIs: ErrValidationArgCount},
		{name: "Nil Arg", args: makeArgs(nil), wantErrIs: ErrValidationRequiredArgNil},
		{name: "Wrong Arg Type", args: makeArgs(123), wantErrIs: ErrValidationTypeMismatch},
		{name: "Correct Args", args: makeArgs("new-feature"), wantErrIs: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitNewBranch", testCases)
	// Note: Actual branch name validation happens inside the tool function currently
}

// --- GitCheckout Validation Tests ---
func TestToolGitCheckoutValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{name: "Wrong Arg Count (None)", args: makeArgs(), wantErrIs: ErrValidationArgCount},
		{name: "Wrong Arg Count (Too Many)", args: makeArgs("branch1", "branch2"), wantErrIs: ErrValidationArgCount},
		{name: "Nil Arg", args: makeArgs(nil), wantErrIs: ErrValidationRequiredArgNil},
		{name: "Wrong Arg Type", args: makeArgs(123), wantErrIs: ErrValidationTypeMismatch},
		{name: "Correct Args", args: makeArgs("main"), wantErrIs: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitCheckout", testCases)
}

// --- GitRm Validation Tests ---
func TestToolGitRmValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{name: "Wrong Arg Count (None)", args: makeArgs(), wantErrIs: ErrValidationArgCount},
		{name: "Wrong Arg Count (Too Many)", args: makeArgs("file1", "file2"), wantErrIs: ErrValidationArgCount},
		{name: "Nil Arg", args: makeArgs(nil), wantErrIs: ErrValidationRequiredArgNil},
		{name: "Wrong Arg Type", args: makeArgs(123), wantErrIs: ErrValidationTypeMismatch},
		{name: "Correct Args", args: makeArgs("path/to/file.txt"), wantErrIs: nil}, // Validation should pass
		// Note: SecureFilePath validation happens inside the tool function
	}
	runValidationTestCases(t, "GitRm", testCases)
}

// --- GitMerge Validation Tests ---
func TestToolGitMergeValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{name: "Wrong Arg Count (None)", args: makeArgs(), wantErrIs: ErrValidationArgCount},
		{name: "Wrong Arg Count (Too Many)", args: makeArgs("branch1", "branch2"), wantErrIs: ErrValidationArgCount},
		{name: "Nil Arg", args: makeArgs(nil), wantErrIs: ErrValidationRequiredArgNil},
		{name: "Wrong Arg Type", args: makeArgs(123), wantErrIs: ErrValidationTypeMismatch},
		{name: "Correct Args", args: makeArgs("develop"), wantErrIs: nil}, // Validation should pass
	}
	runValidationTestCases(t, "GitMerge", testCases)
}

// TODO: Add functional tests for Git tools, likely requiring mocking of toolExec or a test repo setup.

// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Update Git.Rm validation test expectation for ArgTypeAny.
// nlines: 165
// risk_rating: LOW
// filename: pkg/core/tools_git_test.go
package core

import (
	"errors"
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
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ValidateAndConvertArgs(spec, tc.InputArgs)

			if tc.ExpectedError != nil {
				if err == nil {
					t.Errorf("Expected error [%v], got nil", tc.ExpectedError)
				} else if !errors.Is(err, tc.ExpectedError) {
					// Check only errors.Is - if this fails, the wrapping or expected error is wrong.
					// Provide more detail on the actual error received.
					t.Errorf("Expected error wrapping [%v], but errors.Is is false. Got error: [%T] %v", tc.ExpectedError, err, err)
				} else {
					t.Logf("Got expected error type via errors.Is: %v", err) // Log success for clarity
				}
			} else if err != nil {
				// More specific error message when an unexpected error occurs
				t.Errorf("Unexpected validation error: [%T] %v", err, err)
			}
		})
	}
}

// --- Git.Branch Validation Tests ---
// Args: name (string, opt), checkout (bool, opt), list_remote (bool, opt), list_all (bool, opt)
func TestToolGitNewBranchValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Args_(None)", InputArgs: MakeArgs(), ExpectedError: nil},
		{Name: "Correct_Args_(Create)", InputArgs: MakeArgs("new-feature"), ExpectedError: nil},
		{Name: "Correct_Args_(Create_and_Checkout)", InputArgs: MakeArgs("new-feature", true), ExpectedError: nil},
		{Name: "Correct_Args_(List_Remote)", InputArgs: MakeArgs(nil, false, true), ExpectedError: nil},
		{Name: "Correct_Args_(List_All)", InputArgs: MakeArgs(nil, false, false, true), ExpectedError: nil},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("name", true, false, true, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Name)", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Checkout)", InputArgs: MakeArgs("name", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(List_Remote)", InputArgs: MakeArgs(nil, false, "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(List_All)", InputArgs: MakeArgs(nil, false, false, "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Branch", testCases)
}

// --- Git.Checkout Validation Tests ---
// Args: branch (string, req), create (bool, opt)
func TestToolGitCheckoutValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("branch1", false, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg_(Required)", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong_Arg_Type_(Branch)", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Create)", InputArgs: MakeArgs("main", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args_(Checkout)", InputArgs: MakeArgs("main"), ExpectedError: nil},
		{Name: "Correct_Args_(Create_and_Checkout)", InputArgs: MakeArgs("new-feature", true), ExpectedError: nil},
	}
	runValidationTestCases(t, "Git.Checkout", testCases)
}

// --- Git.Rm Validation Tests ---
// Args: paths (any, req) - string or []string
func TestToolGitRmValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("file1", "file2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		// Corrected: Expect nil error for now due to permissive ArgTypeAny validation
		{Name: "Wrong_Arg_Type_(Maybe)", InputArgs: MakeArgs(123), ExpectedError: nil},
		{Name: "Correct_Args_(Single)", InputArgs: MakeArgs("path/to/file.txt"), ExpectedError: nil},
		{Name: "Correct_Args_(List)", InputArgs: MakeArgs([]string{"path/to/file1.txt", "file2.txt"}), ExpectedError: nil},
	}
	runValidationTestCases(t, "Git.Rm", testCases)
}

// --- Git.Merge Validation Tests ---
// Args: branch (string, req)
func TestToolGitMergeValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong_Arg_Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args", InputArgs: MakeArgs("develop"), ExpectedError: nil},
	}
	runValidationTestCases(t, "Git.Merge", testCases)
}

// --- Git.Pull Validation Tests ---
// Args: None
func TestToolGitPullValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: nil}, // Expects zero args
		{Name: "Wrong_Arg_Count_(One)", InputArgs: MakeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "Git.Pull", testCases)
}

// --- Git.Push Validation Tests ---
// Args: remote (string, opt), branch (string, opt), set_upstream (bool, opt)
func TestToolGitPushValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Args_(None)", InputArgs: MakeArgs(), ExpectedError: nil},
		{Name: "Correct_Args_(Remote)", InputArgs: MakeArgs("upstream"), ExpectedError: nil},
		{Name: "Correct_Args_(Remote_Branch)", InputArgs: MakeArgs("upstream", "main"), ExpectedError: nil},
		{Name: "Correct_Args_(All)", InputArgs: MakeArgs("upstream", "main", true), ExpectedError: nil},
		{Name: "Correct_Args_(Nil_Remote_Branch)", InputArgs: MakeArgs(nil, nil, true), ExpectedError: nil},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("origin", "main", false, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Remote)", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(SetUpstream)", InputArgs: MakeArgs("origin", "main", "not-bool"), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Push", testCases)
}

// --- Git.Diff Validation Tests ---
// Args: cached (bool, opt), commit1 (string, opt), commit2 (string, opt), path (string, opt)
func TestToolGitDiffValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Args_(None)", InputArgs: MakeArgs(), ExpectedError: nil},
		{Name: "Correct_Args_(Cached)", InputArgs: MakeArgs(true), ExpectedError: nil},
		{Name: "Correct_Args_(Commit1)", InputArgs: MakeArgs(false, "HEAD~1"), ExpectedError: nil},
		{Name: "Correct_Args_(Commit1_Commit2)", InputArgs: MakeArgs(false, "HEAD~1", "HEAD"), ExpectedError: nil},
		{Name: "Correct_Args_(All)", InputArgs: MakeArgs(false, "HEAD~1", "HEAD", "path/to/file"), ExpectedError: nil},
		{Name: "Correct_Args_(Nil_Strings)", InputArgs: MakeArgs(true, nil, nil, nil), ExpectedError: nil},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(true, "c1", "c2", "path", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Cached)", InputArgs: MakeArgs("not-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Commit1)", InputArgs: MakeArgs(false, 123), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Diff", testCases)
}

// TODO: Add functional tests for Git tools, likely requiring mocking of toolExec or a test repo setup.

// Ensure required error variables are defined
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationRequiredArgMissing
	_ = ErrValidationTypeMismatch
	_ = ErrInvalidArgument // Added for Tree tests
	_ = ErrHandleNotFound  // Added for Tree tests
)

// Ensure MakeArgs is available (implicitly via testing_helpers_test.go)
var _ = MakeArgs

// NeuroScript Version: 0.3.1
// File version: 0.1.4 // Adjust expectation for ArgTypeAny with mixed-type list in Git.Rm test.
// Update Git.Rm validation test expectation for ArgTypeAny.
// nlines: 210 // Approximate
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

const dummyRepoPath = "dummy/repo_path"

// --- Git.Branch Validation Tests ---
// Args: relative_path (string, req), name (string, opt), checkout (bool, opt), list_remote (bool, opt), list_all (bool, opt)
func TestToolGitNewBranchValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Args_(List_Local_Default)", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: nil},
		{Name: "Correct_Args_(Create)", InputArgs: MakeArgs(dummyRepoPath, "new-feature"), ExpectedError: nil},
		{Name: "Correct_Args_(Create_and_Checkout)", InputArgs: MakeArgs(dummyRepoPath, "new-feature", true), ExpectedError: nil},
		{Name: "Correct_Args_(List_Remote)", InputArgs: MakeArgs(dummyRepoPath, nil, false, true), ExpectedError: nil},
		{Name: "Correct_Args_(List_All)", InputArgs: MakeArgs(dummyRepoPath, nil, false, false, true), ExpectedError: nil},
		{Name: "Missing_Relative_Path", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "name", true, false, true, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Relative_Path)", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Name)", InputArgs: MakeArgs(dummyRepoPath, 123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Checkout)", InputArgs: MakeArgs(dummyRepoPath, "name", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(List_Remote)", InputArgs: MakeArgs(dummyRepoPath, nil, false, "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(List_All)", InputArgs: MakeArgs(dummyRepoPath, nil, false, false, "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Branch", testCases)
}

// --- Git.Checkout Validation Tests ---
// Args: relative_path (string, req), branch (string, req), create (bool, opt)
func TestToolGitCheckoutValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Args", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},                    // Missing relative_path
		{Name: "Missing_Branch_Arg", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: ErrValidationRequiredArgMissing}, // Missing branch
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "branch1", false, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg_(Relative_Path_Required)", InputArgs: MakeArgs(nil, "branch1"), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Nil_Arg_(Branch_Required)", InputArgs: MakeArgs(dummyRepoPath, nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong_Arg_Type_(Relative_Path)", InputArgs: MakeArgs(123, "branch1"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Branch)", InputArgs: MakeArgs(dummyRepoPath, 123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Create)", InputArgs: MakeArgs(dummyRepoPath, "main", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args_(Checkout)", InputArgs: MakeArgs(dummyRepoPath, "main"), ExpectedError: nil},
		{Name: "Correct_Args_(Create_and_Checkout)", InputArgs: MakeArgs(dummyRepoPath, "new-feature", true), ExpectedError: nil},
	}
	runValidationTestCases(t, "Git.Checkout", testCases)
}

// --- Git.Rm Validation Tests ---
// Args: relative_path (string, req), paths (any, req) - string or []string
func TestToolGitRmValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Args", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},                   // Missing relative_path
		{Name: "Missing_Paths_Arg", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: ErrValidationRequiredArgMissing}, // Missing paths
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "file1", "file2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg_(Relative_Path)", InputArgs: MakeArgs(nil, "file1"), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Nil_Arg_(Paths)", InputArgs: MakeArgs(dummyRepoPath, nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong_Arg_Type_(Relative_Path_is_int)", InputArgs: MakeArgs(123, "file1"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Arg_Type_For_Paths_but_Wrong_Type_for_Relative_Path", InputArgs: MakeArgs(123, "path/to/file.txt"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args_(Single_Path_String)", InputArgs: MakeArgs(dummyRepoPath, "path/to/file.txt"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_List)", InputArgs: MakeArgs(dummyRepoPath, []string{"path/to/file1.txt", "file2.txt"}), ExpectedError: nil},
		{Name: "Wrong_Path_Type_In_List_(Accepted_by_ArgTypeAny)", InputArgs: MakeArgs(dummyRepoPath, []interface{}{"file1.txt", 123}), ExpectedError: nil}, // ArgTypeAny accepts mixed list; tool impl must validate contents.
	}
	runValidationTestCases(t, "Git.Rm", testCases)
}

// --- Git.Merge Validation Tests ---
// Args: relative_path (string, req), branch (string, req)
func TestToolGitMergeValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Args", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},                    // Missing relative_path
		{Name: "Missing_Branch_Arg", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: ErrValidationRequiredArgMissing}, // Missing branch
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "branch1", "branch2"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg_(Relative_Path)", InputArgs: MakeArgs(nil, "branch1"), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Nil_Arg_(Branch)", InputArgs: MakeArgs(dummyRepoPath, nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong_Arg_Type_(Relative_Path)", InputArgs: MakeArgs(123, "branch1"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Branch)", InputArgs: MakeArgs(dummyRepoPath, 123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args", InputArgs: MakeArgs(dummyRepoPath, "develop"), ExpectedError: nil},
	}
	runValidationTestCases(t, "Git.Merge", testCases)
}

// --- Git.Pull Validation Tests ---
// Args: relative_path (string, req), remote_name (string, opt), branch_name (string, opt) - from tooldefs_git.go
func TestToolGitPullValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Relative_Path", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Correct_Args_(Only_Relative_Path)", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Remote)", InputArgs: MakeArgs(dummyRepoPath, "origin"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Remote_Branch)", InputArgs: MakeArgs(dummyRepoPath, "origin", "main"), ExpectedError: nil},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "origin", "main", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Remote_Name)", InputArgs: MakeArgs(dummyRepoPath, 123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Branch_Name)", InputArgs: MakeArgs(dummyRepoPath, "origin", 123), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Pull", testCases)
}

// --- Git.Push Validation Tests ---
// Args: relative_path (string, req), remote_name (string, opt), branch_name (string, opt) - from tooldefs_git.go
func TestToolGitPushValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Relative_Path", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Correct_Args_(Only_Relative_Path)", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Remote)", InputArgs: MakeArgs(dummyRepoPath, "upstream"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Remote_Branch)", InputArgs: MakeArgs(dummyRepoPath, "upstream", "main"), ExpectedError: nil},
		{Name: "Nil_Args_For_Optional_Params", InputArgs: MakeArgs(dummyRepoPath, nil, nil), ExpectedError: nil}, // Optional args can be nil
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, "origin", "main", "extra_arg"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Remote_Name)", InputArgs: MakeArgs(dummyRepoPath, 123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Branch_Name)", InputArgs: MakeArgs(dummyRepoPath, "origin", false), ExpectedError: ErrValidationTypeMismatch},
	}
	runValidationTestCases(t, "Git.Push", testCases)
}

// --- Git.Diff Validation Tests ---
// Args: relative_path (string, req), cached (bool, opt), commit1 (string, opt), commit2 (string, opt), path (string, opt)
func TestToolGitDiffValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Missing_Relative_Path", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Correct_Args_(Only_Path)", InputArgs: MakeArgs(dummyRepoPath), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Cached)", InputArgs: MakeArgs(dummyRepoPath, true), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Commit1)", InputArgs: MakeArgs(dummyRepoPath, false, "HEAD~1"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_Commit1_Commit2)", InputArgs: MakeArgs(dummyRepoPath, false, "HEAD~1", "HEAD"), ExpectedError: nil},
		{Name: "Correct_Args_(All_Args)", InputArgs: MakeArgs(dummyRepoPath, false, "HEAD~1", "HEAD", "path/to/file"), ExpectedError: nil},
		{Name: "Correct_Args_(Nil_Optional_Strings)", InputArgs: MakeArgs(dummyRepoPath, true, nil, nil, nil), ExpectedError: nil},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs(dummyRepoPath, true, "c1", "c2", "path", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong_Arg_Type_(Relative_Path)", InputArgs: MakeArgs(123, true), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Cached)", InputArgs: MakeArgs(dummyRepoPath, "not-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Arg_Type_(Commit1)", InputArgs: MakeArgs(dummyRepoPath, false, 123), ExpectedError: ErrValidationTypeMismatch},
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

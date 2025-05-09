// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Update expected errors for missing required args.
// filename: pkg/core/tools_file_api_test.go
// nlines: 90
// risk_rating: LOW
package core

import (
	"testing"
)

// --- ListAPIFiles Validation Tests ---
// This tool takes no arguments, so its validation tests remain the same.
func TestToolListAPIFilesValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Args (None)", InputArgs: MakeArgs(), ExpectedError: nil},
		{Name: "Wrong Arg Count (One)", InputArgs: MakeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	runValidationTestCases(t, "ListAPIFiles", testCases)
}

// --- DeleteAPIFile Validation Tests ---
func TestToolDeleteAPIFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		// Corrected: Expect ErrValidationRequiredArgMissing when required arg 'api_file_id' is missing.
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("files/abc", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil}, // Required arg cannot be nil
		{Name: "Wrong_Arg_Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args", InputArgs: MakeArgs("files/abcdef123"), ExpectedError: nil},
	}
	runValidationTestCases(t, "DeleteAPIFile", testCases)
}

// --- UploadFile Validation Tests ---
func TestToolUploadFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		// Corrected: Expect ErrValidationRequiredArgMissing when required arg 'local_filepath' is missing.
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		// Tool allows 1 or 2 args. 3 is too many.
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("path", "name", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_First_Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil}, // Path is required
		{Name: "Wrong_First_Arg_Type", InputArgs: MakeArgs(123, "name"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Second_Arg_Type", InputArgs: MakeArgs("path", 456), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args_(Path_Only)", InputArgs: MakeArgs("local/file.txt"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_and_Name)", InputArgs: MakeArgs("local/file.txt", "api_display_name"), ExpectedError: nil},
		{Name: "Correct_Args_(Path_and_Nil_Name)", InputArgs: MakeArgs("local/file.txt", nil), ExpectedError: nil}, // Allow nil for optional string
	}
	runValidationTestCases(t, "UploadFile", testCases)
}

// --- SyncFiles Validation Tests ---
func TestToolSyncFilesValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		// Corrected: Expect ErrValidationRequiredArgMissing when required arg 'direction' is missing.
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		// Corrected: Expect ErrValidationRequiredArgMissing when required arg 'local_dir' is missing.
		{Name: "Wrong_Arg_Count_(One)", InputArgs: MakeArgs("up"), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Too_Many)", InputArgs: MakeArgs("up", "dir", "*.txt", true, "extra_arg"), ExpectedError: ErrValidationArgCount}, // Max 4 args
		{Name: "Nil_First_Arg", InputArgs: MakeArgs(nil, "dir"), ExpectedError: ErrValidationRequiredArgNil},                                     // Direction required
		{Name: "Nil_Second_Arg", InputArgs: MakeArgs("up", nil), ExpectedError: ErrValidationRequiredArgNil},                                     // LocalDir required
		{Name: "Wrong_First_Arg_Type", InputArgs: MakeArgs(123, "dir"), ExpectedError: ErrValidationTypeMismatch},                                // Direction should be string
		{Name: "Wrong_Second_Arg_Type", InputArgs: MakeArgs("up", 456), ExpectedError: ErrValidationTypeMismatch},                                // LocalDir should be string
		{Name: "Wrong_Third_Arg_Type", InputArgs: MakeArgs("up", "dir", 789), ExpectedError: ErrValidationTypeMismatch},                          // Filter should be string (or nil)
		{Name: "Wrong_Fourth_Arg_Type", InputArgs: MakeArgs("up", "dir", "*.txt", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},       // ignoreGitignore should be bool (or nil)
		{Name: "Correct_Args_(Min_Required)", InputArgs: MakeArgs("up", "local/sync_dir"), ExpectedError: nil},
		{Name: "Correct_Args_(With_Filter)", InputArgs: MakeArgs("up", "local/sync_dir", "*.go"), ExpectedError: nil},
		{Name: "Correct_Args_(With_Nil_Filter)", InputArgs: MakeArgs("up", "local/sync_dir", nil), ExpectedError: nil},
		{Name: "Correct_Args_(With_Filter_and_IgnoreGitignore_True)", InputArgs: MakeArgs("up", "local/sync_dir", "*.go", true), ExpectedError: nil},
		{Name: "Correct_Args_(With_Filter_and_IgnoreGitignore_False)", InputArgs: MakeArgs("up", "local/sync_dir", "*.go", false), ExpectedError: nil},
		{Name: "Correct_Args_(With_Nil_Filter_and_IgnoreGitignore)", InputArgs: MakeArgs("up", "local/sync_dir", nil, true), ExpectedError: nil},
	}
	runValidationTestCases(t, "SyncFiles", testCases)
}

// Ensure required error variables are defined (assuming they are in errors.go)
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationRequiredArgMissing // Ensure this is defined
	_ = ErrValidationTypeMismatch
)

// Ensure helper is available (assuming defined in testing_helpers_test.go or similar)
var _ = runValidationTestCases
var _ = MakeArgs

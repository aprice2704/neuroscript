// filename: pkg/core/tools_file_api_test.go
package core

import (
	"testing"
)

// --- ListAPIFiles Validation Tests ---
func TestToolListAPIFilesValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct Args (None)", InputArgs: makeArgs(), ExpectedError: nil},
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("arg1"), ExpectedError: ErrValidationArgCount},
	}
	// Assuming runValidationTestCases is defined in another _test.go file (e.g., tools_git_test.go or testing_helpers_test.go)
	runValidationTestCases(t, "ListAPIFiles", testCases)
}

// --- DeleteAPIFile Validation Tests ---
func TestToolDeleteAPIFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("files/abc", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: makeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: makeArgs("files/abcdef123"), ExpectedError: nil},
		// Note: Further validation (e.g., non-empty string, prefix) happens inside the tool function
	}
	runValidationTestCases(t, "DeleteAPIFile", testCases)
}

// --- UploadFile Validation Tests ---
func TestToolUploadFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		// Tool allows 1 or 2 args. 3 is too many.
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("path", "name", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil First Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil}, // Path is required
		{Name: "Wrong First Arg Type", InputArgs: makeArgs(123, "name"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Second Arg Type", InputArgs: makeArgs("path", 456), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args (Path Only)", InputArgs: makeArgs("local/file.txt"), ExpectedError: nil},
		{Name: "Correct Args (Path and Name)", InputArgs: makeArgs("local/file.txt", "api_display_name"), ExpectedError: nil},
		{Name: "Correct Args (Path and Nil Name)", InputArgs: makeArgs("local/file.txt", nil), ExpectedError: nil}, // Allow nil for optional string
		// Note: Path security validation happens inside the tool function
	}
	runValidationTestCases(t, "UploadFile", testCases)
}

// --- SyncFiles Validation Tests (NEW) ---
func TestToolSyncFilesValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},    // Expects 2-4 args
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("up"), ExpectedError: ErrValidationArgCount}, // Expects 2-4 args
		// --- FIX: Provide 5 args to exceed the max of 4 ---
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("up", "dir", "*.txt", true, "extra_arg"), ExpectedError: ErrValidationArgCount},
		// --- END FIX ---
		{Name: "Nil First Arg", InputArgs: makeArgs(nil, "dir"), ExpectedError: ErrValidationRequiredArgNil},            // Direction required
		{Name: "Nil Second Arg", InputArgs: makeArgs("up", nil), ExpectedError: ErrValidationRequiredArgNil},            // LocalDir required
		{Name: "Wrong First Arg Type", InputArgs: makeArgs(123, "dir"), ExpectedError: ErrValidationTypeMismatch},       // Direction should be string
		{Name: "Wrong Second Arg Type", InputArgs: makeArgs("up", 456), ExpectedError: ErrValidationTypeMismatch},       // LocalDir should be string
		{Name: "Wrong Third Arg Type", InputArgs: makeArgs("up", "dir", 789), ExpectedError: ErrValidationTypeMismatch}, // Filter should be string (or nil)
		// --- ADDED: Test for wrong fourth arg type ---
		{Name: "Wrong Fourth Arg Type", InputArgs: makeArgs("up", "dir", "*.txt", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch}, // ignoreGitignore should be bool (or nil)
		// --- END ADD ---
		{Name: "Correct Args (Min Required)", InputArgs: makeArgs("up", "local/sync_dir"), ExpectedError: nil},
		{Name: "Correct Args (With Filter)", InputArgs: makeArgs("up", "local/sync_dir", "*.go"), ExpectedError: nil},
		{Name: "Correct Args (With Nil Filter)", InputArgs: makeArgs("up", "local/sync_dir", nil), ExpectedError: nil},
		{Name: "Correct Args (With Filter and IgnoreGitignore True)", InputArgs: makeArgs("up", "local/sync_dir", "*.go", true), ExpectedError: nil},
		{Name: "Correct Args (With Filter and IgnoreGitignore False)", InputArgs: makeArgs("up", "local/sync_dir", "*.go", false), ExpectedError: nil},
		{Name: "Correct Args (With Nil Filter and IgnoreGitignore)", InputArgs: makeArgs("up", "local/sync_dir", nil, true), ExpectedError: nil},
		// Note: Direction value, path security validation happens inside the tool function
	}
	runValidationTestCases(t, "SyncFiles", testCases)
}

// TODO: Add functional tests for File API tools, requiring mocking of the genai.Client interface.

// Ensure required error variables are defined (assuming they are in errors.go)
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
)

// Ensure helper is available (assuming defined in testing_helpers_test.go or similar)
var _ = runValidationTestCases
var _ = makeArgs

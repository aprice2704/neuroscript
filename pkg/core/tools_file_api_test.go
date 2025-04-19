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
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("path", "name", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil First Arg", InputArgs: makeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong First Arg Type", InputArgs: makeArgs(123, "name"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Second Arg Type", InputArgs: makeArgs("path", 456), ExpectedError: ErrValidationTypeMismatch},
		// --- FIX: Expect ArgCount error for 3 args ---
		{Name: "Wrong Third Arg Type", InputArgs: makeArgs("path", "name", 123), ExpectedError: ErrValidationArgCount}, // Tool expects 1-2 args, 3 is invalid count
		// --- END FIX ---
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
		{Name: "Wrong Arg Count (None)", InputArgs: makeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (One)", InputArgs: makeArgs("up"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Too Many)", InputArgs: makeArgs("up", "dir", "*.txt", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil First Arg", InputArgs: makeArgs(nil, "dir"), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Nil Second Arg", InputArgs: makeArgs("up", nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong First Arg Type", InputArgs: makeArgs(123, "dir"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Second Arg Type", InputArgs: makeArgs("up", 456), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Third Arg Type", InputArgs: makeArgs("up", "dir", 789), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args (No Filter)", InputArgs: makeArgs("up", "local/sync_dir"), ExpectedError: nil},
		{Name: "Correct Args (With Filter)", InputArgs: makeArgs("up", "local/sync_dir", "*.go"), ExpectedError: nil},
		{Name: "Correct Args (With Nil Filter)", InputArgs: makeArgs("up", "local/sync_dir", nil), ExpectedError: nil},
		// Note: Direction value, path security validation happens inside the tool function
	}
	runValidationTestCases(t, "SyncFiles", testCases)
}

// TODO: Add functional tests for File API tools, requiring mocking of the genai.Client interface.

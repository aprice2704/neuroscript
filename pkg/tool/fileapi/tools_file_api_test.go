// NeuroScript Version: 0.4.1
// File version: 7
// Purpose: Corrected SyncFiles test to expect ErrLLMNotConfigured instead of a stale stub error.
// filename: pkg/tool/fileapi/tools_file_api_test.go
// nlines: 65
// risk_rating: LOW

package fileapi

import (
	"errors"
	"testing"
)

func testFileAPIToolHelper(t *testing.T, toolName string, tc struct {
	Name		string
	Args		[]interface{}
	WantErrIs	error
}) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		interp, _ := NewDefaultTestInterpreter(t)
		toolImpl, found := interp.ToolRegistry().GetTool(toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", toolName)
		}

		// Since most of these are unimplemented, we just check if the call panics or returns an expected error.
		// A real implementation would have result checking.
		_, err := toolImpl.Func(interp, tc.Args)

		if tc.WantErrIs != nil {
			if !errors.Is(err, tc.WantErrIs) {
				t.Errorf("Expected error wrapping [%v], got: %v", tc.WantErrIs, err)
			}
		} else if err != nil {
			// For unimplemented funcs, we might expect ErrFeatureNotImplemented
			if !errors.Is(err, ErrFeatureNotImplemented) {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	})
}

func TestToolListAPIFiles(t *testing.T) {
	// This tool is currently a stub. Test that it returns the correct error.
	testFileAPIToolHelper(t, "ListAPIFiles", struct {
		Name		string
		Args		[]interface{}
		WantErrIs	error
	}{Name: "Correct Args (None)", Args: MakeArgs(), WantErrIs: ErrFeatureNotImplemented})
}

func TestToolDeleteAPIFile(t *testing.T) {
	// This tool is currently a stub. Test that it returns the correct error.
	testFileAPIToolHelper(t, "DeleteAPIFile", struct {
		Name		string
		Args		[]interface{}
		WantErrIs	error
	}{Name: "Correct_Args", Args: MakeArgs("files/abcdef123"), WantErrIs: ErrFeatureNotImplemented})
}

func TestToolUploadFile(t *testing.T) {
	// This tool is currently a stub. Test that it returns the correct error.
	testFileAPIToolHelper(t, "UploadFile", struct {
		Name		string
		Args		[]interface{}
		WantErrIs	error
	}{Name: "Correct_Args_(Path_Only)", Args: MakeArgs("local/file.txt"), WantErrIs: ErrFeatureNotImplemented})
}

func TestToolSyncFiles(t *testing.T) {
	// FIX: This tool is no longer a stub. It now correctly fails because the test
	// interpreter uses a NoOpLLMClient. The expected error is ErrLLMNotConfigured.
	testFileAPIToolHelper(t, "SyncFiles", struct {
		Name		string
		Args		[]interface{}
		WantErrIs	error
	}{Name: "Correct_Args_(Min_Required)", Args: MakeArgs("up", "local/sync_dir"), WantErrIs: ErrLLMNotConfigured})
}
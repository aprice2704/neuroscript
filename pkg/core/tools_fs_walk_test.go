// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Use correct keys ('path_relative', 'is_dir') in assertWalkResultsEqual.
// nlines: 180 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_fs_walk_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// --- Test Setup Helpers ---
func createWalkTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	interp, sandboxRoot := NewDefaultTestInterpreter(t)
	base := sandboxRoot
	subdir1 := filepath.Join(base, "subdir1")
	subsubdir := filepath.Join(subdir1, "subsubdir")
	subdir2Empty := filepath.Join(base, "subdir2_empty")
	fileAtRoot := filepath.Join(base, "root.txt")
	fileInSubdir1 := filepath.Join(subdir1, "file1.txt")
	fileInSubsubdir := filepath.Join(subsubdir, "nested.txt")
	notADir := filepath.Join(base, "not_a_dir.txt")
	mustMkdir(t, subdir1)
	mustMkdir(t, subsubdir)
	mustMkdir(t, subdir2Empty)
	mustWriteFile(t, fileAtRoot, "root content")
	mustWriteFile(t, fileInSubdir1, "file1 content")
	mustWriteFile(t, fileInSubsubdir, "nested content")
	mustWriteFile(t, notADir, "i am a file")
	return interp, sandboxRoot
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("mustMkdir failed for %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("mustWriteFile failed for %s: %v", path, err)
	}
}

// --- Result Comparison Helper ---
type walkResult struct {
	Path  string
	IsDir bool
}

func assertWalkResultsEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	actualSlice, ok := actual.([]map[string]interface{})
	if !ok {
		t.Fatalf("Actual result is not []map[string]interface{}, got %T", actual)
	}
	expectedSlice, ok := expected.([]walkResult)
	if !ok {
		t.Fatalf("Expected value is not []walkResult, got %T", expected)
	}

	if len(expectedSlice) != len(actualSlice) {
		t.Errorf("Expected %d entries, got %d.\nExpected Paths: %v\nActual Raw: %+v",
			len(expectedSlice), len(actualSlice), getPaths(expectedSlice), actualSlice)
		return
	}

	// Convert actual maps to simplified structs for comparison
	actualSimplified := make([]walkResult, len(actualSlice))
	for i, item := range actualSlice {
		// *** Use keys from the Actual Raw output shown in logs ***
		pathVal, pathOk := item["path_relative"].(string) // <-- CORRECTED KEY
		isDirVal, isDirOk := item["is_dir"].(bool)        // <-- CORRECTED KEY
		if !pathOk || !isDirOk {
			t.Errorf("Failed to extract path_relative (%t) or is_dir (%t) from actual result map at index %d: %+v", pathOk, isDirOk, i, item)
			// Assign default values to avoid panic, comparison will likely fail below
			pathVal = "[EXTRACT_ERROR]"
			isDirVal = false
		}
		actualSimplified[i] = walkResult{Path: pathVal, IsDir: isDirVal}
	}

	// Sort both slices by path for comparison
	sort.Slice(expectedSlice, func(i, j int) bool { return expectedSlice[i].Path < expectedSlice[j].Path })
	sort.Slice(actualSimplified, func(i, j int) bool { return actualSimplified[i].Path < actualSimplified[j].Path })

	if !reflect.DeepEqual(expectedSlice, actualSimplified) {
		t.Errorf("Walk results mismatch (after sorting, ignoring size/modTime):\nExpected: %+v\nActual:   %+v\n---\nActual Raw: %+v",
			expectedSlice, actualSimplified, actualSlice)
	}
}

func getPaths(results []walkResult) []string {
	paths := make([]string, len(results))
	for i, r := range results {
		paths[i] = r.Path
	}
	return paths
}

// --- Test Function ---
func TestToolWalkDir(t *testing.T) {
	interp, _ := createWalkTestInterpreter(t)

	testCases := []struct {
		name          string
		startPath     string
		wantResult    []walkResult
		wantToolErrIs error
		wantErrMsg    string
	}{
		{
			name:      "Walk from root",
			startPath: ".",
			wantResult: []walkResult{
				{Path: "not_a_dir.txt", IsDir: false},
				{Path: "root.txt", IsDir: false},
				{Path: "subdir1", IsDir: true},
				{Path: "subdir1/file1.txt", IsDir: false},
				{Path: "subdir1/subsubdir", IsDir: true},
				{Path: "subdir1/subsubdir/nested.txt", IsDir: false},
				{Path: "subdir2_empty", IsDir: true},
			},
		},
		{
			name:      "Walk from subdir1",
			startPath: "subdir1",
			wantResult: []walkResult{
				{Path: "file1.txt", IsDir: false},
				{Path: "subsubdir", IsDir: true},
				{Path: "subsubdir/nested.txt", IsDir: false},
			},
		},
		{
			name:       "Walk from empty subdir",
			startPath:  "subdir2_empty",
			wantResult: []walkResult{},
		},
		{
			name:          "Walk from file path",
			startPath:     "not_a_dir.txt",
			wantResult:    nil,
			wantToolErrIs: ErrPathNotDirectory,
			wantErrMsg:    "is not a directory",
		},
		{
			name:          "Walk from non-existent path",
			startPath:     "nonexistent_dir",
			wantResult:    nil,
			wantToolErrIs: ErrFileNotFound,
			wantErrMsg:    "start path not found",
		},
		{
			name:          "Path outside sandbox",
			startPath:     "../other",
			wantResult:    nil,
			wantToolErrIs: ErrPathViolation,
			wantErrMsg:    "path resolves outside allowed directory",
		},
		{
			name:          "Null byte in path",
			startPath:     "subdir1\x00/invalid",
			wantResult:    nil,
			wantToolErrIs: ErrNullByteInArgument,
			wantErrMsg:    "input path contains null byte",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsTC := fsTestCase{
				name:          tc.name,
				toolName:      "WalkDir",
				args:          MakeArgs(tc.startPath),
				wantResult:    tc.wantResult, // This is []walkResult for comparison func
				wantToolErrIs: tc.wantToolErrIs,
			}
			// Pass the slice of maps expected by the *tool function* itself
			// runWalkTest needs adjustment if wantResult in fsTestCase isn't the tool output type
			runWalkTest(t, interp, fsTC, tc.wantResult, tc.wantErrMsg)
		})
	}
}

func runWalkTest(t *testing.T, interp *Interpreter, tc fsTestCase, expectedResultStructs []walkResult, wantErrMsg string) {
	t.Helper()
	toolImpl, found := interp.GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found", tc.toolName)
	}
	gotResult, toolErr := toolImpl.Func(interp, tc.args) // gotResult is []map[string]interface{}

	if tc.wantToolErrIs != nil {
		if toolErr == nil {
			t.Errorf("Expected error wrapping [%v], but got nil.", tc.wantToolErrIs)
			return
		}
		if !errors.Is(toolErr, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
		}
		if wantErrMsg != "" && !strings.Contains(toolErr.Error(), wantErrMsg) {
			t.Errorf("Expected error message containing %q, but got: %q", wantErrMsg, toolErr.Error())
		}
		if gotResult != nil {
			t.Errorf("Expected nil result on error, but got: %+v", gotResult)
		}
	} else {
		if toolErr != nil {
			t.Errorf("Unexpected error: %v", toolErr)
			return
		}
		// Compare the expected structs with the actual maps returned by the tool
		assertWalkResultsEqual(t, expectedResultStructs, gotResult)
	}
}

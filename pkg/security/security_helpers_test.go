// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Adjust expectation for Mkdir traversal test based on Clean behavior.
// nlines: 130 // Approximate
// risk_rating: MEDIUM // Tests security-critical functions
// filename: pkg/core/security_helpers_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAndSecurePath(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "test-sandbox-root-")
	if err != nil {
		t.Fatalf("Failed to create temp root dir: %v", err)
	}
	defer os.RemoveAll(tempRoot)

	allowedRoot := filepath.Join(tempRoot, "allowed")
	if err := os.Mkdir(allowedRoot, 0755); err != nil {
		t.Fatalf("Failed to create allowed root dir: %v", err)
	}
	outsideRoot := filepath.Join(tempRoot, "outside")
	if err := os.Mkdir(outsideRoot, 0755); err != nil {
		t.Fatalf("Failed to create outside root dir: %v", err)
	}

	// Define test cases
	testCases := []struct {
		name                 string
		inputPath            string // Path relative to allowedRoot
		wantAbsPathSuffix    string // Expected *suffix* of the absolute path if successful
		wantErrIs            error  // Expected sentinel error (nil for success)
		wantErrorMsgContains string // Substring expected in error message
	}{
		// --- Valid Cases ---
		{
			name:              "Valid Simple Path",
			inputPath:         "file.txt",
			wantAbsPathSuffix: filepath.Join("allowed", "file.txt"),
			wantErrIs:         nil,
		},
		{
			name:              "Valid Subdir Path",
			inputPath:         filepath.Join("subdir", "file.txt"),
			wantAbsPathSuffix: filepath.Join("allowed", "subdir", "file.txt"),
			wantErrIs:         nil,
		},
		{
			name:              "Valid Path With Dot",
			inputPath:         filepath.Join("subdir", ".", "file.txt"),
			wantAbsPathSuffix: filepath.Join("allowed", "subdir", "file.txt"),
			wantErrIs:         nil,
		},
		{
			name:              "Valid Path With Internal Dot Dot",
			inputPath:         filepath.Join("subdir", "nested", "..", "file.txt"),
			wantAbsPathSuffix: filepath.Join("allowed", "subdir", "file.txt"),
			wantErrIs:         nil,
		},
		{
			name:              "Valid Path Resolving To Root",
			inputPath:         filepath.Join("subdir", ".."),
			wantAbsPathSuffix: filepath.Join("allowed"),
			wantErrIs:         nil,
		},
		{
			name:              "Valid Path Using Dot",
			inputPath:         ".",
			wantAbsPathSuffix: filepath.Join("allowed"),
			wantErrIs:         nil,
		},
		// --- Traversal that cleans to *inside* allowedRoot ---
		// This case is considered valid by the current Rel-based check, as Clean resolves
		// allowed/some/dir/../../outside to allowed/outside.
		{
			name:              "Valid Complex Traversal (Cleans to Inside Sibling)",
			inputPath:         "some/dir/../../outside", // Cleans to "outside" relative to allowedRoot
			wantAbsPathSuffix: filepath.Join("allowed", "outside"),
			wantErrIs:         nil, // No error expected as Rel("allowed", "allowed/outside") is "outside"
		},
		// --- Invalid Cases ---
		{
			name:                 "Invalid Empty Path",
			inputPath:            "",
			wantErrIs:            ErrInvalidArgument,
			wantErrorMsgContains: "input path cannot be empty",
		},
		{
			name:                 "Invalid Null Byte",
			inputPath:            "file\x00.txt",
			wantErrIs:            ErrNullByteInArgument,
			wantErrorMsgContains: "input path contains null byte",
		},
		{
			name:                 "Invalid Absolute Path",
			inputPath:            filepath.Join(allowedRoot, "abs.txt"),
			wantErrIs:            ErrPathViolation,
			wantErrorMsgContains: "must be relative, not absolute",
		},
		{
			name:                 "Invalid Simple Traversal Up",
			inputPath:            "..",
			wantErrIs:            ErrPathViolation,
			wantErrorMsgContains: "resolves to", // Expect message about resolving outside
		},
		{
			name:                 "Invalid Simple Traversal File",
			inputPath:            filepath.Join("..", "file.txt"),
			wantErrIs:            ErrPathViolation,
			wantErrorMsgContains: "resolves to", // Expect message about resolving outside
		},
		// This case correctly cleans to parent/outside/file.txt, triggering Rel starting with ..
		{
			name:                 "Invalid Complex Traversal (Cleans to Parent Sibling)",
			inputPath:            filepath.Join("subdir", "..", "..", "outside", "file.txt"),
			wantErrIs:            ErrPathViolation,
			wantErrorMsgContains: "resolves to", // Expect message about resolving outside
		},
		{
			name:                 "Invalid Traversal Leading To Sibling Dir",
			inputPath:            filepath.Join("..", "outside"),
			wantErrIs:            ErrPathViolation,
			wantErrorMsgContains: "resolves to", // Expect message about resolving outside
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAbsPath, gotErr := ResolveAndSecurePath(tc.inputPath, allowedRoot)
			if tc.wantErrIs != nil {
				if gotErr == nil {
					t.Errorf("Expected error wrapping [%v], but got nil", tc.wantErrIs)
				} else if !errors.Is(gotErr, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], but got error: %v (type %T)", tc.wantErrIs, gotErr, gotErr)
				} else {
					// Check contains only if specified and error is correct type
					if tc.wantErrorMsgContains != "" && !strings.Contains(gotErr.Error(), tc.wantErrorMsgContains) {
						t.Errorf("Expected error message containing %q, but got: %q", tc.wantErrorMsgContains, gotErr.Error())
					}
				}
				if gotAbsPath != "" {
					t.Errorf("Expected empty path on error, but got %q", gotAbsPath)
				}
			} else { // Expected success
				if gotErr != nil {
					t.Errorf("Expected no error, but got: %v", gotErr)
				}
				expectedFullPath := filepath.Join(tempRoot, tc.wantAbsPathSuffix)
				expectedFullPath = filepath.Clean(expectedFullPath) // Ensure expected path is clean
				cleanedGotPath := filepath.Clean(gotAbsPath)        // Ensure actual path is clean for comparison

				if cleanedGotPath != expectedFullPath {
					t.Errorf("Path mismatch:\n  Got (Cleaned):  %q\n  Want (Cleaned): %q\n  (Original Got: %q)", cleanedGotPath, expectedFullPath, gotAbsPath)
				}
			}
		})
	}
}

// --- TestIsPathInSandbox unchanged ---
func TestIsPathInSandbox(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "test-sandbox-root-ispath-")
	if err != nil {
		t.Fatalf("Failed to create temp root dir: %v", err)
	}
	defer os.RemoveAll(tempRoot)
	allowedRoot := filepath.Join(tempRoot, "allowed")
	if err := os.Mkdir(allowedRoot, 0755); err != nil {
		t.Fatalf("Failed to create allowed root dir: %v", err)
	}

	testCases := []struct {
		name      string
		inputPath string
		wantIn    bool
		wantErr   bool // Expect errors other than PathViolation
	}{
		{"Valid Simple", "file.txt", true, false},
		{"Valid Subdir", "subdir/file.txt", true, false},
		{"Valid Root", ".", true, false},
		{"Valid Internal ..", "a/../b.txt", true, false},
		{"Valid Complex Internal ..", "a/b/../../c.txt", true, false},             // Resolves to allowed/c.txt
		{"Invalid Simple ..", "..", false, false},                                 // Resolves outside, wantIn=false, no other error
		{"Invalid Traversal", "../outside", false, false},                         // Resolves outside, wantIn=false, no other error
		{"Invalid Complex Traversal", "a/../../outside", false, false},            // Resolves outside, wantIn=false, no other error
		{"Invalid Absolute", filepath.Join(allowedRoot, "abs.txt"), false, false}, // Absolute path violation, wantIn=false, no other error
		{"Invalid Empty", "", false, true},                                        // Other error expected
		{"Invalid Null Byte", "a\x00b", false, true},                              // Other error expected
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotIn, gotErr := IsPathInSandbox(allowedRoot, tc.inputPath)
			if tc.wantErr {
				if gotErr == nil {
					t.Errorf("Expected an error other than PathViolation, but got nil")
				}
				// Check if it's the expected non-PathViolation error if needed
				// For now, just checking if *any* error occurred is sufficient
			} else { // wantErr is false
				if gotErr != nil {
					// Allow ErrPathViolation if wantIn is false, otherwise fail
					isPathViolation := false
					if re, ok := gotErr.(*RuntimeError); ok && errors.Is(re.Wrapped, ErrPathViolation) {
						isPathViolation = true
					}
					if !(!tc.wantIn && isPathViolation) { // Error is only OK if wantIn=false AND it's ErrPathViolation
						t.Errorf("Expected no error (or only PathViolation if wantIn=false), but got: %v (type %T)", gotErr, gotErr)
					}
				}
			}
			if gotIn != tc.wantIn {
				t.Errorf("Expected in sandbox = %t, but got %t (err: %v)", tc.wantIn, gotIn, gotErr)
			}
		})
	}
}

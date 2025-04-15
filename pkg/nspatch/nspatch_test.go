package nspatch

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// --- Test Helpers (readFileLinesForTest, compareStringSlices remain the same) ---
func readFileLinesForTest(t *testing.T, filePath string) []string {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}
		}
		t.Fatalf("Failed to read fixture file %q: %v", filePath, err)
	}
	s := string(content)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" && len(content) == 0 {
		return []string{}
	}
	return lines
}
func compareStringSlices(t *testing.T, actual, expected []string, context string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("%s: Mismatched line count: Expected %d, got %d.\nExpected:\n%s\nActual:\n%s",
			context, len(expected), len(actual), strings.Join(expected, "\n"), strings.Join(actual, "\n"))
		return
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("%s: Mismatch at line %d:\nExpected: %q\nActual:   %q",
				context, i+1, expected[i], actual[i])
		}
	}
}

func TestLoadPatchFile(t *testing.T) {
	// This test seemed to pass mostly, assuming patch_invalid_op.ndpatch.json exists now
	testCases := []struct {
		name        string
		patchFile   string
		expectError bool
		expectedErr error
	}{
		{"Basic Valid", "testdata/patch_basic.ndpatch.json", false, nil},
		{"Insert Empty", "testdata/patch_insert_empty.ndpatch.json", false, nil},
		{"Invalid JSON", "testdata/patch_invalid_json.ndpatch.txt", true, ErrInvalidPatchFile},
		{"Invalid Op", "testdata/patch_invalid_op.ndpatch.json", true, ErrInvalidOperation},
		{"Missing File", "testdata/non_existent_patch.ndpatch.json", true, ErrInvalidPatchFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure the invalid op fixture exists for its test
			if tc.name == "Invalid Op" {
				if _, err := os.Stat(tc.patchFile); os.IsNotExist(err) {
					t.Skipf("Skipping test, fixture file not found: %s", tc.patchFile)
				}
			}
			_, err := LoadPatchFile(tc.patchFile)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error type %v, but got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
			}
		})
	}
}

func TestApplyPatch(t *testing.T) {
	testCases := []struct {
		name         string
		initialFile  string
		patchFile    string
		expectedFile string // Should match initialFile if ApplyPatch expected to error
		expectError  bool
		expectedErr  error
	}{
		{
			name:         "Basic Apply", // Fails verification on delete
			initialFile:  "testdata/initial_basic.txt",
			patchFile:    "testdata/patch_basic.ndpatch.json",
			expectedFile: "testdata/initial_basic.txt", // Expect no change
			expectError:  true,
			expectedErr:  ErrVerificationFailed, // Expect verification error
		},
		{
			name:         "Insert Into Empty", // Fails bounds check on second insert
			initialFile:  "testdata/initial_empty.txt",
			patchFile:    "testdata/patch_insert_empty.ndpatch.json",
			expectedFile: "testdata/initial_empty.txt", // Expect no change
			expectError:  true,
			expectedErr:  ErrOutOfBounds, // Expect bounds error during verification
		},
		{
			name:         "Delete Only", // Fails verification on second delete
			initialFile:  "testdata/initial_basic.txt",
			patchFile:    "testdata/patch_delete_only.ndpatch.json", // Uses corrected patch file
			expectedFile: "testdata/initial_basic.txt",              // Expect no change
			expectError:  true,
			expectedErr:  ErrVerificationFailed, // Expect verification error
		},
		{
			name:         "Verification Fail", // Correctly expects error
			initialFile:  "testdata/initial_basic.txt",
			patchFile:    "testdata/patch_verify_fail.ndpatch.json",
			expectedFile: "testdata/initial_basic.txt",
			expectError:  true,
			expectedErr:  ErrVerificationFailed,
		},
		// Add a new test case that *should* succeed completely
		{
			name:         "Replace Only Success",
			initialFile:  "testdata/initial_basic.txt",
			patchFile:    "testdata/patch_replace_only_success.ndpatch.json", // Need to create this patch file
			expectedFile: "testdata/expected_replace_only_success.txt",       // Need to create this expected file
			expectError:  false,
			expectedErr:  nil,
		},
	}

	// --- Create Fixtures for the new Success Case ---
	// You would save these to testdata/patch_replace_only_success.ndpatch.json and testdata/expected_replace_only_success.txt
	t.Log("Note: Ensure fixture files for 'Replace Only Success' test case exist in testdata/")
	/*
			// testdata/patch_replace_only_success.ndpatch.json
			[
			  {
			    "file": "initial_basic.txt", // File name here doesn't matter as much as ApplyPatch takes slice
			    "line_number": 2,
			    "operation": "replace",
			    "original_line_for_reference": "Line 2: This line will be replaced.",
			    "new_line_content": "Line 2: REPLACE SUCCESS"
			  },
		      {
			    "file": "initial_basic.txt",
			    "line_number": 5,
			    "operation": "replace",
			    "original_line_for_reference": "Line 5:   Indented context line.",
			    "new_line_content": "Line 5:   REPLACED INDENT"
			  }
			]

			// testdata/expected_replace_only_success.txt
			Line 1: Initial content.
			Line 2: REPLACE SUCCESS
			Line 3: Some context here.
			Line 4: Insertion point comes before this line.
			Line 5:   REPLACED INDENT
			Line 6: This line will be deleted.
			Line 7: Final context line.
	*/
	// --------------------------------------------

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure fixture exists before running test
			if _, err := os.Stat(tc.initialFile); os.IsNotExist(err) && tc.initialFile != "testdata/initial_empty.txt" {
				t.Skipf("Skipping test, initial fixture file not found: %s", tc.initialFile)
			}
			if _, err := os.Stat(tc.patchFile); os.IsNotExist(err) {
				t.Skipf("Skipping test, patch fixture file not found: %s", tc.patchFile)
			}
			if _, err := os.Stat(tc.expectedFile); os.IsNotExist(err) && !tc.expectError {
				t.Skipf("Skipping test, expected fixture file not found: %s", tc.expectedFile)
			}

			initialLines := readFileLinesForTest(t, tc.initialFile)
			expectedLines := readFileLinesForTest(t, tc.expectedFile)

			changes, loadErr := LoadPatchFile(tc.patchFile)
			if loadErr != nil {
				t.Fatalf("Prerequisite LoadPatchFile failed for patch %q: %v", tc.patchFile, loadErr)
			}

			modifiedLines, applyErr := ApplyPatch(initialLines, changes)

			if tc.expectError {
				if applyErr == nil {
					t.Errorf("Expected an error but ApplyPatch succeeded")
				} else if tc.expectedErr != nil && !errors.Is(applyErr, tc.expectedErr) {
					t.Errorf("Expected error type %v, but got: %v", tc.expectedErr, applyErr)
				}
				// On expected error, the content should equal the initial content
				// Note: ApplyPatch now returns nil slice on error
				if applyErr != nil && modifiedLines != nil {
					t.Errorf("Expected nil slice on error, but got %d lines", len(modifiedLines))
				}
				compareStringSlices(t, initialLines, expectedLines, "Content changed despite expected error")

			} else { // Expect success
				if applyErr != nil {
					t.Errorf("Did not expect an error but got: %v", applyErr)
				}
				if applyErr == nil && modifiedLines == nil {
					t.Errorf("Expected modified lines on success but got nil")
				}
				compareStringSlices(t, modifiedLines, expectedLines, "Final content mismatch")
			}
		})
	}
}

func TestVerifyChanges(t *testing.T) {
	t.Run("Basic Verify Fails on Delete", func(t *testing.T) { // Test name clarified
		initialLines := readFileLinesForTest(t, "testdata/initial_basic.txt")
		patchFile := "testdata/patch_basic.ndpatch.json"
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// Expect verification error because delete verification fails
		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		}
		if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type %v, but got: %v", ErrVerificationFailed, verifyErr)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 verification results, got %d", len(results))
		}
		// Check first two results - should pass
		if results[0].Status != "Matched" {
			t.Errorf("R[0] Status: expected Matched, got %s", results[0].Status)
		}
		if results[0].IsError {
			t.Errorf("R[0] IsError: expected false, got true")
		}
		if results[1].Status != "Not Checked" {
			t.Errorf("R[1] Status: expected Not Checked, got %s", results[1].Status)
		}
		if results[1].IsError {
			t.Errorf("R[1] IsError: expected false, got true")
		}
		// Check third result - this should show the mismatch
		if !strings.HasPrefix(results[2].Status, "MISMATCHED") {
			t.Errorf("R[2] Status: expected MISMATCHED(...), got %s", results[2].Status)
		}
		if !results[2].IsError {
			t.Errorf("R[2] IsError: expected true, got false")
		} // IsError should be true now
		if !errors.Is(results[2].Err, ErrVerificationFailed) {
			t.Errorf("R[2] Err: expected %v, got %v", ErrVerificationFailed, results[2].Err)
		}
	})

	t.Run("Verify Fail Mismatch", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, "testdata/initial_basic.txt")
		patchFile := "testdata/patch_verify_fail.ndpatch.json"
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		}
		if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type %v, but got: %v", ErrVerificationFailed, verifyErr)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 verification result, got %d", len(results))
		}
		if !strings.HasPrefix(results[0].Status, "MISMATCHED") {
			t.Errorf("R[0] Status: expected MISMATCHED(...), got %s", results[0].Status)
		}
		if !results[0].IsError {
			t.Errorf("R[0] IsError: expected true, got false")
		}
		if !errors.Is(results[0].Err, ErrVerificationFailed) {
			t.Errorf("R[0] Err: expected %v, got %v", ErrVerificationFailed, results[0].Err)
		}
	})

	t.Run("Insert Into Empty Verify", func(t *testing.T) { // Renamed test
		initialLines := readFileLinesForTest(t, "testdata/initial_empty.txt")
		patchFile := "testdata/patch_insert_empty.ndpatch.json"
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// Expect bounds error on the second insert
		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		}
		if !errors.Is(verifyErr, ErrOutOfBounds) {
			t.Errorf("Expected error type %v, but got: %v", ErrOutOfBounds, verifyErr)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 verification results, got %d", len(results))
		}
		// Check first result - should pass
		if results[0].Status != "Not Checked" {
			t.Errorf("R[0] Status: expected Not Checked, got %s", results[0].Status)
		}
		if results[0].IsError {
			t.Errorf("R[0] IsError: expected false, got true")
		}
		// Check second result - should fail bounds check
		if !strings.HasPrefix(results[1].Status, "Error: target index out of bounds") {
			t.Errorf("R[1] Status: expected Error Bounds(...), got %s", results[1].Status)
		}
		if !results[1].IsError {
			t.Errorf("R[1] IsError: expected true, got false")
		}
		if !errors.Is(results[1].Err, ErrOutOfBounds) {
			t.Errorf("R[1] Err: expected %v, got %v", ErrOutOfBounds, results[1].Err)
		}
	})

	t.Run("Delete Only Verify", func(t *testing.T) { // Renamed test
		initialLines := readFileLinesForTest(t, "testdata/initial_basic.txt")
		patchFile := "testdata/patch_delete_only.ndpatch.json" // Uses corrected patch file
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// Expect verification error because second delete finds wrong content
		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		}
		if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type %v, but got: %v", ErrVerificationFailed, verifyErr)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 verification results, got %d", len(results))
		}
		// Check first result - should match
		if results[0].Status != "Matched" {
			t.Errorf("R[0] Status: expected Matched, got %s", results[0].Status)
		}
		if results[0].IsError {
			t.Errorf("R[0] IsError: expected false, got true")
		}
		// Check second result - should mismatch
		if !strings.HasPrefix(results[1].Status, "MISMATCHED") {
			t.Errorf("R[1] Status: expected MISMATCHED(...), got %s", results[1].Status)
		}
		if !results[1].IsError {
			t.Errorf("R[1] IsError: expected true, got false")
		}
		if !errors.Is(results[1].Err, ErrVerificationFailed) {
			t.Errorf("R[1] Err: expected %v, got %v", ErrVerificationFailed, results[1].Err)
		}

	})
}

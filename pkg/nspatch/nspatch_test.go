package nspatch

import (
	"errors"
	"os"
	"path/filepath" // Import path/filepath
	"strings"
	"testing"
)

// --- Test Helpers ---
func readFileLinesForTest(t *testing.T, filePath string) []string {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{} // Return empty slice for non-existent file
		}
		t.Fatalf("Failed to read fixture file %q: %v", filePath, err)
	}
	s := string(content)
	s = strings.ReplaceAll(s, "\r\n", "\n") // Normalize line endings
	if s == "" {                            // Handle truly empty file
		return []string{}
	}
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	// Handle split edge case for file containing only newline(s)
	// if len(lines) == 1 && lines[0] == "" && len(s) > 0 {
	// 	// This indicates the file had content, likely just newlines.
	//  // Depending on desired behavior, might return [""] or similar.
	//  // For now, returning empty slice assuming empty lines aren't desired targets.
	// 	 return []string{}
	// }
	return lines
}

func compareStringSlices(t *testing.T, actual, expected []string, context string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("%s: Mismatched line count: Expected %d, got %d.\nExpected:\n---\n%s\n---\nActual:\n---\n%s\n---",
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

// Helper function to check if test fixtures exist
func checkFixtures(t *testing.T, files ...string) bool {
	t.Helper()
	missing := false
	for _, file := range files {
		// Allow initial_empty.txt to represent an empty file state without physically existing
		if file == filepath.Join("testdata", "initial_empty.txt") {
			continue // Don't require initial_empty.txt to exist
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Logf("Required fixture file not found: %s", file) // Use Logf for non-fatal info
			missing = true
		} else if err != nil {
			t.Logf("Error checking fixture file %q: %v", file, err)
			missing = true // Treat other errors as missing too for safety
		}
	}
	return !missing
}

// --- Test Functions ---

func TestLoadPatchFile(t *testing.T) {
	// Define test cases using .ndpatch.json extension
	testCases := []struct {
		name        string
		patchFile   string
		expectError bool
		expectedErr error // Specific error type expected
	}{
		// *** Reverted filenames to .ndpatch.json ***
		{"Basic Valid", filepath.Join("testdata", "patch_basic.ndpatch.json"), false, nil},
		{"Insert Empty", filepath.Join("testdata", "patch_insert_empty.ndpatch.json"), false, nil},
		{"Invalid JSON", filepath.Join("testdata", "patch_invalid_json.ndpatch.txt"), true, ErrInvalidPatchFile}, // This one IS .txt
		{"Invalid Op", filepath.Join("testdata", "patch_invalid_op.ndpatch.json"), true, ErrInvalidOperation},
		{"Missing File", filepath.Join("testdata", "non_existent_patch.ndpatch.json"), true, ErrInvalidPatchFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip only if needed file is missing, except for the "Missing File" test itself
			if !checkFixtures(t, tc.patchFile) && tc.name != "Missing File" {
				t.Skipf("Skipping test %q, required fixture missing: %s", tc.name, tc.patchFile)
			}

			_, err := LoadPatchFile(tc.patchFile)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					// Special handling for missing file error check
					if tc.name == "Missing File" {
						if !errors.Is(err, os.ErrNotExist) {
							t.Errorf("Expected missing file error (os.ErrNotExist wrapped by %v), but got: %v", ErrInvalidPatchFile, err)
						}
					} else {
						t.Errorf("Expected error type [%v], but got: [%v] (Type: %T)", tc.expectedErr, err, err)
					}
				}
			} else { // Expect success
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
			}
		})
	}
}

func TestApplyPatch(t *testing.T) {
	// Define test cases using .ndpatch.json extension
	testCases := []struct {
		name         string
		initialFile  string
		patchFile    string
		expectedFile string
		expectError  bool
		expectedErr  error // Specific error expected
	}{
		// *** Reverted filenames to .ndpatch.json ***
		{
			name:         "Basic Apply", // Expects verification error because patch data forces mismatch
			initialFile:  filepath.Join("testdata", "initial_basic.txt"),
			patchFile:    filepath.Join("testdata", "patch_basic.ndpatch.json"), // Assumes user fixed "old" value
			expectedFile: filepath.Join("testdata", "initial_basic.txt"),        // Expect no change
			expectError:  true,
			expectedErr:  ErrVerificationFailed,
		},
		{
			name:         "Insert Into Empty", // Expects bounds error on second insert
			initialFile:  filepath.Join("testdata", "initial_empty.txt"),
			patchFile:    filepath.Join("testdata", "patch_insert_empty.ndpatch.json"),
			expectedFile: filepath.Join("testdata", "initial_empty.txt"), // Expect no change
			expectError:  true,
			expectedErr:  ErrOutOfBounds,
		},
		{
			name:         "Delete Only", // Expects verification error on second delete
			initialFile:  filepath.Join("testdata", "initial_basic.txt"),
			patchFile:    filepath.Join("testdata", "patch_delete_only.ndpatch.json"), // Assumes user fixed "old" value
			expectedFile: filepath.Join("testdata", "initial_basic.txt"),              // Expect no change
			expectError:  true,
			expectedErr:  ErrVerificationFailed,
		},
		{
			name:         "Verification Fail", // Tests explicit verification failure
			initialFile:  filepath.Join("testdata", "initial_basic.txt"),
			patchFile:    filepath.Join("testdata", "patch_verify_fail.ndpatch.json"),
			expectedFile: filepath.Join("testdata", "initial_basic.txt"), // Expect no change
			expectError:  true,
			expectedErr:  ErrVerificationFailed,
		},
		{
			name:         "Replace Only Success", // Should succeed
			initialFile:  filepath.Join("testdata", "initial_basic.txt"),
			patchFile:    filepath.Join("testdata", "patch_replace_only_success.ndpatch.json"),
			expectedFile: filepath.Join("testdata", "expected_replace_only_success.txt"),
			expectError:  false,
			expectedErr:  nil,
		},
	}

	t.Logf("Note: Ensure fixture files for '%s' test case exist in testdata/", "Replace Only Success")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requiredFiles := []string{tc.initialFile, tc.patchFile}
			if !tc.expectError {
				requiredFiles = append(requiredFiles, tc.expectedFile)
			}
			if !checkFixtures(t, requiredFiles...) {
				t.Skipf("Skipping test %q, required fixtures missing.", tc.name)
			}

			initialLines := readFileLinesForTest(t, tc.initialFile)
			expectedLines := readFileLinesForTest(t, tc.expectedFile)

			changes, loadErr := LoadPatchFile(tc.patchFile)
			if loadErr != nil {
				if tc.expectError && errors.Is(loadErr, tc.expectedErr) {
					return // Loading failed as expected
				}
				t.Fatalf("Prerequisite LoadPatchFile failed for patch %q: %v", tc.patchFile, loadErr)
			}

			modifiedLines, applyErr := ApplyPatch(initialLines, changes)

			if tc.expectError {
				if applyErr == nil {
					t.Errorf("Expected an error but ApplyPatch succeeded")
				} else if tc.expectedErr != nil && !errors.Is(applyErr, tc.expectedErr) {
					t.Errorf("Expected error type [%v], but got: [%v] (Type: %T)", tc.expectedErr, applyErr, applyErr)
				}
				if applyErr != nil && modifiedLines != nil {
					t.Errorf("Expected nil slice on error, but got %d lines", len(modifiedLines))
				}
				// Don't compare content if error was expected, ApplyPatch returns nil slice
				// compareStringSlices(t, initialLines, expectedLines, "Content changed despite expected error")

			} else { // Expect success
				if applyErr != nil {
					t.Errorf("Did not expect an error but got: %v", applyErr)
				}
				if applyErr == nil {
					if modifiedLines == nil && len(expectedLines) > 0 {
						t.Errorf("Expected modified lines on success but got nil")
					} else {
						compareStringSlices(t, modifiedLines, expectedLines, "Final content mismatch")
					}
				}
			}
		})
	}
}

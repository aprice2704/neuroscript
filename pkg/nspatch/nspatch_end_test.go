package nspatch

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// NOTE: readFileLinesForTest, compareStringSlices, checkFixtures assumed to be in nspatch_test.go

func TestVerifyChanges(t *testing.T) {
	t.Run("Basic_Verify_Fails_on_Mismatch", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// Assumes "old" for line 6 in patch_basic.ndpatch.json causes a mismatch vs initial_basic.txt
		// User MUST ensure patch_basic.ndpatch.json contains an 'old' value for change 2 (line 6)
		// that DOES NOT match "Line 6: This line will be deleted."
		patchFile := filepath.Join("testdata", "patch_basic.ndpatch.json")
		if !checkFixtures(t, patchFile, filepath.Join("testdata", "initial_basic.txt")) {
			t.Skipf("Skipping test, fixture missing.")
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// Expect verification error because delete verification fails due to mismatch
		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		} else if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type [%v], but got: [%v]", ErrVerificationFailed, verifyErr)
		}

		// Assuming patch_basic has 3 changes: replace line 2 (match), insert line 4 (ok), delete line 6 (mismatch)
		// User needs to adjust expectedResults count and details based on actual patch_basic.ndpatch.json
		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			{StatusPrefix: "Matched", IsError: false, ErrIs: nil},                     // Assuming change 0 (replace line 2) matches
			{StatusPrefix: "OK (No Verification Needed)", IsError: false, ErrIs: nil}, // Assuming change 1 (insert line 4) is OK
			{StatusPrefix: "MISMATCHED", IsError: true, ErrIs: ErrVerificationFailed}, // Assuming change 2 (delete line 6) mismatches
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}

		for i, expected := range expectedResults {
			if i >= len(results) { // Avoid index out of bounds if results length is wrong
				t.Errorf("Result index %d out of bounds (only %d results)", i, len(results))
				continue
			}
			actual := results[i]
			// Use HasPrefix for error statuses which might include details
			if !strings.HasPrefix(actual.Status, expected.StatusPrefix) {
				t.Errorf("R[%d] Status: expected prefix %q, got %q", i, expected.StatusPrefix, actual.Status)
			}
			if actual.IsError != expected.IsError {
				t.Errorf("R[%d] IsError: expected %t, got %t (Err: %v)", i, expected.IsError, actual.IsError, actual.Err)
			}
			// Check error type if expected
			if expected.ErrIs != nil {
				if actual.Err == nil {
					t.Errorf("R[%d] Err: expected type [%v], got nil", i, expected.ErrIs)
				} else if !errors.Is(actual.Err, expected.ErrIs) {
					t.Errorf("R[%d] Err: expected type [%v], got [%v]", i, expected.ErrIs, actual.Err)
				}
			} else if actual.Err != nil { // Check if error is present when not expected
				t.Errorf("R[%d] Err: expected nil, got [%v]", i, actual.Err)
			}
		}
	})

	t.Run("Verify_Fail_Direct_Mismatch", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// Assumes patch_verify_fail.ndpatch.json has one change with a deliberately wrong 'old' value
		patchFile := filepath.Join("testdata", "patch_verify_fail.ndpatch.json")
		if !checkFixtures(t, patchFile, filepath.Join("testdata", "initial_basic.txt")) {
			t.Skipf("Skipping test, fixture missing.")
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		} else if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type %v, but got: %v", ErrVerificationFailed, verifyErr)
		}

		// Assuming patch_verify_fail has only one change that mismatches
		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			{StatusPrefix: "MISMATCHED", IsError: true, ErrIs: ErrVerificationFailed},
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification result, got %d", len(expectedResults), len(results))
		}
		for i, expected := range expectedResults {
			if i >= len(results) {
				t.Errorf("Result index %d out of bounds (only %d results)", i, len(results))
				continue
			}
			actual := results[i]
			if !strings.HasPrefix(actual.Status, expected.StatusPrefix) {
				t.Errorf("R[%d] Status: expected prefix %q, got %q", i, expected.StatusPrefix, actual.Status)
			}
			if actual.IsError != expected.IsError {
				t.Errorf("R[%d] IsError: expected %t, got %t (Err: %v)", i, expected.IsError, actual.IsError, actual.Err)
			}
			if expected.ErrIs != nil && !errors.Is(actual.Err, expected.ErrIs) {
				t.Errorf("R[%d] Err: expected type [%v], got [%v]", i, expected.ErrIs, actual.Err)
			} else if expected.ErrIs == nil && actual.Err != nil {
				t.Errorf("R[%d] Err: expected nil, got [%v]", i, actual.Err)
			}
		}
	})

	t.Run("Insert_Into_Empty_Verify", func(t *testing.T) {
		// *** THIS TEST CASE IS CORRECTED ***
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_empty.txt"))
		patchFile := filepath.Join("testdata", "patch_insert_empty.ndpatch.json")
		// No need to check initial_empty.txt existence
		if !checkFixtures(t, patchFile) {
			t.Skipf("Skipping test, fixture missing: %s", patchFile)
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// *** CHANGE: Expect SUCCESS (nil error) ***
		if verifyErr != nil {
			t.Errorf("Expected verification to succeed, but got error: %v", verifyErr)
		}

		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			// *** CHANGE: Expect OK for both inserts ***
			{StatusPrefix: "OK (No Verification Needed)", IsError: false, ErrIs: nil},
			{StatusPrefix: "OK (No Verification Needed)", IsError: false, ErrIs: nil},
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}
		for i, expected := range expectedResults {
			if i >= len(results) {
				t.Errorf("Result index %d out of bounds (only %d results)", i, len(results))
				continue
			}
			actual := results[i]
			if !strings.HasPrefix(actual.Status, expected.StatusPrefix) {
				t.Errorf("R[%d] Status: expected prefix %q, got %q", i, expected.StatusPrefix, actual.Status)
			}
			if actual.IsError != expected.IsError {
				t.Errorf("R[%d] IsError: expected %t, got %t (Err: %v)", i, expected.IsError, actual.IsError, actual.Err)
			}
			if expected.ErrIs != nil && !errors.Is(actual.Err, expected.ErrIs) {
				t.Errorf("R[%d] Err: expected type [%v], got [%v]", i, expected.ErrIs, actual.Err)
			} else if expected.ErrIs == nil && actual.Err != nil {
				t.Errorf("R[%d] Err: expected nil, got [%v]", i, actual.Err)
			}
		}
	})

	t.Run("Delete_Only_Verify_Mismatch", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// Assumes patch_delete_only.ndpatch.json forces mismatch on second delete
		// User MUST ensure this patch file contains 'old' for change 1 (line 6)
		// that DOES NOT match "Line 6: This line will be deleted."
		patchFile := filepath.Join("testdata", "patch_delete_only.ndpatch.json")
		if !checkFixtures(t, patchFile, filepath.Join("testdata", "initial_basic.txt")) {
			t.Skipf("Skipping test, fixture missing.")
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		} else if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type %v, but got: %v", ErrVerificationFailed, verifyErr)
		}

		// Assuming patch_delete_only has 2 changes: delete line 1 (match), delete line 6 (mismatch)
		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			{StatusPrefix: "Matched", IsError: false, ErrIs: nil},                     // Assuming change 0 (delete line 1) matches
			{StatusPrefix: "MISMATCHED", IsError: true, ErrIs: ErrVerificationFailed}, // Assuming change 1 (delete line 6) mismatches
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}
		for i, expected := range expectedResults {
			if i >= len(results) {
				t.Errorf("Result index %d out of bounds (only %d results)", i, len(results))
				continue
			}
			actual := results[i]
			if !strings.HasPrefix(actual.Status, expected.StatusPrefix) {
				t.Errorf("R[%d] Status: expected prefix %q, got %q", i, expected.StatusPrefix, actual.Status)
			}
			if actual.IsError != expected.IsError {
				t.Errorf("R[%d] IsError: expected %t, got %t (Err: %v)", i, expected.IsError, actual.IsError, actual.Err)
			}
			if expected.ErrIs != nil && !errors.Is(actual.Err, expected.ErrIs) {
				t.Errorf("R[%d] Err: expected type [%v], got [%v]", i, expected.ErrIs, actual.Err)
			} else if expected.ErrIs == nil && actual.Err != nil {
				t.Errorf("R[%d] Err: expected nil, got [%v]", i, actual.Err)
			}
		}
	})
}

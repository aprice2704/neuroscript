package nspatch

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyChanges(t *testing.T) {
	t.Run("Basic_Verify_Fails_on_Delete", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// *** Reverted filenames to .ndpatch.json ***
		patchFile := filepath.Join("testdata", "patch_basic.ndpatch.json") // Assumes "old" for line 6 forces mismatch now
		if !checkFixtures(t, patchFile) {
			t.Skipf("Skipping test, fixture missing: %s", patchFile)
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		// Expect verification error because delete verification fails
		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		} else if !errors.Is(verifyErr, ErrVerificationFailed) {
			t.Errorf("Expected error type [%v], but got: [%v]", ErrVerificationFailed, verifyErr)
		}

		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			{StatusPrefix: "Matched", IsError: false, ErrIs: nil},
			// *** CORRECTED EXPECTED STATUS ***
			{StatusPrefix: "OK (No Verification Needed)", IsError: false, ErrIs: nil},
			{StatusPrefix: "MISMATCHED", IsError: true, ErrIs: ErrVerificationFailed},
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}

		for i, expected := range expectedResults {
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

	t.Run("Verify_Fail_Mismatch", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// *** Reverted filenames to .ndpatch.json ***
		patchFile := filepath.Join("testdata", "patch_verify_fail.ndpatch.json")
		if !checkFixtures(t, patchFile) {
			t.Skipf("Skipping test, fixture missing: %s", patchFile)
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
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_empty.txt"))
		// *** Reverted filenames to .ndpatch.json ***
		patchFile := filepath.Join("testdata", "patch_insert_empty.ndpatch.json")
		if !checkFixtures(t, patchFile) {
			t.Skipf("Skipping test, fixture missing: %s", patchFile)
		}
		changes, err := LoadPatchFile(patchFile)
		if err != nil {
			t.Fatalf("Failed to load patch: %v", err)
		}

		results, verifyErr := VerifyChanges(initialLines, changes)

		if verifyErr == nil {
			t.Errorf("Expected verification error, but got nil")
		} else if !errors.Is(verifyErr, ErrOutOfBounds) {
			t.Errorf("Expected error type %v, but got: %v", ErrOutOfBounds, verifyErr)
		}

		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			// *** CORRECTED EXPECTED STATUS ***
			{StatusPrefix: "OK (No Verification Needed)", IsError: false, ErrIs: nil},
			{StatusPrefix: "Error: target index out of bounds", IsError: true, ErrIs: ErrOutOfBounds},
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}
		for i, expected := range expectedResults {
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

	t.Run("Delete_Only_Verify", func(t *testing.T) {
		initialLines := readFileLinesForTest(t, filepath.Join("testdata", "initial_basic.txt"))
		// *** Reverted filenames to .ndpatch.json ***
		patchFile := filepath.Join("testdata", "patch_delete_only.ndpatch.json") // Assumes "old" forces mismatch on second delete
		if !checkFixtures(t, patchFile) {
			t.Skipf("Skipping test, fixture missing: %s", patchFile)
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

		expectedResults := []struct {
			StatusPrefix string
			IsError      bool
			ErrIs        error
		}{
			{StatusPrefix: "Matched", IsError: false, ErrIs: nil},
			{StatusPrefix: "MISMATCHED", IsError: true, ErrIs: ErrVerificationFailed},
		}

		if len(results) != len(expectedResults) {
			t.Fatalf("Expected %d verification results, got %d", len(expectedResults), len(results))
		}
		for i, expected := range expectedResults {
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

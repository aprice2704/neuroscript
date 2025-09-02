// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 host-side ProgressTracker.
// filename: aeiou/progress_tracker_test.go
// nlines: 60
// risk_rating: LOW

package aeiou

import "testing"

func TestProgressTracker(t *testing.T) {
	t.Run("Halts after default N repeats", func(t *testing.T) {
		tracker := NewProgressTracker(0) // Use default
		digest := "digest-a"

		// Simulate turns
		if tracker.CheckAndRecord(digest) {
			t.Error("Should not halt after 1st occurrence")
		}
		if tracker.CheckAndRecord(digest) {
			t.Error("Should not halt after 2nd occurrence")
		}
		if !tracker.CheckAndRecord(digest) {
			t.Error("Should halt after 3rd occurrence")
		}
	})

	t.Run("Halts after custom N repeats", func(t *testing.T) {
		tracker := NewProgressTracker(2)
		digest := "digest-b"

		if tracker.CheckAndRecord(digest) {
			t.Error("Should not halt after 1st occurrence")
		}
		if !tracker.CheckAndRecord(digest) {
			t.Error("Should halt after 2nd occurrence")
		}
	})

	t.Run("Resets when digest changes", func(t *testing.T) {
		tracker := NewProgressTracker(3)
		digest1 := "digest-c1"
		digest2 := "digest-c2"

		tracker.CheckAndRecord(digest1)
		tracker.CheckAndRecord(digest1) // 2 hits

		// The digest changes, resetting the counter
		if tracker.CheckAndRecord(digest2) {
			t.Error("Should not halt when digest changes")
		}
		if tracker.consecutiveHits != 1 {
			t.Errorf("Expected consecutive hits to be 1, got %d", tracker.consecutiveHits)
		}

		// It should now take 2 more identical digests to halt
		tracker.CheckAndRecord(digest2)
		if !tracker.CheckAndRecord(digest2) {
			t.Error("Should halt after 3 consecutive occurrences of the new digest")
		}
	})
}

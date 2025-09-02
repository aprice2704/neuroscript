// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the host-side progress guard tracking logic.
// filename: aeiou/progress_tracker.go
// nlines: 39
// risk_rating: LOW

package aeiou

const (
	// DefaultNoProgressHalt is the default number of times a digest can be repeated.
	DefaultNoProgressHalt = 3
)

// ProgressTracker monitors the host-observed digest across turns to detect
// repetitive loops where the agent is making no progress.
type ProgressTracker struct {
	maxRepeats      int
	lastDigest      string
	consecutiveHits int
}

// NewProgressTracker creates a new tracker with a configurable repeat limit.
func NewProgressTracker(maxRepeats int) *ProgressTracker {
	if maxRepeats <= 0 {
		maxRepeats = DefaultNoProgressHalt
	}
	return &ProgressTracker{
		maxRepeats: maxRepeats,
	}
}

// CheckAndRecord updates the tracker with the latest digest and returns true if
// the no-progress limit has been exceeded.
func (pt *ProgressTracker) CheckAndRecord(digest string) bool {
	if digest == pt.lastDigest {
		pt.consecutiveHits++
	} else {
		pt.lastDigest = digest
		pt.consecutiveHits = 1 // The first time we see a digest counts as 1 hit.
	}
	return pt.consecutiveHits >= pt.maxRepeats
}

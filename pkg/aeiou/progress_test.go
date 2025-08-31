// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 host-side progress guard.
// filename: aeiou/progress_test.go
// nlines: 60
// risk_rating: LOW

package aeiou

import (
	"fmt"
	"testing"
)

func TestComputeHostDigest(t *testing.T) {
	// A sample control token to be stripped
	controlToken := fmt.Sprintf("%s:%s:%s.%s%s",
		TokenMarkerPrefix,
		KindLoop,
		"payload_b64",
		"tag_b64",
		TokenMarkerSuffix,
	)

	testCases := []struct {
		name         string
		output       string
		scratchpad   string
		wantSameAs   *struct{ out, scr string } // Pointer to compare against another case
		wantDiffFrom *struct{ out, scr string } // Pointer to compare against another case
	}{
		{
			name:       "Base case",
			output:     "line 1\nline 2",
			scratchpad: "note 1",
		},
		{
			name:       "Identical content produces same digest",
			output:     "line 1\nline 2",
			scratchpad: "note 1",
			wantSameAs: &struct{ out, scr string }{"line 1\nline 2", "note 1"},
		},
		{
			name:         "Different content produces different digest",
			output:       "line 1\nline 2 changed",
			scratchpad:   "note 1",
			wantDiffFrom: &struct{ out, scr string }{"line 1\nline 2", "note 1"},
		},
		{
			name:       "Control tokens are stripped",
			output:     fmt.Sprintf("line 1\n%s\nline 2", controlToken),
			scratchpad: "note 1",
			wantSameAs: &struct{ out, scr string }{"line 1\nline 2", "note 1"},
		},
		{
			name:       "Trailing whitespace is normalized",
			output:     "line 1  \nline 2\t",
			scratchpad: "note 1",
			wantSameAs: &struct{ out, scr string }{"line 1\nline 2", "note 1"},
		},
	}

	digests := make(map[string]string)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			digest := ComputeHostDigest(tc.output, tc.scratchpad)
			digests[tc.name] = digest

			if tc.wantSameAs != nil {
				baseDigest := ComputeHostDigest(tc.wantSameAs.out, tc.wantSameAs.scr)
				if digest != baseDigest {
					t.Errorf("Expected digest to be the same, but they were different.\n- Got:  %s\n- Want: %s", digest, baseDigest)
				}
			}
			if tc.wantDiffFrom != nil {
				baseDigest := ComputeHostDigest(tc.wantDiffFrom.out, tc.wantDiffFrom.scr)
				if digest == baseDigest {
					t.Errorf("Expected digest to be different, but they were the same: %s", digest)
				}
			}
		})
	}
}

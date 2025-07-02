// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 01:05:00 AM PDT
// filename: pkg/neurodata/checklist/checklist_status_test.go

package checklist

import (
	"testing"
)

func TestCalculateAutomaticStatus(t *testing.T) {
	testCases := []struct {
		name           string
		childStatuses  []string
		childSymbols   map[int]string // Only needed for 'special' cases
		expectedStatus string
		expectedSymbol string
		expectError    bool
	}{
		// Rule 5 (Handled by caller, but test edge case)
		{name: "No children", childStatuses: []string{}, expectedStatus: "open", expectedSymbol: ""},

		// Rule 4: All Open
		{name: "All Open (1)", childStatuses: []string{"open"}, expectedStatus: "open", expectedSymbol: ""},
		{name: "All Open (3)", childStatuses: []string{"open", "open", "open"}, expectedStatus: "open", expectedSymbol: ""},

		// Rule 3: All Done
		{name: "All Done (1)", childStatuses: []string{"done"}, expectedStatus: "done", expectedSymbol: ""}, // Rule 3 takes precedence over Rule 2 here
		{name: "All Done (3)", childStatuses: []string{"done", "done", "done"}, expectedStatus: "done", expectedSymbol: ""},

		// Rule 2: Partial Triggers
		{name: "Partial Trigger (Done)", childStatuses: []string{"open", "done"}, expectedStatus: "partial", expectedSymbol: ""},
		{name: "Partial Trigger (Skipped)", childStatuses: []string{"open", "skipped", "open"}, expectedStatus: "partial", expectedSymbol: ""},
		{name: "Partial Trigger (Partial)", childStatuses: []string{"open", "partial"}, expectedStatus: "partial", expectedSymbol: ""},
		{name: "Partial Trigger (Mixed)", childStatuses: []string{"done", "skipped", "partial", "open"}, expectedStatus: "partial", expectedSymbol: ""},
		{name: "Partial Trigger (Done+Skipped)", childStatuses: []string{"done", "skipped"}, expectedStatus: "partial", expectedSymbol: ""},

		// Rule 1: Priority Blocked
		{name: "Priority Blocked (Alone)", childStatuses: []string{"blocked"}, expectedStatus: "blocked", expectedSymbol: ""},
		{name: "Priority Blocked (Mixed)", childStatuses: []string{"open", "done", "blocked", "question"}, expectedStatus: "blocked", expectedSymbol: ""},

		// Rule 1: Priority Question
		{name: "Priority Question (Alone)", childStatuses: []string{"question"}, expectedStatus: "question", expectedSymbol: ""},
		{name: "Priority Question (Mixed)", childStatuses: []string{"open", "done", "question", "inprogress"}, expectedStatus: "question", expectedSymbol: ""},

		// Rule 1: Priority InProgress
		{name: "Priority InProgress (Alone)", childStatuses: []string{"inprogress"}, expectedStatus: "inprogress", expectedSymbol: ""},
		{name: "Priority InProgress (Mixed)", childStatuses: []string{"open", "done", "inprogress", "special"}, childSymbols: map[int]string{3: "*"}, expectedStatus: "inprogress", expectedSymbol: ""},

		// Rule 1: Priority Special
		{name: "Priority Special (Alone)", childStatuses: []string{"special"}, childSymbols: map[int]string{0: "*"}, expectedStatus: "special", expectedSymbol: "*"},
		{name: "Priority Special (Mixed)", childStatuses: []string{"open", "done", "special", "open"}, childSymbols: map[int]string{2: "A"}, expectedStatus: "special", expectedSymbol: "A"},
		{name: "Priority Special (Multiple, First Wins)", childStatuses: []string{"open", "special", "special", "done"}, childSymbols: map[int]string{1: "1", 2: "2"}, expectedStatus: "special", expectedSymbol: "1"},

		// Error Cases
		{name: "Error: Special Status, Symbol Missing", childStatuses: []string{"open", "special"}, childSymbols: map[int]string{0: "?"} /* Missing symbol for index 1 */, expectError: true},

		// Mixed non-triggering
		{name: "Mixed Open and Unknown", childStatuses: []string{"open", "unknown"}, expectedStatus: "open"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, symbol, err := calculateAutomaticStatus(tc.childStatuses, tc.childSymbols)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else {
					t.Logf("Got expected error: %v", err)
				}
				return // Don't check status/symbol if error was expected
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if status != tc.expectedStatus {
				t.Errorf("Status mismatch: want %q, got %q (child statuses: %v)", tc.expectedStatus, status, tc.childStatuses)
			}

			if symbol != tc.expectedSymbol {
				t.Errorf("Symbol mismatch: want %q, got %q (child statuses: %v)", tc.expectedSymbol, symbol, tc.childStatuses)
			}
		})
	}
}

// NeuroScript Version: 0.3.0
// File version: 3 // Bumped version
// Purpose: Unit tests for capability matching logic, separated from test case data. Fixed panic by constructing invalid grants directly.
// filename: pkg/policy/capability/matcher_test.go
// nlines: 70 // Adjusted line count
// risk_rating: LOW

package capability

import "testing"

// Test cases are defined in matcher_test_cases.go

func TestCapsSatisfied(t *testing.T) {
	for _, tc := range capsSatisfiedTestCases { // Use test cases from separate file
		t.Run(tc.name, func(t *testing.T) {
			// Handle cases that would cause MustParse to panic by creating structs directly
			needs := make([]Capability, len(tc.needs))
			for i, s := range tc.needs {
				if s == "*" || s == "*:*" { // Check for invalid strings for MustParse
					// Handle reconstruction if needed, though needs should generally be valid
					t.Fatalf("Test setup error: Invalid capability string '%s' found in 'needs' for test case '%s'. Needs must be parsable.", s, tc.name)
				}
				needs[i] = MustParse(s)
			}

			grants := make([]Capability, len(tc.grants))
			for i, s := range tc.grants {
				if s == "*" {
					grants[i] = Capability{Resource: "*"} // Create incomplete struct directly
				} else if s == "*:*" {
					grants[i] = Capability{Resource: "*", Verbs: []string{"*"}} // Create incomplete struct directly
				} else {
					// Use MustParse for valid grant strings
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Fatalf("Test setup error: MustParse panicked for grant string '%s' in test case '%s': %v", s, tc.name, r)
							}
						}()
						grants[i] = MustParse(s)
					}()
				}
			}

			satisfied := CapsSatisfied(needs, grants)
			if satisfied != tc.expectSatisfied {
				t.Errorf("CapsSatisfied() = %v, want %v", satisfied, tc.expectSatisfied)
				t.Logf("  Needs (parsed): %v", needs)
				t.Logf("  Grants (parsed): %v", grants)
				t.Logf("  Needs (original): %v", tc.needs)
				t.Logf("  Grants (original): %v", tc.grants)
			}
		})
	}
}

// --- Keep existing Limits tests ---

func TestLimits_BudgetRunAndPerCall(t *testing.T) {
	gs := GrantSet{
		Limits: Limits{
			BudgetPerRunCents:  map[string]int{"CAD": 100},
			BudgetPerCallCents: map[string]int{"CAD": 60},
		},
		Counters: NewCounters(),
	}
	if err := gs.CheckPerCallBudget("CAD", 61); err != ErrBudgetExceeded {
		t.Errorf("expected per-call budget exceed error, got %v", err)
	}
	if err := gs.ChargeBudget("CAD", 90); err != nil {
		t.Errorf("unexpected error charging budget: %v", err)
	}
	if err := gs.ChargeBudget("CAD", 20); err != ErrBudgetExceeded {
		t.Errorf("expected run budget exceed error, got %v", err)
	}
}

func TestLimits_NetAndFS(t *testing.T) {
	gs := GrantSet{
		Limits: Limits{
			NetMaxCalls: 1,
			NetMaxBytes: 10,
			FSMaxCalls:  1,
			FSMaxBytes:  5,
		},
		Counters: NewCounters(),
	}
	if err := gs.CountNet(5); err != nil {
		t.Errorf("unexpected net count error: %v", err)
	}
	if err := gs.CountNet(1); err != ErrNetExceeded {
		t.Errorf("expected net exceeded, got %v", err)
	}
	if err := gs.CountFS(3); err != nil {
		t.Errorf("unexpected fs count error: %v", err)
	}
	if err := gs.CountFS(3); err != ErrFSExceeded {
		t.Errorf("expected fs exceeded, got %v", err)
	}
}

func TestLimits_ToolCalls(t *testing.T) {
	gs := GrantSet{
		Limits: Limits{
			ToolMaxCalls: map[string]int{"dangerTool": 1},
		},
		Counters: NewCounters(),
	}
	if err := gs.CountToolCall("dangerTool"); err != nil {
		t.Errorf("unexpected tool call error: %v", err)
	}
	if err := gs.CountToolCall("dangerTool"); err != ErrToolExceeded {
		t.Errorf("expected tool exceeded, got %v", err)
	}
}

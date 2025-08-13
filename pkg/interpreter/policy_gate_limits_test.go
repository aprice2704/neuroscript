// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains policy gate tests for resource limits and budget enforcement.
// filename: pkg/interpreter/policy_gate_limits_test.go
// nlines: 150
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

func TestPolicyGate_Limits(t *testing.T) {
	// --- Per-Tool Call Limit ---
	t.Run("[Limits] Tool call count", func(t *testing.T) {
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextNormal,
			Allow:   []string{"*"},
			Grants: capability.NewGrantSet(nil, capability.Limits{
				ToolMaxCalls: map[string]int{"tool.limited.call": 2},
			}),
		}
		tool := runtime.ToolMeta{Name: "tool.limited.call"}

		// First two calls should succeed
		if err := policy.CanCall(tool); err != nil {
			t.Fatalf("Expected 1st call to succeed, got %v", err)
		}
		if err := policy.CanCall(tool); err != nil {
			t.Fatalf("Expected 2nd call to succeed, got %v", err)
		}

		// Third call should fail
		err := policy.CanCall(tool)
		if !errors.Is(err, capability.ErrToolExceeded) {
			t.Errorf("Expected 3rd call to fail with ErrToolExceeded, got %v", err)
		}
	})

	// --- Network Limits ---
	t.Run("[Limits] Network call and byte limits", func(t *testing.T) {
		grants := capability.NewGrantSet(nil, capability.Limits{
			NetMaxCalls: 2,
			NetMaxBytes: 1000,
		})

		// Call 1 (500 bytes) - OK
		if err := grants.CountNet(500); err != nil {
			t.Fatalf("Net count 1 failed unexpectedly: %v", err)
		}

		// Call 2 (600 bytes) - Fail on bytes
		err := grants.CountNet(600)
		if !errors.Is(err, capability.ErrNetExceeded) {
			t.Errorf("Expected ErrNetExceeded on byte limit, got %v", err)
		}

		// Reset counters for next test
		grants.Counters = capability.NewCounters()

		// Call 1 (100 bytes) - OK
		if err := grants.CountNet(100); err != nil {
			t.Fatalf("Net count 3 failed unexpectedly: %v", err)
		}
		// Call 2 (100 bytes) - OK
		if err := grants.CountNet(100); err != nil {
			t.Fatalf("Net count 4 failed unexpectedly: %v", err)
		}
		// Call 3 (100 bytes) - Fail on calls
		err = grants.CountNet(100)
		if !errors.Is(err, capability.ErrNetExceeded) {
			t.Errorf("Expected ErrNetExceeded on call limit, got %v", err)
		}
	})

	// --- Filesystem Limits ---
	t.Run("[Limits] Filesystem byte limit", func(t *testing.T) {
		grants := capability.NewGrantSet(nil, capability.Limits{
			FSMaxBytes: 512,
		})
		if err := grants.CountFS(256); err != nil {
			t.Fatalf("FS count 1 failed unexpectedly: %v", err)
		}
		if err := grants.CountFS(256); err != nil {
			t.Fatalf("FS count 2 failed unexpectedly: %v", err)
		}
		err := grants.CountFS(1)
		if !errors.Is(err, capability.ErrFSExceeded) {
			t.Errorf("Expected ErrFSExceeded on byte limit, got %v", err)
		}
	})

	// --- Budget Limits ---
	t.Run("[Limits] Budget per-call and per-run", func(t *testing.T) {
		grants := capability.NewGrantSet(nil, capability.Limits{
			BudgetPerCallCents: map[string]int{"CAD": 50},
			BudgetPerRunCents:  map[string]int{"CAD": 120},
		})

		// Per-call check
		err := grants.CheckPerCallBudget("CAD", 51)
		if !errors.Is(err, capability.ErrBudgetExceeded) {
			t.Errorf("Expected ErrBudgetExceeded for per-call check, got %v", err)
		}
		if err := grants.CheckPerCallBudget("CAD", 50); err != nil {
			t.Errorf("Did not expect error for valid per-call check, got %v", err)
		}

		// Per-run charges
		if err := grants.ChargeBudget("CAD", 40); err != nil {
			t.Fatalf("Charge 1 failed: %v", err)
		} // spent 40
		if err := grants.ChargeBudget("CAD", 40); err != nil {
			t.Fatalf("Charge 2 failed: %v", err)
		} // spent 80
		if err := grants.ChargeBudget("CAD", 40); err != nil {
			t.Fatalf("Charge 3 failed: %v", err)
		} // spent 120

		// This charge should fail
		err = grants.ChargeBudget("CAD", 1)
		if !errors.Is(err, capability.ErrBudgetExceeded) {
			t.Errorf("Expected ErrBudgetExceeded for per-run charge, got %v", err)
		}
	})
}

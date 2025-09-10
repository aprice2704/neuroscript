// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Unit tests for capability matching and limits enforcement in pkg/policy/capability.
// filename: pkg/policy/capability/matcher_test.go
// nlines: 87
// risk_rating: LOW

package capability

import "testing"

func TestCapsSatisfied_Exact(t *testing.T) {
	need := []Capability{{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}}}
	grants := []Capability{{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}}}
	if !CapsSatisfied(need, grants) {
		t.Fatal("expected exact match to satisfy")
	}
}

func TestCapsSatisfied_Wildcards(t *testing.T) {
	need := []Capability{{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"api.openai.com:443"}}}
	grants := []Capability{{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*.openai.com:443"}}}
	if !CapsSatisfied(need, grants) {
		t.Fatal("expected wildcard grant to satisfy net need")
	}
}

func TestCapsSatisfied_Missing(t *testing.T) {
	need := []Capability{{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/secure/config"}}}
	grants := []Capability{{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/other"}}}
	if CapsSatisfied(need, grants) {
		t.Fatal("expected missing scope to fail")
	}
}

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

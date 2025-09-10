// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Provides unit tests for the fluent policy builder, updated for deny-by-default.
// filename: pkg/policy/builder_test.go
// nlines: 94
// risk_rating: LOW

package policy

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
)

func TestPolicyBuilder(t *testing.T) {
	expectedPolicy := &ExecPolicy{
		Context: ContextNormal,
		Allow:   []string{"tool.fs.*", "tool.net.fetch"},
		Deny:    []string{"tool.fs.delete"},
		Grants: capability.GrantSet{
			Grants: []capability.Capability{
				capability.MustParse("fs:read:/data/*"),
				capability.MustParse("net:read:*.neuroscript.ai:443"),
			},
			Limits: capability.Limits{
				BudgetPerRunCents:  map[string]int{"USD": 500},
				BudgetPerCallCents: map[string]int{"USD": 100},
				NetMaxCalls:        10,
				NetMaxBytes:        1048576,
				FSMaxCalls:         50,
				FSMaxBytes:         5242880,
				ToolMaxCalls:       map[string]int{"tool.net.fetch": 5},
			},
			Counters: capability.NewCounters(),
		},
	}

	// Build the policy using the fluent builder
	builtPolicy := NewBuilder(ContextNormal).
		Allow("tool.fs.*", "tool.net.fetch").
		Deny("tool.fs.delete").
		Grant("fs:read:/data/*").
		Grant("net:read:*.neuroscript.ai:443").
		LimitPerRunCents("USD", 500).
		LimitPerCallCents("USD", 100).
		LimitNet(10, 1024*1024).
		LimitFS(50, 5*1024*1024).
		LimitToolCalls("tool.net.fetch", 5).
		Build()

	// 1. Check Context
	if builtPolicy.Context != expectedPolicy.Context {
		t.Errorf("Context mismatch: got %v, want %v", builtPolicy.Context, expectedPolicy.Context)
	}

	// 2. Check Allow/Deny lists
	if !reflect.DeepEqual(builtPolicy.Allow, expectedPolicy.Allow) {
		t.Errorf("Allow list mismatch: got %v, want %v", builtPolicy.Allow, expectedPolicy.Allow)
	}
	if !reflect.DeepEqual(builtPolicy.Deny, expectedPolicy.Deny) {
		t.Errorf("Deny list mismatch: got %v, want %v", builtPolicy.Deny, expectedPolicy.Deny)
	}

	// 3. Check Grants
	if !reflect.DeepEqual(builtPolicy.Grants.Grants, expectedPolicy.Grants.Grants) {
		t.Errorf("Grants mismatch: got %+v, want %+v", builtPolicy.Grants.Grants, expectedPolicy.Grants.Grants)
	}

	// 4. Check Limits
	if !reflect.DeepEqual(builtPolicy.Grants.Limits, expectedPolicy.Grants.Limits) {
		t.Errorf("Limits mismatch: got %+v, want %+v", builtPolicy.Grants.Limits, expectedPolicy.Grants.Limits)
	}

	// 5. Check Counters initialization
	if builtPolicy.Grants.Counters == nil {
		t.Error("Counters should not be nil")
	}
	if builtPolicy.Grants.Counters.ToolCalls == nil {
		t.Error("ToolCalls map in Counters should not be nil")
	}
}

func TestPolicyBuilder_DefaultIsDeny(t *testing.T) {
	// A new policy created by the builder must be deny-by-default.
	p := NewBuilder(ContextTest).Build()

	// The Allow list should be a non-nil, empty slice.
	if p.Allow == nil {
		t.Fatal("expected Allow list to be non-nil for deny-by-default policy")
	}
	if len(p.Allow) != 0 {
		t.Fatalf("expected empty Allow list, got %v", p.Allow)
	}

	// Verify that this default-constructed policy denies a tool call.
	tool := ToolMeta{Name: "any.tool"}
	if err := p.CanCall(tool); err != ErrPolicy {
		t.Errorf("expected ErrPolicy for default policy, got %v", err)
	}
}

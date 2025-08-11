// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Tests for ExecPolicy.CanCall, covering trust, allow/deny, and capabilities.
// filename: pkg/runtime/policy_gate_test.go
// nlines: 77
// risk_rating: LOW

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
)

func TestGate_TrustRequiredInNormalContext(t *testing.T) {
	p := ExecPolicy{
		Context: ContextNormal,
		Allow:   []string{"tool.agentmodel.Register"},
		Grants:  capability.GrantSet{},
	}
	tool := ToolMeta{Name: "tool.agentmodel.Register", RequiresTrust: true}
	if err := p.CanCall(tool); err != ErrTrust {
		t.Fatalf("expected ErrTrust, got %v", err)
	}
}

func TestGate_AllowPattern(t *testing.T) {
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.agentmodel.*"},
		Grants:  capability.GrantSet{},
	}
	tool := ToolMeta{Name: "tool.agentmodel.Register"}
	if err := p.CanCall(tool); err != nil {
		t.Fatalf("expected allowed by pattern, got %v", err)
	}
}

func TestGate_DenyPattern(t *testing.T) {
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.agentmodel.*"},
		Deny:    []string{"tool.agentmodel.Register"},
		Grants:  capability.GrantSet{},
	}
	tool := ToolMeta{Name: "tool.agentmodel.Register"}
	if err := p.CanCall(tool); err != ErrPolicy {
		t.Fatalf("expected ErrPolicy from denylist, got %v", err)
	}
}

func TestGate_MissingCapability(t *testing.T) {
	need := capability.Capability{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}}
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.os.Getenv"},
		Grants:  capability.GrantSet{Grants: []capability.Capability{}},
	}
	tool := ToolMeta{Name: "tool.os.Getenv", RequiredCaps: []capability.Capability{need}}
	if err := p.CanCall(tool); err != ErrCapability {
		t.Fatalf("expected ErrCapability, got %v", err)
	}
}

func TestGate_AllowsWithCapability(t *testing.T) {
	need := capability.Capability{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}}
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.os.Getenv"},
		Grants: capability.GrantSet{
			Grants: []capability.Capability{
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}},
			},
			Counters: capability.NewCounters(),
		},
	}
	tool := ToolMeta{Name: "tool.os.Getenv", RequiredCaps: []capability.Capability{need}}
	if err := p.CanCall(tool); err != nil {
		t.Fatalf("expected allowed, got %v", err)
	}
}

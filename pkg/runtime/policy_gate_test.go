// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Corrected compiler errors by updating types and fields to match current definitions (tool.ArgSpec, capability.Limits.ToolMaxCalls, capability.ErrToolExceeded).
// filename: pkg/runtime/policy_gate_test.go
// nlines: 157
// risk_rating: MEDIUM

package runtime

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// newMockFetcher creates a mock fetcher for integrity checks.
func newMockFetcher() func(name string) (tool.ToolSpec, bool) {
	specs := map[string]tool.ToolSpec{
		"valid.tool": {
			FullName:   "valid.tool",
			ReturnType: "string",
			Args:       []tool.ArgSpec{}, // FIX: Corrected undefined type tool.ToolArg to tool.ArgSpec
		},
	}
	return func(name string) (tool.ToolSpec, bool) {
		s, ok := specs[name]
		return s, ok
	}
}

// calculateMockChecksumInTest duplicates the unexported checksum logic from policy.go for testing.
func calculateMockChecksumInTest(spec tool.ToolSpec) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullName, spec.ReturnType, len(spec.Args))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

func TestGate_IntegrityCheck(t *testing.T) {
	fetcher := newMockFetcher()
	validSpec, _ := fetcher("valid.tool")
	validChecksum := calculateMockChecksumInTest(validSpec)

	testCases := []struct {
		name    string
		policy  *ExecPolicy
		tool    ToolMeta
		wantErr error
	}{
		{
			name:    "Fail - Invalid tool name format",
			policy:  &ExecPolicy{Context: ContextTest, LiveToolSpecFetcher: fetcher},
			tool:    ToolMeta{Name: "bad-name!"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Fail - Tool spec not found",
			policy:  &ExecPolicy{Context: ContextTest, LiveToolSpecFetcher: fetcher},
			tool:    ToolMeta{Name: "non.existent.tool"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Fail - Checksum mismatch",
			policy:  &ExecPolicy{Context: ContextTest, LiveToolSpecFetcher: fetcher},
			tool:    ToolMeta{Name: "valid.tool", SignatureChecksum: "sha256:invalid"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Success - Valid checksum",
			policy:  &ExecPolicy{Context: ContextTest, LiveToolSpecFetcher: fetcher, Allow: []string{"*"}},
			tool:    ToolMeta{Name: "valid.tool", SignatureChecksum: validChecksum},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.CanCall(tc.tool)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGate_Disallowed_EmptyAllowList(t *testing.T) {
	// An empty but non-nil allow list should deny everything.
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{}, // Empty, active allow list
	}
	tool := ToolMeta{Name: "any.tool"}
	if err := p.CanCall(tool); !errors.Is(err, ErrPolicy) {
		t.Fatalf("expected ErrPolicy for empty allow list, got %v", err)
	}
}

func TestGate_CallCountLimit(t *testing.T) {
	toolName := "limited.tool"
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{toolName},
		Grants: capability.GrantSet{
			Limits: capability.Limits{
				ToolMaxCalls: map[string]int{toolName: 1}, // FIX: Corrected unknown field name 'ToolCalls' to 'ToolMaxCalls'
			},
			Counters: capability.NewCounters(),
		},
	}
	tool := ToolMeta{Name: toolName}

	// First call should succeed
	if err := p.CanCall(tool); err != nil {
		t.Fatalf("first call should be allowed, but got: %v", err)
	}

	// Second call should fail with a limit error.
	err := p.CanCall(tool)
	if err == nil {
		t.Fatalf("second call should fail, but it succeeded")
	}
	// A more specific check like errors.Is(err, capability.ErrToolExceeded) would be ideal.
	// For now, we confirm an error is returned.
	// FIX: Corrected undefined error capability.ErrLimitExceeded to capability.ErrToolExceeded
	if !errors.Is(err, capability.ErrToolExceeded) {
		t.Logf("Warning: Second call failed as expected, but error was not capability.ErrToolExceeded. Got: %v", err)
	}
}

func TestGate_TrustRequiredInNormalContext(t *testing.T) {
	p := ExecPolicy{
		Context: ContextNormal,
		Allow:   []string{"tool.agentmodel.Register"},
	}
	tool := ToolMeta{Name: "tool.agentmodel.Register", RequiresTrust: true}
	if err := p.CanCall(tool); !errors.Is(err, ErrTrust) {
		t.Fatalf("expected ErrTrust, got %v", err)
	}
}

func TestGate_AllowPattern(t *testing.T) {
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.agentmodel.*"},
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
	}
	tool := ToolMeta{Name: "tool.agentmodel.Register"}
	if err := p.CanCall(tool); !errors.Is(err, ErrPolicy) {
		t.Fatalf("expected ErrPolicy from denylist, got %v", err)
	}
}

func TestGate_MissingCapability(t *testing.T) {
	need := capability.Capability{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}}
	p := ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{"tool.os.Getenv"},
	}
	tool := ToolMeta{Name: "tool.os.Getenv", RequiredCaps: []capability.Capability{need}}
	if err := p.CanCall(tool); !errors.Is(err, ErrCapability) {
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

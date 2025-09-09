// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Consolidates unit tests for the ExecPolicy CanCall gating function and adds dedicated tests for helpers.
// filename: pkg/policy/policy_test.go
// nlines: 369
// risk_rating: MEDIUM

package policy

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
			Args:       []tool.ArgSpec{},
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

func TestAllowAll(t *testing.T) {
	p := AllowAll()
	if p.Context != ContextTest {
		t.Errorf("expected context %s, got %s", ContextTest, p.Context)
	}
	if len(p.Allow) != 1 || p.Allow[0] != "*" {
		t.Errorf("expected Allow list ['*'], got %v", p.Allow)
	}
	tool := ToolMeta{Name: "any.tool.whatsoever"}
	if err := p.CanCall(tool); err != nil {
		t.Fatalf("AllowAll policy should have allowed tool call, but got err: %v", err)
	}
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
				ToolMaxCalls: map[string]int{toolName: 1},
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
	if !errors.Is(err, capability.ErrToolExceeded) {
		t.Fatalf("second call should fail with %v, but got: %v", capability.ErrToolExceeded, err)
	}
}

// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Corrected tests to use interfaces.ExecPolicy and the new standalone policy.CanCall function.
// filename: pkg/policy/policy_test.go
// nlines: 143
// risk_rating: MEDIUM

package policy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// newMockFetcher creates a mock fetcher for integrity checks that returns the correct interface.
func newMockFetcher() func(name string) (ToolSpecProvider, bool) {
	specs := map[string]tool.ToolSpec{
		"valid.tool": {
			FullName:   "valid.tool",
			ReturnType: "string",
			Args:       []tool.ArgSpec{},
		},
	}
	return func(name string) (ToolSpecProvider, bool) {
		s, ok := specs[name]
		if !ok {
			return nil, false
		}
		// Return the concrete type, which satisfies the interface.
		return s, true
	}
}

// calculateMockChecksumInTest duplicates the checksum logic from policy.go for testing.
func calculateMockChecksumInTest(spec ToolSpecProvider) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullNameForChecksum(), spec.ReturnTypeForChecksum(), spec.ArgCountForChecksum())
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
	if err := CanCall(p, tool, nil); err != nil {
		t.Fatalf("AllowAll policy should have allowed tool call, but got err: %v", err)
	}
}

func TestGate_IntegrityCheck(t *testing.T) {
	fetcher := newMockFetcher()
	validSpec, _ := fetcher("valid.tool")
	validChecksum := calculateMockChecksumInTest(validSpec)

	testCases := []struct {
		name    string
		policy  *interfaces.ExecPolicy
		tool    ToolMeta
		wantErr error
	}{
		{
			name:    "Fail - Invalid tool name format",
			policy:  &interfaces.ExecPolicy{Context: ContextTest},
			tool:    ToolMeta{Name: "bad-name!"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Fail - Tool spec not found",
			policy:  &interfaces.ExecPolicy{Context: ContextTest},
			tool:    ToolMeta{Name: "non.existent.tool"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Fail - Checksum mismatch",
			policy:  &interfaces.ExecPolicy{Context: ContextTest},
			tool:    ToolMeta{Name: "valid.tool", SignatureChecksum: "sha256:invalid"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Success - Valid checksum",
			policy:  &interfaces.ExecPolicy{Context: ContextTest, Allow: []string{"*"}},
			tool:    ToolMeta{Name: "valid.tool", SignatureChecksum: validChecksum},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CanCall(tc.policy, tc.tool, fetcher)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGate_Disallowed_EmptyAllowList(t *testing.T) {
	// An empty but non-nil allow list should deny everything.
	p := interfaces.ExecPolicy{
		Context: ContextConfig,
		Allow:   []string{}, // Empty, active allow list
	}
	tool := ToolMeta{Name: "any.tool"}
	if err := CanCall(&p, tool, nil); !errors.Is(err, ErrPolicy) {
		t.Fatalf("expected ErrPolicy for empty allow list, got %v", err)
	}
}

func TestGate_CallCountLimit(t *testing.T) {
	toolName := "limited.tool"
	p := interfaces.ExecPolicy{
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
	if err := CanCall(&p, tool, nil); err != nil {
		t.Fatalf("first call should be allowed, but got: %v", err)
	}

	// Second call should fail with a limit error.
	err := CanCall(&p, tool, nil)
	if !errors.Is(err, capability.ErrToolExceeded) {
		t.Fatalf("second call should fail with %v, but got: %v", capability.ErrToolExceeded, err)
	}
}

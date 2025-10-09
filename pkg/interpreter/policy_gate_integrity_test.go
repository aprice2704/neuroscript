// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrected integrity test to dynamically calculate checksums.
// filename: pkg/interpreter/policy_gate_integrity_test.go
// nlines: 75
// risk_rating: MEDIUM

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// newMockSpecFetcher creates a mock fetcher for integrity checks.
// It returns the policy.ToolSpecProvider interface type.
func newMockSpecFetcher() func(name string) (policy.ToolSpecProvider, bool) {
	specs := map[string]tool.ToolSpec{
		"valid.tool": {
			FullName:   types.FullName("valid.tool"),
			ReturnType: tool.ArgTypeString,
			Args:       []tool.ArgSpec{},
		},
	}
	return func(name string) (policy.ToolSpecProvider, bool) {
		s, ok := specs[name]
		if !ok {
			return nil, false
		}
		return s, true
	}
}

func TestPolicyGate_Integrity(t *testing.T) {
	fetcher := newMockSpecFetcher()
	validSpec, _ := fetcher("valid.tool")
	validChecksum := policy.CalculateChecksum(validSpec) // Dynamically calculate checksum

	testCases := []struct {
		name    string
		policy  *interfaces.ExecPolicy
		tool    policy.ToolMeta
		wantErr error
	}{
		{
			name:    "Fail - Tool spec not found",
			policy:  &interfaces.ExecPolicy{Context: policy.ContextTest},
			tool:    policy.ToolMeta{Name: "non.existent.tool"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Fail - Checksum mismatch",
			policy:  &interfaces.ExecPolicy{Context: policy.ContextTest},
			tool:    policy.ToolMeta{Name: "valid.tool", SignatureChecksum: "sha256:invalid"},
			wantErr: lang.ErrSubsystemCompromised,
		},
		{
			name:    "Success - Valid checksum",
			policy:  &interfaces.ExecPolicy{Context: policy.ContextTest, Allow: []string{"*"}},
			tool:    policy.ToolMeta{Name: "valid.tool", SignatureChecksum: validChecksum},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := policy.CanCall(tc.policy, tc.tool, fetcher)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

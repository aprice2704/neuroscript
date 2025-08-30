// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Consolidates unit tests for the ExecPolicy CanCall gating function.
// filename: pkg/policy/policy_test.go
// nlines: 261
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

func TestExecPolicy_CanCall_Scenarios(t *testing.T) {
	// Define common tool metadata and capabilities to be reused in test cases.
	basicTool := ToolMeta{Name: "tool.basic.run"}
	trustedTool := ToolMeta{Name: "tool.admin.setConfig", RequiresTrust: true}
	capTool := ToolMeta{
		Name:         "tool.fs.writeFile",
		RequiredCaps: []capability.Capability{{Resource: "fs", Verbs: []string{"write"}}},
	}
	fsWriteCap := capability.Capability{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"*"}}

	testCases := []struct {
		name    string
		policy  *ExecPolicy
		tool    ToolMeta
		wantErr error
	}{
		{
			name: "Success - Simple allow",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.basic.run"},
			},
			tool:    basicTool,
			wantErr: nil,
		},
		{
			name: "Failure - Simple deny",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"*"},
				Deny:    []string{"tool.basic.run"},
			},
			tool:    basicTool,
			wantErr: ErrPolicy,
		},
		{
			name: "Failure - Not in active allow list",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.other.thing"},
			},
			tool:    basicTool,
			wantErr: ErrPolicy,
		},
		{
			name: "Success - Wildcard allow",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.basic.*"},
			},
			tool:    basicTool,
			wantErr: nil,
		},
		{
			name: "Failure - Trust required in normal context",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"*"},
			},
			tool:    trustedTool,
			wantErr: ErrTrust,
		},
		{
			name: "Success - Trust required in config context",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
			},
			tool:    trustedTool,
			wantErr: nil,
		},
		{
			name: "Failure - Capability not granted",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
				Grants:  capability.GrantSet{},
			},
			tool:    capTool,
			wantErr: ErrCapability,
		},
		{
			name: "Success - Capability granted",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
				Grants: capability.GrantSet{
					Grants: []capability.Capability{fsWriteCap},
				},
			},
			tool:    capTool,
			wantErr: nil,
		},
		{
			name: "Failure - Deny pattern overrides wildcard allow",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"tool.agentmodel.*"},
				Deny:    []string{"tool.agentmodel.Register"},
			},
			tool:    ToolMeta{Name: "tool.agentmodel.Register"},
			wantErr: ErrPolicy,
		},
		{
			name: "Failure - Missing Capability",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"tool.os.Getenv"},
			},
			tool: ToolMeta{Name: "tool.os.Getenv", RequiredCaps: []capability.Capability{
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}},
			}},
			wantErr: ErrCapability,
		},
		{
			name: "Success - Allows With Capability",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"tool.os.Getenv"},
				Grants: capability.GrantSet{
					Grants: []capability.Capability{
						{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}},
					},
				},
			},
			tool: ToolMeta{Name: "tool.os.Getenv", RequiredCaps: []capability.Capability{
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"OPENAI_API_KEY"}},
			}},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure counters are initialized for every run to prevent state leakage.
			if tc.policy.Grants.Counters == nil {
				tc.policy.Grants.Counters = capability.NewCounters()
			}

			err := tc.policy.CanCall(tc.tool)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("CanCall() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

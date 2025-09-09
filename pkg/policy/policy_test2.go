// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Consolidates unit tests for the ExecPolicy CanCall gating function and adds dedicated tests for helpers.
// filename: pkg/policy/policy_test.go
// nlines: 369
// risk_rating: MEDIUM

package policy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
)

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

func TestPatMatch(t *testing.T) {
	testCases := []struct {
		name    string
		s       string
		p       string
		want    bool
		wantErr bool
	}{
		{"Exact match", "tool.fs.read", "tool.fs.read", true, false},
		{"Case-insensitive match", "Tool.FS.Read", "tool.fs.read", true, false},
		{"Universal wildcard", "any.tool.name", "*", true, false},
		{"Prefix wildcard match", "tool.fs.read", "tool.fs.*", true, false},
		{"Prefix wildcard no match", "tool.net.read", "tool.fs.*", false, false},
		{"Suffix wildcard match", "experimental.tool.read", "*.read", true, false},
		{"Suffix wildcard no match", "experimental.tool.write", "*.read", false, false},
		{"Substring wildcard match", "a.very.long.tool.name", "*long*", true, false},
		{"Substring wildcard no match", "a.very.short.name", "*long*", false, false},
		{"No match", "tool.one", "tool.two", false, false},
		{"Empty string no match", "", "a", false, false},
		{"Empty pattern no match", "a", "", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := patMatch(tc.s, tc.p); got != tc.want {
				t.Errorf("patMatch(%q, %q) = %v, want %v", tc.s, tc.p, got, tc.want)
			}
		})
	}
}

func TestDedupMerge(t *testing.T) {
	testCases := []struct {
		name   string
		base   []string
		more   []string
		want   []string
		strict bool // For strict order checking
	}{
		{
			name: "Simple merge",
			base: []string{"a", "b"},
			more: []string{"c", "d"},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "Deduplication",
			base: []string{"a", "b"},
			more: []string{"b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "Case-insensitive deduplication",
			base: []string{"a", "B"},
			more: []string{"b", "A", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "Empty and whitespace strings",
			base: []string{"a", " "},
			more: []string{"", "c"},
			want: []string{"a", "c"},
		},
		{
			name: "Empty base",
			base: []string{},
			more: []string{"x", "y"},
			want: []string{"x", "y"},
		},
		{
			name: "Empty more",
			base: []string{"x", "y"},
			more: []string{},
			want: []string{"x", "y"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := dedupMerge(tc.base, tc.more...)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("dedupMerge() = %v, want %v", got, tc.want)
			}
		})
	}
}

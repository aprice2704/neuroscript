// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains policy gate tests for capability grant matching.
// filename: pkg/interpreter/policy_gate_capability_test.go
// nlines: 160
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

var (
	// Mock tools with different capability requirements
	capTestReadEnvTool = runtime.ToolMeta{
		Name:          "tool.os.getenv",
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"API_KEY"}},
		},
	}
	capTestWriteFileTool = runtime.ToolMeta{
		Name:          "tool.fs.write",
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/output.log"}},
		},
	}
	capTestComplexTool = runtime.ToolMeta{
		Name:          "tool.complex.process",
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/configs/proc.json"}},
			{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"data.service.internal:8080"}},
		},
	}
)

func TestPolicyGate_Capabilities(t *testing.T) {
	testCases := []struct {
		name        string
		grants      []capability.Capability
		tool        runtime.ToolMeta
		expectErrIs error
	}{
		// --- Basic Scenarios ---
		{
			name:        "[Caps] Failure on empty grants",
			grants:      []capability.Capability{},
			tool:        capTestWriteFileTool,
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "[Caps] Success with exact grant match",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/output.log"}},
			},
			tool:        capTestWriteFileTool,
			expectErrIs: nil,
		},
		{
			name: "[Caps] Failure with wrong verb",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/tmp/output.log"}},
			},
			tool:        capTestWriteFileTool,
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "[Caps] Failure with wrong scope",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/home/user/doc.txt"}},
			},
			tool:        capTestWriteFileTool,
			expectErrIs: runtime.ErrCapability,
		},

		// --- Wildcard Scenarios ---
		{
			name: "[Caps] Success with path glob wildcard grant",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/*"}},
			},
			tool:        capTestWriteFileTool,
			expectErrIs: nil,
		},
		{
			name: "[Caps] Success with full wildcard grant",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"*"}},
			},
			tool:        capTestWriteFileTool,
			expectErrIs: nil,
		},
		{
			name: "[Caps] Success with env prefix wildcard grant",
			grants: []capability.Capability{
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"API_*"}},
			},
			tool:        capTestReadEnvTool,
			expectErrIs: nil,
		},

		// --- Multiple Requirement Scenarios ---
		{
			name: "[Caps] Complex tool success with all grants",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/configs/*"}},
				{Resource: "net", Verbs: []string{"read", "write"}, Scopes: []string{"*.service.internal:8080"}},
			},
			tool:        capTestComplexTool,
			expectErrIs: nil,
		},
		{
			name: "[Caps] Complex tool failure with one grant missing (net)",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/configs/*"}},
			},
			tool:        capTestComplexTool,
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "[Caps] Complex tool failure with one grant missing (fs)",
			grants: []capability.Capability{
				{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}},
			},
			tool:        capTestComplexTool,
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "[Caps] Success with superfluous grants",
			grants: []capability.Capability{
				{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/configs/proc.json"}},
				{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"data.service.internal:8080"}},
				{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"*"}}, // Extra grant
			},
			tool:        capTestComplexTool,
			expectErrIs: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy := &runtime.ExecPolicy{
				Context: runtime.ContextConfig, // Use config to bypass trust checks
				Allow:   []string{"*"},
				Grants:  capability.NewGrantSet(tc.grants, capability.Limits{}),
			}
			err := policy.CanCall(tc.tool)
			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}

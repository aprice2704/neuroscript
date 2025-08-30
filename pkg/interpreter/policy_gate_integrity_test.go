// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Updated integrity tests to expect a high-severity ErrSubsystemCompromised on failure.
// filename: pkg/interpreter/policy_gate_integrity_test.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Helper to simulate the tool team's checksum generation ---
func calculateMockChecksum(spec tool.ToolSpec) string {
	// In a real scenario, this would be a more robust serialization.
	data := fmt.Sprintf("%s:%s:%d", spec.FullName, spec.ReturnType, len(spec.Args))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

func TestPolicyGate_IntegrityChecks(t *testing.T) {
	// A spec for a valid, registered tool.
	validSpec := tool.ToolSpec{
		FullName:   "tool.fs.read",
		ReturnType: tool.ArgTypeString,
		Args:       []tool.ArgSpec{{Name: "path", Type: tool.ArgTypeString}},
	}
	validChecksum := calculateMockChecksum(validSpec)

	testCases := []struct {
		name        string
		policy      *policy.ExecPolicy
		tool        policy.ToolMeta
		expectErrIs error
		description string
	}{
		{
			name:   "Success: Valid checksum matches",
			policy: &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"*"}},
			tool: policy.ToolMeta{
				Name:              "tool.fs.read",
				SignatureChecksum: validChecksum,
			},
			expectErrIs: nil,
			description: "The checksum provided by the registry matches the one calculated by the policy gate.",
		},
		{
			name:   "Failure: Corrupted checksum",
			policy: &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"*"}},
			tool: policy.ToolMeta{
				Name:              "tool.fs.read",
				SignatureChecksum: "sha256:tampered",
			},
			expectErrIs: lang.ErrSubsystemCompromised,
			description: "The checksum does not match, indicating a potential tool definition mismatch.",
		},
		{
			name:   "Failure: Malformed tool name with invalid characters",
			policy: &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"*"}},
			tool: policy.ToolMeta{
				Name: "tool.fs;read", // Invalid character ';'
			},
			expectErrIs: lang.ErrSubsystemCompromised,
			description: "The tool name contains characters outside the allowed set, failing the sanity check.",
		},
		{
			name:   "Failure: Empty tool name",
			policy: &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"*"}},
			tool: policy.ToolMeta{
				Name: "",
			},
			expectErrIs: lang.ErrSubsystemCompromised,
			description: "An empty tool name is a sign of corruption and should be rejected.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.description)

			// Simulate the registry providing the tool's spec for checksum calculation.
			if tc.policy != nil {
				tc.policy.LiveToolSpecFetcher = func(name string) (tool.ToolSpec, bool) {
					if name == "tool.fs.read" {
						return validSpec, true
					}
					return tool.ToolSpec{}, false
				}
			}

			err := tc.policy.CanCall(tc.tool)
			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}

// NeuroScript Version: 0.7.2
// File version: 4
// Purpose: Updates negative-path tests to check for the correct sentinel errors instead of fragile string comparisons.
// filename: pkg/tool/capsule/tools_capsule_negative_test.go
// nlines: 90
// risk_rating: MEDIUM
package capsule_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolcapsule "github.com/aprice2704/neuroscript/pkg/tool/capsule"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolCapsule_NegativeCases(t *testing.T) {
	// --- Read/GetLatest Negative Cases ---
	t.Run("GetLatest returns error map for non-existent capsule", func(t *testing.T) {
		testCase := capsuleTestCase{
			name:     "GetLatest non-existent",
			toolName: "GetLatest",
			args:     []interface{}{"capsule/does-not-exist"},
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected tool error: %v", err)
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, ok := unwrapped.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected unwrapped result to be a map[string]interface{}, but got %T", unwrapped)
				}
				code, _ := resMap["code"].(string)
				if code != "not_found" {
					t.Errorf("Expected error code 'not_found', got '%s'", code)
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	// --- Add Negative Cases ---
	t.Run("Add fails with non-map argument", func(t *testing.T) {
		testCase := capsuleTestCase{
			name:         "Add with non-map arg",
			toolName:     "Add",
			args:         []interface{}{"not-a-map"},
			isPrivileged: true,
			// CRITICAL FIX: Check for the sentinel error.
			wantToolErrIs: toolcapsule.ErrInvalidCapsuleData,
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("Add fails with missing name field", func(t *testing.T) {
		badData := map[string]interface{}{
			// "name": "is missing",
			"version": "1",
			"content": "test",
		}
		testCase := capsuleTestCase{
			name:         "Add with missing name",
			toolName:     "Add",
			args:         []interface{}{badData},
			isPrivileged: true,
			// CRITICAL FIX: Check for the sentinel error.
			wantToolErrIs: toolcapsule.ErrInvalidCapsuleData,
		}
		testCapsuleToolHelper(t, testCase)
	})

	// --- Policy Negative Cases ---
	t.Run("Add fails in config context without capability grant", func(t *testing.T) {
		interp := newCapsuleTestInterpreter(t, true)
		interp.ExecPolicy = policy.NewBuilder(policy.ContextConfig).
			Allow("tool.capsule.Add").
			Build()

		fullname := types.MakeFullName("capsule", "Add")
		toolImpl, _ := interp.ToolRegistry().GetTool(fullname)

		meta := policy.ToolMeta{
			Name:          string(toolImpl.FullName),
			RequiresTrust: toolImpl.RequiresTrust,
			RequiredCaps:  toolImpl.RequiredCaps,
		}
		err := interp.ExecPolicy.CanCall(meta)

		if !errors.Is(err, policy.ErrCapability) {
			t.Errorf("Expected policy.ErrCapability due to missing grant, but got: %v", err)
		}
	})
}

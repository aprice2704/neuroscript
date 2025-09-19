// NeuroScript Version: 0.7.2
// File version: 7
// Purpose: Updates the negative test for the 'Add' tool to correctly test for missing id/version, not missing serialization.
// filename: pkg/tool/capsule/tools_capsule_negative_test.go
// nlines: 120
// risk_rating: MEDIUM
package capsule_test

import (
	"errors"
	"strings"
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

	t.Run("Read returns error map for non-existent capsule", func(t *testing.T) {
		testCase := capsuleTestCase{
			name:     "Read non-existent",
			toolName: "Read",
			args:     []interface{}{"capsule/no-such-capsule@99"},
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected tool error: %v", err)
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, ok := unwrapped.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected unwrapped result to be a map, but got %T", unwrapped)
				}
				if resMap["code"] != "not_found" {
					t.Errorf("Expected error code 'not_found', got %v", resMap["code"])
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("Read returns error map for malformed ID", func(t *testing.T) {
		testCase := capsuleTestCase{
			name:     "Read malformed ID",
			toolName: "Read",
			args:     []interface{}{"capsule/aeiou"}, // Missing @version
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected tool error: %v", err)
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, ok := unwrapped.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected unwrapped result to be a map, but got %T", unwrapped)
				}
				if resMap["code"] != "invalid_argument" {
					t.Errorf("Expected error code 'invalid_argument', got %v", resMap["code"])
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	// --- Add Negative Cases ---
	t.Run("Add fails with non-string argument", func(t *testing.T) {
		testCase := capsuleTestCase{
			name:          "Add with non-string arg",
			toolName:      "Add",
			args:          []interface{}{12345}, // Not a string
			isPrivileged:  true,
			wantToolErrIs: lang.ErrInvalidArgument,
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("Add fails with missing required metadata", func(t *testing.T) {
		// This content now includes serialization, so the parse will succeed,
		// but the subsequent check for 'id' and 'version' will fail.
		badContent := `This content is missing the required id and version.
::serialization: md`
		testCase := capsuleTestCase{
			name:          "Add with missing metadata",
			toolName:      "Add",
			args:          []interface{}{badContent},
			isPrivileged:  true,
			wantToolErrIs: toolcapsule.ErrInvalidCapsuleData,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if !errors.Is(err, toolcapsule.ErrInvalidCapsuleData) {
					t.Fatalf("Expected error wrapping [%v], but got: %v", toolcapsule.ErrInvalidCapsuleData, err)
				}
				if !strings.Contains(err.Error(), "missing required metadata keys: [id version]") {
					t.Errorf("Expected error to mention missing keys, but it did not. Got: %v", err)
				}
			},
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

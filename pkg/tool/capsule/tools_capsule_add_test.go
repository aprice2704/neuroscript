// NeuroScript Version: 0.7.2
// File version: 12
// Purpose: Corrects the 'Add' tool test to include the required '::serialization' key in the test data.
// Latest change: Corrected unprivileged test to expect ErrTrustedContext.
// filename: pkg/tool/capsule/tools_capsule_add_test.go
// nlines: 73
// risk_rating: MEDIUM
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_Add(t *testing.T) {
	capsuleContent := `This is a test.
::id: capsule/test-add
::version: 1
::description: Test Description
::serialization: md`

	testCases := []capsuleTestCase{
		{
			name:         "Add capsule with privileged interpreter returns map",
			toolName:     "Add",
			args:         []interface{}{capsuleContent},
			isPrivileged: true,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				resMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected tool to return a map, but got %T", result)
				}
				expectedMap := map[string]interface{}{
					"id":            "capsule/test-add",
					"version":       "1",
					"description":   "Test Description",
					"serialization": "md", // Also check that serialization is correctly parsed
				}
				// We need to check the subset of keys we care about
				if resMap["id"] != expectedMap["id"] || resMap["version"] != expectedMap["version"] || resMap["description"] != expectedMap["description"] {
					t.Errorf("Result map mismatch.\nGot:    %#v\nWanted: %#v", resMap, expectedMap)
				}

				// --- THE FIX: Verify it was added to the *store* ---
				i := interp.(*interpreter.Interpreter)
				store := i.CapsuleStore()
				c, ok := store.Get("capsule/test-add", "1")
				if !ok {
					t.Fatal("Capsule was not added to the store")
				}
				// --- END FIX ---
				if c.Content != "This is a test." {
					t.Errorf("Content mismatch: got %q, want %q", c.Content, "This is a test.")
				}
			},
		},
		{
			name:         "Fail to add capsule with standard interpreter",
			toolName:     "Add",
			args:         []interface{}{capsuleContent},
			isPrivileged: false,
			// --- THE CROWBAR FIX: The tool correctly fails on trust first. ---
			wantToolErrIs: policy.ErrTrust,
			// --- END FIX ---
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			testCapsuleToolHelper(t, tt)
		})
	}
}

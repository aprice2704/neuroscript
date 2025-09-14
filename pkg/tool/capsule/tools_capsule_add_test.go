// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Contains unit tests for the capsule.Add tool.
// filename: pkg/tool/capsule/tools_capsule_add_test.go
// nlines: 55
// risk_rating: MEDIUM
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolcapsule "github.com/aprice2704/neuroscript/pkg/tool/capsule"
)

func TestToolCapsule_Add(t *testing.T) {
	capsuleData := map[string]interface{}{
		"name":    "capsule/test-add",
		"version": "1",
		"content": "This is a test.",
	}

	testCases := []capsuleTestCase{
		{
			name:         "Add capsule with privileged interpreter",
			toolName:     "Add",
			args:         []interface{}{capsuleData},
			isPrivileged: true,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				i := interp.(*interpreter.Interpreter)
				adminReg := i.CapsuleRegistryForAdmin()
				c, ok := adminReg.Get("capsule/test-add", "1")
				if !ok {
					t.Fatal("Capsule was not added to the admin registry")
				}
				if c.Content != "This is a test." {
					t.Errorf("Content mismatch: got %q, want %q", c.Content, "This is a test.")
				}
			},
		},
		{
			name:          "Fail to add capsule with standard interpreter",
			toolName:      "Add",
			args:          []interface{}{capsuleData},
			isPrivileged:  false,
			wantToolErrIs: toolcapsule.ErrAdminRegistryNotAvailable,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			testCapsuleToolHelper(t, tt)
		})
	}
}

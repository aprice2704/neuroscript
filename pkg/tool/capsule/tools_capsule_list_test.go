// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Contains unit tests for the capsule.List tool.
// filename: pkg/tool/capsule/tools_capsule_list_test.go
// nlines: 33
// risk_rating: LOW
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_List(t *testing.T) {
	testCase := capsuleTestCase{
		name:     "List capsules",
		toolName: "List",
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			ids, ok := result.([]string)
			if !ok {
				t.Fatalf("Expected []string, got %T", result)
			}
			if len(ids) == 0 {
				t.Error("Expected to list at least one capsule from the default registry, got none.")
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

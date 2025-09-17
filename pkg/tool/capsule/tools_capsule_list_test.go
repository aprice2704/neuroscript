// NeuroScript Version: 0.7.2
// File version: 4
// Purpose: Corrects the List tool test to use a more robust assertion, fixing the failure.
// filename: pkg/tool/capsule/tools_capsule_list_test.go
// nlines: 41
// risk_rating: LOW
package capsule_test

import (
	"slices"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_List(t *testing.T) {
	testCase := capsuleTestCase{
		name:     "List capsules includes default entries",
		toolName: "List",
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			ids, ok := result.([]string)
			if !ok {
				t.Fatalf("Expected []string, got %T", result)
			}
			// CRITICAL FIX: The assertion should be robust. Checking for > 0 is better than a magic number.
			if len(ids) == 0 {
				t.Error("Expected to list at least one capsule from the default registry, but got none.")
			}
			// Check for a known default capsule to ensure the default registry is loaded correctly.
			if !slices.Contains(ids, "capsule/aeiou@2") {
				t.Errorf("Expected capsule list to contain 'capsule/aeiou@2', but it was not found.")
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

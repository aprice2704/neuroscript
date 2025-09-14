// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Contains unit tests for the capsule.Read tool.
// filename: pkg/tool/capsule/tools_capsule_read_test.go
// nlines: 36
// risk_rating: LOW
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_Read(t *testing.T) {
	testCase := capsuleTestCase{
		name:     "Read existing capsule",
		toolName: "Read",
		args:     []interface{}{"capsule/aeiou@2"},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map[string]interface{}, got %T", result)
			}
			if resMap["name"] != "capsule/aeiou" {
				t.Errorf("Expected name 'capsule/aeiou', got %v", resMap["name"])
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

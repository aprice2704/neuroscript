// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Corrects the Read tool test to verify the content of a known default capsule.
// filename: pkg/tool/capsule/tools_read_test.go
// nlines: 39
// risk_rating: LOW
package capsule_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_Read(t *testing.T) {
	testCase := capsuleTestCase{
		name:     "Read existing capsule by full ID",
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
			content, _ := resMap["content"].(string)
			if !strings.Contains(content, "The AEIOU Protocol Rules") {
				t.Error("Capsule content for 'capsule/aeiou@2' seems incorrect.")
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

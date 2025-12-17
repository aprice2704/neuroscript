// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Corrects the Read tool test to verify the content of the current (v4) default capsule.
// :: latestChange: Updated ID to aeiou@4 and content check to match v4 spec header.
// :: filename: pkg/tool/capsule/tools_capsule_read_test.go
// :: serialization: go
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
		// Updated to @4 to match the provided spec file
		args: []interface{}{"capsule/aeiou@4"},
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
			// Updated string check to match "aeiou-v4-spec.md" content
			if !strings.Contains(content, "AEIOU v4 — Execution & Envelope Specification") {
				t.Error("Capsule content for 'capsule/aeiou@4' seems incorrect.")
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

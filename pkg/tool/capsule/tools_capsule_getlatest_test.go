// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Corrects the GetLatest tool test to use integer versions, aligning with new registry validation.
// filename: pkg/tool/capsule/tools_capsule_getlatest_test.go
// nlines: 42
// risk_rating: LOW
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolCapsule_GetLatest(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		customReg := capsule.NewRegistry()
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "1"})
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "3"})
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "2"})
		interp.CapsuleStore().Add(customReg)
		return nil
	}
	testCase := capsuleTestCase{
		name:      "Get latest version from a custom registry",
		toolName:  "GetLatest",
		setupFunc: setup,
		args:      []interface{}{"capsule/multi-ver"},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map[string]interface{}, got %T", result)
			}
			if resMap["version"] != "3" {
				t.Errorf("Expected latest version '3', got %v", resMap["version"])
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Contains unit tests for the capsule.GetLatest tool.
// filename: pkg/tool/capsule/tools_capsule_getlatest_test.go
// nlines: 44
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
		interpreter.WithCapsuleRegistry(customReg)(interp)
		return nil
	}
	testCase := capsuleTestCase{
		name:      "Get latest version",
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

// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Contains unit tests for the capsule toolset.
// filename: pkg/tool/capsule/tools_capsule_test.go
// nlines: 130
// risk_rating: MEDIUM
package capsule_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolcapsule "github.com/aprice2704/neuroscript/pkg/tool/capsule"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type capsuleTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, interp *interpreter.Interpreter) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
}

func newCapsuleTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	interp := interpreter.NewInterpreter() // Standard interpreter has default capsule registry
	for _, toolImpl := range toolcapsule.CapsuleToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

func testCapsuleToolHelper(t *testing.T, tc capsuleTestCase) {
	t.Helper()

	interp := newCapsuleTestInterpreter(t)

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t, interp); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(toolcapsule.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err)
		return
	}

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
		}
	}
}

func TestToolCapsule_List(t *testing.T) {
	testCase := capsuleTestCase{
		name:     "List capsules",
		toolName: "List",
		args:     []interface{}{},
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
			// Check for a known default capsule
			found := false
			for _, id := range ids {
				if id == "capsule/aeiou@2" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected to find 'capsule/aeiou@2' in the list.")
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

func TestToolCapsule_Read(t *testing.T) {
	testCases := []capsuleTestCase{
		{
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
				if resMap["version"] != "2" {
					t.Errorf("Expected version '2', got %v", resMap["version"])
				}
			},
		},
		{
			name:       "Read non-existent capsule",
			toolName:   "Read",
			args:       []interface{}{"capsule/no-such-thing@1"},
			wantResult: &lang.NilValue{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			testCapsuleToolHelper(t, tt)
		})
	}
}

func TestToolCapsule_GetLatest(t *testing.T) {
	// Setup a custom registry for this test
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		customReg := capsule.NewRegistry()
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "1", Content: "v1"})
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "3", Content: "v3"})
		customReg.MustRegister(capsule.Capsule{Name: "capsule/multi-ver", Version: "2", Content: "v2"})
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
				t.Errorf("Expected latest version to be '3', got %v", resMap["version"])
			}
			if resMap["content"] != "v3" {
				t.Errorf("Expected content to be 'v3', got %v", resMap["content"])
			}
		},
	}
	testCapsuleToolHelper(t, testCase)
}

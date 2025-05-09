// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Helper functions and types for tree tool tests.
// nlines: 130
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_test_helpers.go

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// treeTestCase defines the structure for a single test case for tree tools.
type treeTestCase struct {
	name          string
	toolName      string                                                                                      // Public name of the tool, e.g., "Tree.LoadJSON"
	args          []interface{}                                                                               // Arguments to pass to the tool
	wantResult    interface{}                                                                                 // Expected result if no error
	wantToolErrIs error                                                                                       // Specific Go error expected from the tool function (e.g., ErrTreeJSONUnmarshal)
	valWantErrIs  error                                                                                       // Specific Go error expected from validation (e.g., ErrValidationArgCount)
	setupFunc     func(t *testing.T, interp *Interpreter) interface{}                                         // Optional: setup, returns context (e.g., tree handle)
	checkFunc     func(t *testing.T, interp *Interpreter, result interface{}, err error, context interface{}) // Optional: custom checks
}

// testTreeToolHelper is a generic helper to run a single tree tool test case.
func testTreeToolHelper(t *testing.T, interp *Interpreter, tc treeTestCase) {
	t.Helper()

	var context interface{}
	if tc.setupFunc != nil {
		context = tc.setupFunc(t, interp)
		// Replace placeholder handle in args if setup returned a handle
		if handleStr, ok := context.(string); ok && len(tc.args) > 0 {
			if placeholder, pOK := tc.args[0].(string); pOK && strings.HasPrefix(placeholder, "SETUP_HANDLE:") {
				actualArgs := make([]interface{}, len(tc.args))
				actualArgs[0] = handleStr
				copy(actualArgs[1:], tc.args[1:])
				tc.args = actualArgs // Modify args in place for this run
			}
		}
	}

	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}
		spec := toolImpl.Spec
		convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)

		if tc.valWantErrIs != nil {
			if valErr == nil {
				t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
			} else if !errors.Is(valErr, tc.valWantErrIs) {
				t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
			}
			return
		}
		if valErr != nil {
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		if tc.checkFunc != nil {
			tc.checkFunc(t, interp, gotResult, toolErr, context)
			return
		}

		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				var rtError *RuntimeError
				if errors.As(toolErr, &rtError) {
					if !errors.Is(rtError.Wrapped, tc.wantToolErrIs) {
						t.Errorf("Tool function expected wrapped error [%v], but got wrapped [%v] in error: %v", tc.wantToolErrIs, rtError.Wrapped, toolErr)
					} else {
						t.Logf("Got expected wrapped tool error: %v", toolErr)
					}
				} else {
					t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
				}
			} else {
				t.Logf("Got expected tool error: %v", toolErr)
			}
			return
		}
		if toolErr != nil {
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

// setupTreeWithJSON simplifies creating a tree from a JSON string for tests and returns its handle.
func setupTreeWithJSON(t *testing.T, interp *Interpreter, jsonStr string) string {
	t.Helper()
	loadTool, found := interp.ToolRegistry().GetTool("Tree.LoadJSON")
	if !found {
		t.Fatalf("setupTreeWithJSON: Tool Tree.LoadJSON not found in registry")
	}
	args := MakeArgs(jsonStr)
	result, err := loadTool.Func(interp, args)
	if err != nil {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON failed: %v", err)
	}
	handle, ok := result.(string)
	if !ok {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON did not return a string handle, got %T", result)
	}
	return handle
}

// callGetNode is a helper to call the Tree.GetNode tool within tests.
func callGetNode(t *testing.T, interp *Interpreter, handle, nodeID string) (map[string]interface{}, error) {
	t.Helper()
	getNodeTool, found := interp.ToolRegistry().GetTool("Tree.GetNode")
	if !found {
		return nil, fmt.Errorf("callGetNode: Tool Tree.GetNode not found")
	}
	result, err := getNodeTool.Func(interp, MakeArgs(handle, nodeID))
	if err != nil {
		return nil, err
	}
	nodeMap, ok := result.(map[string]interface{})
	if !ok {
		if result == nil && err == nil { // It's possible for a GetNode on a non-existent node to return nil, nil before error check
			return nil, fmt.Errorf("callGetNode: Tree.GetNode returned nil result for handle %q, nodeID %q", handle, nodeID)
		}
		return nil, fmt.Errorf("callGetNode: Tree.GetNode did not return a map, got %T", result)
	}
	return nodeMap, nil
}

// callSetMetadata is a helper to call Tree.SetNodeMetadata within tests.
func callSetMetadata(t *testing.T, interp *Interpreter, handle, nodeID, key, value string) error {
	t.Helper()
	setMetaTool, found := interp.ToolRegistry().GetTool("Tree.SetNodeMetadata")
	if !found {
		return fmt.Errorf("callSetMetadata: Tool Tree.SetNodeMetadata not found")
	}
	_, err := setMetaTool.Func(interp, MakeArgs(handle, nodeID, key, value))
	return err
}

// callGetChildren is a helper to call Tree.GetChildren within tests.
func callGetChildren(t *testing.T, interp *Interpreter, handle, nodeID string) ([]interface{}, error) {
	t.Helper()
	getChildrenTool, found := interp.ToolRegistry().GetTool("Tree.GetChildren")
	if !found {
		return nil, fmt.Errorf("callGetChildren: Tool Tree.GetChildren not found")
	}
	result, err := getChildrenTool.Func(interp, MakeArgs(handle, nodeID))
	if err != nil {
		return nil, err
	}
	children, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("callGetChildren: Tree.GetChildren did not return []interface{}, got %T", result)
	}
	return children, nil
}

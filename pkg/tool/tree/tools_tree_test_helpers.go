// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Made helper functions more robust by adding proper checking for type assertions.
// filename: pkg/tool/tree/tools_tree_test_helpers.go
// nlines: 80
// risk_rating: MEDIUM

package tree

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// treeTestCase defines the structure for a single tree tool test case.
type treeTestCase struct {
	name      string
	toolName  string
	args      []interface{}
	setupFunc func(t *testing.T, interp tool.RunTime) interface{}                                          // Returns a context, like a handle string
	checkFunc func(t *testing.T, interp tool.Runtime, result interface{}, err error, setupCtx interface{}) // Custom check logic
	wantErr   error                                                                                        // For simple error checks
}

// testTreeToolHelper runs a single tree tool test case.
func testTreeToolHelper(t *testing.T, interp tool.Runtime, tc treeTestCase) {
	t.Helper()

	t.Run(tc.name, func(t *testing.T) {
		var setupCtx interface{}
		if tc.setupFunc != nil {
			setupCtx = tc.setupFunc(t, interp)
		}

		// Prepare args, replacing placeholder with the handle from setup
		finalArgs := make([]interface{}, len(tc.args))
		for i, arg := range tc.args {
			if strArg, ok := arg.(string); ok && strings.HasPrefix(strArg, "SETUP_HANDLE:") {
				finalArgs[i] = setupCtx
			} else {
				finalArgs[i] = arg
			}
		}

		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}

		// Execute the tool function directly with primitive args
		result, err := toolImpl.Func(interp, finalArgs)

		// Perform checks
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErr, err)
			}
		} else if tc.checkFunc != nil {
			tc.checkFunc(t, interp, result, err, setupCtx)
		} else if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

// setupTreeWithJSON is a helper to load a tree from a JSON string and return its handle.
func setupTreeWithJSON(t *testing.T, interp tool.Runtime, jsonStr string) string {
	t.Helper()
	loadTool, _ := interp.ToolRegistry().GetTool("Tree.LoadJSON")
	handle, err := loadTool.Func(interp, []interface{}{jsonStr})
	if err != nil {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON failed: %v", err)
	}
	handleStr, ok := handle.(string)
	if !ok {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON did not return a string handle, got %T", handle)
	}
	return handleStr
}

// callGetNode is a helper to simplify getting node data within tests.
func callGetNode(t *testing.T, interp tool.Runtime, handle, nodeID string) (map[string]interface{}, error) {
	t.Helper()
	getTool, _ := interp.ToolRegistry().GetTool("Tree.GetNode")
	result, err := getTool.Func(interp, []interface{}{handle, nodeID})
	if err != nil {
		return nil, err
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("callGetNode: Tree.GetNode did not return a map[string]interface{}, but %T", result)
	}
	return resultMap, nil
}

// callSetMetadata is a helper for setting metadata during test setups.
func callSetMetadata(t *testing.T, interp tool.Runtime, handle, nodeID, key, value string) error {
	t.Helper()
	setTool, _ := interp.ToolRegistry().GetTool("Tree.SetNodeMetadata")
	_, err := setTool.Func(interp, []interface{}{handle, nodeID, key, value})
	return err
}

// callGetChildren is a helper to simplify getting child node IDs within tests.
func callGetChildren(t *testing.T, interp tool.Runtime, handle, nodeID string) ([]string, error) {
	t.Helper()
	getTool, _ := interp.ToolRegistry().GetTool("Tree.GetChildren")
	result, err := getTool.Func(interp, []interface{}{handle, nodeID})
	if err != nil {
		return nil, err
	}
	resultSlice, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("GetChildren did not return a slice of interface{}, got %T", result)
	}
	ids := make([]string, len(resultSlice))
	for i, v := range resultSlice {
		id, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("GetChildren: element at index %d is not a string, got %T", i, v)
		}
		ids[i] = id
	}
	return ids, nil
}

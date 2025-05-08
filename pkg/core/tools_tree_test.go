// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Fix multi-value GetTool calls inside checkFuncs.
// nlines: 460 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_test.go

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// --- Test Case Structure for Tree Tools ---
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

// --- Generic Tree Tool Test Helper ---
func testTreeToolHelper(t *testing.T, interp *Interpreter, tc treeTestCase) {
	t.Helper()

	var context interface{}
	if tc.setupFunc != nil {
		context = tc.setupFunc(t, interp)
		if handleStr, ok := context.(string); ok && len(tc.args) > 0 {
			if placeholder, pOK := tc.args[0].(string); pOK && strings.HasPrefix(placeholder, "SETUP_HANDLE:") {
				actualArgs := make([]interface{}, len(tc.args))
				actualArgs[0] = handleStr
				copy(actualArgs[1:], tc.args[1:])
				tc.args = actualArgs
			}
		}
	}

	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}
		spec := toolImpl.Spec

		// --- Validation ---
		convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)

		if tc.valWantErrIs != nil {
			if valErr == nil {
				t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
			} else if !errors.Is(valErr, tc.valWantErrIs) {
				t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
			}
			return // Stop if validation error was expected
		}
		if valErr != nil { // Unexpected validation error
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution ---
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// --- Custom Check (if provided, it takes precedence) ---
		if tc.checkFunc != nil {
			tc.checkFunc(t, interp, gotResult, toolErr, context)
			return
		}

		// --- Tool Error Check ---
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				var rtError *RuntimeError
				if errors.As(toolErr, &rtError) {
					if !errors.Is(rtError.Wrapped, tc.wantToolErrIs) {
						t.Errorf("Tool function expected wrapped error [%v], but got wrapped [%v] in error: %v", tc.wantToolErrIs, rtError.Wrapped, toolErr)
					}
				} else {
					t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
				}
			}
			return // Stop if tool error was expected
		}
		if toolErr != nil { // Unexpected tool error
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison ---
		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

// Helper to simplify creating a tree for tests and returning its handle
func setupTreeWithJSON(t *testing.T, interp *Interpreter, jsonStr string) string {
	t.Helper()
	// Need to get the LoadJSON tool correctly first
	loadTool, found := interp.ToolRegistry().GetTool("Tree.LoadJSON")
	if !found {
		t.Fatalf("setupTreeWithJSON: Tool Tree.LoadJSON not found in registry")
	}
	args := MakeArgs(jsonStr)
	result, err := loadTool.Func(interp, args) // Call the function via the implementation
	if err != nil {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON failed: %v", err)
	}
	handle, ok := result.(string)
	if !ok {
		t.Fatalf("setupTreeWithJSON: Tree.LoadJSON did not return a string handle, got %T", result)
	}
	return handle
}

// Helper to call Tree.GetNode within tests (handles GetTool correctly)
func callGetNode(t *testing.T, interp *Interpreter, handle, nodeID string) (map[string]interface{}, error) {
	t.Helper()
	getNodeTool, found := interp.ToolRegistry().GetTool("Tree.GetNode")
	if !found {
		return nil, fmt.Errorf("callGetNode: Tool Tree.GetNode not found")
	}
	result, err := getNodeTool.Func(interp, MakeArgs(handle, nodeID))
	if err != nil {
		return nil, err // Return the error from the tool call
	}
	nodeMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("callGetNode: Tree.GetNode did not return a map, got %T", result)
	}
	return nodeMap, nil
}

// Helper to call Tree.SetNodeMetadata within tests
func callSetMetadata(t *testing.T, interp *Interpreter, handle, nodeID, key, value string) error {
	t.Helper()
	setMetaTool, found := interp.ToolRegistry().GetTool("Tree.SetNodeMetadata")
	if !found {
		return fmt.Errorf("callSetMetadata: Tool Tree.SetNodeMetadata not found")
	}
	_, err := setMetaTool.Func(interp, MakeArgs(handle, nodeID, key, value))
	return err
}

// Helper to call Tree.GetChildren within tests
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

// --- Tests ---

func TestTreeLoadJSONAndToJSON(t *testing.T) {
	//interp, _ := NewDefaultTestInterpreter(t)
	validJSONSimple := `{"key":"value","num":123}`
	validJSONNested := `{"a":[1,{"b":null}],"c":true}`
	validJSONArray := `[1,"two",true]`

	testCases := []treeTestCase{
		// Tree.LoadJSON
		{name: "LoadJSON Simple Object", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONSimple), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			handleStr, ok := result.(string)
			if !ok || !strings.HasPrefix(handleStr, GenericTreeHandleType+"::") {
				t.Errorf("Expected valid handle string, got %T: %v", result, result)
			}
		}},
		{name: "LoadJSON Nested Structure", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONNested), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Simple Array", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONArray), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Empty Object", toolName: "Tree.LoadJSON", args: MakeArgs(`{}`), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Empty Array", toolName: "Tree.LoadJSON", args: MakeArgs(`[]`), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Invalid JSON", toolName: "Tree.LoadJSON", args: MakeArgs(`{"key": "value`), wantToolErrIs: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Empty Input", toolName: "Tree.LoadJSON", args: MakeArgs(``), wantToolErrIs: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Wrong Arg Type", toolName: "Tree.LoadJSON", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},

		// Tree.ToJSON
		{name: "ToJSON Simple Object", toolName: "Tree.ToJSON",
			setupFunc: func(t *testing.T, interp *Interpreter) interface{} {
				return setupTreeWithJSON(t, interp, validJSONSimple)
			},
			args: MakeArgs("SETUP_HANDLE:tree1"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
				if err != nil {
					t.Fatalf("ToJSON failed: %v", err)
				}
				jsonStr, ok := result.(string)
				if !ok {
					t.Fatalf("ToJSON did not return a string, got %T", result)
				}
				if !strings.Contains(jsonStr, `"key": "value"`) || !strings.Contains(jsonStr, `"num": 123`) {
					t.Errorf("ToJSON output mismatch for simple object. Got: %s", jsonStr)
				}
			}},
		{name: "ToJSON Invalid Handle", toolName: "Tree.ToJSON", args: MakeArgs("invalid-handle"), wantToolErrIs: ErrNotFound},
	}

	for _, tc := range testCases {
		// Use a fresh interpreter for each Load/ToJSON test case to avoid handle ID collisions from setup
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}

func TestTreeNavigationTools(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Shared interp might be ok if setup is careful
	jsonInput := `{
		"name": "root_obj",
		"type": "directory",
		"children": [
			{"name": "file1.txt", "type": "file", "size": 100},
			{"name": "subdir", "type": "directory", "children": [
				{"name": "file2.txt", "type": "file", "size": 50}
			]}
		],
		"metadata": {"owner": "admin"}
	}`
	rootHandle := setupTreeWithJSON(t, interp, jsonInput) // Load once for all nav tests
	rootNodeID := "node-1"                                // Assume root is node-1

	testCases := []treeTestCase{
		// Tree.GetNode
		{
			name:     "GetNode Root",
			toolName: "Tree.GetNode",
			args:     MakeArgs(rootHandle, rootNodeID), // Use the actual handle
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("GetNode Root failed: %v", err)
				}
				nodeMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("GetNode Root: expected map, got %T", result)
				}
				if nodeMap["id"] != rootNodeID {
					t.Errorf("GetNode Root: ID mismatch, got %v", nodeMap["id"])
				}
				if nodeMap["type"] != "object" {
					t.Errorf("GetNode Root: type mismatch, got %v", nodeMap["type"])
				}
			},
		},
		{
			name:     "GetNode Nested (file1 name)",
			toolName: "Tree.GetNode",
			args:     MakeArgs(rootHandle, ""), // Node ID determined dynamically in checkFunc
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				// --- Dynamic ID Discovery ---
				// Use helper function `callGetNode` which handles GetTool correctly
				rootMap, err := callGetNode(t, interp, rootHandle, rootNodeID)
				if err != nil {
					t.Fatalf("CheckFunc setup: Failed to get root node: %v", err)
				}
				childrenArrNodeID := rootMap["attributes"].(map[string]interface{})["children"].(string)
				childrenArrMap, err := callGetNode(t, interp, rootHandle, childrenArrNodeID)
				if err != nil {
					t.Fatalf("CheckFunc setup: Failed to get children array node: %v", err)
				}
				file1ObjNodeID := childrenArrMap["children"].([]interface{})[0].(string)
				file1ObjMap, err := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				if err != nil {
					t.Fatalf("CheckFunc setup: Failed to get file1 object node: %v", err)
				}
				fileNameNodeID := file1ObjMap["attributes"].(map[string]interface{})["name"].(string)
				// --- End Discovery ---

				// Test GetNode on the dynamically found ID
				nodeMap, err := callGetNode(t, interp, rootHandle, fileNameNodeID)
				if err != nil {
					t.Fatalf("GetNode for file1 name failed: %v", err)
				}
				if nodeMap["value"] != "file1.txt" {
					t.Errorf("Expected 'file1.txt', got %v", nodeMap["value"])
				}
				if nodeMap["type"] != "string" {
					t.Errorf("Expected type string, got %v", nodeMap["type"])
				}
			},
		},
		{name: "GetNode NonExistent Node", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, "node-999"), wantToolErrIs: ErrNotFound},

		// Tree.GetChildren
		{
			name:     "GetChildren of Array Node",
			toolName: "Tree.GetChildren",
			args:     MakeArgs(rootHandle, ""), // Node ID determined dynamically
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				// --- Dynamic ID Discovery ---
				rootMap, err := callGetNode(t, interp, rootHandle, rootNodeID)
				if err != nil {
					t.Fatalf("CheckFunc setup: Failed to get root node: %v", err)
				}
				childrenArrNodeID := rootMap["attributes"].(map[string]interface{})["children"].(string)
				// --- End Discovery ---

				childrenIDs, err := callGetChildren(t, interp, rootHandle, childrenArrNodeID) // Use helper
				if err != nil {
					t.Fatalf("GetChildren for array failed: %v", err)
				}
				if len(childrenIDs) != 2 {
					t.Errorf("Expected 2 children, got %d", len(childrenIDs))
				}
			},
		},
		// Need to discover a leaf node ID first for this test reliably
		// {name: "GetChildren of Leaf Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, leafNodeID), wantResult: []interface{}{}},

		// Tree.GetParent
		{
			name:       "GetParent of Root",
			toolName:   "Tree.GetParent",
			args:       MakeArgs(rootHandle, rootNodeID),
			wantResult: "",
		},
		// Need dynamic discovery for reliable parent tests of non-root nodes
	}

	for _, tc := range testCases {
		// Use the shared interpreter for navigation tests as they don't modify the tree
		testTreeToolHelper(t, interp, tc)
	}
}

func TestTreeModificationTools(t *testing.T) {
	jsonSimple := `{"name": "item", "value": 10}`

	setupInitialTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple)
	}

	testCases := []treeTestCase{
		// Tree.SetValue
		{
			name: "SetValue Valid Leaf", toolName: "Tree.SetValue",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-3", "new_value_for_value_node"), // Assumes "value" is node-3
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("SetValue Valid Leaf failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, err := callGetNode(t, interp, handle, "node-3") // Use helper
				if err != nil {
					t.Fatalf("CheckFunc: Failed to get node-3 after SetValue: %v", err)
				}
				if nodeMap["value"] != "new_value_for_value_node" {
					t.Errorf("SetValue did not update node. Got: %v", nodeMap["value"])
				}
			},
		},
		{name: "SetValue On Object Node", toolName: "Tree.SetValue", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1", "should_fail"), wantToolErrIs: ErrCannotSetValueOnType},

		// Tree.AddChildNode
		{
			name: "AddChildNode To Root Object", toolName: "Tree.AddChildNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1", "newChild1", "string", "hello", "newKey"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("AddChildNode failed: %v", err)
				}
				newID, ok := result.(string)
				if !ok || newID == "" {
					t.Fatalf("AddChildNode did not return new node ID string, got %T: %v", result, result)
				}
				if newID != "newChild1" {
					t.Errorf("Expected new ID newChild1, got %s", newID)
				}

				handle := ctx.(string)
				parentNodeMap, err := callGetNode(t, interp, handle, "node-1") // Use helper
				if err != nil {
					t.Fatalf("CheckFunc: Failed to get parent node-1 after AddChildNode: %v", err)
				}
				attrs := parentNodeMap["attributes"].(map[string]interface{})
				if attrs["newKey"] != newID {
					t.Errorf("AddChildNode: newKey not pointing to new child ID. Attrs: %v", attrs)
				}
				childNodeMap, err := callGetNode(t, interp, handle, newID) // Use helper
				if err != nil {
					t.Fatalf("CheckFunc: Failed to get new child node %s: %v", newID, err)
				}
				if childNodeMap["value"] != "hello" {
					t.Errorf("Added child has wrong value: %v", childNodeMap["value"])
				}
			},
		},
		{name: "AddChildNode ID Exists", toolName: "Tree.AddChildNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1", "node-2", "string", "fail", "anotherKey"), wantToolErrIs: ErrNodeIDExists},
	}
	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}

func TestTreeMetadataTools(t *testing.T) {
	jsonSimple := `{"key":"value"}`
	setupMetaTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple)
	}

	testCases := []treeTestCase{
		// Tree.SetNodeMetadata
		{
			name: "SetNodeMetadata New Key", toolName: "Tree.SetNodeMetadata",
			setupFunc: setupMetaTree,
			args:      MakeArgs("SETUP_HANDLE:mTree", "node-1", "metaKey1", "metaValue1"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("SetNodeMetadata failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, err := callGetNode(t, interp, handle, "node-1") // Use helper
				if err != nil {
					t.Fatalf("CheckFunc: Failed to get node-1 after SetNodeMetadata: %v", err)
				}
				attrs := nodeMap["attributes"].(map[string]interface{})
				if attrs["metaKey1"] != "metaValue1" {
					t.Errorf("Metadata not set correctly. Got: %v", attrs["metaKey1"])
				}
			},
		},
		{name: "SetNodeMetadata Empty Key", toolName: "Tree.SetNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "", "val"), wantToolErrIs: ErrInvalidArgument},

		// Tree.RemoveNodeMetadata
		{
			name: "RemoveNodeMetadata Existing Key", toolName: "Tree.RemoveNodeMetadata",
			setupFunc: func(t *testing.T, interp *Interpreter) interface{} {
				handle := setupTreeWithJSON(t, interp, jsonSimple)
				err := callSetMetadata(t, interp, handle, "node-1", "toRemove", "val") // Use helper
				if err != nil {
					t.Fatalf("Setup for RemoveNodeMetadata failed: %v", err)
				}
				return handle
			},
			args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "toRemove"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("RemoveNodeMetadata failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, err := callGetNode(t, interp, handle, "node-1") // Use helper
				if err != nil {
					t.Fatalf("CheckFunc: Failed to get node-1 after RemoveNodeMetadata: %v", err)
				}
				attrs := nodeMap["attributes"].(map[string]interface{})
				if _, exists := attrs["toRemove"]; exists {
					t.Errorf("Metadata key 'toRemove' still exists after removal.")
				}
			},
		},
		{name: "RemoveNodeMetadata NonExistent Key", toolName: "Tree.RemoveNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "nonKey"), wantToolErrIs: ErrAttributeNotFound},
	}
	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}

func TestTreeFindAndRenderTools(t *testing.T) {
	jsonComplex := `{"id":"root","type":"folder","meta":{"status":"active"},"children":[{"id":"child1","type":"file","value":"content1.txt","meta":{"size":100}},{"id":"child2","type":"folder","children":[{"id":"grandchild1","type":"file","value":"content2.dat","meta":{"size":200}}]}]}`
	setupFindRenderTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonComplex)
	}

	testCases := []treeTestCase{
		// Tree.FindNodes
		{
			name: "FindNodes By Type 'file'", toolName: "Tree.FindNodes",
			setupFunc: setupFindRenderTree,
			args:      MakeArgs("SETUP_HANDLE:frTree", "node-1", map[string]interface{}{"type": "file"}, int64(-1), int64(-1)),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("FindNodes by type 'file' failed: %v", err)
				}
				ids, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T", result)
				}
				if len(ids) != 2 {
					t.Errorf("Expected 2 'file' nodes, got %d: %v", len(ids), ids)
				}
				handle := ctx.(string)
				for _, idUnk := range ids {
					idStr := idUnk.(string)
					nodeMap, err := callGetNode(t, interp, handle, idStr) // Use helper
					if err != nil {
						t.Errorf("CheckFunc: Failed to get node %s: %v", idStr, err)
						continue
					}
					if nodeMap["type"] != "file" {
						t.Errorf("Found node %s is not type 'file', but '%s'", idStr, nodeMap["type"])
					}
				}
			},
		},
		{
			name: "FindNodes By Value", toolName: "Tree.FindNodes",
			setupFunc: setupFindRenderTree,
			args:      MakeArgs("SETUP_HANDLE:frTree", "node-1", map[string]interface{}{"value": "content1.txt"}, int64(-1), int64(-1)),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("FindNodes by value failed: %v", err)
				}
				ids, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T", result)
				}
				if len(ids) != 1 {
					t.Errorf("Expected 1 node with value 'content1.txt', got %d", len(ids))
				}
			},
		},

		// Tree.RenderText
		{
			name: "RenderText Basic", toolName: "Tree.RenderText",
			setupFunc: func(t *testing.T, interp *Interpreter) interface{} { return setupTreeWithJSON(t, interp, `{"a":"b"}`) },
			args:      MakeArgs("SETUP_HANDLE:renderTree"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("RenderText failed: %v", err)
				}
				s, ok := result.(string)
				if !ok {
					t.Fatalf("RenderText did not return string, got %T", result)
				}
				if !strings.Contains(s, "- (object)") || !strings.Contains(s, `* Key: "a"`) || !strings.Contains(s, `- (string): "b"`) {
					t.Errorf("RenderText output seems incorrect. Got:\n%s", s)
				}
			},
		},
	}

	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}

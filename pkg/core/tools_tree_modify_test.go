// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 20:29:00 PM PDT // Update error expectations after interpreter fix
// filename: pkg/core/tools_tree_modify_test.go

package core

import (
	"errors"
	"reflect" // Needed for DeepEqual in verification
	"testing"
	// Using cmp for better diffs
)

// --- Test Setup Helper (using standard testing) ---

// setupTreeTest loads JSON and returns interpreter + handle. Fails test on error.
func setupTreeTestStd(t *testing.T, jsonStr string) (*Interpreter, string) {
	// ... (implementation unchanged) ...
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t)
	handle, err := toolTreeLoadJSON(interp, MakeArgs(jsonStr))
	if err != nil {
		t.Fatalf("setupTreeTest: toolTreeLoadJSON failed: %v", err)
	}
	handleStr, ok := handle.(string)
	if !ok || handleStr == "" {
		t.Fatalf("setupTreeTest: handle is not a string or is empty: %T %v", handle, handle)
	}
	return interp, handleStr
}

// Helper function to get a pointer to an int
func Pint(i int) *int {
	return &i
}

// getNodeValueStd is a helper to get the 'value' field using standard testing checks
func getNodeValueStd(t *testing.T, nodeMap map[string]interface{}) interface{} {
	// ... (implementation unchanged) ...
	t.Helper()
	val, exists := nodeMap["value"]
	if !exists {
		t.Fatalf("getNodeValueStd: 'value' field missing from node map: %v", nodeMap)
	}
	return val
}

// getNodeAttributesStd is a helper to get the 'attributes' map using standard testing checks
func getNodeAttributesStd(t *testing.T, nodeMap map[string]interface{}) map[string]interface{} {
	// ... (implementation unchanged) ...
	t.Helper()
	attrsVal, exists := nodeMap["attributes"]
	if !exists {
		return nil
	}
	if attrsVal == nil {
		return nil
	}
	attrsMap, ok := attrsVal.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeAttributesStd: 'attributes' field is not a map[string]interface{}: %T", attrsVal)
	}
	return attrsMap
}

// getNodeChildrenStd is a helper to get the 'children' slice using standard testing checks
func getNodeChildrenStd(t *testing.T, nodeMap map[string]interface{}) []interface{} {
	// ... (implementation unchanged) ...
	t.Helper()
	childrenVal, exists := nodeMap["children"]
	if !exists {
		return nil
	}
	if childrenVal == nil {
		return nil
	}
	childrenSlice, ok := childrenVal.([]interface{})
	if !ok {
		t.Fatalf("getNodeChildrenStd: 'children' field is not a []interface{}: %T", childrenVal)
	}
	return childrenSlice
}

// --- Test Cases (using standard testing) ---

func TestTreeModifyNodeStd(t *testing.T) {
	jsonFixture := `{ "stringVal": "initial", "numberVal": 123.0, "boolVal": true, "nullVal": null, "objVal": {}, "arrVal": [] }`
	interp, handle := setupTreeTestStd(t, jsonFixture)

	getNodeID := func(key string) string {
		rootNode := getNodeHelper(t, interp, handle, "node-1")
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if rootAttrs == nil {
			t.Fatalf("Cannot get root attributes in getNodeID helper")
		}
		return getAttrNodeID(t, rootAttrs, key)
	}

	stringNodeID := getNodeID("stringVal")
	numberNodeID := getNodeID("numberVal")
	boolNodeID := getNodeID("boolVal")
	nullNodeID := getNodeID("nullVal")
	objNodeID := getNodeID("objVal")
	arrNodeID := getNodeID("arrVal")

	testCases := []struct {
		name          string
		nodeID        string
		modifications map[string]interface{}
		expectError   bool
		wantErrIs     error
		verifyValue   interface{}
		skipVerify    bool
	}{
		{name: "Modify String Value", nodeID: stringNodeID, modifications: map[string]interface{}{"value": "updated"}, expectError: false, verifyValue: "updated"},
		{name: "Modify Number Value", nodeID: numberNodeID, modifications: map[string]interface{}{"value": 456.7}, expectError: false, verifyValue: 456.7},
		{name: "Modify Boolean Value (true->false)", nodeID: boolNodeID, modifications: map[string]interface{}{"value": false}, expectError: false, verifyValue: false},
		{name: "Modify Null Value (null->string)", nodeID: nullNodeID, modifications: map[string]interface{}{"value": "not null anymore"}, expectError: false, verifyValue: "not null anymore"},
		{name: "Modify Back To Null", nodeID: stringNodeID, modifications: map[string]interface{}{"value": nil}, expectError: false, verifyValue: nil},
		{name: "Error: Target is Object Node", nodeID: objNodeID, modifications: map[string]interface{}{"value": "should fail"}, expectError: true, wantErrIs: ErrTreeCannotSetValueOnType, skipVerify: true},
		{name: "Error: Target is Array Node", nodeID: arrNodeID, modifications: map[string]interface{}{"value": "should also fail"}, expectError: true, wantErrIs: ErrTreeCannotSetValueOnType, skipVerify: true},
		{name: "Error: Modifications map missing 'value'", nodeID: numberNodeID, modifications: map[string]interface{}{"wrong_key": 999}, expectError: true, wantErrIs: ErrInvalidArgument, skipVerify: true},
		{name: "Error: Invalid Node ID", nodeID: "node-999", modifications: map[string]interface{}{"value": 1}, expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		// UPDATED Error Expectation for Invalid Handle
		{name: "Error: Invalid Handle", nodeID: stringNodeID, modifications: map[string]interface{}{"value": 1}, expectError: true, wantErrIs: ErrInvalidArgument, skipVerify: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentHandle := handle
			if tc.name == "Error: Invalid Handle" {
				currentHandle = "bad-handle"
			}

			_, err := toolTreeModifyNode(interp, MakeArgs(currentHandle, tc.nodeID, tc.modifications))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.wantErrIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got one: %v", err)
				}
				if !tc.skipVerify {
					modifiedNode := getNodeHelper(t, interp, handle, tc.nodeID)
					actualValue := getNodeValueStd(t, modifiedNode)
					if !reflect.DeepEqual(tc.verifyValue, actualValue) {
						t.Errorf("Value was not modified correctly. got = %v (%T), want = %v (%T)", actualValue, actualValue, tc.verifyValue, tc.verifyValue)
					} else {
						t.Logf("Value modified correctly to %v", actualValue)
					}
				}
			}
		})
	}
}

func TestTreeSetAttributeStd(t *testing.T) {
	jsonFixture := `{ "targetObject": {"existing": "initial child"}, "valueNode1": "hello", "valueNode2": 100, "nonObjectNode": "i am string" }`
	interp, handle := setupTreeTestStd(t, jsonFixture)

	getNodeID := func(key string) string {
		rootNode := getNodeHelper(t, interp, handle, "node-1")
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if rootAttrs == nil {
			t.Fatalf("Cannot get root attributes in getNodeID helper")
		}
		return getAttrNodeID(t, rootAttrs, key)
	}

	targetObjID := getNodeID("targetObject")
	val1ID := getNodeID("valueNode1")
	val2ID := getNodeID("valueNode2")
	nonObjID := getNodeID("nonObjectNode")

	initialTargetNode := getNodeHelper(t, interp, handle, targetObjID)
	initialTargetAttrs := getNodeAttributesStd(t, initialTargetNode)
	if initialTargetAttrs == nil {
		t.Fatalf("Initial target object attributes are nil")
	}
	existingChildID := getAttrNodeID(t, initialTargetAttrs, "existing")
	if existingChildID == "" {
		t.Fatalf("Could not fetch initial child ID for 'existing' attribute")
	}

	testCases := []struct {
		name        string
		nodeID      string
		attrKey     string
		childNodeID string
		expectError bool
		wantErrIs   error
		verifyFunc  func(t *testing.T)
		skipVerify  bool
	}{
		{
			name: "Add new attribute", nodeID: targetObjID, attrKey: "newKey", childNodeID: val1ID, expectError: false,
			verifyFunc: func(t *testing.T) {
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should have 2 attributes, got: %v", attrs)
				}
				if val, ok := attrs["newKey"].(string); !ok || val != val1ID {
					t.Errorf("Attribute 'newKey' has wrong value. got = %v (%T), want = %v", attrs["newKey"], attrs["newKey"], val1ID)
				}
				if val, ok := attrs["existing"].(string); !ok || val != existingChildID {
					t.Errorf("Attribute 'existing' changed unexpectedly. got = %v (%T), want = %v", attrs["existing"], attrs["existing"], existingChildID)
				}
			},
		},
		{
			name: "Update existing attribute", nodeID: targetObjID, attrKey: "existing", childNodeID: val2ID, expectError: false,
			verifyFunc: func(t *testing.T) {
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should have 2 attributes, got: %v", attrs)
				}
				if val, ok := attrs["existing"].(string); !ok || val != val2ID {
					t.Errorf("Attribute 'existing' was not updated correctly. got = %v (%T), want = %v", attrs["existing"], attrs["existing"], val2ID)
				}
				if val, ok := attrs["newKey"].(string); !ok || val != val1ID {
					t.Errorf("Attribute 'newKey' changed unexpectedly. got = %v (%T), want = %v", attrs["newKey"], attrs["newKey"], val1ID)
				}
			},
		},
		{name: "Error: Target node not object", nodeID: nonObjID, attrKey: "testKey", childNodeID: val1ID, expectError: true, wantErrIs: ErrTreeNodeNotObject, skipVerify: true},
		{name: "Error: Empty attribute key", nodeID: targetObjID, attrKey: "", childNodeID: val1ID, expectError: true, wantErrIs: ErrInvalidArgument, skipVerify: true},
		{name: "Error: Child node ID does not exist", nodeID: targetObjID, attrKey: "testKey", childNodeID: "node-999", expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		{name: "Error: Invalid Node ID", nodeID: "node-invalid", attrKey: "test", childNodeID: val1ID, expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		// UPDATED Error Expectation for Invalid Handle
		{name: "Error: Invalid Handle", nodeID: targetObjID, attrKey: "placeholderKey", childNodeID: val1ID, skipVerify: true, expectError: true, wantErrIs: ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentHandle := handle
			currentNodeID := tc.nodeID
			if tc.name == "Error: Invalid Handle" {
				currentHandle = "bad-handle"
			}
			if tc.name == "Error: Invalid Node ID" {
				currentNodeID = "node-invalid"
			} // Added this check

			_, err := toolTreeSetAttribute(interp, MakeArgs(currentHandle, currentNodeID, tc.attrKey, tc.childNodeID))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.wantErrIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error during modification: %v", err)
				}
				if !tc.skipVerify && tc.verifyFunc != nil {
					tc.verifyFunc(t)
				}
			}
		})
	}
}

// --- TestTreeRemoveAttributeStd ---

func TestTreeRemoveAttributeStd(t *testing.T) {
	jsonFixture := `{ "targetObject": {"key1": "val1", "keyToRemove": "valRemove", "key3": "val3"}, "otherObject": {}, "nonObjectNode": "i am string" }`
	interp, handle := setupTreeTestStd(t, jsonFixture)

	getNodeID := func(key string) string {
		rootNode := getNodeHelper(t, interp, handle, "node-1")
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if rootAttrs == nil {
			t.Fatalf("Cannot get root attributes in getNodeID helper")
		}
		return getAttrNodeID(t, rootAttrs, key)
	}

	targetObjID := getNodeID("targetObject")
	nonObjID := getNodeID("nonObjectNode")

	testCases := []struct {
		name        string
		nodeID      string
		attrKey     string
		expectError bool
		wantErrIs   error
		verifyFunc  func(t *testing.T)
		skipVerify  bool
	}{
		{
			name: "Remove existing attribute", nodeID: targetObjID, attrKey: "keyToRemove", expectError: false,
			verifyFunc: func(t *testing.T) {
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should have 2 attributes after removal, got: %v", attrs)
				}
				if _, exists := attrs["keyToRemove"]; exists {
					t.Errorf("Attribute 'keyToRemove' was not removed.")
				}
				if _, exists := attrs["key1"]; !exists {
					t.Errorf("Attribute 'key1' was removed unexpectedly.")
				}
				if _, exists := attrs["key3"]; !exists {
					t.Errorf("Attribute 'key3' was removed unexpectedly.")
				}
			},
		},
		{
			name: "Attempt remove non-existent attribute", nodeID: targetObjID, attrKey: "doesNotExist", expectError: true, wantErrIs: ErrAttributeNotFound,
			verifyFunc: func(t *testing.T) {
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should still have 2 attributes, got: %v", attrs)
				} // Verify state didn't change
			},
		},
		{name: "Attempt remove from non-object node", nodeID: nonObjID, attrKey: "anyKey", expectError: true, wantErrIs: ErrTreeNodeNotObject, skipVerify: true},
		{name: "Attempt remove with empty key", nodeID: targetObjID, attrKey: "", expectError: true, wantErrIs: ErrInvalidArgument, skipVerify: true},
		{name: "Error: Invalid Node ID", nodeID: "node-invalid", attrKey: "key1", expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		// UPDATED Error Expectation for Invalid Handle
		{name: "Error: Invalid Handle", nodeID: targetObjID, attrKey: "key1", skipVerify: true, expectError: true, wantErrIs: ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentHandle := handle
			currentNodeID := tc.nodeID
			if tc.name == "Error: Invalid Handle" {
				currentHandle = "bad-handle"
			}
			if tc.name == "Error: Invalid Node ID" {
				currentNodeID = "node-invalid"
			}

			_, err := toolTreeRemoveAttribute(interp, MakeArgs(currentHandle, currentNodeID, tc.attrKey))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.wantErrIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
				if !tc.skipVerify && tc.verifyFunc != nil {
					t.Logf("Verifying state after expected error...")
					tc.verifyFunc(t)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error during removal: %v", err)
				}
				if !tc.skipVerify && tc.verifyFunc != nil {
					tc.verifyFunc(t)
				}
			}
		})
	}
}

// --- TestTreeAddNodeStd ---

func TestTreeAddNodeStd(t *testing.T) {
	jsonFixture := `{ "rootObject": {"key": "value"}, "rootArray": ["existing1", "existing2"] }`
	interp, handle := setupTreeTestStd(t, jsonFixture)

	getNodeID := func(key string) string {
		rootNode := getNodeHelper(t, interp, handle, "node-1")
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if rootAttrs == nil {
			t.Fatalf("Cannot get root attributes in getNodeID helper")
		}
		return getAttrNodeID(t, rootAttrs, key)
	}

	rootObjID := getNodeID("rootObject")
	rootArrID := getNodeID("rootArray")
	rootID := "node-1"

	// --- Verification Helper for Added Nodes ---
	verifyAddedNode := func(t *testing.T, parentID, newNodeID, expectedType string, expectedValue interface{}, index int) {
		// ... (implementation unchanged) ...
		t.Helper()
		newNodeData := getNodeHelper(t, interp, handle, newNodeID)
		if newNodeData == nil {
			t.Fatalf("verifyAddedNode: Could not retrieve newly added node %q", newNodeID)
		}
		if newNodeData["id"] != newNodeID {
			t.Errorf("verifyAddedNode: New node ID mismatch. got=%q, want=%q", newNodeData["id"], newNodeID)
		}
		if newNodeData["type"] != expectedType {
			t.Errorf("verifyAddedNode: New node type mismatch for %q. got=%q, want=%q", newNodeID, newNodeData["type"], expectedType)
		}
		if newNodeData["parentId"] != parentID {
			t.Errorf("verifyAddedNode: New node parent ID mismatch for %q. got=%q, want=%q", newNodeID, newNodeData["parentId"], parentID)
		}
		if expectedValue != nil || expectedType == "null" {
			if !reflect.DeepEqual(newNodeData["value"], expectedValue) {
				t.Errorf("verifyAddedNode: New node value mismatch for %q. got=%v (%T), want=%v (%T)", newNodeID, newNodeData["value"], newNodeData["value"], expectedValue, expectedValue)
			}
		} else if newNodeData["value"] != nil && expectedType != "object" && expectedType != "array" {
			t.Errorf("verifyAddedNode: New node value should be nil for complex type %q, got=%v", expectedType, newNodeData["value"])
		}
		parentNodeData := getNodeHelper(t, interp, handle, parentID)
		parentChildren := getNodeChildrenStd(t, parentNodeData)
		if parentChildren == nil {
			t.Fatalf("verifyAddedNode: Parent %q children list is nil", parentID)
		}
		found := false
		actualIndex := -1
		for i, childIntf := range parentChildren {
			childID, ok := childIntf.(string)
			if !ok {
				t.Errorf("verifyAddedNode: Child in parent %q list is not a string ID: %v", parentID, childIntf)
				continue
			}
			if childID == newNodeID {
				found = true
				actualIndex = i
				break
			}
		}
		if !found {
			t.Errorf("verifyAddedNode: New node ID %q not found in parent %q children list: %v", newNodeID, parentID, parentChildren)
			return
		}
		if index >= 0 {
			expectedIndex := index
			if actualIndex != expectedIndex {
				t.Errorf("verifyAddedNode: New node %q found at wrong index in parent %q children. got=%d, want=%d, children=%v", newNodeID, parentID, actualIndex, expectedIndex, parentChildren)
			}
		} else {
			if actualIndex != len(parentChildren)-1 {
				t.Errorf("verifyAddedNode: New node %q was appended but is not the last element in parent %q children. got index=%d, len=%d, children=%v", newNodeID, parentID, actualIndex, len(parentChildren), parentChildren)
			}
		}
	}
	// --- End Verification Helper ---

	testCases := []struct {
		name        string
		parentID    string
		newNodeID   string
		nodeType    string
		nodeValue   interface{}
		index       *int
		expectError bool
		wantErrIs   error
	}{
		// --- Success Cases ---
		{name: "Add string node to root object", parentID: rootID, newNodeID: "new-string-node", nodeType: "string", nodeValue: "I am new", index: nil, expectError: false},
		{name: "Add number node to root object", parentID: rootID, newNodeID: "new-number-node", nodeType: "number", nodeValue: 999.9, index: nil, expectError: false},
		{name: "Add boolean node to root object", parentID: rootID, newNodeID: "new-bool-node", nodeType: "boolean", nodeValue: false, index: nil, expectError: false},
		{name: "Add null node to root object", parentID: rootID, newNodeID: "new-null-node", nodeType: "null", nodeValue: nil, index: nil, expectError: false},
		{name: "Add empty object node to root object", parentID: rootID, newNodeID: "new-object-node", nodeType: "object", nodeValue: nil, index: nil, expectError: false},
		{name: "Add empty array node to root object", parentID: rootID, newNodeID: "new-array-node", nodeType: "array", nodeValue: nil, index: nil, expectError: false},
		{name: "Append node to existing array", parentID: rootArrID, newNodeID: "appended-node", nodeType: "string", nodeValue: "appended", index: nil, expectError: false},
		{name: "Insert node at start of array", parentID: rootArrID, newNodeID: "inserted-start-node", nodeType: "string", nodeValue: "inserted-start", index: Pint(0), expectError: false},
		{name: "Insert node in middle of array", parentID: rootArrID, newNodeID: "inserted-middle-node", nodeType: "string", nodeValue: "inserted-middle", index: Pint(2), expectError: false},

		// --- Error Cases ---
		{name: "Error: Node ID already exists", parentID: rootID, newNodeID: rootObjID, nodeType: "string", nodeValue: "fail", index: nil, expectError: true, wantErrIs: ErrNodeIDExists},
		{name: "Error: Parent Node ID does not exist", parentID: "node-999", newNodeID: "fail-node", nodeType: "string", nodeValue: "fail", index: nil, expectError: true, wantErrIs: ErrNotFound},
		{name: "Error: Empty New Node ID", parentID: rootID, newNodeID: "", nodeType: "string", nodeValue: "fail", index: nil, expectError: true, wantErrIs: ErrInvalidArgument},
		{name: "Error: Invalid Node Type", parentID: rootID, newNodeID: "fail-node-type", nodeType: "invalid-type", nodeValue: "fail", index: nil, expectError: true, wantErrIs: ErrInvalidArgument},
		{name: "Error: Index out of bounds (negative)", parentID: rootArrID, newNodeID: "fail-idx-neg", nodeType: "string", nodeValue: "fail", index: Pint(-5), expectError: false}, // Should append if negative index
		{name: "Error: Index out of bounds (too large)", parentID: rootArrID, newNodeID: "fail-idx-large", nodeType: "string", nodeValue: "fail", index: Pint(100), expectError: true, wantErrIs: ErrListIndexOutOfBounds},
		// UPDATED Error Expectation for Invalid Handle
		{name: "Error: Invalid Handle", parentID: rootID, newNodeID: "fail-handle", nodeType: "string", nodeValue: "fail", index: nil, expectError: true, wantErrIs: ErrInvalidArgument},
	}

	// Execute tests sequentially as they modify the tree state
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentHandle := handle
			if tc.name == "Error: Invalid Handle" {
				currentHandle = "bad-handle"
			}

			args := []interface{}{currentHandle, tc.parentID, tc.newNodeID, tc.nodeType}
			args = append(args, tc.nodeValue)
			if tc.index != nil {
				args = append(args, *tc.index)
			} else {
				args = append(args, nil)
			}

			_, err := toolTreeAddNode(interp, args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.wantErrIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					idxVal := -1
					if tc.index != nil {
						idxVal = *tc.index
						if idxVal < 0 {
							idxVal = -1
						}
					}
					verifyAddedNode(t, tc.parentID, tc.newNodeID, tc.nodeType, tc.nodeValue, idxVal)
				}
			}
		})
	}
}

// --- TestTreeRemoveNodeStd ---

func TestTreeRemoveNodeStd(t *testing.T) {
	jsonFixture := `{ "rootKey": "rootValue", "leafToRemove": "simple leaf", "emptyObjectToRemove": {}, "emptyArrayToRemove": [], "objectToRemove": { "child1": "objChildVal1", "child2": [ "objChildArrVal" ] }, "arrayToRemove": [ "arrChildVal1", { "arrChildKey": "arrChildObjVal" } ] }`

	var interp *Interpreter
	var handle string
	var rootID string
	var leafID, emptyObjID, emptyArrID, objID, objChild1ID, objChild2ID, objChildArrValID, arrID, arrChild1ID, arrChild2ID, arrChildObjValID string

	// Helper to reset state and get IDs
	setupSubTest := func(t *testing.T) {
		t.Helper()
		interp, handle = setupTreeTestStd(t, jsonFixture)
		rootNode := getNodeHelper(t, interp, handle, "node-1")
		rootID = "node-1"
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if rootAttrs == nil {
			t.Fatalf("Cannot get root attributes in setupSubTest helper")
		}
		leafID = getAttrNodeID(t, rootAttrs, "leafToRemove")
		emptyObjID = getAttrNodeID(t, rootAttrs, "emptyObjectToRemove")
		emptyArrID = getAttrNodeID(t, rootAttrs, "emptyArrayToRemove")
		objID = getAttrNodeID(t, rootAttrs, "objectToRemove")
		arrID = getAttrNodeID(t, rootAttrs, "arrayToRemove")
		objNode := getNodeHelper(t, interp, handle, objID)
		objAttrs := getNodeAttributesStd(t, objNode)
		objChild1ID = getAttrNodeID(t, objAttrs, "child1")
		objChild2ID = getAttrNodeID(t, objAttrs, "child2")
		objChild2Node := getNodeHelper(t, interp, handle, objChild2ID)
		objChild2Children := getNodeChildrenStd(t, objChild2Node)
		if len(objChild2Children) > 0 {
			objChildArrValID = getChildNodeID(t, objChild2Children, 0)
		} else {
			t.Logf("Warning: objChild2 node %s had no children initially", objChild2ID)
		}
		arrNode := getNodeHelper(t, interp, handle, arrID)
		arrChildren := getNodeChildrenStd(t, arrNode)
		arrChild1ID = getChildNodeID(t, arrChildren, 0)
		arrChild2ID = getChildNodeID(t, arrChildren, 1)
		arrChild2Node := getNodeHelper(t, interp, handle, arrChild2ID)
		arrChild2Attrs := getNodeAttributesStd(t, arrChild2Node)
		arrChildObjValID = getAttrNodeID(t, arrChild2Attrs, "arrChildKey") // ID of the value node associated with the key
	}

	nodeExists := func(t *testing.T, nodeID string) bool {
		t.Helper()
		_, _, err := getNodeFromHandle(interp, handle, nodeID, "nodeExistsCheck")
		if err == nil {
			return true
		}
		if errors.Is(err, ErrNotFound) {
			return false
		}
		t.Errorf("nodeExists: Unexpected error checking for node %q: %v", nodeID, err)
		return false
	}

	// --- Subtests ---
	t.Run("RemoveLeafNode", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, leafID))
		if err != nil {
			t.Fatalf("Failed to remove leaf node %q: %v", leafID, err)
		}
		if nodeExists(t, leafID) {
			t.Errorf("Leaf node %q still exists", leafID)
		}
		rootNode := getNodeHelper(t, interp, handle, rootID)
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if _, exists := rootAttrs["leafToRemove"]; exists {
			t.Errorf("Parent node %q still has attribute 'leafToRemove'", rootID)
		}
	})
	t.Run("RemoveEmptyObject", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, emptyObjID))
		if err != nil {
			t.Fatalf("Failed to remove empty object node %q: %v", emptyObjID, err)
		}
		if nodeExists(t, emptyObjID) {
			t.Errorf("Empty object node %q still exists", emptyObjID)
		}
		rootNode := getNodeHelper(t, interp, handle, rootID)
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if _, exists := rootAttrs["emptyObjectToRemove"]; exists {
			t.Errorf("Parent node %q still has attribute 'emptyObjectToRemove'", rootID)
		}
	})
	t.Run("RemoveEmptyArray", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, emptyArrID))
		if err != nil {
			t.Fatalf("Failed to remove empty array node %q: %v", emptyArrID, err)
		}
		if nodeExists(t, emptyArrID) {
			t.Errorf("Empty array node %q still exists", emptyArrID)
		}
		rootNode := getNodeHelper(t, interp, handle, rootID)
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if _, exists := rootAttrs["emptyArrayToRemove"]; exists {
			t.Errorf("Parent node %q still has attribute 'emptyArrayToRemove'", rootID)
		}
	})
	t.Run("RemoveObjectWithChildren", func(t *testing.T) {
		setupSubTest(t)
		if !nodeExists(t, objChild1ID) {
			t.Fatalf("Pre-check fail: objChild1ID %q missing", objChild1ID)
		}
		if !nodeExists(t, objChild2ID) {
			t.Fatalf("Pre-check fail: objChild2ID %q missing", objChild2ID)
		}
		if !nodeExists(t, objChildArrValID) {
			t.Fatalf("Pre-check fail: objChildArrValID %q missing", objChildArrValID)
		}
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, objID))
		if err != nil {
			t.Fatalf("Failed to remove object node %q: %v", objID, err)
		}
		if nodeExists(t, objID) {
			t.Errorf("Object node %q still exists", objID)
		}
		if nodeExists(t, objChild1ID) {
			t.Errorf("Descendant node %q still exists", objChild1ID)
		}
		if nodeExists(t, objChild2ID) {
			t.Errorf("Descendant node %q still exists", objChild2ID)
		}
		if nodeExists(t, objChildArrValID) {
			t.Errorf("Descendant node %q still exists", objChildArrValID)
		}
		rootNode := getNodeHelper(t, interp, handle, rootID)
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if _, exists := rootAttrs["objectToRemove"]; exists {
			t.Errorf("Parent node %q still has attribute 'objectToRemove'", rootID)
		}
	})
	t.Run("RemoveArrayWithChildren", func(t *testing.T) {
		setupSubTest(t)
		if !nodeExists(t, arrChild1ID) {
			t.Fatalf("Pre-check fail: arrChild1ID %q missing", arrChild1ID)
		}
		if !nodeExists(t, arrChild2ID) {
			t.Fatalf("Pre-check fail: arrChild2ID %q missing", arrChild2ID)
		}
		if !nodeExists(t, arrChildObjValID) {
			t.Fatalf("Pre-check fail: arrChildObjValID %q missing", arrChildObjValID)
		}
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, arrID))
		if err != nil {
			t.Fatalf("Failed to remove array node %q: %v", arrID, err)
		}
		if nodeExists(t, arrID) {
			t.Errorf("Array node %q still exists", arrID)
		}
		if nodeExists(t, arrChild1ID) {
			t.Errorf("Descendant node %q still exists", arrChild1ID)
		}
		if nodeExists(t, arrChild2ID) {
			t.Errorf("Descendant node %q still exists", arrChild2ID)
		}
		if nodeExists(t, arrChildObjValID) {
			t.Errorf("Descendant node %q still exists", arrChildObjValID)
		}
		rootNode := getNodeHelper(t, interp, handle, rootID)
		rootAttrs := getNodeAttributesStd(t, rootNode)
		if _, exists := rootAttrs["arrayToRemove"]; exists {
			t.Errorf("Parent node %q still has attribute 'arrayToRemove'", rootID)
		}
	})
	t.Run("Error_RemoveRoot", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, rootID))
		if !errors.Is(err, ErrCannotRemoveRoot) {
			t.Errorf("Expected ErrCannotRemoveRoot, got %v", err)
		}
	})
	t.Run("Error_RemoveNonExistent", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs(handle, "node-does-not-exist"))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
	t.Run("Error_InvalidHandle", func(t *testing.T) {
		setupSubTest(t)
		_, err := toolTreeRemoveNode(interp, MakeArgs("bad-handle", leafID))
		// UPDATED Error Expectation for Invalid Handle
		if !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument for bad handle, got %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrInvalidArgument for bad handle: %v", err)
		}
	})
}

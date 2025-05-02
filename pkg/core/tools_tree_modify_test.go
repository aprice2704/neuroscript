// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:59:05 PDT // Fix: Expect ErrCacheObjectWrongType for bad handle prefix
// filename: pkg/core/tools_tree_modify_test.go

package core

import (
	"errors"
	"reflect" // Needed for DeepEqual in verification
	"testing"
)

// --- Test Setup Helper (using standard testing) ---

// setupTreeTest loads JSON and returns interpreter + handle. Fails test on error.
func setupTreeTestStd(t *testing.T, jsonStr string) (*Interpreter, string) {
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

// getNodeValueStd is a helper to get the 'value' field using standard testing checks
func getNodeValueStd(t *testing.T, nodeMap map[string]interface{}) interface{} {
	t.Helper()
	val, exists := nodeMap["value"]
	if !exists {
		t.Fatalf("getNodeValueStd: 'value' field missing from node map: %v", nodeMap)
	}
	return val
}

// getNodeAttributesStd is a helper to get the 'attributes' map using standard testing checks
func getNodeAttributesStd(t *testing.T, nodeMap map[string]interface{}) map[string]interface{} {
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
		// *** FIXED: Expect ErrCacheObjectWrongType for bad handle prefix ***
		{name: "Error: Invalid Handle", nodeID: stringNodeID, modifications: map[string]interface{}{"value": 1}, expectError: true, wantErrIs: ErrCacheObjectWrongType, skipVerify: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Error: Invalid Handle" {
				_, err := toolTreeModifyNode(interp, MakeArgs("bad-handle", tc.nodeID, tc.modifications))
				if err == nil {
					t.Errorf("Expected an error for bad handle, got nil")
				} else if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error type [%v] for bad handle, got: %v", tc.wantErrIs, err) // Use specific error
				} else {
					t.Logf("Got expected error for bad handle: %v", err)
				}
				return
			}

			_, err := toolTreeModifyNode(interp, MakeArgs(handle, tc.nodeID, tc.modifications))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error type [%v], got: %v", tc.wantErrIs, err)
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
	jsonFixture := `{ "targetObject": {}, "valueNode1": "hello", "valueNode2": 100, "nonObjectNode": "i am string" }`
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
			name: "Add first attribute", nodeID: targetObjID, attrKey: "firstKey", childNodeID: val1ID, expectError: false,
			verifyFunc: func(t *testing.T) { /* ... verification ... */
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 1 {
					t.Fatalf("Should have 1 attribute, got: %v", attrs)
				}
				if val, ok := attrs["firstKey"].(string); !ok || val != val1ID {
					t.Errorf("Attribute 'firstKey' has wrong value. got = %v (%T), want = %v", attrs["firstKey"], attrs["firstKey"], val1ID)
				}
			},
		},
		{
			name: "Add second attribute", nodeID: targetObjID, attrKey: "secondKey", childNodeID: val2ID, expectError: false,
			verifyFunc: func(t *testing.T) { /* ... verification ... */
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should have 2 attributes, got: %v", attrs)
				}
				if val, ok := attrs["firstKey"].(string); !ok || val != val1ID {
					t.Errorf("Attribute 'firstKey' changed unexpectedly. got = %v (%T), want = %v", attrs["firstKey"], attrs["firstKey"], val1ID)
				}
				if val, ok := attrs["secondKey"].(string); !ok || val != val2ID {
					t.Errorf("Attribute 'secondKey' has wrong value. got = %v (%T), want = %v", attrs["secondKey"], attrs["secondKey"], val2ID)
				}
			},
		},
		{
			name: "Update existing attribute", nodeID: targetObjID, attrKey: "firstKey", childNodeID: val2ID, expectError: false,
			verifyFunc: func(t *testing.T) { /* ... verification ... */
				node := getNodeHelper(t, interp, handle, targetObjID)
				attrs := getNodeAttributesStd(t, node)
				if attrs == nil || len(attrs) != 2 {
					t.Fatalf("Should still have 2 attributes, got: %v", attrs)
				}
				if val, ok := attrs["firstKey"].(string); !ok || val != val2ID {
					t.Errorf("Attribute 'firstKey' was not updated correctly. got = %v (%T), want = %v", attrs["firstKey"], attrs["firstKey"], val2ID)
				}
				if val, ok := attrs["secondKey"].(string); !ok || val != val2ID {
					t.Errorf("Attribute 'secondKey' changed unexpectedly. got = %v (%T), want = %v", attrs["secondKey"], attrs["secondKey"], val2ID)
				}
			},
		},
		{name: "Error: Target node not object", nodeID: nonObjID, attrKey: "testKey", childNodeID: val1ID, expectError: true, wantErrIs: ErrTreeNodeNotObject, skipVerify: true},
		{name: "Error: Empty attribute key", nodeID: targetObjID, attrKey: "", childNodeID: val1ID, expectError: true, wantErrIs: ErrInvalidArgument, skipVerify: true},
		{name: "Error: Child node ID does not exist", nodeID: targetObjID, attrKey: "testKey", childNodeID: "node-999", expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		{name: "Error: Invalid Node ID", nodeID: "node-invalid", attrKey: "test", childNodeID: val1ID, expectError: true, wantErrIs: ErrNotFound, skipVerify: true},
		// *** FIXED: Expect ErrCacheObjectWrongType for bad handle prefix ***
		{name: "Error: Invalid Handle", skipVerify: true, expectError: true, wantErrIs: ErrCacheObjectWrongType}, // Changed expected error
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Error: Invalid Handle" {
				_, err := toolTreeSetAttribute(interp, MakeArgs("bad-handle", targetObjID, "testKey", val1ID))
				if err == nil {
					t.Errorf("Expected an error for bad handle, got nil")
					// *** FIXED: Check against tc.wantErrIs ***
				} else if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error type [%v] for bad handle, got: %v", tc.wantErrIs, err)
				} else {
					t.Logf("Got expected error for bad handle: %v", err)
				}
				return
			}

			_, err := toolTreeSetAttribute(interp, MakeArgs(handle, tc.nodeID, tc.attrKey, tc.childNodeID))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error type [%v], got: %v", tc.wantErrIs, err)
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

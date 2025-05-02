// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 22:14:05 PDT // Reformat loadJSONHelper to fix syntax error
// filename: pkg/core/tools_tree_test.go

package core

import (
	"errors"
	"strings"
	"testing"
)

// --- Test Helpers ---

// Helper to load JSON and return handle (fails test on error)
func loadJSONHelper(t *testing.T, interp *Interpreter, jsonStr string) string {
	t.Helper()
	handle, err := toolTreeLoadJSON(interp, MakeArgs(jsonStr))
	if err != nil {
		t.Fatalf("loadJSONHelper failed: %v", err)
	}
	handleStr, ok := handle.(string)
	if !ok || handleStr == "" {
		t.Fatalf("loadJSONHelper bad handle: %T %v", handle, handle)
	}
	// Check prefix on a separate line for clarity
	if !strings.Contains(handleStr, GenericTreeHandleType+"::") {
		t.Logf("Warning: Handle format might have changed. Expected prefix %q, got %q", GenericTreeHandleType+"::", handleStr)
	} // <<< Semicolon removed, closing brace on newline (Go style)
	return handleStr
}

// --- (Other helpers getNodeHelper, getNodeAttributes, etc. remain unchanged) ---
func getNodeHelper(t *testing.T, interp *Interpreter, handle, nodeID string) map[string]interface{} {
	t.Helper()
	nodeData, err := toolTreeGetNode(interp, MakeArgs(handle, nodeID))
	if err != nil {
		t.Fatalf("getNodeHelper failed for %s/%s: %v", handle, nodeID, err)
	}
	nodeMap, ok := nodeData.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeHelper bad return type for %s/%s: %T", handle, nodeID, nodeData)
	}
	return nodeMap
}
func getNodeAttributes(t *testing.T, nodeData map[string]interface{}) map[string]interface{} {
	t.Helper()
	attrs, ok := nodeData["attributes"].(map[string]interface{})
	if !ok {
		t.Fatalf("Could not get attributes map: %v", nodeData)
	}
	return attrs
}
func getNodeChildren(t *testing.T, nodeData map[string]interface{}) []interface{} {
	t.Helper()
	children, ok := nodeData["children"].([]interface{})
	if !ok {
		t.Fatalf("Could not get children slice: %v", nodeData)
	}
	return children
}
func getAttrNodeID(t *testing.T, attrs map[string]interface{}, key string) string {
	t.Helper()
	nodeIDIntf, ok := attrs[key]
	if !ok {
		t.Fatalf("Attribute key %q not found: %v", key, attrs)
	}
	nodeID, ok := nodeIDIntf.(string)
	if !ok {
		t.Fatalf("Attribute value for key %q is not string node ID: %T", key, nodeIDIntf)
	}
	return nodeID
}
func getChildNodeID(t *testing.T, children []interface{}, index int) string {
	t.Helper()
	if index < 0 || index >= len(children) {
		t.Fatalf("Index %d out of bounds for children (len %d)", index, len(children))
	}
	nodeIDIntf := children[index]
	nodeID, ok := nodeIDIntf.(string)
	if !ok {
		t.Fatalf("Child at index %d is not string node ID: %T", index, nodeIDIntf)
	}
	return nodeID
}

// --- Tests ---

func TestTreeLoadJSON(t *testing.T) {
	// --- (TestTreeLoadJSON remains unchanged from previous fixed version) ---
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name        string
		jsonInput   string
		expectError bool
		wantErrIs   error
	}{
		{"Simple Object", `{"key": "value", "num": 123}`, false, nil},
		{"Simple Array", `[1, "two", true]`, false, nil},
		{"Nested Structure", `{"a": [1, {"b": null}], "c": true}`, false, nil},
		{"Empty Object", `{}`, false, nil},
		{"Empty Array", `[]`, false, nil},
		{"Just String", `"hello"`, false, nil},
		{"Just Number", `123.45`, false, nil},
		{"Just Boolean", `true`, false, nil},
		{"Just Null", `null`, false, nil},
		{"Invalid JSON", `{"key": "value`, true, ErrTreeJSONUnmarshal},
		{"Empty Input", ``, true, ErrTreeJSONUnmarshal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle, err := toolTreeLoadJSON(interp, MakeArgs(tt.jsonInput))
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, got nil")
				} else if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Expected error type [%v], got error: %v", tt.wantErrIs, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				handleStr, ok := handle.(string)
				if !ok || !strings.Contains(handleStr, GenericTreeHandleType+"::") {
					t.Errorf("Expected valid handle string, got %T: %v", handle, handle)
				} else {
					_, getErr := interp.GetHandleValue(handleStr, GenericTreeHandleType)
					if getErr != nil {
						t.Errorf("Failed to retrieve handle %q: %v", handleStr, getErr)
					} else {
						t.Logf("OK handle %q", handleStr)
					}
				}
			}
		})
	}
	t.Run("Validation_Wrong_Arg_Type", func(t *testing.T) {
		_, err := toolTreeLoadJSON(interp, MakeArgs(123))
		if err == nil {
			t.Error("Expected validation error, got nil")
		} else if !errors.Is(err, ErrValidationTypeMismatch) {
			t.Errorf("Expected ErrValidationTypeMismatch, got %v", err)
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})
}

func TestTreeNavigation(t *testing.T) {
	// --- (TestTreeNavigation remains unchanged from previous fixed version) ---
	interp, _ := NewDefaultTestInterpreter(t)
	jsonInput := `{ "z_name": "root", "a_enabled": true, "items": [ {"v_num": 10, "x_id": "item1"}, {"x_id": "item2", "v_num": 20} ], "b_config": { "y_nested": "yes" } }`
	handle := loadJSONHelper(t, interp, jsonInput)
	rootID := "node-1"
	rootNode := getNodeHelper(t, interp, handle, rootID)
	rootAttrs := getNodeAttributes(t, rootNode)

	t.Run("GetNode_Root", func(t *testing.T) {
		if rootNode["type"] != "object" {
			t.Errorf("Bad root type")
		}
		if rootNode["id"] != rootID {
			t.Errorf("Bad root id")
		}
		if rootNode["parentId"] != "" {
			t.Errorf("Bad root parentId")
		}
		if len(rootAttrs) != 4 {
			t.Errorf("Bad root attr count: %d", len(rootAttrs))
		}
	})
	t.Run("GetNode_LeafNodes", func(t *testing.T) {
		nameNodeID := getAttrNodeID(t, rootAttrs, "z_name")
		nameNode := getNodeHelper(t, interp, handle, nameNodeID)
		if nameNode["type"] != "string" || nameNode["value"] != "root" {
			t.Errorf("Bad name node: %v", nameNode)
		}
		if nameNode["parentId"] != rootID {
			t.Errorf("Bad name parent")
		}
		enabledNodeID := getAttrNodeID(t, rootAttrs, "a_enabled")
		enabledNode := getNodeHelper(t, interp, handle, enabledNodeID)
		if enabledNode["type"] != "boolean" || enabledNode["value"] != true {
			t.Errorf("Bad enabled node: %v", enabledNode)
		}
		if enabledNode["parentId"] != rootID {
			t.Errorf("Bad enabled parent")
		}
	})
	t.Run("GetNode_Array_And_Items", func(t *testing.T) {
		itemsNodeID := getAttrNodeID(t, rootAttrs, "items")
		itemsNode := getNodeHelper(t, interp, handle, itemsNodeID)
		if itemsNode["type"] != "array" {
			t.Errorf("Bad items type")
		}
		if itemsNode["parentId"] != rootID {
			t.Errorf("Bad items parent")
		}
		itemsChildren := getNodeChildren(t, itemsNode)
		if len(itemsChildren) != 2 {
			t.Fatalf("Bad items child count: %d", len(itemsChildren))
		}
		item0NodeID := getChildNodeID(t, itemsChildren, 0)
		item0Node := getNodeHelper(t, interp, handle, item0NodeID)
		if item0Node["type"] != "object" {
			t.Errorf("Bad item0 type")
		}
		item0Attrs := getNodeAttributes(t, item0Node)
		if len(item0Attrs) != 2 {
			t.Errorf("Bad item0 attr count: %d", len(item0Attrs))
		}
		item0ValNodeID := getAttrNodeID(t, item0Attrs, "v_num")
		item0ValNode := getNodeHelper(t, interp, handle, item0ValNodeID)
		if item0ValNode["type"] != "number" || item0ValNode["value"] != 10.0 {
			t.Errorf("Bad item0 value node: %v", item0ValNode)
		}
		item0IdNodeID := getAttrNodeID(t, item0Attrs, "x_id")
		item0IdNode := getNodeHelper(t, interp, handle, item0IdNodeID)
		if item0IdNode["type"] != "string" || item0IdNode["value"] != "item1" {
			t.Errorf("Bad item0 id node: %v", item0IdNode)
		}
		item1NodeID := getChildNodeID(t, itemsChildren, 1)
		item1Node := getNodeHelper(t, interp, handle, item1NodeID)
		if item1Node["type"] != "object" {
			t.Errorf("Bad item1 type")
		}
		item1Attrs := getNodeAttributes(t, item1Node)
		if len(item1Attrs) != 2 {
			t.Errorf("Bad item1 attr count: %d", len(item1Attrs))
		}
		item1ValNodeID := getAttrNodeID(t, item1Attrs, "v_num")
		item1ValNode := getNodeHelper(t, interp, handle, item1ValNodeID)
		if item1ValNode["type"] != "number" || item1ValNode["value"] != 20.0 {
			t.Errorf("Bad item1 value node: %v", item1ValNode)
		}
		item1IdNodeID := getAttrNodeID(t, item1Attrs, "x_id")
		item1IdNode := getNodeHelper(t, interp, handle, item1IdNodeID)
		if item1IdNode["type"] != "string" || item1IdNode["value"] != "item2" {
			t.Errorf("Bad item1 id node: %v", item1IdNode)
		}
	})
	t.Run("GetNode_ConfigObject_And_Leaf", func(t *testing.T) {
		configNodeID := getAttrNodeID(t, rootAttrs, "b_config")
		configNode := getNodeHelper(t, interp, handle, configNodeID)
		if configNode["type"] != "object" {
			t.Errorf("Bad config type")
		}
		configAttrs := getNodeAttributes(t, configNode)
		if len(configAttrs) != 1 {
			t.Errorf("Bad config attr count: %d", len(configAttrs))
		}
		nestedNodeID := getAttrNodeID(t, configAttrs, "y_nested")
		nestedNode := getNodeHelper(t, interp, handle, nestedNodeID)
		if nestedNode["type"] != "string" || nestedNode["value"] != "yes" {
			t.Errorf("Bad nested node: %v", nestedNode)
		}
	})
	t.Run("GetNode_NonExistent", func(t *testing.T) {
		_, err := toolTreeGetNode(interp, MakeArgs(handle, "node-999"))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
	t.Run("GetNode_InvalidHandle", func(t *testing.T) {
		_, err := toolTreeGetNode(interp, MakeArgs("bad-handle", rootID))
		if err == nil {
			t.Error("Expected error for invalid handle, got nil")
		} else {
			t.Logf("Got expected error for invalid handle: %v", err)
		}
	})

	t.Run("GetChildren_RootObject", func(t *testing.T) {
		childrenData, err := toolTreeGetChildren(interp, MakeArgs(handle, rootID))
		if err != nil {
			t.Fatalf("GetChildren failed for root: %v", err)
		}
		children, ok := childrenData.([]interface{})
		if !ok {
			t.Fatalf("Expected slice, got %T", childrenData)
		}
		if len(children) != 0 {
			t.Errorf("Expected 0 children, got %d", len(children))
		}
	})
	t.Run("GetChildren_Array", func(t *testing.T) {
		itemsNodeID := getAttrNodeID(t, rootAttrs, "items")
		childrenData, err := toolTreeGetChildren(interp, MakeArgs(handle, itemsNodeID))
		if err != nil {
			t.Fatalf("GetChildren failed for items node: %v", err)
		}
		children := childrenData.([]interface{})
		if len(children) != 2 {
			t.Fatalf("Expected 2 children, got %d", len(children))
		}
		if _, ok := children[0].(string); !ok {
			t.Error("Child 0 not string ID")
		}
		if _, ok := children[1].(string); !ok {
			t.Error("Child 1 not string ID")
		}
	})
	t.Run("GetChildren_Leaf", func(t *testing.T) {
		nameNodeID := getAttrNodeID(t, rootAttrs, "z_name")
		childrenData, err := toolTreeGetChildren(interp, MakeArgs(handle, nameNodeID))
		if err != nil {
			t.Fatalf("GetChildren failed for leaf node: %v", err)
		}
		children := childrenData.([]interface{})
		if len(children) != 0 {
			t.Errorf("Expected 0 children, got %d", len(children))
		}
	})
	t.Run("GetChildren_NonExistent", func(t *testing.T) {
		_, err := toolTreeGetChildren(interp, MakeArgs(handle, "node-999"))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("GetParent_Root", func(t *testing.T) {
		parentData, err := toolTreeGetParent(interp, MakeArgs(handle, rootID))
		if err != nil {
			t.Fatalf("GetParent failed for root: %v", err)
		}
		parentID := parentData.(string)
		if parentID != "" {
			t.Errorf("Expected empty parent ID, got %q", parentID)
		}
	})
	t.Run("GetParent_Level1", func(t *testing.T) {
		nameNodeID := getAttrNodeID(t, rootAttrs, "z_name")
		parentData, err := toolTreeGetParent(interp, MakeArgs(handle, nameNodeID))
		if err != nil {
			t.Fatalf("GetParent failed for name node: %v", err)
		}
		parentID := parentData.(string)
		if parentID != rootID {
			t.Errorf("Expected parent %q, got %q", rootID, parentID)
		}
	})
	t.Run("GetParent_ArrayElement", func(t *testing.T) {
		itemsNodeID := getAttrNodeID(t, rootAttrs, "items")
		itemsNode := getNodeHelper(t, interp, handle, itemsNodeID)
		itemsChildren := getNodeChildren(t, itemsNode)
		item0NodeID := getChildNodeID(t, itemsChildren, 0)
		parentData, err := toolTreeGetParent(interp, MakeArgs(handle, item0NodeID))
		if err != nil {
			t.Fatalf("GetParent failed for item0 node: %v", err)
		}
		parentID := parentData.(string)
		if parentID != itemsNodeID {
			t.Errorf("Expected parent %q (array node), got %q", itemsNodeID, parentID)
		}
	})
	t.Run("GetParent_Deeper", func(t *testing.T) {
		itemsNodeID := getAttrNodeID(t, rootAttrs, "items")
		itemsNode := getNodeHelper(t, interp, handle, itemsNodeID)
		itemsChildren := getNodeChildren(t, itemsNode)
		item0NodeID := getChildNodeID(t, itemsChildren, 0)
		item0Node := getNodeHelper(t, interp, handle, item0NodeID)
		item0Attrs := getNodeAttributes(t, item0Node)
		item0IdNodeID := getAttrNodeID(t, item0Attrs, "x_id")
		parentData, err := toolTreeGetParent(interp, MakeArgs(handle, item0IdNodeID))
		if err != nil {
			t.Fatalf("GetParent failed for item0 id node: %v", err)
		}
		parentID := parentData.(string)
		if parentID != item0NodeID {
			t.Errorf("Expected parent %q (item0 object), got %q", item0NodeID, parentID)
		}
	})
	t.Run("GetParent_NonExistent", func(t *testing.T) {
		_, err := toolTreeGetParent(interp, MakeArgs(handle, "node-999"))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

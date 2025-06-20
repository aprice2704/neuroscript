// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Corrected type assertions to use the specific `TreeAttrs` type instead of a generic map.
// nlines: 60
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_metadata_test.go

package core

import (
	"testing"
)

func TestTreeMetadataTools(t *testing.T) {
	jsonSimple := `{"key":"value"}` // Root node (node-1) type: object, attributes: {"key": "node-2"}, node-2 type: string, value: "value"
	setupMetaTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple)
	}

	testCases := []treeTestCase{
		// Tree.SetNodeMetadata
		{name: "SetNodeMetadata New Key", toolName: "Tree.SetNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "metaKey1", "metaValue1"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("SetNodeMetadata failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, _ := callGetNode(t, interp, handle, "node-1")
				// Metadata is stored in node.Attributes for GenericTree
				attrs, ok := nodeMap["attributes"].(TreeAttrs)
				if !ok {
					t.Fatalf("Node attributes are not TreeAttrs, got %T", nodeMap["attributes"])
				}
				if val, vOK := attrs["metaKey1"].(string); !vOK || val != "metaValue1" {
					t.Errorf("Metadata not set correctly. Got: %v, expected 'metaValue1'", attrs["metaKey1"])
				}
			}},
		{name: "SetNodeMetadata Empty Key", toolName: "Tree.SetNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "", "val"), wantErr: ErrInvalidArgument}, // Empty key for metadata
		{name: "SetNodeMetadata NonExistent Node", toolName: "Tree.SetNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-999", "key", "val"), wantErr: ErrNotFound},

		// Tree.RemoveNodeMetadata
		{name: "RemoveNodeMetadata Existing Key", toolName: "Tree.RemoveNodeMetadata",
			setupFunc: func(t *testing.T, interp *Interpreter) interface{} {
				handle := setupTreeWithJSON(t, interp, jsonSimple)
				// Set a metadata key to remove
				err := callSetMetadata(t, interp, handle, "node-1", "toRemove", "val")
				if err != nil {
					t.Fatalf("Setup failed for RemoveNodeMetadata: %v", err)
				}
				return handle
			},
			args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "toRemove"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("RemoveNodeMetadata failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, _ := callGetNode(t, interp, handle, "node-1")
				attrs, ok := nodeMap["attributes"].(TreeAttrs)
				if !ok {
					t.Fatalf("Node attributes are not TreeAttrs, got %T", nodeMap["attributes"])
				}
				if _, exists := attrs["toRemove"]; exists {
					t.Errorf("Metadata key 'toRemove' still exists after removal.")
				}
			}},
		{name: "RemoveNodeMetadata NonExistent Key", toolName: "Tree.RemoveNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-1", "nonKey"), wantErr: ErrAttributeNotFound}, // Metadata key not found
		{name: "RemoveNodeMetadata NonExistent Node", toolName: "Tree.RemoveNodeMetadata", setupFunc: setupMetaTree, args: MakeArgs("SETUP_HANDLE:mTree", "node-999", "key"), wantErr: ErrNotFound},
	}
	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}

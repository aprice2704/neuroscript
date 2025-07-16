// NeuroScript Version: 0.6.5
// File version: 1
// Purpose: Adds test coverage for object attribute and node metadata manipulation tools.
// filename: pkg/tool/tree/tools_tree_attributes_test.go
// nlines: 105
// risk_rating: LOW

package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeObjectAttributes(t *testing.T) {
	const testJSON = `{"parentNode": {}, "childNode": "a value"}`

	testTreeToolHelper(t, "Set and Remove Object Attribute", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		parentID, err := getNodeIDByPath(t, interp, handle, "parentNode")
		if err != nil {
			t.Fatal(err)
		}
		childID, err := getNodeIDByPath(t, interp, handle, "childNode")
		if err != nil {
			t.Fatal(err)
		}

		// 1. Set the attribute
		_, err = runTool(t, interp, "SetObjectAttribute", handle, parentID, "childLink", childID)
		if err != nil {
			t.Fatalf("SetObjectAttribute failed unexpectedly: %v", err)
		}

		// 2. Verify the attribute was set
		parentNodeData, err := callGetNode(t, interp, handle, parentID)
		if err != nil {
			t.Fatal(err)
		}
		attrs := parentNodeData.(map[string]interface{})["attributes"].(map[string]interface{})
		if linkedID, ok := attrs["childLink"]; !ok || linkedID != childID {
			t.Fatalf("attribute 'childLink' was not set correctly. Got: %v", attrs)
		}

		// 3. Remove the attribute
		_, err = runTool(t, interp, "RemoveObjectAttribute", handle, parentID, "childLink")
		if err != nil {
			t.Fatalf("RemoveObjectAttribute failed unexpectedly: %v", err)
		}

		// 4. Verify the attribute was removed
		parentNodeData, err = callGetNode(t, interp, handle, parentID)
		if err != nil {
			t.Fatal(err)
		}
		attrs = parentNodeData.(map[string]interface{})["attributes"].(map[string]interface{})
		if _, ok := attrs["childLink"]; ok {
			t.Fatal("attribute 'childLink' was not removed.")
		}
	})

	testTreeToolHelper(t, "Attribute Errors", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		parentID, err := getNodeIDByPath(t, interp, handle, "parentNode")
		if err != nil {
			t.Fatal(err)
		}
		childID, err := getNodeIDByPath(t, interp, handle, "childNode")
		if err != nil {
			t.Fatal(err)
		}

		// Error on removing a key that doesn't exist
		_, err = runTool(t, interp, "RemoveObjectAttribute", handle, parentID, "nonexistentKey")
		assertResult(t, nil, err, nil, lang.ErrAttributeNotFound)

		// Error on setting an attribute on a non-object node
		_, err = runTool(t, interp, "SetObjectAttribute", handle, childID, "childLink", parentID)
		assertResult(t, nil, err, nil, lang.ErrTreeNodeNotObject)
	})
}

func TestTreeMetadataAttributes(t *testing.T) {
	const testJSON = `{"node": "some value"}`

	testTreeToolHelper(t, "Remove Node Metadata", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		nodeID, err := getNodeIDByPath(t, interp, handle, "node")
		if err != nil {
			t.Fatal(err)
		}

		// 1. Set metadata
		_, err = callSetNodeMetadata(t, interp, handle, nodeID, "status", "active")
		if err != nil {
			t.Fatalf("SetNodeMetadata failed unexpectedly: %v", err)
		}

		// 2. Remove it
		_, err = runTool(t, interp, "RemoveNodeMetadata", handle, nodeID, "status")
		if err != nil {
			t.Fatalf("RemoveNodeMetadata failed unexpectedly: %v", err)
		}

		// 3. Verify it's gone
		metadata, err := callGetMetadata(t, interp, handle, nodeID)
		if err != nil {
			t.Fatal(err)
		}
		metaMap := metadata.(map[string]interface{})
		if _, ok := metaMap["status"]; ok {
			t.Fatal("metadata 'status' was not removed")
		}

		// 4. Error on removing a key that doesn't exist
		_, err = runTool(t, interp, "RemoveNodeMetadata", handle, nodeID, "nonexistentKey")
		assertResult(t, nil, err, nil, lang.ErrAttributeNotFound)
	})
}

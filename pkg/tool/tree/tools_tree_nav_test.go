// NeuroScript Version: 0.6.5
// File version: 17
// Purpose: Corrected assertions and logic to handle the map[string]interface{} type returned for nodes.
// filename: pkg/tool/tree/tools_tree_nav_test.go
// nlines: 100
// risk_rating: LOW
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeNav(t *testing.T) {
	const testJSON = `{"name": "root", "ports": [80, 443], "child": {"name": "child1"}}`

	setup := func(t *testing.T, interp tool.Runtime) string {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("Tree setup failed unexpectedly: %v", err)
		}
		return handle
	}

	testTreeToolHelper(t, "Get Node Root", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		node, err := callGetNode(t, interp, handle, rootID)
		assertResult(t, nil, err, nil, nil)

		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			t.Fatalf("expected node to be a map, got %T", node)
		}
		if nodeMap["id"] != rootID {
			t.Errorf("expected node ID to be '%s', got '%s'", rootID, nodeMap["id"])
		}
	})

	testTreeToolHelper(t, "Get Node Child", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		childID, err := getNodeIDByPath(t, interp, handle, "child")
		if err != nil {
			t.Fatal(err)
		}

		node, err := callGetNode(t, interp, handle, childID)
		assertResult(t, nil, err, nil, nil)

		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			t.Fatalf("expected node to be a map, got %T", node)
		}

		nameAttrNodeID, ok := nodeMap["attributes"].(map[string]interface{})["name"].(string)
		if !ok {
			t.Fatal("could not get name attribute ID")
		}
		value, err := callGetValue(t, interp, handle, nameAttrNodeID)
		assertResult(t, value, err, "child1", nil)
	})

	testTreeToolHelper(t, "Get Node Invalid ID", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		_, err := callGetNode(t, interp, handle, "invalid-node-id")
		assertResult(t, nil, err, nil, lang.ErrNotFound)
	})

	testTreeToolHelper(t, "Get Children Of Array", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		portsID, err := getNodeIDByPath(t, interp, handle, "ports")
		if err != nil {
			t.Fatal(err)
		}

		children, err := callGetChildren(t, interp, handle, portsID)
		assertResult(t, nil, err, nil, nil)

		childrenSlice, ok := children.([]interface{})
		if !ok {
			t.Fatalf("expected children to be a slice, got %T", children)
		}
		if len(childrenSlice) != 2 {
			t.Errorf("expected 2 children, got %d", len(childrenSlice))
		}
	})

	testTreeToolHelper(t, "Get Parent", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}
		childID, err := getNodeIDByPath(t, interp, handle, "child")
		if err != nil {
			t.Fatal(err)
		}

		parent, err := runTool(t, interp, "GetParent", handle, childID)
		assertResult(t, nil, err, nil, nil)

		parentMap, ok := parent.(map[string]interface{})
		if !ok {
			t.Fatalf("expected parent to be a map, got %T", parent)
		}
		if parentMap["id"] != rootID {
			t.Errorf("expected parent ID to be '%s', got '%s'", rootID, parentMap["id"])
		}

		parentOfRoot, err := runTool(t, interp, "GetParent", handle, rootID)
		assertResult(t, parentOfRoot, err, nil, nil)
	})
}

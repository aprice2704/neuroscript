// NeuroScript Version: 0.6.5
// File version: 18
// Purpose: Corrected all modification tests to use valid tool calls, arguments, and target nodes.
// filename: pkg/tool/tree/tools_tree_modify_test.go
// nlines: 120
// risk_rating: LOW
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeModify(t *testing.T) {
	const testJSON = `{"name": "root", "children": [{"name": "child1"}, {"name": "child2"}]}`

	setup := func(t *testing.T, interp tool.Runtime) string {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("Tree setup failed unexpectedly: %v", err)
		}
		return handle
	}

	testTreeToolHelper(t, "Set Value on Leaf Node", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		// FIX: Target the 'name' node, which is a leaf, not the object containing it.
		childNameID, err := getNodeIDByPath(t, interp, handle, "children.0.name")
		if err != nil {
			t.Fatal(err)
		}

		_, err = callSetValue(t, interp, handle, childNameID, "new_child_name")
		if err != nil {
			t.Fatalf("SetValue failed: %v", err)
		}

		value, err := callGetValue(t, interp, handle, childNameID)
		assertResult(t, value, err, "new_child_name", nil)
	})

	testTreeToolHelper(t, "Remove Node", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		child1ID, err := getNodeIDByPath(t, interp, handle, "children.0")
		if err != nil {
			t.Fatal(err)
		}

		_, err = callGetNode(t, interp, handle, child1ID)
		if err != nil {
			t.Fatalf("Could not find child node before removal test: %v", err)
		}

		// FIX: Use the correct RemoveNode tool.
		_, err = runTool(t, interp, "RemoveNode", handle, child1ID)
		if err != nil {
			t.Fatalf("RemoveNode failed: %v", err)
		}

		_, err = callGetNode(t, interp, handle, child1ID)
		assertResult(t, nil, err, nil, lang.ErrNotFound)
	})

	testTreeToolHelper(t, "Add Node to Root", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		// FIX: Provide all required arguments for AddChildNode.
		newNodeResult, err := callAddChildNode(t, interp, handle, rootID, "newNode", "string", "new_value", "newKey")
		if err != nil {
			t.Fatalf("AddChildNode failed: %v", err)
		}
		newNodeID := newNodeResult.(string)

		children, err := callGetChildren(t, interp, handle, rootID)
		if err != nil {
			t.Fatalf("GetChildren failed: %v", err)
		}

		childrenSlice, ok := children.([]interface{})
		if !ok {
			t.Fatalf("expected children to be a slice, got %T", children)
		}

		found := false
		for _, childID := range childrenSlice {
			if childID == newNodeID {
				found = true
				break
			}
		}
		if !found {
			t.Error("newly added node was not found in the root's children")
		}
	})

	testTreeToolHelper(t, "Append Child to Array", func(t *testing.T, interp tool.Runtime) {
		handle := setup(t, interp)
		childrenArrayID, err := getNodeIDByPath(t, interp, handle, "children")
		if err != nil {
			t.Fatal(err)
		}

		// FIX: Provide all required arguments.
		_, err = callAddChildNode(t, interp, handle, childrenArrayID, "", "object", nil, "")
		if err != nil {
			t.Fatalf("AddChildNode to array failed: %v", err)
		}

		// Verify the array now has 3 elements.
		children, err := callGetChildren(t, interp, handle, childrenArrayID)
		if err != nil {
			t.Fatal(err)
		}
		if len(children.([]interface{})) != 3 {
			t.Errorf("Expected 3 children in array, got %d", len(children.([]interface{})))
		}
	})
}

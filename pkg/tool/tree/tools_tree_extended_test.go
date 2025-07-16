// NeuroScript Version: 0.6.5
// File version: 2
// Purpose: Corrected the assertion for the max_depth test to reflect the correct number of expected nodes.
// filename: pkg/tool/tree/tools_tree_extended_test.go
// nlines: 120
// risk_rating: LOW

package tree_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeRender(t *testing.T) {
	const testJSON = `{"name": "root", "child": {"value": 123}}`

	testTreeToolHelper(t, "Render Text Tree", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Call RenderText
		renderedText, err := runTool(t, interp, "RenderText", handle)
		if err != nil {
			t.Fatalf("RenderText failed unexpectedly: %v", err)
		}

		renderedStr, ok := renderedText.(string)
		if !ok {
			t.Fatalf("RenderText did not return a string, got %T", renderedText)
		}

		// Basic sanity checks on the output
		if !strings.Contains(renderedStr, `(object)`) {
			t.Error("Rendered text missing expected content: '(object)'")
		}
		if !strings.Contains(renderedStr, `: "root"`) {
			t.Error("Rendered text missing expected content: ': \"root\"'")
		}
		if !strings.Contains(renderedStr, `: 123`) {
			t.Error("Rendered text missing expected content: ': 123'")
		}
	})
}

func TestTreeModificationErrors(t *testing.T) {
	const testJSON = `{"root": {"leaf": "a"}, "other": {}}`

	testTreeToolHelper(t, "Modification Error Conditions", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		rootID, err := getNodeIDByPath(t, interp, handle, "root")
		if err != nil {
			t.Fatal(err)
		}
		leafID, err := getNodeIDByPath(t, interp, handle, "root.leaf")
		if err != nil {
			t.Fatal(err)
		}

		// Attempt to set value on an object node
		_, err = callSetValue(t, interp, handle, rootID, "new value")
		assertResult(t, nil, err, nil, lang.ErrCannotSetValueOnType)

		// Attempt to add a child to a leaf node
		_, err = callAddChildNode(t, interp, handle, leafID, "new", "string", "v", "k")
		assertResult(t, nil, err, nil, lang.ErrNodeWrongType)

		// Attempt to add a node with a duplicate ID suggestion
		_, err = callAddChildNode(t, interp, handle, rootID, leafID, "string", "v", "k")
		assertResult(t, nil, err, nil, lang.ErrNodeIDExists)

		// Attempt to remove the root node
		rootNodeID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}
		_, err = runTool(t, interp, "RemoveNode", handle, rootNodeID)
		assertResult(t, nil, err, nil, lang.ErrCannotRemoveRoot)
	})
}

func TestTreeFindAndPathErrors(t *testing.T) {
	const testJSON = `{"a": {"b": [{}, {"c": "found"}]}}`

	testTreeToolHelper(t, "Find and Path Error Conditions", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		// FindNodes with multiple criteria
		query := map[string]interface{}{"type": "string", "value": "found"}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		if err != nil {
			t.Fatalf("FindNodes with multiple criteria failed: %v", err)
		}
		if len(result.([]interface{})) != 1 {
			t.Fatalf("Expected 1 node from multi-criteria search, got %d", len(result.([]interface{})))
		}

		// GetNodeByPath with an invalid intermediate path
		_, err = runTool(t, interp, "GetNodeByPath", handle, "a.b.99.c")
		assertResult(t, nil, err, nil, lang.ErrNotFound)

		// FindNodes with max_depth
		query = map[string]interface{}{"type": "object"}
		// FIX: The search starts at depth 0 (the root). A max_depth of 1 includes the root (depth 0)
		// and its direct children (depth 1). In this case, the root and the node at "a" are both objects.
		result, err = runTool(t, interp, "FindNodes", handle, rootID, query, int64(1), int64(-1))
		if err != nil {
			t.Fatalf("FindNodes with max_depth failed: %v", err)
		}
		if len(result.([]interface{})) != 2 {
			t.Fatalf("Expected 2 nodes with max_depth=1, got %d", len(result.([]interface{})))
		}
	})
}

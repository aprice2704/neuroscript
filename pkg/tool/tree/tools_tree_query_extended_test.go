// NeuroScript Version: 0.6.5
// File version: 1
// Purpose: Adds test coverage for query edge cases, including 'value' and 'max_results' criteria for FindNodes, and error conditions for GetChildren.
// filename: pkg/tool/tree/tools_tree_query_extended_test.go
// nlines: 70
// risk_rating: LOW

package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeFindExtended(t *testing.T) {
	const testJSON = `[
		{"name": "fileA", "value": 100},
		{"name": "fileB", "value": 200},
		{"name": "fileC", "value": 100}
	]`

	testTreeToolHelper(t, "Find by Exact Value", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		// Find nodes whose child 'value' node has a value of 100
		query := map[string]interface{}{"value": float64(100)}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		if err != nil {
			t.Fatalf("FindNodes by value failed: %v", err)
		}

		results := result.([]interface{})
		if len(results) != 2 {
			t.Fatalf("Expected 2 nodes with value 100, got %d", len(results))
		}
	})

	testTreeToolHelper(t, "Find with max_results", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		// Use the same query but limit results
		query := map[string]interface{}{"value": float64(100)}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(1))
		if err != nil {
			t.Fatalf("FindNodes with max_results failed: %v", err)
		}

		results := result.([]interface{})
		if len(results) != 1 {
			t.Fatalf("Expected 1 node with max_results=1, got %d", len(results))
		}
	})
}

func TestTreeGetChildrenError(t *testing.T) {
	const testJSON = `{"object_node": {"a": 1}}`
	testTreeToolHelper(t, "GetChildren on Non-Array Node", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		// The root node itself is the object
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}

		// Attempt to get children from an object node, which should fail
		_, err = runTool(t, interp, "GetChildren", handle, rootID)
		assertResult(t, nil, err, nil, lang.ErrNodeWrongType)
	})
}

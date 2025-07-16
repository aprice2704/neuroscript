// NeuroScript Version: 0.6.5
// File version: 18
// Purpose: Corrected argument types for FindNodes calls to pass int64 instead of int, resolving test failures.
// filename: pkg/tool/tree/tools_tree_find_test.go
// nlines: 100
// risk_rating: LOW
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeFind(t *testing.T) {
	const testJSON = `[
		{"type": "file", "name": "a.txt", "size": 100},
		{"type": "dir", "name": "d1", "children": [
			{"type": "file", "name": "b.log", "size": 250}
		]},
		{"type": "file", "name": "c.out", "size": 100}
	]`

	setup := func(t *testing.T, interp tool.Runtime) (string, string) {
		handle, err := setupTreeWithJSON(t, interp, testJSON)
		if err != nil {
			t.Fatalf("Tree setup failed unexpectedly: %v", err)
		}
		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatalf("GetRootID failed: %v", err)
		}
		return handle, rootID
	}

	testTreeToolHelper(t, "Find All Files By Type", func(t *testing.T, interp tool.Runtime) {
		handle, rootID := setup(t, interp)
		query := map[string]interface{}{"type": "file"}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		assertResult(t, nil, err, nil, nil)

		results, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected result to be a slice, got %T", result)
		}
		if len(results) != 3 {
			t.Fatalf("expected 3 file nodes, got %d", len(results))
		}
	})

	testTreeToolHelper(t, "Find By Metadata Attribute", func(t *testing.T, interp tool.Runtime) {
		handle, rootID := setup(t, interp)
		fileID, err := getNodeIDByPath(t, interp, handle, "0")
		if err != nil {
			t.Fatal(err)
		}
		_, err = callSetNodeMetadata(t, interp, handle, fileID, "tag", "important")
		if err != nil {
			t.Fatal(err)
		}

		query := map[string]interface{}{"metadata": map[string]interface{}{"tag": "important"}}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		assertResult(t, nil, err, nil, nil)

		results, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected result to be a slice, got %T", result)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 tagged node, got %d", len(results))
		}
	})

	testTreeToolHelper(t, "FindNodes Invalid Query Map", func(t *testing.T, interp tool.Runtime) {
		handle, rootID := setup(t, interp)
		query := map[string]interface{}{"type": 123}
		_, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		assertResult(t, nil, err, nil, lang.ErrTreeInvalidQuery)
	})
}

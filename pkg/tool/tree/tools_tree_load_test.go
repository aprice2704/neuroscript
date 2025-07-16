// NeuroScript Version: 0.6.5
// File version: 20
// Purpose: Corrected FindNodes call to use int64 arguments for max_depth and max_results.
// filename: pkg/tool/tree/tools_tree_load_test.go
// nlines: 120
// risk_rating: LOW
package tree_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/go-cmp/cmp"
)

const simpleJSON = `{
    "name": "root_obj",
    "enabled": true,
    "ports": [80, 443],
    "metadata": {
        "version": "1.0"
    }
}`

const complexJSON = `[
    {"id": "user1", "type": "user", "details": {"name": "Alice", "role": "admin"}},
    {"id": "user2", "type": "user", "details": {"name": "Bob", "role": "editor"}},
    {"id": "group1", "type": "group", "members": ["user1"]}
]`

func TestTreeLoadJSON(t *testing.T) {
	testTreeToolHelper(t, "Load Simple JSON Object", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "LoadJSON", simpleJSON)
		if err != nil {
			t.Fatalf("LoadJSON failed unexpectedly: %v", err)
		}
		if _, ok := result.(string); !ok {
			t.Fatalf("expected a string handle, got %T", result)
		}
	})

	testTreeToolHelper(t, "Load Complex JSON Array", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, complexJSON)
		if err != nil {
			t.Fatalf("setupTreeWithJSON failed: %v", err)
		}
		if handle == "" {
			t.Fatal("expected a valid handle, got empty string")
		}
	})

	testTreeToolHelper(t, "Load Invalid JSON", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "LoadJSON", `{"key": "no_close_quote}`)
		assertResult(t, nil, err, nil, lang.ErrTreeJSONUnmarshal)
	})
}

func TestTreeToJSON(t *testing.T) {
	testTreeToolHelper(t, "Roundtrip Simple JSON", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, simpleJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		jsonResult, err := callToJSON(t, interp, handle)
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		jsonStr, ok := jsonResult.(string)
		if !ok {
			t.Fatalf("ToJSON did not return a string, got %T", jsonResult)
		}

		var original, roundtripped interface{}
		if err := json.NewDecoder(strings.NewReader(simpleJSON)).Decode(&original); err != nil {
			t.Fatalf("could not decode original json: %v", err)
		}
		if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&roundtripped); err != nil {
			t.Fatalf("could not decode roundtripped json: %v", err)
		}
		if diff := cmp.Diff(original, roundtripped); diff != "" {
			t.Errorf("JSON content mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestTreeGetRoot(t *testing.T) {
	testTreeToolHelper(t, "Get Root Node", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, simpleJSON)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		rootNode, err := runTool(t, interp, "GetRoot", handle)
		if err != nil {
			t.Fatalf("GetRoot failed unexpectedly: %v", err)
		}

		nodeMap, ok := rootNode.(map[string]interface{})
		if !ok {
			t.Fatalf("expected root node to be a map, got %T", rootNode)
		}
		if nodeMap["type"] != "object" {
			t.Errorf("expected root type to be 'object', got '%s'", nodeMap["type"])
		}
	})
}

func TestFindNodes(t *testing.T) {
	testTreeToolHelper(t, "Find Nodes By Metadata", func(t *testing.T, interp tool.Runtime) {
		handle, err := setupTreeWithJSON(t, interp, complexJSON)
		if err != nil {
			t.Fatalf("Failed to load initial JSON: %v", err)
		}

		rootID, err := getRootID(t, interp, handle)
		if err != nil {
			t.Fatal(err)
		}
		user1ID, err := getNodeIDByPath(t, interp, handle, "0")
		if err != nil {
			t.Fatal(err)
		}
		user2ID, err := getNodeIDByPath(t, interp, handle, "1")
		if err != nil {
			t.Fatal(err)
		}

		_, err = callSetNodeMetadata(t, interp, handle, user1ID, "status", "active")
		if err != nil {
			t.Fatalf("Failed to set metadata for user1: %v", err)
		}
		_, err = callSetNodeMetadata(t, interp, handle, user2ID, "status", "inactive")
		if err != nil {
			t.Fatalf("Failed to set metadata for user2: %v", err)
		}

		query := map[string]interface{}{"metadata": map[string]interface{}{"status": "active"}}
		result, err := runTool(t, interp, "FindNodes", handle, rootID, query, int64(-1), int64(-1))
		assertResult(t, nil, err, nil, nil)

		nodeIDs, ok := result.([]interface{})
		if !ok {
			t.Fatalf("FindNodes did not return a slice, got %T", result)
		}
		if len(nodeIDs) != 1 {
			t.Fatalf("expected 1 node, got %d", len(nodeIDs))
		}
		if nodeIDs[0] != user1ID {
			t.Errorf("expected found node to be '%s', got '%s'", user1ID, nodeIDs[0])
		}
	})
}

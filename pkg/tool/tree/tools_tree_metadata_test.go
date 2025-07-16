// NeuroScript Version: 0.6.5
// File version: 18
// Purpose: Corrected metadata tests to assert against a map[string]interface{} and account for initial attributes from JSON load.
// filename: pkg/tool/tree/tools_tree_metadata_test.go
// nlines: 80
// risk_rating: LOW
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestTreeMetadata(t *testing.T) {
	const testJSON = `{"key": "value", "nested": {"num": 123}}`

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

	testTreeToolHelper(t, "Get Root Metadata Initially", func(t *testing.T, interp tool.Runtime) {
		handle, rootID := setup(t, interp)
		metadata, err := callGetMetadata(t, interp, handle, rootID)
		if err != nil {
			t.Fatal(err)
		}

		metaMap, ok := metadata.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected metadata to be a map[string]interface{}, got %T", metadata)
		}

		if len(metaMap) != 2 {
			t.Errorf("Expected 2 initial attributes, got %d", len(metaMap))
		}
		if _, ok := metaMap["key"]; !ok {
			t.Error("Expected 'key' in initial metadata")
		}
		if _, ok := metaMap["nested"]; !ok {
			t.Error("Expected 'nested' in initial metadata")
		}
	})

	testTreeToolHelper(t, "Set and Get Metadata", func(t *testing.T, interp tool.Runtime) {
		handle, rootID := setup(t, interp)

		_, err := callSetNodeMetadata(t, interp, handle, rootID, "status", "active")
		if err != nil {
			t.Fatalf("SetNodeMetadata failed: %v", err)
		}
		_, err = callSetNodeMetadata(t, interp, handle, rootID, "priority", "10")
		if err != nil {
			t.Fatalf("SetNodeMetadata failed: %v", err)
		}

		metadata, err := callGetMetadata(t, interp, handle, rootID)
		expectedMeta := map[string]interface{}{
			"key":      "node-2",
			"nested":   "node-3",
			"status":   "active",
			"priority": "10",
		}
		assertResult(t, metadata, err, expectedMeta, nil)

		_, err = callSetNodeMetadata(t, interp, handle, rootID, "status", "inactive")
		if err != nil {
			t.Fatalf("SetNodeMetadata failed on overwrite: %v", err)
		}

		metadata, err = callGetMetadata(t, interp, handle, rootID)
		expectedMeta["status"] = "inactive"
		assertResult(t, metadata, err, expectedMeta, nil)
	})

	testTreeToolHelper(t, "Get Metadata On Invalid Node ID", func(t *testing.T, interp tool.Runtime) {
		handle, _ := setup(t, interp)
		_, err := callGetMetadata(t, interp, handle, "invalid-node-id")
		assertResult(t, nil, err, nil, lang.ErrNotFound)
	})
}

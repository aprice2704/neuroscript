// NeuroScript Version: 0.5.4
// File version: 14
// Purpose: Corrected all function signatures in test cases to use tool.Runtime, resolving compiler errors.
// filename: pkg/tool/tree/tools_tree_query_test.go
// nlines: 130
// risk_rating: LOW
package tree_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

func TestTreeQuery(t *testing.T) {
	baseJSON := `{
		"name": "root",
		"files": [
			{"name": "file1.txt", "size": 100},
			{"name": "file2.txt", "size": 200, "tags": ["important", "text"]},
			{"name": "image.jpg", "size": 1500, "tags": ["image"]}
		],
		"config": {"enabled": true, "version": "1.2.3"}
	}`

	testCases := []treeTestCase{
		{
			Name:      "Query_All_Files_By_Type",
			JSONInput: baseJSON,
			ToolName:  "FindNodes",
			Args:      []interface{}{nil, "placeholder_root", map[string]interface{}{"type": "object"}, int64(-1), int64(-1)},
			Validation: func(t *testing.T, interp tool.Runtime, treeHandle string, result interface{}) {
				results, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T", result)
				}
				if len(results) != 5 {
					t.Errorf("Expected 5 object nodes, got %d", len(results))
				}
			},
		},
		{
			Name:      "Query_By_Metadata_Attribute",
			JSONInput: baseJSON,
			ToolName:  "FindNodes",
			Args:      []interface{}{nil, "placeholder_root", map[string]interface{}{"metadata": map[string]interface{}{"name": "file2.txt"}}},
			Validation: func(t *testing.T, interp tool.Runtime, treeHandle string, result interface{}) {
				results, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T", result)
				}
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
				nodeID := results[0].(string)
				node, err := callGetNode(t, interp, treeHandle, nodeID)
				if err != nil {
					t.Fatalf("Could not get node from query result: %v", err)
				}
				nodeMap, ok := node.(map[string]interface{})
				if !ok {
					t.Fatalf("node is not a map, but %T", node)
				}
				attributes, ok := nodeMap["attributes"].(utils.TreeAttrs)
				if !ok {
					t.Fatalf("attributes is not utils.TreeAttrs, but %T", nodeMap["attributes"])
				}

				sizeNodeID := attributes["size"].(string)
				sizeNodeValue, err := callGetValue(t, interp, treeHandle, sizeNodeID)
				if err != nil {
					t.Fatalf("could not get size node value: %v", err)
				}

				if sizeNodeValue != float64(200) {
					t.Errorf("Expected size of file2.txt to be 200, got %v", sizeNodeValue)
				}
			},
		},
		{
			Name:        "Query_Invalid_Syntax_in_Map",
			JSONInput:   baseJSON,
			ToolName:    "FindNodes",
			Args:        []interface{}{nil, "placeholder_root", map[string]interface{}{"attributes": "not-a-map"}},
			ExpectedErr: lang.ErrTreeInvalidQuery,
		},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, tc.Name, func(t *testing.T, interp tool.Runtime) {
			treeHandle, err := setupTreeWithJSON(t, interp, tc.JSONInput)
			if err != nil {
				t.Fatalf("Tree setup failed unexpectedly: %v", err)
			}

			rootID := getRootID(t, interp, treeHandle)

			args := tc.Args
			if len(args) > 0 {
				if args[0] == nil {
					args[0] = treeHandle
				}
				if args[1] == "placeholder_root" {
					args[1] = rootID
				}
			}

			result, err := runTool(t, interp, tc.ToolName, args...)

			if tc.Validation != nil {
				if err != nil && tc.ExpectedErr == nil {
					t.Fatalf("Tool execution failed unexpectedly: %v", err)
				}
				tc.Validation(t, interp, treeHandle, result)
			} else {
				assertResult(t, result, err, tc.Expected, tc.ExpectedErr)
			}
		})
	}
}

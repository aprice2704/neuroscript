// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Purpose: Corrected type assertion to use TreeAttrs, completing the tree refactor in the test suite.
// nlines: 160
// risk_rating: MEDIUM
// filename: pkg/tool/tree/tools_tree_query_test.go

package tree

import (
	"reflect"
	"strings"
	"testing"
)

const treeJSONForFindRenderNested = `{
    "name": "root_obj",
    "type": "directory",
    "children": [
        {"name": "file1.txt", "type": "file", "size": 100},
        {"name": "subdir", "type": "directory", "children": [
            {"name": "file2.txt", "type": "file", "size": 50, "meta_deep": {"is_special": true}}
        ]}
    ],
    "metadata": {"owner": "admin", "status": "active"}
}`

func TestTreeFindAndRenderTools(t *testing.T) {
	setupFindRenderTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, treeJSONForFindRenderNested)
	}

	testCases := []treeTestCase{
		// Tree.FindNodes
		{name: "FindNodes_By_Type_'file'", toolName: "Tree.FindNodes",
			setupFunc:	setupFindRenderTree,
			args:		MakeArgs("SETUP_HANDLE:frTree", "node-1", map[string]interface{}{"value": "file"}, int64(-1), int64(-1)),	// Query for string nodes with value "file"
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("FindNodes by value 'file' failed: %v", err)
				}
				ids, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T. Result: %#v", result, result)
				}

				handle := ctx.(string)
				var actualFileObjectNodeIDs []string

				if len(ids) != 2 {	// Should find two string nodes with value "file"
					t.Errorf("Expected 2 string nodes with value 'file', got %d: %v", len(ids), ids)
				}

				for _, idInterface := range ids {
					idStr, ok := idInterface.(string)
					if !ok {
						t.Errorf("Found ID is not a string: %T", idInterface)
						continue
					}
					stringNodeMap, errGetNode := callGetNode(t, interp, handle, idStr)
					if errGetNode != nil {
						t.Errorf("Failed to get string node %s: %v", idStr, errGetNode)
						continue
					}

					if typeVal, _ := stringNodeMap["type"].(string); typeVal != "string" {
						t.Errorf("Expected node %s to be type 'string', got '%s'", idStr, typeVal)
						continue
					}
					if valueVal, _ := stringNodeMap["value"].(string); valueVal != "file" {
						t.Errorf("Expected node %s to have value 'file', got '%s'", idStr, valueVal)
						continue
					}

					parentNodeID, parentOK := stringNodeMap["parent_id"].(string)
					parentAttrKey, keyOK := stringNodeMap["parent_attribute_key"].(string)

					if !parentOK || !keyOK || parentAttrKey != "type" {
						t.Errorf("String node %s (value 'file') is not a 'type' attribute as expected. ParentID: %v, ParentAttrKey: %v", idStr, parentNodeID, parentAttrKey)
						continue
					}

					parentObjectNodeMap, errGetParent := callGetNode(t, interp, handle, parentNodeID)
					if errGetParent != nil {
						t.Errorf("Failed to get parent object node %s: %v", parentNodeID, errGetParent)
						continue
					}
					if parentType, _ := parentObjectNodeMap["type"].(string); parentType != "object" {
						t.Errorf("Parent node %s of string node %s is not an 'object'. Got type %s", parentNodeID, idStr, parentType)
						continue
					}

					alreadyFound := false
					for _, foundID := range actualFileObjectNodeIDs {
						if foundID == parentNodeID {
							alreadyFound = true
							break
						}
					}
					if !alreadyFound {
						actualFileObjectNodeIDs = append(actualFileObjectNodeIDs, parentNodeID)
					}
				}

				if len(actualFileObjectNodeIDs) != 2 {
					t.Errorf("Expected to find 2 distinct file object nodes, but found %d: %v. Original string node IDs: %v", len(actualFileObjectNodeIDs), actualFileObjectNodeIDs, ids)
				}
			}},
		{name: "FindNodes_By_Value_of_Name_Attribute", toolName: "Tree.FindNodes", setupFunc: setupFindRenderTree, args: MakeArgs("SETUP_HANDLE:frTree", "node-1", map[string]interface{}{"value": "file1.txt"}, int64(-1), int64(-1)),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("FindNodes by value failed: %v", err)
				}
				ids, ok := result.([]interface{})
				if !ok {
					t.Fatalf("FindNodes did not return a slice, got %T", result)
				}
				if len(ids) != 1 {
					t.Errorf("Expected 1 node with value 'file1.txt', got %d: %v", len(ids), ids)
				}
				if len(ids) == 1 {
					handle := ctx.(string)
					nodeMap, _ := callGetNode(t, interp, handle, ids[0].(string))
					if nodeMap["type"] != "string" || nodeMap["value"] != "file1.txt" {
						t.Errorf("Found node is not the expected string 'file1.txt'. Got: %#v", nodeMap)
					}
				}
			}},
		{name: "FindNodes_By_Metadata_Deep", toolName: "Tree.FindNodes",
			setupFunc:	setupFindRenderTree,
			args:		MakeArgs("SETUP_HANDLE:frTree", "node-1", map[string]interface{}{"comment": "query_is_dynamic_in_checkfunc"}, int64(-1), int64(-1)),
			checkFunc: func(t *testing.T, interp *Interpreter, _ interface{}, _ error, ctx interface{}) {	// Ignored result from static args
				handle := ctx.(string)
				findNodesTool, toolFound := interp.ToolRegistry().GetTool("Tree.FindNodes")
				if !toolFound {
					t.Fatalf("Tree.FindNodes tool not found")
				}

				var file2ObjNodeID string
				var actualMetaDeepTargetNodeID string

				nameQuery := map[string]interface{}{"value": "file2.txt"}
				nameNodeIDsResult, errFindName := findNodesTool.Func(interp, MakeArgs(handle, "node-1", nameQuery, int64(-1), int64(-1)))
				if errFindName != nil {
					t.Fatalf("Error finding 'file2.txt' string node: %v", errFindName)
				}
				nameNodeIDList, _ := nameNodeIDsResult.([]interface{})
				if len(nameNodeIDList) != 1 {
					t.Fatalf("Could not uniquely find 'file2.txt' string node, found %d", len(nameNodeIDList))
				}

				nameStrNodeID := nameNodeIDList[0].(string)
				nameStrNode, _ := callGetNode(t, interp, handle, nameStrNodeID)
				file2ObjNodeID = nameStrNode["parent_id"].(string)

				file2ObjNode, _ := callGetNode(t, interp, handle, file2ObjNodeID)
				attrsFile2Obj, ok := file2ObjNode["attributes"].(TreeAttrs)
				if !ok {
					t.Fatalf("Attributes of node %s not TreeAttrs, got %T", file2ObjNodeID, file2ObjNode["attributes"])
				}
				actualMetaDeepTargetNodeID, ok = attrsFile2Obj["meta_deep"].(string)
				if !ok {
					t.Fatalf("'meta_deep' attribute not found on node %s", file2ObjNodeID)
				}

				dynamicQuery := map[string]interface{}{"attributes": map[string]interface{}{"meta_deep": actualMetaDeepTargetNodeID}}

				dynamicallyFoundResult, errFindDynamic := findNodesTool.Func(interp, MakeArgs(handle, "node-1", dynamicQuery, int64(-1), int64(-1)))
				if errFindDynamic != nil {
					t.Fatalf("Tree.FindNodes with dynamic query failed: %v", errFindDynamic)
				}
				dynamicallyFoundIDs, ok := dynamicallyFoundResult.([]interface{})
				if !ok {
					t.Fatalf("Tree.FindNodes dynamic query not slice: %T", dynamicallyFoundResult)
				}

				expectedToFind := []string{file2ObjNodeID}	// Expect to find the file2.txt object node itself
				var foundIDsStr []string
				for _, id := range dynamicallyFoundIDs {
					foundIDsStr = append(foundIDsStr, id.(string))
				}

				if !reflect.DeepEqual(foundIDsStr, expectedToFind) {
					t.Errorf("FindNodes_By_Metadata_Deep mismatch.\n   Query was: %#v\n   Got:       %#v\n   Wanted:    %#v", dynamicQuery, foundIDsStr, expectedToFind)
				}
			}},
		{name: "FindNodes_Invalid_Query_Type", toolName: "Tree.FindNodes", setupFunc: setupFindRenderTree, args: MakeArgs("SETUP_HANDLE:frTree", "node-1", "not-a-map", int64(-1), int64(-1)), wantErr: ErrInvalidArgument},

		// Tree.RenderText
		{name: "RenderText Basic", toolName: "Tree.RenderText",
			setupFunc:	func(t *testing.T, interp *Interpreter) interface{} { return setupTreeWithJSON(t, interp, `{"a":"b"}`) },
			args:		MakeArgs("SETUP_HANDLE:renderTree"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("RenderText failed: %v", err)
				}
				s, ok := result.(string)
				if !ok {
					t.Fatalf("RenderText did not return string, got %T", result)
				}
				if !strings.Contains(s, "- (object)") {
					t.Errorf("RenderText output missing '- (object)'. Got:\n%s", s)
				}
				if !strings.Contains(s, `Key: "a"`) {
					t.Errorf("RenderText output missing 'Key: \"a\"'. Got:\n%s", s)
				}
				if !strings.Contains(s, `(string): "b"`) {
					t.Errorf("RenderText output missing '(string): \"b\"'. Got:\n%s", s)
				}
			}},
		{name: "RenderText Invalid Handle", toolName: "Tree.RenderText", args: MakeArgs("bad-handle"), wantErr: ErrInvalidArgument},
	}

	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}
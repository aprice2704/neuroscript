// NeuroScript Version: 0.5.4
// File version: 14
// Purpose: Corrected the function signatures within the treeTestCase struct to use the tool.Runtime interface, resolving the final compiler error.
// filename: pkg/tool/tree/tools_tree_test_helpers.go
// nlines: 135
// risk_rating: LOW

package tree_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

const group = "tree"

type treeTestCase struct {
	Name      string
	ToolName  types.ToolName
	Args      []interface{}
	JSONInput string
	// FIX: Updated function signatures to use tool.Runtime interface.
	SetupFunc   func(t *testing.T, interp tool.Runtime, treeHandle string)
	Validation  func(t *testing.T, interp tool.Runtime, treeHandle string, result interface{})
	Expected    interface{}
	ExpectedErr error
}

func testTreeToolHelper(t *testing.T, testName string, testFunc func(t *testing.T, interp tool.Runtime)) {
	t.Run(testName, func(t *testing.T) {
		// This must be the actual interpreter to satisfy the runtime needs for handles.
		interp := interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))
		if err := tool.RegisterExtendedTools(interp.ToolRegistry()); err != nil {
			t.Fatalf("Failed to register extended tools: %v", err)
		}
		testFunc(t, interp)
	})
}

func runTool(t *testing.T, interp tool.Runtime, toolName types.ToolName, args ...interface{}) (interface{}, error) {
	t.Helper()
	fullName := types.MakeFullName(group, string(toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}
	return toolImpl.Func(interp, args)
}

func assertResult(t *testing.T, result interface{}, err error, expected interface{}, expectedErrIs error) {
	t.Helper()
	if expectedErrIs != nil {
		if err == nil {
			t.Errorf("expected an error wrapping '%v', but got nil", expectedErrIs)
		} else if !errors.Is(err, expectedErrIs) {
			t.Errorf("expected error to wrap '%v', but got: %v", expectedErrIs, err)
		}
	} else {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("result does not match expected: got %#v, want %#v", result, expected)
		}
	}
}

func setupTreeWithJSON(t *testing.T, interp tool.Runtime, jsonStr string) (string, error) {
	t.Helper()
	result, err := runTool(t, interp, "LoadJSON", jsonStr)
	if err != nil {
		return "", err
	}
	handle, ok := result.(string)
	if !ok {
		t.Fatalf("LoadJSON did not return a string handle, got %T", result)
	}
	return handle, nil
}

func callGetNode(t *testing.T, interp tool.Runtime, treeHandle string, nodeID string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "GetNode", treeHandle, nodeID)
}

func callGetChildren(t *testing.T, interp tool.Runtime, treeHandle string, nodeID string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "GetChildren", treeHandle, nodeID)
}

func callSetNodeMetadata(t *testing.T, interp tool.Runtime, treeHandle string, nodeID string, key string, value string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "SetNodeMetadata", treeHandle, nodeID, key, value)
}

func callGetValue(t *testing.T, interp tool.Runtime, treeHandle string, nodeID string) (interface{}, error) {
	t.Helper()
	nodeInfo, err := callGetNode(t, interp, treeHandle, nodeID)
	if err != nil {
		return nil, err
	}
	nodeMap, ok := nodeInfo.(map[string]interface{})
	if !ok {
		t.Fatalf("GetNode did not return a map, got %T", nodeInfo)
	}
	return nodeMap["value"], nil
}

func callGetMetadata(t *testing.T, interp tool.Runtime, treeHandle string, nodeID string) (interface{}, error) {
	t.Helper()
	nodeInfo, err := callGetNode(t, interp, treeHandle, nodeID)
	if err != nil {
		return nil, err
	}
	nodeMap, ok := nodeInfo.(map[string]interface{})
	if !ok {
		t.Fatalf("GetNode did not return a map, got %T", nodeInfo)
	}
	attributes, ok := nodeMap["attributes"].(utils.TreeAttrs)
	if !ok {
		attrMap, ok := nodeMap["attributes"].(map[string]interface{})
		if !ok {
			t.Fatalf("attributes is not a map, but %T", nodeMap["attributes"])
		}
		attributes = utils.TreeAttrs(attrMap)
	}

	metadata := make(utils.TreeAttrs)
	nodeType := nodeMap["type"].(string)

	if nodeType != "object" {
		return attributes, nil
	}

	for k, v := range attributes {
		if _, ok := v.(string); !ok {
			metadata[k] = v
		}
	}

	return metadata, nil
}

func callToJSON(t *testing.T, interp tool.Runtime, treeHandle string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "ToJSON", treeHandle)
}

func callAddChildNode(t *testing.T, interp tool.Runtime, args ...interface{}) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "AddChildNode", args...)
}

func getRootNode(t *testing.T, interp tool.Runtime, treeHandle string) map[string]interface{} {
	t.Helper()
	node, err := callGetNode(t, interp, treeHandle, "node-1")
	if err != nil {
		t.Fatalf("could not get root node: %v", err)
	}
	return node.(map[string]interface{})
}

func getRootID(t *testing.T, interp tool.Runtime, treeHandle string) string {
	t.Helper()
	return getRootNode(t, interp, treeHandle)["id"].(string)
}

func getNodeIDByPath(t *testing.T, interp tool.Runtime, treeHandle string, path string) (string, error) {
	t.Helper()
	rootID := getRootID(t, interp, treeHandle)
	if path == "" || path == "root" {
		return rootID, nil
	}

	parts := strings.Split(path, ".")
	currentNodeID := rootID

	for i, part := range parts {
		nodeInfo, err := callGetNode(t, interp, treeHandle, currentNodeID)
		if err != nil {
			return "", fmt.Errorf("could not get node '%s' in path '%s': %w", currentNodeID, path, err)
		}
		nodeMap := nodeInfo.(map[string]interface{})
		nodeType := nodeMap["type"].(string)

		if nodeType == "object" {
			attributes, ok := nodeMap["attributes"].(utils.TreeAttrs)
			if !ok {
				return "", fmt.Errorf("attributes of node '%s' have unexpected type %T", currentNodeID, nodeMap["attributes"])
			}
			childNodeID, ok := attributes[part].(string)
			if !ok {
				return "", fmt.Errorf("path part '%s' not found in object node '%s'", part, currentNodeID)
			}
			currentNodeID = childNodeID
		} else if nodeType == "array" {
			index, err := strconv.Atoi(part)
			if err != nil {
				return "", fmt.Errorf("invalid array index '%s' in path '%s'", part, path)
			}
			children, err := callGetChildren(t, interp, treeHandle, currentNodeID)
			if err != nil {
				return "", fmt.Errorf("could not get children of array node '%s': %w", currentNodeID, err)
			}
			childIDs := children.([]interface{})
			if index >= len(childIDs) {
				return "", fmt.Errorf("index %d out of bounds for array node '%s'", index, currentNodeID)
			}
			currentNodeID = childIDs[index].(string)
		} else {
			return "", fmt.Errorf("cannot traverse path further from leaf node '%s' of type '%s' at part %d ('%s')", currentNodeID, nodeType, i, part)
		}
	}

	return currentNodeID, nil
}

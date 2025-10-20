// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Removed redundant tool.RegisterGlobalToolsets call to fix duplicate key errors.
// filename: pkg/tool/tree/tools_tree_helpers_test.go
// nlines: 163
package tree_test

import (
	"bytes"
	"errors"
	"fmt"
	"log" // DEBUG
	"os"  // DEBUG
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/go-cmp/cmp"
)

// treeTestCase defines the structure for a single tree tool test case.
type treeTestCase struct {
	Name         string
	ToolName     types.ToolName
	Args         []interface{}
	JSONInput    string
	SetupFunc    func(t *testing.T, interp tool.Runtime, treeHandle string)
	Validation   func(t *testing.T, interp tool.Runtime, treeHandle string, result interface{})
	Expected     interface{}
	ExpectedErr  error
	ExpectedLogs []string
}

func testTreeToolHelper(t *testing.T, testName string, testFunc func(t *testing.T, interp tool.Runtime)) {
	t.Run(testName, func(t *testing.T) {
		// DEBUG: Per AGENTS.md Rule 1b
		log.Printf("[DEBUG] START %s", testName)
		hostCtx, err := interpreter.NewHostContextBuilder().
			WithLogger(logging.NewTestLogger(t)).
			WithStdout(&bytes.Buffer{}).
			WithStdin(&bytes.Buffer{}).
			WithStderr(os.Stderr). // DEBUG: Send stderr to os.Stderr
			Build()
		if err != nil {
			t.Fatalf("Failed to create host context: %v", err)
		}
		// DEBUG: Per AGENTS.md Rule 1b
		log.Printf("[DEBUG] %s: HostContext created.", testName)
		interp := interpreter.NewInterpreter(interpreter.WithHostContext(hostCtx))
		// DEBUG: Per AGENTS.md Rule 1b
		log.Printf("[DEBUG] %s: NewInterpreter created. Tool registry should be populated.", testName)

		// XXX: This call is redundant. NewInterpreter already calls
		// RegisterStandardTools, which in turn calls RegisterGlobalToolsets.
		// Calling it again causes the "duplicate key" errors.
		// if err := tool.RegisterGlobalToolsets(interp.ToolRegistry()); err != nil {
		// 	t.Fatalf("Failed to register extended tools: %v", err)
		// }
		testFunc(t, interp)
		// DEBUG: Per AGENTS.md Rule 1b
		log.Printf("[DEBUG] END %s", testName)
	})
}

func runTool(t *testing.T, interp tool.Runtime, toolName types.ToolName, args ...interface{}) (interface{}, error) {
	t.Helper()
	fullName := types.MakeFullName("tree", string(toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}
	return toolImpl.Func(interp, args)
}

func assertResult(t *testing.T, result interface{}, err error, expectedResult interface{}, expectedErr error) {
	t.Helper()
	if expectedErr != nil {
		if err == nil {
			t.Fatalf("expected error '%v', but got nil", expectedErr)
		}
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error to wrap '%v', but got: %v", expectedErr, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(expectedResult, result); diff != "" {
		t.Errorf("result does not match expected (-want +got):\n%s", diff)
	}
}

func setupTreeWithJSON(t *testing.T, interp tool.Runtime, jsonContent string) (string, error) {
	t.Helper()
	handle, err := runTool(t, interp, "LoadJSON", jsonContent)
	if err != nil {
		return "", fmt.Errorf("failed to load tree from JSON: %w", err)
	}
	handleStr, ok := handle.(string)
	if !ok {
		return "", fmt.Errorf("expected handle to be a string, but got %T", handle)
	}
	return handleStr, nil
}

func callToJSON(t *testing.T, interp tool.Runtime, handle string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "ToJSON", handle)
}

func callGetNode(t *testing.T, interp tool.Runtime, handle, nodeID string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "GetNode", handle, nodeID)
}

func callGetMetadata(t *testing.T, interp tool.Runtime, handle, nodeID string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "GetNodeMetadata", handle, nodeID)
}

func callSetNodeMetadata(t *testing.T, interp tool.Runtime, handle, nodeID, key, value string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "SetNodeMetadata", handle, nodeID, key, value)
}

// FIX: Implement callGetValue helper to get a node and return its 'value' field.
func callGetValue(t *testing.T, interp tool.Runtime, handle, nodeID string) (interface{}, error) {
	t.Helper()
	nodeData, err := callGetNode(t, interp, handle, nodeID)
	if err != nil {
		return nil, err
	}
	nodeMap, ok := nodeData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected GetNode to return a map, got %T", nodeData)
	}
	return nodeMap["value"], nil
}

func callSetValue(t *testing.T, interp tool.Runtime, handle, nodeID string, value interface{}) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "SetValue", handle, nodeID, value)
}

func callAddChildNode(t *testing.T, interp tool.Runtime, handle, parentID, idSuggestion, nodeType string, value interface{}, key string) (interface{}, error) {
	t.Helper()
	return runTool(t, interp, "AddChildNode", handle, parentID, idSuggestion, nodeType, value, key)
}

func callGetChildren(t *testing.T, interp tool.Runtime, handle, nodeID string) (interface{}, error) {
	t.Helper()
	nodeInfo, err := callGetNode(t, interp, handle, nodeID)
	if err != nil {
		return nil, err
	}
	nodeMap, ok := nodeInfo.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected node to be a map, got %T", nodeInfo)
	}
	// In our tree, the "children" of an object are the node IDs in its attributes map.
	if nodeMap["type"] == "object" {
		attrs, ok := nodeMap["attributes"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("attributes not a map")
		}
		var children []interface{}
		for _, v := range attrs {
			children = append(children, v)
		}
		return children, nil
	}
	// For arrays, it's the ChildIDs list.
	return nodeMap["children"], nil
}

func getRootNode(t *testing.T, interp tool.Runtime, handle string) (map[string]interface{}, error) {
	t.Helper()
	rootNode, err := runTool(t, interp, "GetRoot", handle)
	if err != nil {
		return nil, err
	}
	nodeMap, ok := rootNode.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected root node to be a map, got %T", rootNode)
	}
	return nodeMap, nil
}

func getRootID(t *testing.T, interp tool.Runtime, handle string) (string, error) {
	t.Helper()
	rootNode, err := getRootNode(t, interp, handle)
	if err != nil {
		return "", err
	}
	id, ok := rootNode["id"].(string)
	if !ok {
		return "", fmt.Errorf("expected root node to have a string 'id', got %T", rootNode["id"])
	}
	return id, nil
}

func getNodeIDByPath(t *testing.T, interp tool.Runtime, handle string, path string) (string, error) {
	t.Helper()
	result, err := runTool(t, interp, "GetNodeByPath", handle, path)
	if err != nil {
		return "", fmt.Errorf("GetNodeByPath failed for path %q: %w", path, err)
	}
	nodeMap, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected GetNodeByPath to return a map, got %T", result)
	}
	id, ok := nodeMap["id"].(string)
	if !ok {
		return "", fmt.Errorf("expected node map to have a string 'id', but got %T", nodeMap["id"])
	}
	return id, nil
}

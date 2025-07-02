// NeuroScript Version: 0.3.1
// File version: 0.2.0
// Purpose: Updated test helpers to return map[string]interface{} for attributes, aligning with  TreeAttrs.
// filename: pkg/neurodata/checklist/test_helpers.go

package checklist

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

func newTestInterpreterWithAllTools(t *testing.T) (*neurogo.Interpreter, tool.ToolRegistry) {
	t.Helper()
	tempDir := t.TempDir()
	logger, errLog := adapters.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	assertNoErrorSetup(t, errLog, "Failed to create logger")
	if logger == nil {
		t.Fatalf("Setup Error: Failed to create logger using SimpleTestLogger, returned nil unexpectedly")
	}
	llmClient := adapters.NewNoOpLLMClient()
	interp, errInterp := NewInterpreter(logger, llmClient, tempDir, nil, nil)
	assertNoErrorSetup(t, errInterp, "Failed to create  Interpreter")
	errExt := toolsets.RegisterExtendedTools(interp)
	assertNoErrorSetup(t, errExt, "Failed to register extended toolsets")
	errSandbox := interp.SetSandboxDir(tempDir)
	assertNoErrorSetup(t, errSandbox, "Failed to set sandbox dir")
	return interp, interp.ToolRegistry()
}

func getNodeViaTool(t *testing.T, interp *neurogo.Interpreter, handleID string, nodeID string) map[string]interface{} {
	t.Helper()
	toolReg := interp.ToolRegistry()
	impl, exists := toolReg.GetTool("Tree.GetNode")
	if !exists {
		t.Fatalf("getNodeViaTool: Prerequisite tool 'Tree.GetNode' not registered.")
	}
	if impl.Func == nil {
		t.Fatalf("getNodeViaTool: Tool 'Tree.GetNode' has nil function.")
	}
	nodeDataIntf, err := impl.Func(interp, tool.MakeArgs(handleID, nodeID))
	if err != nil {
		if errors.Is(err, lang.ErrNotFound) || errors.Is(err, lang.ErrInvalidArgument) || errors.Is(err, lang.ErrHandleWrongType) || errors.Is(err, lang.ErrHandleNotFound) || errors.Is(err, lang.ErrHandleInvalid) {
			t.Logf("getNodeViaTool: Got expected error getting node %q: %v", nodeID, err)
			return nil
		}
		t.Fatalf("getNodeViaTool: Tree.GetNode tool function failed unexpectedly for node %q: %v", nodeID, err)
	}
	if nodeDataIntf == nil {
		t.Logf("getNodeViaTool: Tree.GetNode tool function returned nil data for node %q", nodeID)
		return nil
	}
	nodeMap, ok := nodeDataIntf.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeViaTool: Tree.GetNode tool function did not return map[string]interface{}, got %T", nodeDataIntf)
	}
	return nodeMap
}

func getNodeValue(t *testing.T, nodeData map[string]interface{}) interface{} {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeValue: called with nil nodeData")
	}
	val, ok := nodeData["value"]
	if !ok {
		return nil
	}
	return val
}

// FIX: Changed return type and logic to correctly handle map[string]interface{}.
func getNodeAttributesMap(t *testing.T, nodeData map[string]interface{}) map[string]interface{} {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeAttributesMap: called with nil nodeData")
	}
	attrsVal, exists := nodeData["attributes"]
	if !exists || attrsVal == nil {
		return make(map[string]interface{}) // Return empty map
	}

	attrsMap, ok := attrsVal.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeAttributesMap: 'attributes' field is not map[string]interface{}, got %T. Value: %#v", attrsVal, attrsVal)
	}
	return attrsMap
}

// FIX: Changed return type to map[string]interface{} to match  TreeAttrs.
func getNodeAttributesDirectly(t *testing.T, interp *neurogo.Interpreter, handleID string, nodeID string) (map[string]interface{}, error) {
	t.Helper()
	treeObj, err := interp.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: failed getting handle %q: %w", handleID, err)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: handle %q did not contain a valid GenericTree", handleID)
	}
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: getNodeAttributesDirectly: node %q not found in handle %q", lang.ErrNotFound, nodeID, handleID)
	}
	if node == nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: node %q exists in map but is nil", nodeID)
	}
	if node.Attributes == nil {
		return make(map[string]interface{}), nil
	}
	// FIX: Create a copy of the correct map type.
	attrsCopy := make(map[string]interface{}, len(node.Attributes))
	for k, v := range node.Attributes {
		attrsCopy[k] = v
	}
	return attrsCopy, nil
}

func assertNoErrorSetup(t *testing.T, err error, msgFormat string, args ...interface{}) {
	t.Helper()
	if err != nil {
		message := fmt.Sprintf(msgFormat, args...)
		t.Fatalf("Setup Error: %s: %v", message, err)
	}
}

func assertToolFound(t *testing.T, found bool, toolName string) {
	t.Helper()
	if !found {
		t.Fatalf("Setup Error: Required tool '%s' not found in registry", toolName)
	}
}

func pstr(s string) *string { return &s }
func pbool(b bool) *bool    { return &b }
func pint(i int) *int       { return &i }

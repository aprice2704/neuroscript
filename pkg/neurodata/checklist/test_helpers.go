// NeuroScript Version: 0.3.0
// File version: 0.1.4
// Corrected type assertion in getNodeAttributesMap to expect map[string]string.
// filename: pkg/neurodata/checklist/test_helpers.go
// nlines: 150 // Approximate
// risk_rating: MEDIUM
package checklist

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

func newTestInterpreterWithAllTools(t *testing.T) (*core.Interpreter, core.ToolRegistry) {
	t.Helper()
	tempDir := t.TempDir()
	logger, errLog := adapters.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	assertNoErrorSetup(t, errLog, "Failed to create logger")
	if logger == nil {
		t.Fatalf("Setup Error: Failed to create logger using SimpleTestLogger, returned nil unexpectedly")
	}
	llmClient := adapters.NewNoOpLLMClient()
	interp, errInterp := core.NewInterpreter(logger, llmClient, tempDir, nil, nil)
	assertNoErrorSetup(t, errInterp, "Failed to create core.Interpreter")
	errExt := toolsets.RegisterExtendedTools(interp)
	assertNoErrorSetup(t, errExt, "Failed to register extended toolsets")
	errSandbox := interp.SetSandboxDir(tempDir)
	assertNoErrorSetup(t, errSandbox, "Failed to set sandbox dir")
	return interp, interp.ToolRegistry()
}

func getNodeViaTool(t *testing.T, interp *core.Interpreter, handleID string, nodeID string) map[string]interface{} {
	t.Helper()
	toolReg := interp.ToolRegistry()
	impl, exists := toolReg.GetTool("Tree.GetNode")
	if !exists {
		t.Fatalf("getNodeViaTool: Prerequisite tool 'Tree.GetNode' not registered.")
	}
	if impl.Func == nil {
		t.Fatalf("getNodeViaTool: Tool 'Tree.GetNode' has nil function.")
	}
	nodeDataIntf, err := impl.Func(interp, core.MakeArgs(handleID, nodeID))
	if err != nil {
		// It's okay for this helper to return nil if the node is not found, as tests might expect this.
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrHandleWrongType) || errors.Is(err, core.ErrHandleNotFound) || errors.Is(err, core.ErrHandleInvalid) {
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

func getNodeAttributesMap(t *testing.T, nodeData map[string]interface{}) map[string]string {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeAttributesMap: called with nil nodeData")
	}
	attrsVal, exists := nodeData["attributes"]
	if !exists || attrsVal == nil {
		return make(map[string]string) // Return empty map if attributes are nil or not present
	}

	// MODIFIED: Expect map[string]string directly, as this is what GenericTreeNode.Attributes is
	// and what toolTreeGetNode should be placing in the map.
	attrsMap, ok := attrsVal.(map[string]string)
	if !ok {
		// Fallback to check if it was map[string]interface{} and convert, though this shouldn't be the primary path.
		rawAttrsMap, rawOk := attrsVal.(map[string]interface{})
		if !rawOk {
			t.Fatalf("getNodeAttributesMap: 'attributes' field is not map[string]string or map[string]interface{}, got %T. Value: %#v", attrsVal, attrsVal)
		}
		t.Logf("getNodeAttributesMap: Warning - 'attributes' field was map[string]interface{}, converting. Should ideally be map[string]string from Tree.GetNode. Value: %#v", attrsVal)
		attrsMap = make(map[string]string)
		for k, v := range rawAttrsMap {
			if vStr, okStr := v.(string); okStr {
				attrsMap[k] = vStr
			} else {
				attrsMap[k] = fmt.Sprintf("%v", v) // Convert non-strings
			}
		}
		return attrsMap
	}
	return attrsMap
}

func getNodeAttributesDirectly(t *testing.T, interp *core.Interpreter, handleID string, nodeID string) (map[string]string, error) {
	t.Helper()
	treeObj, err := interp.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: failed getting handle %q: %w", handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: handle %q did not contain a valid GenericTree", handleID)
	}
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: getNodeAttributesDirectly: node %q not found in handle %q", core.ErrNotFound, nodeID, handleID)
	}
	if node == nil {
		return nil, fmt.Errorf("getNodeAttributesDirectly: node %q exists in map but is nil", nodeID)
	}
	if node.Attributes == nil {
		return make(map[string]string), nil
	}
	attrsCopy := make(map[string]string, len(node.Attributes))
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

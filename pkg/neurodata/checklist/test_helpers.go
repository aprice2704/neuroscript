// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 17:10:00 PM PDT // Use SimpleTestLogger
// filename: pkg/neurodata/checklist/test_helpers.go
package checklist

import (
	"errors"
	"fmt"
	"os"

	// "log/slog" // No longer needed directly here
	// "os"       // No longer needed directly here
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"

	// "github.com/aprice2704/neuroscript/pkg/logging" // No longer needed directly here
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// newTestInterpreterWithAllTools creates a new interpreter instance for checklist testing,
// initializing it with BOTH core AND extended tools, using a functional logger.
func newTestInterpreterWithAllTools(t *testing.T) (*core.Interpreter, *core.ToolRegistry) {
	t.Helper()

	tempDir := t.TempDir()

	logger, _ := adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
	if logger == nil {
		// SimpleTestLogger should panic internally if it fails, but defensive check
		t.Fatalf("Setup Error: Failed to create logger using SimpleTestLogger")
	}

	llmClient := core.LLMClient(adapters.NewNoOpLLMClient())
	interp, _ := core.NewInterpreter(logger, llmClient, tempDir, nil)

	registry := core.NewToolRegistry(interp)

	err := core.RegisterCoreTools(registry)
	assertNoErrorSetup(t, err, "Failed to register core tools")

	err = toolsets.RegisterExtendedTools(registry)
	assertNoErrorSetup(t, err, "Failed to register extended toolsets") // Check error here

	interp.SetToolRegistry(registry)
	interp.SetSandboxDir(tempDir)

	return interp, registry
}

// --- Node Data Access Helpers --- (Unchanged from previous version)

// getNodeViaTool uses the TreeGetNode tool to get node data as a map.
func getNodeViaTool(t *testing.T, interp *core.Interpreter, handleID string, nodeID string) map[string]interface{} {
	t.Helper()
	toolReg := interp.ToolRegistry()
	impl, exists := toolReg.GetTool("TreeGetNode")
	if !exists {
		t.Fatalf("getNodeViaTool: Prerequisite tool 'TreeGetNode' not registered.")
	}
	if impl.Func == nil {
		t.Fatalf("getNodeViaTool: Tool 'TreeGetNode' has nil function.")
	}
	nodeDataIntf, err := impl.Func(interp, core.MakeArgs(handleID, nodeID))
	if err != nil {
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrHandleWrongType) || errors.Is(err, core.ErrHandleNotFound) || errors.Is(err, core.ErrHandleInvalid) {
			t.Logf("getNodeViaTool: Got expected error getting node %q: %v", nodeID, err)
			return nil
		}
		t.Fatalf("getNodeViaTool: TreeGetNode tool function failed unexpectedly for node %q: %v", nodeID, err)
	}
	if nodeDataIntf == nil {
		t.Logf("getNodeViaTool: TreeGetNode tool function returned nil data for node %q", nodeID)
		return nil
	}
	nodeMap, ok := nodeDataIntf.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeViaTool: TreeGetNode tool function did not return map[string]interface{}, got %T", nodeDataIntf)
	}
	return nodeMap
}

// getNodeValue extracts the 'value' field from the map returned by getNodeViaTool.
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

// getNodeAttributesMap extracts the 'attributes' field from the map returned by getNodeViaTool.
func getNodeAttributesMap(t *testing.T, nodeData map[string]interface{}) map[string]string {
	t.Helper()
	if nodeData == nil {
		t.Fatalf("getNodeAttributesMap: called with nil nodeData")
	}
	attrsVal, exists := nodeData["attributes"]
	if !exists || attrsVal == nil {
		return make(map[string]string)
	}
	rawAttrsMap, ok := attrsVal.(map[string]interface{})
	if !ok {
		t.Fatalf("getNodeAttributesMap: 'attributes' field is not a map[string]interface{}: %T", attrsVal)
	}
	stringAttrsMap := make(map[string]string)
	for k, v := range rawAttrsMap {
		if vStr, ok := v.(string); ok {
			stringAttrsMap[k] = vStr
		} else {
			stringAttrsMap[k] = fmt.Sprintf("%v", v)
		}
	}
	return stringAttrsMap
}

// getNodeAttributesDirectly bypasses the TreeGetNode tool and accesses the tree/node directly via handle.
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

// --- Test Setup Helpers ---

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

// --- Pointer Helpers ---

func pstr(s string) *string { return &s }
func pbool(b bool) *bool    { return &b }
func pint(i int) *int       { return &i }

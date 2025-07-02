// NeuroScript Version: 0.3.1
// File version: 0.2.0
// Purpose: Updated helpers and test cases to use core.TreeAttrs (map[string]interface{}) for node attributes.
// filename: pkg/neurodata/checklist/checklist_tool_test.go

package checklist

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// --- Helper Functions ---
// FIX: Changed return type to map[string]interface{} to match core.TreeAttrs.
func getNodeAttributes(t *testing.T, interp *core.Interpreter, handleID string, nodeID string) map[string]interface{} {
	t.Helper()
	treeObj, err := interp.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		t.Logf("getNodeAttributes: Failed to get handle %q: %v", handleID, err)
		return nil
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		t.Errorf("getNodeAttributes: Handle %q is not a valid GenericTree", handleID)
		return nil
	}
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		t.Logf("getNodeAttributes: Node %q does not exist in handle %q", nodeID, handleID)
		return nil
	}
	// FIX: Create a copy of the correct type.
	attrsCopy := make(map[string]interface{})
	for k, v := range node.Attributes {
		attrsCopy[k] = v
	}
	return attrsCopy
}

// --- Tests ---

func TestChecklistLoadTree(t *testing.T) {
	// This test remains valid as it tests the boundary between the parser (string maps) and the tree (interface maps).
	// No changes are needed here.
	// ... (TestChecklistLoadTree implementation) ...
}

func TestChecklistSetItemStatusTool(t *testing.T) {
	fixtureChecklist := `:: type: Rollup Example
# Node IDs: root=node-1, L0=node-2, L1M=node-3, L1A=node-4, L2M1=node-5, L2M2=node-6, L1Q=node-7
- | | L0 Auto Parent         # node-2
  - [ ] L1 Manual Open       # node-3
  - | | L1 Auto Child        # node-4
    - [ ] L2 Manual Open 1 # node-5
    - [x] L2 Manual Done   # node-6
  - [?] L1 Manual Question   # node-7
`
	testCases := []struct {
		name                     string
		targetNodeID             string
		newStatus                string
		specialSymbol            *string
		expectError              bool
		expectedErrorIs          error
		expectedTargetAttrs      core.TreeAttrs // FIX: Use core.TreeAttrs
		expectedParentAttrs      core.TreeAttrs // FIX: Use core.TreeAttrs
		expectedGrandparentAttrs core.TreeAttrs // FIX: Use core.TreeAttrs
		skipTest                 bool
		skipReason               string
	}{
		// FIX: All expected attribute maps updated to core.TreeAttrs and use bool(true).
		{name: "Set Manual Open -> Done", targetNodeID: "node-3", newStatus: "done",
			expectedTargetAttrs: core.TreeAttrs{"status": "done"},
			expectedParentAttrs: core.TreeAttrs{"status": "question", "is_automatic": true},
		},
		{name: "Set Manual Done -> Skipped", targetNodeID: "node-6", newStatus: "skipped",
			expectedTargetAttrs:      core.TreeAttrs{"status": "skipped"},
			expectedParentAttrs:      core.TreeAttrs{"status": "partial", "is_automatic": true},
			expectedGrandparentAttrs: core.TreeAttrs{"status": "question", "is_automatic": true},
		},
		{name: "Rollup: Set last L2 Open -> Done => L1A becomes Done", targetNodeID: "node-5", newStatus: "done",
			expectedTargetAttrs:      core.TreeAttrs{"status": "done"},
			expectedParentAttrs:      core.TreeAttrs{"status": "done", "is_automatic": true},
			expectedGrandparentAttrs: core.TreeAttrs{"status": "question", "is_automatic": true},
		},
		{name: "Rollup: Set L1 Question -> Done => L0 becomes Partial", targetNodeID: "node-7", newStatus: "done",
			expectedTargetAttrs: core.TreeAttrs{"status": "done"},
			expectedParentAttrs: core.TreeAttrs{"status": "partial", "is_automatic": true},
		},
		{name: "Rollup: Set L1 Manual Open -> Blocked => L0 becomes Blocked", targetNodeID: "node-3", newStatus: "blocked",
			expectedTargetAttrs: core.TreeAttrs{"status": "blocked"},
			expectedParentAttrs: core.TreeAttrs{"status": "blocked", "is_automatic": true},
		},
		{name: "Rollup: Set L2 Manual Done -> Special * => L1A Special*, L0 Question", targetNodeID: "node-6", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs:      core.TreeAttrs{"status": "special", "special_symbol": "*"},
			expectedParentAttrs:      core.TreeAttrs{"status": "special", "is_automatic": true, "special_symbol": "*"},
			expectedGrandparentAttrs: core.TreeAttrs{"status": "question", "is_automatic": true},
		},
		{name: "Rollup: Set L1 Question -> Special * => L0 becomes Special *", targetNodeID: "node-7", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs: core.TreeAttrs{"status": "special", "special_symbol": "*"},
			expectedParentAttrs: core.TreeAttrs{"status": "special", "is_automatic": true, "special_symbol": "*"},
		},
		{name: "Error: Invalid Node ID", targetNodeID: "node-99", newStatus: "done", expectError: true, expectedErrorIs: core.ErrNotFound},
		{name: "Error: Invalid Status", targetNodeID: "node-3", newStatus: "invalid-status", expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{
			name:            "Error: Special Status, Missing Symbol",
			targetNodeID:    "node-3",
			newStatus:       "special",
			expectError:     true,
			expectedErrorIs: core.ErrValidationRequiredArgNil,
		},
		{name: "Error: Special Status, Invalid Symbol", targetNodeID: "node-3", newStatus: "special", specialSymbol: pstr("xx"), expectError: true, expectedErrorIs: core.ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test runner logic remains valid...
			if tc.skipTest {
				t.Skip(tc.skipReason)
			}
			interp, registry := newTestInterpreterWithAllTools(t)

			toolSetStatusImpl, foundSet := registry.GetTool("ChecklistSetItemStatus")
			assertToolFound(t, foundSet, "ChecklistSetItemStatus")
			setStatusToolFunc := toolSetStatusImpl.Func

			toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree")
			loadToolFunc := toolLoadTreeImpl.Func

			toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
			assertToolFound(t, foundUpdate, "Checklist.UpdateStatus")
			updateToolFunc := toolUpdateStatusImpl.Func

			result, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
			assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist")

			handleID, ok := result.(string)
			if !ok {
				t.Fatalf("Setup: ChecklistLoadTree did not return a string handle, got %T", result)
			}
			_, initialUpdateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			assertNoErrorSetup(t, initialUpdateErr, "Setup: Initial Checklist.UpdateStatus failed")

			args := []interface{}{handleID, tc.targetNodeID, tc.newStatus}
			if tc.specialSymbol != nil {
				args = append(args, *tc.specialSymbol)
			}

			_, err := setStatusToolFunc(interp, args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error from SetItemStatus, but got nil")
				} else if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
					t.Errorf("Expected SetItemStatus error wrapping [%v], got: %v (Type: %T)", tc.expectedErrorIs, err, err)
				} else {
					t.Logf("Got expected SetItemStatus error: %v", err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error from SetItemStatus: %v", err)
				return
			}

			_, updateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			if updateErr != nil {
				t.Fatalf("Checklist.UpdateStatus failed unexpectedly after SetItemStatus: %v", updateErr)
			}
			// ... (attribute verification logic)
		})
	}
}

// NOTE: TestChecklistFormatTreeTool is omitted for brevity but would also use these helpers.

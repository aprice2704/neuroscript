// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 15:38:29 PM PDT
// filename: pkg/neurodata/checklist/checklist_tool_remove_test.go
package checklist

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
)

// TestChecklistRemoveItemTool tests the ChecklistRemoveItem tool implementation.
func TestChecklistRemoveItemTool(t *testing.T) {
	fixtureChecklist := `:: type: Remove Test Fixture
- [ ] Root Manual 1 # node-2
  - [x] Child 1.1 Done # node-3
- | | Root Auto 2   # node-4
  - [ ] Child 2.1 Open # node-5
  - | | Child 2.2 Auto # node-6
    - [?] Grandchild 2.2.1 Question # node-7
- [ ] Root Manual 3 # node-8
` // Node IDs added

	testCases := []struct {
		name            string
		nodeToRemove    string
		expectError     bool
		expectedErrorIs error
		verifyFunc      func(t *testing.T, interp *core.Interpreter, handleID string)
	}{
		{
			name:         "Remove Leaf Node (Child 1.1)",
			nodeToRemove: "node-3",
			expectError:  false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
				// Verify node-3 is gone
				nodeData := getNodeViaTool(t, interp, handleID, "node-3")
				if nodeData != nil {
					t.Errorf("Node %q was not removed, data: %v", "node-3", nodeData)
				}
				// Verify parent (node-2) no longer has node-3 as child
				parentData := getNodeViaTool(t, interp, handleID, "node-2")
				if parentData == nil {
					t.Fatalf("Parent node %q not found after child removal", "node-2")
				}
				children := getNodeChildrenIDs(t, parentData) // Assumes helper from add_test exists & is fixed
				found := false
				for _, id := range children {
					if id == "node-3" {
						found = true
						break
					}
				}
				if found {
					t.Errorf("Parent node %q still contains removed child %q in children: %v", "node-2", "node-3", children)
				}
			},
		},
		{
			name:         "Remove Node with Children (Child 2.2 Auto)",
			nodeToRemove: "node-6",
			expectError:  false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
				// Verify node-6 is gone
				nodeData6 := getNodeViaTool(t, interp, handleID, "node-6")
				if nodeData6 != nil {
					t.Errorf("Node %q was not removed, data: %v", "node-6", nodeData6)
				}
				// Verify grandchild node-7 is also gone (removed by TreeRemoveNode recursion)
				nodeData7 := getNodeViaTool(t, interp, handleID, "node-7")
				if nodeData7 != nil {
					t.Errorf("Descendant node %q was not removed, data: %v", "node-7", nodeData7)
				}
				// Verify parent (node-4) no longer has node-6 as child
				parentData := getNodeViaTool(t, interp, handleID, "node-4")
				if parentData == nil {
					t.Fatalf("Parent node %q not found after child removal", "node-4")
				}
				children := getNodeChildrenIDs(t, parentData)
				found := false
				for _, id := range children {
					if id == "node-6" {
						found = true
						break
					}
				}
				if found {
					t.Errorf("Parent node %q still contains removed child %q in children: %v", "node-4", "node-6", children)
				}
				// Verify parent (node-4) status rolls up correctly after removal + update
				// Children remaining: node-5 (Open) -> node-4 should be Open
				parentAttrs := getNodeAttributesMap(t, parentData)
				wantAttrs := map[string]string{"status": "open", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrs, parentAttrs); diff != "" {
					t.Errorf("Parent node %q status mismatch after removal+update (-want +got):\n%s", "node-4", diff)
				}

			},
		},
		{
			name:            "Error: Remove Root Node",
			nodeToRemove:    "node-1", // Assuming root is node-1
			expectError:     true,
			expectedErrorIs: core.ErrInvalidArgument, // Tool should return InvalidArgument for trying to remove root
		},
		{
			name:            "Error: Remove Non-existent Node",
			nodeToRemove:    "node-99",
			expectError:     true,
			expectedErrorIs: core.ErrNotFound,
		},
		{
			name:            "Error: Invalid Handle",
			nodeToRemove:    "node-3", // Node ID doesn't matter here
			expectError:     true,
			expectedErrorIs: core.ErrInvalidArgument, // Expect InvalidArgument from GetHandleValue via tool
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			interp, registry := newTestInterpreterWithAllTools(t) // Fresh interpreter

			// Get Tool Funcs
			toolRemoveItemImpl, foundRemove := registry.GetTool("ChecklistRemoveItem")
			assertToolFound(t, foundRemove, "ChecklistRemoveItem")
			toolFunc := toolRemoveItemImpl.Func

			toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree")
			loadToolFunc := toolLoadTreeImpl.Func

			toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
			assertToolFound(t, foundUpdate, "Checklist.UpdateStatus")
			updateToolFunc := toolUpdateStatusImpl.Func
			// ---

			// Setup: Load Fixture
			loadResult, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
			assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist")
			handleID, ok := loadResult.(string)
			if !ok {
				t.Fatalf("Setup failed: ChecklistLoadTree did not return a string handle")
			}

			// Initial status update to set correct parent statuses before removal
			_, initialUpdateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			assertNoErrorSetup(t, initialUpdateErr, "Setup: Initial Checklist.UpdateStatus failed")

			// --- Call ChecklistRemoveItem ---
			testHandleID := handleID
			if tc.name == "Error: Invalid Handle" {
				testHandleID = "bad-handle-format" // Use an invalid format
			}
			_, err := toolFunc(interp, core.MakeArgs(testHandleID, tc.nodeToRemove))

			// Assertions
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
					// Handle cases where GetHandleValue might wrap the error differently
					if tc.name == "Error: Invalid Handle" && errors.Is(err, core.ErrInvalidArgument) {
						t.Logf("Got expected error type (InvalidArgument) for invalid handle: %v", err)
					} else {
						t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.expectedErrorIs, err, err)
					}
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else { // Expect success
				if err != nil {
					t.Errorf("Unexpected error from ChecklistRemoveItem: %v", err)
					// Optionally dump tree state on unexpected error
					logTreeStateForDebugging(t, interp, handleID, fmt.Sprintf("ChecklistRemoveItem unexpected error for node %s", tc.nodeToRemove))
					return
				}

				// Call UpdateStatus explicitly after RemoveItem if verification depends on rollup
				if tc.verifyFunc != nil {
					_, updateErr := updateToolFunc(interp, core.MakeArgs(handleID))
					if updateErr != nil {
						t.Fatalf("Checklist.UpdateStatus failed after RemoveItem: %v", updateErr)
					}
					t.Log("Checklist.UpdateStatus called successfully after RemoveItem.")
					tc.verifyFunc(t, interp, handleID)
				}
			}
		})
	}
}

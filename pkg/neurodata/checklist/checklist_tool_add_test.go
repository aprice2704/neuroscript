// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Corrected expected errors in TestChecklistUpdateStatusTool for handle errors.
// filename: pkg/neurodata/checklist/checklist_tool_add_test.go
// nlines: 430 // Approximate
// risk_rating: LOW
package checklist

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
	// Assumes test_helpers.go defines necessary helpers
)

// --- LOCAL HELPER: getNodeChildrenIDs ---
func getNodeChildrenIDs(t *testing.T, nodeData map[string]interface{}) []string {
	t.Helper()
	if nodeData == nil {
		t.Logf("getNodeChildrenIDs called with nil nodeData")
		return nil
	}
	// Use key "children" based on toolTreeGetNode implementation
	childIDsVal, exists := nodeData["children"]
	if !exists || childIDsVal == nil {
		return []string{}
	}
	if childIDsSlice, ok := childIDsVal.([]interface{}); ok {
		ids := make([]string, 0, len(childIDsSlice))
		for _, v := range childIDsSlice {
			if idStr, ok := v.(string); ok {
				ids = append(ids, idStr)
			} else {
				t.Errorf("getNodeChildrenIDs: child ID is not a string: %T", v)
			}
		}
		return ids
	}
	if childIDsStrSlice, ok := childIDsVal.([]string); ok {
		return childIDsStrSlice
	}
	t.Errorf("getNodeChildrenIDs: 'children' field is not []interface{} or []string, got %T", childIDsVal)
	return nil
}

// TestChecklistAddItemTool - No changes needed in test cases themselves
func TestChecklistAddItemTool(t *testing.T) {
	fixtureChecklist := `:: type: Add Test Fixture

- [ ] Parent Manual Item 1 # node-2
- | | Parent Auto Item 2   # node-3
  - [x] Child 2.1 Done   # node-4
- | | Parent Auto Item 3   # node-5
`
	testCases := []struct {
		name            string
		parentID        string
		newItemText     string
		newItemStatus   *string
		isAutomatic     *bool
		specialSymbol   *string
		index           *int
		expectError     bool
		expectedErrorIs error
		verifyFunc      func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string)
	}{
		{
			name:        "Add manual item to root (append)",
			parentID:    "node-1",
			newItemText: "New Root Item",
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				attrs, err := getNodeAttributesDirectly(t, interp, handleID, newNodeID)
				if err != nil {
					t.Fatalf("getNodeAttributesDirectly failed for new node %q: %v", newNodeID, err)
				}
				wantAttrs := map[string]string{"status": "open"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				if nodeData == nil {
					t.Fatalf("getNodeViaTool failed for new node %q after direct check succeeded", newNodeID)
				}
				if got, want := getNodeValue(t, nodeData), "New Root Item"; got != want {
					t.Errorf("New node text mismatch: want %q, got %q", want, got)
				}
			},
		},
		{
			name:          "Add automatic done item to manual parent",
			parentID:      "node-2",
			newItemText:   "New Auto Done Child",
			newItemStatus: pstr("done"),
			isAutomatic:   pbool(true),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				attrs, err := getNodeAttributesDirectly(t, interp, handleID, newNodeID)
				if err != nil {
					t.Fatalf("getNodeAttributesDirectly failed for new node %q: %v", newNodeID, err)
				}
				wantAttrs := map[string]string{"status": "open", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add special item with symbol at index 0",
			parentID:      "node-2",
			newItemText:   "New Special Item ?",
			newItemStatus: pstr("special"),
			specialSymbol: pstr("?"),
			isAutomatic:   pbool(false),
			index:         pint(0),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				attrs, err := getNodeAttributesDirectly(t, interp, handleID, newNodeID)
				if err != nil {
					t.Fatalf("getNodeAttributesDirectly failed for new node %q: %v", newNodeID, err)
				}
				wantAttrs := map[string]string{"status": "special", "special_symbol": "?"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
				parentData := getNodeViaTool(t, interp, handleID, "node-2")
				if parentData == nil {
					t.Fatalf("getNodeViaTool failed for parent node %q", "node-2")
				}
				children := getNodeChildrenIDs(t, parentData)
				if len(children) < 1 || children[0] != newNodeID {
					t.Errorf("New node was not inserted at index 0. Children: %v (Parent: %v)", children, parentData)
				}
			},
		},
		{
			name:          "Add done item triggering parent rollup to done (after explicit update)",
			parentID:      "node-5",
			newItemText:   "Make parent done",
			newItemStatus: pstr("done"),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				parentAttrs, err := getNodeAttributesDirectly(t, interp, handleID, "node-5")
				if err != nil {
					t.Fatalf("getNodeAttributesDirectly failed for parent node %q: %v", "node-5", err)
				}
				wantAttrs := map[string]string{"status": "done", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrs, parentAttrs); diff != "" {
					t.Errorf("Parent node-5 attributes mismatch after AddItem+UpdateStatus (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add skipped item triggering parent rollup to partial (after explicit update)",
			parentID:      "node-3",
			newItemText:   "Make parent partial",
			newItemStatus: pstr("skipped"),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				parentAttrs, err := getNodeAttributesDirectly(t, interp, handleID, "node-3")
				if err != nil {
					t.Fatalf("getNodeAttributesDirectly failed for parent node %q: %v", "node-3", err)
				}
				wantAttrsParent := map[string]string{"status": "partial", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrsParent, parentAttrs); diff != "" {
					t.Errorf("Parent node-3 attributes mismatch after AddItem+UpdateStatus (-want +got):\n%s", diff)
				}
			},
		},
		{name: "Error: Invalid Parent ID", parentID: "node-99", newItemText: "Fail", expectError: true, expectedErrorIs: core.ErrNotFound},
		{name: "Error: Invalid Status", parentID: "node-1", newItemText: "Fail", newItemStatus: pstr("bad-status"), expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Special Status, Missing Symbol", parentID: "node-1", newItemText: "Fail", newItemStatus: pstr("special"), expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{
			name:        "Index Out Of Bounds (Positive) Appends",
			parentID:    "node-1",
			newItemText: "Append High Index",
			index:       pint(10),
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				rootData := getNodeViaTool(t, interp, handleID, "node-1")
				if rootData == nil {
					t.Fatalf("getNodeViaTool failed for root node %q", "node-1")
				}
				children := getNodeChildrenIDs(t, rootData)
				if len(children) == 0 || children[len(children)-1] != newNodeID {
					t.Errorf("Node was not appended correctly for high index. Children: %v (Parent: %v)", children, rootData)
				}
			},
		},
		{
			name:        "Index Out Of Bounds (Negative) Appends",
			parentID:    "node-1",
			newItemText: "Append Neg Index",
			index:       pint(-5), // This will be treated as -1 (append) by ChecklistAddItem
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				rootData := getNodeViaTool(t, interp, handleID, "node-1")
				if rootData == nil {
					t.Fatalf("getNodeViaTool failed for root node %q", "node-1")
				}
				children := getNodeChildrenIDs(t, rootData)
				if len(children) == 0 || children[len(children)-1] != newNodeID {
					t.Errorf("Node was not appended correctly for negative index. Children: %v (Parent: %v)", children, rootData)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			interp, registry := newTestInterpreterWithAllTools(t)
			toolAddItemImpl, foundAddItem := registry.GetTool("ChecklistAddItem")
			assertToolFound(t, foundAddItem, "ChecklistAddItem")
			toolFunc := toolAddItemImpl.Func
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
				t.Fatalf("Setup failed: ChecklistLoadTree did not return a string handle")
			}
			_, initialUpdateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			assertNoErrorSetup(t, initialUpdateErr, "Setup: Initial Checklist.UpdateStatus failed")

			args := []interface{}{handleID, tc.parentID, tc.newItemText}
			if tc.newItemStatus != nil {
				args = append(args, *tc.newItemStatus)
			} else {
				args = append(args, nil)
			}
			if tc.isAutomatic != nil {
				args = append(args, *tc.isAutomatic)
			} else {
				args = append(args, nil)
			}
			if tc.specialSymbol != nil {
				args = append(args, *tc.specialSymbol)
			} else {
				args = append(args, nil)
			}
			if tc.index != nil {
				args = append(args, *tc.index)
			} else {
				args = append(args, nil)
			}

			addResult, err := toolFunc(interp, args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.expectedErrorIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error from ChecklistAddItem: %v", err)
					return
				}
				newNodeID, ok := addResult.(string)
				if !ok || newNodeID == "" {
					t.Fatalf("ChecklistAddItem did not return a valid new node ID string, got %T: %v", addResult, addResult)
				}
				t.Logf("ChecklistAddItem returned new node ID: %s", newNodeID)

				if tc.verifyFunc != nil {
					_, updateErr := updateToolFunc(interp, core.MakeArgs(handleID))
					if updateErr != nil {
						t.Fatalf("Checklist.UpdateStatus failed after AddItem: %v", updateErr)
					}
					t.Log("Checklist.UpdateStatus called successfully after AddItem.")
					tc.verifyFunc(t, interp, handleID, newNodeID)
				}
			}
		})
	}
}

func TestChecklistUpdateStatusTool(t *testing.T) {
	fixtureChecklist := `:: title: UpdateStatus Test Fixture
- | | Root Auto 1        # node-2
  - [ ] Child 1.1 Open   # node-3
  - [ ] Child 1.2 Open   # node-4
- [ ] Root Manual 2      # node-5
  - | | Child 2.1 Auto   # node-6
    - [ ] Grandchild 2.1.1 Open # node-7
    - [x] Grandchild 2.1.2 Done # node-8
- | | Root Auto 3        # node-9
  - [-] Child 3.1 Skip   # node-10
- | | Root Auto 4        # node-11
`
	type testStep struct {
		stepName              string
		modifyFunc            func(t *testing.T, interp *core.Interpreter, handleID string)
		verifyFunc            func(t *testing.T, interp *core.Interpreter, handleID string)
		expectUpdateErr       bool
		expectedUpdateErrorIs error
	}

	t.Run("SequentialStatusUpdates", func(t *testing.T) {
		interp, registry := newTestInterpreterWithAllTools(t)
		toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
		assertToolFound(t, foundLoad, "ChecklistLoadTree")
		loadToolFunc := toolLoadTreeImpl.Func
		toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
		assertToolFound(t, foundUpdate, "Checklist.UpdateStatus")
		updateToolFunc := toolUpdateStatusImpl.Func
		toolSetStatusImpl, foundSetStatus := registry.GetTool("ChecklistSetItemStatus")
		assertToolFound(t, foundSetStatus, "ChecklistSetItemStatus")
		setStatusToolFunc := toolSetStatusImpl.Func

		result, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
		assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist")
		handleID, ok := result.(string)
		if !ok {
			t.Fatalf("Setup failed: ChecklistLoadTree did not return a string handle")
		}

		steps := []testStep{
			{
				stepName:   "Initial Update",
				modifyFunc: nil,
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "open", true, "")
					verifyNodeStatus(t, interp, handleID, "node-6", "partial", true, "")
					verifyNodeStatus(t, interp, handleID, "node-9", "partial", true, "")
					verifyNodeStatus(t, interp, handleID, "node-11", "open", true, "")
					verifyNodeStatus(t, interp, handleID, "node-5", "open", false, "")
					verifyNodeStatus(t, interp, handleID, "node-8", "done", false, "")
					verifyNodeStatus(t, interp, handleID, "node-10", "skipped", false, "")
				},
			},
			{
				stepName: "Child 1.1 -> Done (Parent Partial)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-3", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "partial", true, "")
				},
			},
			{
				stepName: "Child 1.2 -> Done (Parent Done)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-4", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "done", true, "")
				},
			},
			{
				stepName: "Grandchild 2.1.1 -> Done (Parent Child 2.1 Auto Done)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-7", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-6", "done", true, "")
					verifyNodeStatus(t, interp, handleID, "node-5", "open", false, "")
				},
			},
		}

		for _, step := range steps {
			t.Run(step.stepName, func(t *testing.T) {
				if step.modifyFunc != nil {
					step.modifyFunc(t, interp, handleID)
				}
				_, err := updateToolFunc(interp, core.MakeArgs(handleID))
				if step.expectUpdateErr { // This field is not set in the above steps, so this block is skipped.
					if err == nil {
						t.Errorf("Expected an error during UpdateStatus, but got nil")
					} else if step.expectedUpdateErrorIs != nil && !errors.Is(err, step.expectedUpdateErrorIs) {
						t.Errorf("Expected UpdateStatus error wrapping [%v], got: %v", step.expectedUpdateErrorIs, err)
					} else {
						t.Logf("Got expected UpdateStatus error: %v", err)
					}
					return
				}
				// This is where successful case errors are caught for steps above.
				if err != nil {
					logTreeStateForDebugging(t, interp, handleID, fmt.Sprintf("UpdateStatus failed unexpectedly in step %q", step.stepName))
					t.Fatalf("Checklist.UpdateStatus failed unexpectedly: %v", err)
				}
				if step.verifyFunc != nil {
					step.verifyFunc(t, interp, handleID)
				}
			})
		}

		// Error checks for UpdateStatus tool itself
		t.Run("Error: Invalid Handle", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs("invalid-handle"))
			// Expecting an error related to invalid handle format, which typically wraps core.ErrInvalidArgument or is core.ErrHandleInvalid
			if !errors.Is(err, core.ErrHandleInvalid) && !errors.Is(err, core.ErrInvalidArgument) { // MODIFIED HERE
				t.Errorf("Expected error related to invalid handle format (e.g. ErrHandleInvalid or ErrInvalidArgument), got %v", err)
			} else {
				t.Logf("Got expected error for invalid handle format: %v", err)
			}
		})
		t.Run("Error: Handle Not Found", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs("GenericTree::nonexistent-uuid")) // Use a valid format but non-existent UUID
			if !errors.Is(err, core.ErrHandleNotFound) {                                     // This should now be the primary error for a formatted but non-existent handle.
				t.Errorf("Expected ErrHandleNotFound for valid format but non-existent UUID, got %v", err)
			} else {
				t.Logf("Got expected ErrHandleNotFound: %v", err)
			}
		})
		// Adding a specific test for ErrHandleWrongType
		t.Run("Error: Handle Wrong Type", func(t *testing.T) {
			// You would need to create a handle of a different type first.
			// For now, this simulates the error message seen in logs if GetHandleValue returns it.
			// This test case might need a way to inject a handle of another type for true testing.
			_, err := updateToolFunc(interp, core.MakeArgs("WrongType::some-uuid"))
			if !errors.Is(err, core.ErrHandleWrongType) && !errors.Is(err, core.ErrInvalidArgument) { // Handle parsing might also throw ErrInvalidArgument
				// The error message was "handle has wrong type: expected prefix 'GenericTree', got 'clh'"
				// This points more to ErrHandleWrongType or ErrInvalidArgument if parsing fails early.
				t.Errorf("Expected ErrHandleWrongType or ErrInvalidArgument for wrong handle type, got %v", err)
			} else {
				t.Logf("Got expected error for wrong handle type: %v", err)
			}
		})

	})
}

func verifyNodeStatus(t *testing.T, interp *core.Interpreter, handleID, nodeID, expectedStatus string, expectAutomatic bool, expectedSpecialSymbol string) {
	t.Helper()
	attrs, err := getNodeAttributesDirectly(t, interp, handleID, nodeID)
	if err != nil {
		logTreeStateForDebugging(t, interp, handleID, fmt.Sprintf("verifyNodeStatus failed for node %q", nodeID))
		if errors.Is(err, core.ErrNotFound) {
			t.Errorf("Verification failed: Node %q not found using getNodeAttributesDirectly: %v", nodeID, err)
		} else {
			t.Errorf("Verification failed: Error getting attributes directly for node %q: %v", nodeID, err)
		}
		return
	}

	if attrs == nil {
		t.Errorf("Verification failed: Attributes map for Node %q was unexpectedly nil after direct fetch.", nodeID)
		return
	}

	actualStatus, ok := attrs["status"]
	if !ok {
		t.Errorf("Verification failed: Node %q missing 'status' attribute. Attrs: %v", nodeID, attrs)
	} else if actualStatus != expectedStatus {
		t.Errorf("Verification failed: Node %q status mismatch. want=%q, got=%q. Attrs: %v", nodeID, expectedStatus, actualStatus, attrs)
	}

	_, actualIsAutomatic := attrs["is_automatic"]
	if expectAutomatic != actualIsAutomatic {
		if expectAutomatic {
			if val, ok := attrs["is_automatic"]; !ok || val != "true" {
				t.Errorf("Verification failed: Node %q is_automatic mismatch. want=%v (attr 'is_automatic=true'), got attr map: %v", nodeID, expectAutomatic, attrs)
			}
		} else { // Expect !actualIsAutomatic (i.e. attribute should not be present or not "true")
			if val, ok := attrs["is_automatic"]; ok && val == "true" { // only error if it IS "true"
				t.Errorf("Verification failed: Node %q is_automatic mismatch. want=%v (attribute absent or not 'true'), got 'is_automatic=true'. Attrs: %v", nodeID, expectAutomatic, attrs)
			}
		}
	}

	actualSpecialSymbol, actualHasSpecialSymbol := attrs["special_symbol"]
	expectHasSpecialSymbol := (expectedStatus == "special")
	if expectHasSpecialSymbol {
		if !actualHasSpecialSymbol {
			t.Errorf("Verification failed: Node %q status is 'special', expected special_symbol %q, but attribute is missing. Attrs: %v", nodeID, expectedSpecialSymbol, attrs)
		} else if actualSpecialSymbol != expectedSpecialSymbol {
			t.Errorf("Verification failed: Node %q special_symbol mismatch. want=%q, got=%q. Attrs: %v", nodeID, expectedSpecialSymbol, actualSpecialSymbol, attrs)
		}
	} else {
		if actualHasSpecialSymbol {
			t.Errorf("Verification failed: Node %q status is %q, but unexpected special_symbol %q found. Attrs: %v", nodeID, actualStatus, actualSpecialSymbol, attrs)
		}
	}
}

func logTreeStateForDebugging(t *testing.T, interp *core.Interpreter, handleID string, contextMsg string) {
	t.Helper()
	toolReg := interp.ToolRegistry()
	formatToolImpl, exists := toolReg.GetTool("ChecklistFormatTree")
	if !exists || formatToolImpl.Func == nil {
		t.Logf("DEBUG: Could not find ChecklistFormatTree tool to dump state for handle %s (%s)", handleID, contextMsg)
		return
	}
	formattedTree, err := formatToolImpl.Func(interp, core.MakeArgs(handleID))
	if err != nil {
		t.Logf("DEBUG: Error formatting tree state for handle %s (%s): %v", handleID, contextMsg, err)
	} else {
		formattedStr, _ := formattedTree.(string)
		t.Logf("DEBUG: Tree State for Handle %s (%s):\n---\n%s\n---", handleID, contextMsg, formattedStr)
	}
}

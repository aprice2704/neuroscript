// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 15:33:00 PM PDT // Fix getNodeChildrenIDs key (use 'children'); Correct test expectation
// filename: pkg/neurodata/checklist/checklist_tool_add_test.go
package checklist

import (
	"errors"
	"fmt"
	"testing"

	// Only standard library and approved imports
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
	// *** Make helper import assumption explicit ***
	// Assumes test_helpers.go defines:
	// - newTestInterpreterWithAllTools
	// - assertNoErrorSetup, assertToolFound
	// - pstr, pbool, pint
	// - getNodeViaTool, getNodeValue, getNodeAttributesMap (local getNodeChildrenIDs used instead)
)

// --- REMOVED LOCAL HELPER DEFINITIONS for getNodeViaTool etc. ---
// --- THEY ARE ASSUMED TO BE IN test_helpers.go ---

// --- LOCAL HELPER: getNodeChildrenIDs ---
// NOTE: This helper remains local to this test file.
func getNodeChildrenIDs(t *testing.T, nodeData map[string]interface{}) []string {
	t.Helper()
	if nodeData == nil {
		t.Logf("getNodeChildrenIDs called with nil nodeData")
		return nil
	}

	// <<< FIX: Use correct key "children" based on toolTreeGetNode implementation >>>
	childIDsVal, exists := nodeData["children"]
	if !exists {
		// If key is missing, assume no children (toolTreeGetNode might return nil)
		return []string{}
	}
	if childIDsVal == nil {
		// If key exists but value is nil, also no children
		return []string{}
	}

	// Attempt to assert the type to []interface{} first, common for JSON unmarshaling
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

	// Handle case where it might already be []string (less likely but possible)
	if childIDsStrSlice, ok := childIDsVal.([]string); ok {
		return childIDsStrSlice
	}

	// Handle unexpected type
	t.Errorf("getNodeChildrenIDs: 'children' field is not []interface{} or []string, got %T", childIDsVal)
	return nil
}

// TestChecklistAddItemTool tests the ChecklistAddItem tool implementation.
func TestChecklistAddItemTool(t *testing.T) {
	fixtureChecklist := `:: type: Add Test Fixture

- [ ] Parent Manual Item 1 # node-2
- | | Parent Auto Item 2   # node-3
  - [x] Child 2.1 Done   # node-4
- | | Parent Auto Item 3   # node-5
` // Added node IDs for clarity

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
			parentID:    "node-1", // Assuming root is node-1 based on typical parsing
			newItemText: "New Root Item",
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				if nodeData == nil {
					t.Fatalf("Newly added node %q not found", newNodeID)
				}
				if got, want := getNodeValue(t, nodeData), "New Root Item"; got != want {
					t.Errorf("New node text mismatch: want %q, got %q", want, got)
				}
				attrs := getNodeAttributesMap(t, nodeData)
				wantAttrs := map[string]string{"status": "open"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
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
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				attrs := getNodeAttributesMap(t, nodeData)
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
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				attrs := getNodeAttributesMap(t, nodeData)
				wantAttrs := map[string]string{"status": "special", "special_symbol": "?"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
				parentData := getNodeViaTool(t, interp, handleID, "node-2")
				// <<< Use the fixed local helper >>>
				children := getNodeChildrenIDs(t, parentData)
				if len(children) < 1 || children[0] != newNodeID {
					t.Errorf("New node was not inserted at index 0. Children: %v (Parent: %v)", children, parentData) // Log parent data for context
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
				parentData := getNodeViaTool(t, interp, handleID, "node-5")
				parentAttrs := getNodeAttributesMap(t, parentData)
				wantAttrs := map[string]string{"status": "done", "is_automatic": "true"} // Corrected expectation
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
				parentData := getNodeViaTool(t, interp, handleID, "node-3")
				parentAttrs := getNodeAttributesMap(t, parentData)
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
			index:       pint(10), // Use a higher index
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				rootData := getNodeViaTool(t, interp, handleID, "node-1")
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
			index:       pint(-5),
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				rootData := getNodeViaTool(t, interp, handleID, "node-1")
				children := getNodeChildrenIDs(t, rootData)
				if len(children) == 0 || children[len(children)-1] != newNodeID {
					t.Errorf("Node was not appended correctly for negative index. Children: %v (Parent: %v)", children, rootData)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
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

// TestChecklistUpdateStatusTool tests the Checklist.UpdateStatus tool implementation.
// NOTE: Expectations updated based on bug fixes.
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
` // Added node IDs

	type testStep struct {
		stepName              string
		modifyFunc            func(t *testing.T, interp *core.Interpreter, handleID string) // Func to change item status before update
		verifyFunc            func(t *testing.T, interp *core.Interpreter, handleID string) // Func to check statuses after update
		expectUpdateErr       bool
		expectedUpdateErrorIs error
	}

	// Test Execution
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
					verifyNodeStatus(t, interp, handleID, "node-6", "partial", true, "") // Expect partial based on children open, done
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
					verifyNodeStatus(t, interp, handleID, "node-6", "done", true, "") // Expect done based on children done, done
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
				if step.expectUpdateErr {
					if err == nil {
						t.Errorf("Expected an error during UpdateStatus, but got nil")
					} else if step.expectedUpdateErrorIs != nil && !errors.Is(err, step.expectedUpdateErrorIs) {
						t.Errorf("Expected UpdateStatus error wrapping [%v], got: %v", step.expectedUpdateErrorIs, err)
					} else {
						t.Logf("Got expected UpdateStatus error: %v", err)
					}
					return
				}
				if err != nil {
					// <<< ADDED: Log tree state on unexpected error >>>
					logTreeStateForDebugging(t, interp, handleID, fmt.Sprintf("UpdateStatus failed unexpectedly in step %q", step.stepName))
					t.Fatalf("Checklist.UpdateStatus failed unexpectedly: %v", err)
				}
				if step.verifyFunc != nil {
					step.verifyFunc(t, interp, handleID)
				}
			})
		}

		t.Run("Error: Invalid Handle", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs("invalid-handle-format"))
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Errorf("Expected error wrapping [%v] for invalid handle format, got: %v", core.ErrInvalidArgument, err)
			} else {
				t.Logf("Got expected error for invalid handle format: %v", err)
			}
		})
		t.Run("Error: Handle Not Found", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs(core.GenericTreeHandleType+"::no-such-uuid"))
			if !errors.Is(err, core.ErrNotFound) {
				t.Errorf("Expected error wrapping [%v] for handle not found, got: %v", core.ErrNotFound, err)
			} else {
				t.Logf("Got expected error for handle not found: %v", err)
			}
		})
	})
}

// verifyNodeStatus helper - Assumes getNodeViaTool and getNodeAttributesMap exist in test_helpers.go
func verifyNodeStatus(t *testing.T, interp *core.Interpreter, handleID, nodeID, expectedStatus string, expectAutomatic bool, expectedSpecialSymbol string) {
	t.Helper()
	nodeData := getNodeViaTool(t, interp, handleID, nodeID)
	if nodeData == nil {
		// <<< ADDED: Log tree state on verification failure >>>
		logTreeStateForDebugging(t, interp, handleID, fmt.Sprintf("verifyNodeStatus failed to find node %q", nodeID))
		t.Errorf("Verification failed: Node %q not found using getNodeViaTool", nodeID)
		return
	}
	attrs := getNodeAttributesMap(t, nodeData)
	if attrs == nil {
		t.Logf("Attributes map for Node %q was nil or not a map", nodeID)
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
			if val, ok := attrs["is_automatic"]; !ok || val != "true" { // Check value is "true" if present
				t.Errorf("Verification failed: Node %q is_automatic mismatch. want=%v (attr 'is_automatic=true'), got attr map: %v", nodeID, expectAutomatic, attrs)
			}
		} else {
			t.Errorf("Verification failed: Node %q is_automatic mismatch. want=%v (attribute absent), got attribute present. Attrs: %v", nodeID, expectAutomatic, attrs)
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

// <<< ADDED: Helper to log tree state for debugging failures >>>
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

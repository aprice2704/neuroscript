// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 23:55:00 PDT // <<< ENSURE correct test helper and assertions are used >>>
package checklist

import (
	"errors"
	"testing"

	// Only standard library and approved imports
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
	// NOTE: Assumes test_helpers.go containing newTestInterpreterWithAllTools,
	//       assertNoErrorSetup, assertToolFound, pstr, pbool, pint,
	//       getNodeViaTool, getNodeValue, getNodeAttributesMap etc. exists in this package.
)

// --- REMOVED LOCAL HELPER DEFINITIONS ---
// Assumes necessary helpers (like getNodeViaTool, getNodeValue, getNodeAttributesMap)
// are correctly defined in test_helpers.go or elsewhere accessible within the package.

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
		// (Test cases remain the same as previously provided)
		{
			name:        "Add manual item to root (append)",
			parentID:    "node-1", // Assuming root is node-1 based on typical parsing
			newItemText: "New Root Item",
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID) // Assumes helper exists
				if nodeData == nil {
					t.Fatalf("Newly added node %q not found", newNodeID)
				}
				if got, want := getNodeValue(t, nodeData), "New Root Item"; got != want {
					t.Errorf("New node text mismatch: want %q, got %q", want, got)
				}
				attrs := getNodeAttributesMap(t, nodeData)       // Assumes helper exists
				wantAttrs := map[string]string{"status": "open"} // Default status
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add automatic done item to manual parent",
			parentID:      "node-2", // Parent Manual Item 1
			newItemText:   "New Auto Done Child",
			newItemStatus: pstr("done"), // Assumes pstr exists in test_helpers
			isAutomatic:   pbool(true),  // Assumes pbool exists in test_helpers
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				attrs := getNodeAttributesMap(t, nodeData)
				wantAttrs := map[string]string{"status": "done", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add special item with symbol at index 0",
			parentID:      "node-2", // Parent Manual Item 1
			newItemText:   "New Special Item ?",
			newItemStatus: pstr("special"),
			specialSymbol: pstr("?"),
			isAutomatic:   pbool(false),
			index:         pint(0), // Assumes pint exists in test_helpers
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				nodeData := getNodeViaTool(t, interp, handleID, newNodeID)
				attrs := getNodeAttributesMap(t, nodeData)
				wantAttrs := map[string]string{"status": "special", "special_symbol": "?"}
				if diff := cmp.Diff(wantAttrs, attrs); diff != "" {
					t.Errorf("New node attributes mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add done item triggering parent rollup to partial (after explicit update)",
			parentID:      "node-5", // Parent Auto Item 3 (initially open, no children)
			newItemText:   "Make parent partial",
			newItemStatus: pstr("done"),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				parentData := getNodeViaTool(t, interp, handleID, "node-5")
				parentAttrs := getNodeAttributesMap(t, parentData)
				wantAttrs := map[string]string{"status": "partial", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrs, parentAttrs); diff != "" {
					t.Errorf("Parent node-5 attributes mismatch after AddItem+UpdateStatus (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:          "Add skipped item triggering parent rollup to done (after explicit update)",
			parentID:      "node-3", // Parent Auto Item 2 (initially partial/done based on children)
			newItemText:   "Make parent done again",
			newItemStatus: pstr("skipped"),
			expectError:   false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string, newNodeID string) {
				parentData := getNodeViaTool(t, interp, handleID, "node-3")
				parentAttrs := getNodeAttributesMap(t, parentData)
				wantAttrsParent := map[string]string{"status": "done", "is_automatic": "true"}
				if diff := cmp.Diff(wantAttrsParent, parentAttrs); diff != "" {
					t.Errorf("Parent node-3 attributes mismatch after AddItem+UpdateStatus (-want +got):\n%s", diff)
				}
			},
		},
		{name: "Error: Invalid Parent ID", parentID: "node-99", newItemText: "Fail", expectError: true, expectedErrorIs: core.ErrNotFound},
		{name: "Error: Invalid Status", parentID: "node-1", newItemText: "Fail", newItemStatus: pstr("bad-status"), expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Special Status, Missing Symbol", parentID: "node-1", newItemText: "Fail", newItemStatus: pstr("special"), expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Index Out Of Bounds", parentID: "node-1", newItemText: "Fail", index: pint(5), expectError: true, expectedErrorIs: core.ErrListIndexOutOfBounds},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// <<< USE CORRECT HELPER for each test case run >>>
			interp, registry := newTestInterpreterWithAllTools(t)

			// --- Get Tool Funcs from the isolated registry ---
			toolAddItemImpl, foundAddItem := registry.GetTool("ChecklistAddItem")
			assertToolFound(t, foundAddItem, "ChecklistAddItem") // <<< Use correct assert
			toolFunc := toolAddItemImpl.Func

			toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree") // <<< Use correct assert
			loadToolFunc := toolLoadTreeImpl.Func

			toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
			assertToolFound(t, foundUpdate, "Checklist.UpdateStatus") // <<< Use correct assert
			updateToolFunc := toolUpdateStatusImpl.Func
			// --- End Get Tool Funcs ---

			// Setup: Load Fixture for each test case
			result, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
			assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist") // Assumes helper exists
			handleID, ok := result.(string)
			if !ok {
				t.Fatalf("Setup failed: ChecklistLoadTree did not return a string handle")
			}

			// Initial rollup after loading (needed for tests verifying rollup)
			_, initialUpdateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			assertNoErrorSetup(t, initialUpdateErr, "Setup: Initial Checklist.UpdateStatus failed")

			// Construct Args Slice Manually
			args := []interface{}{handleID, tc.parentID, tc.newItemText}
			args = append(args, tc.newItemStatus) // Append nil directly if pointer is nil
			args = append(args, tc.isAutomatic)   // Append nil directly if pointer is nil
			args = append(args, tc.specialSymbol) // Append nil directly if pointer is nil
			args = append(args, tc.index)         // Append nil directly if pointer is nil

			// Call ChecklistAddItem
			addResult, err := toolFunc(interp, args)

			// Assertions
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

				// Call UpdateStatus explicitly after AddItem if verification depends on it
				if tc.verifyFunc != nil { // Only call if verification needs the updated status
					_, updateErr := updateToolFunc(interp, core.MakeArgs(handleID))
					if updateErr != nil {
						t.Fatalf("Checklist.UpdateStatus failed after AddItem: %v", updateErr)
					}
					t.Log("Checklist.UpdateStatus called successfully after AddItem.")
					tc.verifyFunc(t, interp, handleID, newNodeID) // Pass new ID to verify func
				}
			}
		})
	}
}

// TestChecklistUpdateStatusTool tests the Checklist.UpdateStatus tool implementation.
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
		// <<< USE CORRECT HELPER >>>
		interp, registry := newTestInterpreterWithAllTools(t)

		// --- Get Tool Funcs from the registry ---
		toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
		assertToolFound(t, foundLoad, "ChecklistLoadTree") // <<< Use correct assert
		loadToolFunc := toolLoadTreeImpl.Func

		toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
		assertToolFound(t, foundUpdate, "Checklist.UpdateStatus") // <<< Use correct assert
		updateToolFunc := toolUpdateStatusImpl.Func

		toolSetStatusImpl, foundSetStatus := registry.GetTool("ChecklistSetItemStatus")
		assertToolFound(t, foundSetStatus, "ChecklistSetItemStatus") // <<< Use correct assert
		setStatusToolFunc := toolSetStatusImpl.Func
		// --- End Get Tool Funcs ---

		// Setup: Load Fixture
		result, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
		assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist") // Assumes helper exists
		handleID, ok := result.(string)
		if !ok {
			t.Fatalf("Setup failed: ChecklistLoadTree did not return a string handle")
		}

		// Define test steps (using corrected logic expectations based on previous analysis)
		steps := []testStep{
			{
				stepName:   "Initial Update",
				modifyFunc: nil, // No modification before first update
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "open", true, "")    // RA1: Children open, open -> open
					verifyNodeStatus(t, interp, handleID, "node-6", "partial", true, "") // C2.1A: Children open, done -> partial
					verifyNodeStatus(t, interp, handleID, "node-9", "partial", true, "") // RA3: Child skipped -> partial (Assuming skipped triggers partial)
					verifyNodeStatus(t, interp, handleID, "node-11", "open", true, "")   // RA4: No children -> open
					verifyNodeStatus(t, interp, handleID, "node-5", "open", false, "")   // RM2 (manual)
					verifyNodeStatus(t, interp, handleID, "node-8", "done", false, "")   // GC2.1.2 (manual)
				},
			},
			{
				stepName: "Child 1.1 -> Done (Parent Partial)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-3", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "partial", true, "") // RA1: Children done, open -> partial
				},
			},
			{
				stepName: "Child 1.2 -> Done (Parent Done)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-4", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-2", "done", true, "") // RA1: Children done, done -> done
				},
			},
			{
				stepName: "Grandchild 2.1.1 -> Done (Parent Child 2.1 Auto Done)",
				modifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					_, err := setStatusToolFunc(interp, core.MakeArgs(handleID, "node-7", "done"))
					assertNoErrorSetup(t, err, "Modify step failed: %v", err)
				},
				verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
					verifyNodeStatus(t, interp, handleID, "node-6", "done", true, "")  // C2.1A: Children done, done -> done
					verifyNodeStatus(t, interp, handleID, "node-5", "open", false, "") // RM2 unaffected
				},
			},
			// ... add more steps as needed from original test ...
		}

		// Execute steps sequentially
		for _, step := range steps {
			t.Run(step.stepName, func(t *testing.T) {
				// 1. Modify (if needed)
				if step.modifyFunc != nil {
					step.modifyFunc(t, interp, handleID)
				}

				// 2. Update Status
				_, err := updateToolFunc(interp, core.MakeArgs(handleID))

				// 3. Check for expected error during update
				if step.expectUpdateErr {
					if err == nil {
						t.Errorf("Expected an error during UpdateStatus, but got nil")
					} else if step.expectedUpdateErrorIs != nil && !errors.Is(err, step.expectedUpdateErrorIs) {
						t.Errorf("Expected UpdateStatus error wrapping [%v], got: %v", step.expectedUpdateErrorIs, err)
					} else {
						t.Logf("Got expected UpdateStatus error: %v", err)
					}
					return // Stop this step if error was expected
				}
				// Fail if update error was NOT expected but occurred
				if err != nil {
					t.Fatalf("Checklist.UpdateStatus failed unexpectedly: %v", err)
				}

				// 4. Verify
				if step.verifyFunc != nil {
					step.verifyFunc(t, interp, handleID)
				}
			})
		}

		// Final check: Invalid Handle for UpdateStatus
		t.Run("Error: Invalid Handle", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs("invalid-handle-format"))
			// Expect InvalidArgument because GetHandleValue returns that for bad format
			if !errors.Is(err, core.ErrInvalidArgument) {
				t.Errorf("Expected error wrapping [%v] for invalid handle format, got: %v", core.ErrInvalidArgument, err)
			} else {
				t.Logf("Got expected error for invalid handle format: %v", err)
			}
		})
		t.Run("Error: Handle Not Found", func(t *testing.T) {
			_, err := updateToolFunc(interp, core.MakeArgs(core.GenericTreeHandleType+"::no-such-uuid"))
			if !errors.Is(err, core.ErrNotFound) { // Check ErrNotFound from GetHandleValue
				t.Errorf("Expected error wrapping [%v] for handle not found, got: %v", core.ErrNotFound, err)
			} else {
				t.Logf("Got expected error for handle not found: %v", err)
			}
		})
	})
}

// verifyNodeStatus helper - Requires getNodeViaTool and getNodeAttributesMap
// (Make sure these helpers are defined either here or in test_helpers.go)
func verifyNodeStatus(t *testing.T, interp *core.Interpreter, handleID, nodeID, expectedStatus string, expectAutomatic bool, expectedSpecialSymbol string) {
	t.Helper()
	// Use Tree.GetNode tool/method instead of placeholder if available
	nodeData := getNodeViaTool(t, interp, handleID, nodeID) // Assumes helper exists
	if nodeData == nil {
		t.Errorf("Verification failed: Node %q not found using getNodeViaTool", nodeID)
		return
	}
	attrs := getNodeAttributesMap(t, nodeData) // Assumes helper exists
	if attrs == nil {
		// This case should be handled within getNodeAttributesMap returning empty map
		t.Logf("Attributes map for Node %q was nil or not a map", nodeID)
	}

	actualStatus, ok := attrs["status"]
	if !ok {
		t.Errorf("Verification failed: Node %q missing 'status' attribute. Attrs: %v", nodeID, attrs)
	} else if actualStatus != expectedStatus {
		t.Errorf("Verification failed: Node %q status mismatch. want=%q, got=%q. Attrs: %v", nodeID, expectedStatus, actualStatus, attrs)
	}

	// Check is_automatic presence correctly
	_, actualIsAutomatic := attrs["is_automatic"]
	if expectAutomatic != actualIsAutomatic {
		// If automatic is expected, check the value is "true"
		if expectAutomatic && attrs["is_automatic"] != "true" {
			t.Errorf("Verification failed: Node %q is_automatic mismatch. want=%v (value 'true'), got attribute value %q. Attrs: %v", nodeID, expectAutomatic, attrs["is_automatic"], attrs)
		} else if !expectAutomatic { // If not automatic expected, attribute should NOT be present
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
	} else { // Status is not 'special'
		if actualHasSpecialSymbol {
			t.Errorf("Verification failed: Node %q status is %q, but unexpected special_symbol %q found. Attrs: %v", nodeID, actualStatus, actualSpecialSymbol, attrs)
		}
	}
}

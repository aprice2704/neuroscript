// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 17:10:00 PM PDT // Remove redeclared getNodeViaTool
// filename: pkg/neurodata/checklist/checklist_modify_tool_test.go

package checklist

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// <<< REMOVED local getNodeViaTool function definition >>>
// It is now defined in test_helpers.go

func TestChecklistSetItemTextTool(t *testing.T) {
	fixtureChecklist := `:: type: Rollup Example

- | | L0 Auto Parent
  - [ ] L1 Manual Open
  - | | L1 Auto Child
    - [ ] L2 Manual Open 1
    - [x] L2 Manual Done
  - [?] L1 Manual Question
`
	// Node IDs: root=node-1, L0=node-2, L1M=node-3, L1A=node-4, L2M1=node-5, L2M2=node-6, L1Q=node-7

	testCases := []struct {
		name            string
		targetNodeID    string
		newText         string // Value to set
		expectError     bool
		expectedErrorIs error
		initialStatus   string // EXPECTED STATUS AFTER ChecklistToTree (based on its *own* line, not rollup)
	}{
		{
			name:          "Set text on leaf node",
			targetNodeID:  "node-3", // L1 Manual Open ([ ])
			newText:       "L1 Manual Open - Updated",
			expectError:   false,
			initialStatus: "open", // Status from [ ]
		},
		{
			name:          "Set text on node with children",
			targetNodeID:  "node-4", // L1 Auto Child (| |)
			newText:       "L1 Auto Child - Renamed",
			expectError:   false,
			initialStatus: "open", // Status from | | is open initially
		},
		{
			name:          "Set text to empty string",
			targetNodeID:  "node-5", // L2 Manual Open 1 ([ ])
			newText:       "",
			expectError:   false,
			initialStatus: "open", // Status from [ ]
		},
		{
			name:            "Error: Invalid Handle",
			targetNodeID:    "node-3",
			newText:         "Fail",
			expectError:     true,
			expectedErrorIs: core.ErrInvalidArgument, // Invalid handle format
		},
		{
			name:            "Error: Invalid Node ID",
			targetNodeID:    "node-99",
			newText:         "Fail",
			expectError:     true,
			expectedErrorIs: core.ErrNotFound,
		},
		{
			name:            "Error: Target is not checklist_item (Root)",
			targetNodeID:    "node-1", // Root node
			newText:         "Fail",
			expectError:     true,
			expectedErrorIs: core.ErrInvalidArgument, // Should fail because node type is wrong
		},
	}

	// Get tool implementation directly (assuming it's registered)
	// No need to call newTestInterpreterWithAllTools just to get the func if static
	toolImpl := toolChecklistSetItemTextImpl // Assuming this var is accessible or defined globally/package-level
	toolFunc := toolImpl.Func
	if toolFunc == nil {
		t.Fatal("toolChecklistSetItemTextImpl.Func is nil")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Need interpreter for handle management and potential tool calls inside tested func
			interp, _ := newTestInterpreterWithAllTools(t) // Use helper for consistent setup

			// Setup: Load the fixture using the specific load tool
			loadToolImpl, foundLoad := interp.ToolRegistry().GetTool("ChecklistLoadTree")
			if !foundLoad || loadToolImpl.Func == nil {
				t.Fatalf("Prerequisite tool ChecklistLoadTree not found or invalid")
			}
			result, loadErr := loadToolImpl.Func(interp, core.MakeArgs(fixtureChecklist))
			if loadErr != nil {
				t.Fatalf("Failed to load fixture checklist: %v", loadErr)
			}
			handleID, ok := result.(string)
			if !ok {
				t.Fatalf("ChecklistLoadTree did not return a string handle")
			}

			testHandleID := handleID
			if tc.name == "Error: Invalid Handle" {
				testHandleID = "bad-handle"
			}

			// --- Call the tool function ---
			args := core.MakeArgs(testHandleID, tc.targetNodeID, tc.newText)
			_, err := toolFunc(interp, args)

			// --- Assertions ---
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
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify the text (Value) was updated using the tool-based helper from test_helpers
				nodeData := getNodeViaTool(t, interp, handleID, tc.targetNodeID)
				if nodeData == nil {
					t.Fatalf("Node %s not found after presumably successful update", tc.targetNodeID)
				}
				actualValue := getNodeValue(t, nodeData) // Assumes helper exists

				if !reflect.DeepEqual(tc.newText, actualValue) {
					t.Errorf("Node text mismatch. got = %v (%T), want = %v (%T)", actualValue, actualValue, tc.newText, tc.newText)
				}

				// Verify status wasn't changed (use direct access helper for robustness)
				actualAttrs, attrErr := getNodeAttributesDirectly(t, interp, handleID, tc.targetNodeID)
				if attrErr != nil {
					t.Fatalf("Failed to get attributes directly for node %q: %v", tc.targetNodeID, attrErr)
				}
				currentStatus := actualAttrs["status"]
				if currentStatus != tc.initialStatus {
					t.Errorf("Node status changed unexpectedly or initial expectation wrong. got = %q, want = %q", currentStatus, tc.initialStatus)
				}
			}
		})
	}
}

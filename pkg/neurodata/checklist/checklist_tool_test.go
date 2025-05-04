// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 00:25:06 AM PDT // Correct test expectations based on priority rules
// filename: pkg/neurodata/checklist/checklist_tool_test.go

package checklist

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
)

// --- Helper Functions ---
// (getNodeAttributes remains the same)
func getNodeAttributes(t *testing.T, interp *core.Interpreter, handleID string, nodeID string) map[string]string {
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
	attrsCopy := make(map[string]string)
	for k, v := range node.Attributes {
		attrsCopy[k] = v
	}
	return attrsCopy
}

// --- Tests ---

// TestChecklistLoadTree - Assumes standard registration works via new helper
func TestChecklistLoadTree(t *testing.T) {
	// Use the NEW helper defined in checklist/test_helpers.go
	interp, registry := newTestInterpreterWithAllTools(t) // Setup interpreter once

	// (Fixture content remains the same)
	validChecklistContent := `
:: title: Sample Checklist
:: version: 1.1
# Section 1
- [x] Item 1.1 *(Anno)*
- [ ] Item 1.2
  - [?] Item 1.2.1
# Section 2
- | | Item 2.1 (Auto)
  - [-] Item 2.1.1 (Skipped)
`
	malformedChecklistContent := `
- [x] Good item
- [xx] Bad item
`
	emptyChecklistContent := `
# Only comments and blank lines
`

	testCases := []struct {
		name        string
		content     interface{}
		expectError bool
		wantErrIs   error
		verifyFunc  func(t *testing.T, interp *core.Interpreter, handleID string) // Pass interp
	}{
		// (Test cases remain the same)
		{
			name:        "Valid Checklist",
			content:     validChecklistContent,
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) { /* verification */
				if !strings.HasPrefix(handleID, core.GenericTreeHandleType+"::") {
					t.Errorf("Expected handle prefix %q, got %q", core.GenericTreeHandleType+"::", handleID)
				}
				obj, err := interp.GetHandleValue(handleID, core.GenericTreeHandleType)
				if err != nil {
					t.Fatalf("Failed to get handle value for %q: %v", handleID, err)
				}
				tree, ok := obj.(*core.GenericTree)
				if !ok || tree == nil || tree.RootID == "" || tree.NodeMap == nil {
					t.Errorf("Handle %q did not return a valid *core.GenericTree", handleID)
					return
				}
				if len(tree.NodeMap) == 0 {
					t.Error("Tree NodeMap is empty after loading valid checklist")
				}
				rootAttrs := getNodeAttributes(t, interp, handleID, tree.RootID)
				if rootAttrs == nil {
					t.Error("Could not get root node attributes")
				} else if rootAttrs["title"] != "Sample Checklist" {
					t.Errorf("Root attribute 'title' mismatch: want %q, got %q", "Sample Checklist", rootAttrs["title"])
				}
				t.Logf("Successfully retrieved tree handle %q with root %q and %d nodes", handleID, tree.RootID, len(tree.NodeMap))
			},
		},
		{name: "Empty Content", content: emptyChecklistContent, expectError: true, wantErrIs: core.ErrInvalidArgument},
		{name: "No Content String", content: "", expectError: true, wantErrIs: core.ErrInvalidArgument},
		{name: "Malformed Item", content: malformedChecklistContent, expectError: true, wantErrIs: core.ErrInvalidArgument},
		{name: "Invalid Argument Type", content: 12345, expectError: true, wantErrIs: core.ErrValidationTypeMismatch},
	}

	// Get tool from the registry created by the helper
	toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
	assertToolFound(t, foundLoad, "ChecklistLoadTree") // <<< CORRECTED HELPER
	toolFunc := toolLoadTreeImpl.Func

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Each test run uses the SAME interpreter created ONCE for the parent test.
			result, err := toolFunc(interp, core.MakeArgs(tc.content))
			// (Assertions remain the same)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], got: %v (Type: %T)", tc.wantErrIs, err, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				handleID, ok := result.(string)
				if !ok {
					t.Fatalf("Expected string handleID result, got %T: %v", result, result)
				}
				if tc.verifyFunc != nil {
					tc.verifyFunc(t, interp, handleID)
				}
			}
		})
	}
}

// TestChecklistSetItemStatusTool - MODIFIED to use assertToolFound and pstr
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
		specialSymbol            *string // Use pointer for optional arg
		expectError              bool
		expectedErrorIs          error
		expectedTargetAttrs      map[string]string
		expectedParentAttrs      map[string]string // Status AFTER UpdateStatus call
		expectedGrandparentAttrs map[string]string // Status AFTER UpdateStatus call
		skipTest                 bool
		skipReason               string
	}{
		{name: "Set Manual Open -> Done", targetNodeID: "node-3", newStatus: "done",
			expectedTargetAttrs: map[string]string{"status": "done"},
			// Parent node-2 has children [done (node-3), partial (node-4), question (node-7)] -> question takes priority
			expectedParentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Set Manual Done -> Skipped", targetNodeID: "node-6", newStatus: "skipped",
			expectedTargetAttrs: map[string]string{"status": "skipped"},
			// Parent node-4 children: [open (node-5), skipped (node-6)] -> partial (Rule 2 trigger)
			expectedParentAttrs: map[string]string{"status": "partial", "is_automatic": "true"}, // << CORRECTED EXPECTATION (was open)
			// Grandparent node-2 children: [open (node-3), partial (node-4), question (node-7)] -> question takes priority
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set last L2 Open -> Done => L1A becomes Done", targetNodeID: "node-5", newStatus: "done",
			expectedTargetAttrs: map[string]string{"status": "done"},
			// Parent node-4 children: [done (node-5), done (node-6)] -> done (Rule 3)
			expectedParentAttrs: map[string]string{"status": "done", "is_automatic": "true"}, // << CORRECTED EXPECTATION (was partial)
			// Grandparent node-2 children: [open (node-3), done (node-4), question (node-7)] -> question takes priority
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set L1 Question -> Done => L0 becomes Partial", targetNodeID: "node-7", newStatus: "done",
			expectedTargetAttrs: map[string]string{"status": "done"},
			// Parent node-2 children: [open (node-3), partial (node-4), done (node-7)] -> partial (Rule 2 trigger)
			expectedParentAttrs: map[string]string{"status": "partial", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Rollup: Set L1 Manual Open -> Blocked => L0 becomes Blocked", targetNodeID: "node-3", newStatus: "blocked",
			expectedTargetAttrs: map[string]string{"status": "blocked"},
			// Parent node-2 children: [blocked (node-3), partial (node-4), question (node-7)] -> blocked takes priority
			expectedParentAttrs: map[string]string{"status": "blocked", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Rollup: Set L2 Manual Done -> Special * => L1A Special*, L0 Question", targetNodeID: "node-6", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs: map[string]string{"status": "special", "special_symbol": "*"},
			// Parent node-4 children: [open (node-5), special * (node-6)] -> special * takes priority
			expectedParentAttrs: map[string]string{"status": "special", "is_automatic": "true", "special_symbol": "*"},
			// Grandparent node-2 children: [open (node-3), special * (node-4), question (node-7)] -> question takes priority
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"}, // << CORRECTED EXPECTATION (was blocked)
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set L1 Question -> Special * => L0 becomes Special *", targetNodeID: "node-7", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs: map[string]string{"status": "special", "special_symbol": "*"},
			// Parent node-2 children: [open (node-3), partial (node-4), special * (node-7)] -> special * takes priority
			expectedParentAttrs: map[string]string{"status": "special", "is_automatic": "true", "special_symbol": "*"}, // << CORRECTED EXPECTATION (was blocked)
			skipTest:            false, skipReason: ""},
		{name: "Error: Invalid Node ID", targetNodeID: "node-99", newStatus: "done", expectError: true, expectedErrorIs: core.ErrNotFound, skipTest: false, skipReason: ""},
		{name: "Error: Invalid Status", targetNodeID: "node-3", newStatus: "invalid-status", expectError: true, expectedErrorIs: core.ErrInvalidArgument, skipTest: false, skipReason: ""},
		{
			name:            "Error: Special Status, Missing Symbol",
			targetNodeID:    "node-3",
			newStatus:       "special",
			expectError:     true,
			expectedErrorIs: core.ErrValidationRequiredArgNil, // <<< FIXED EXPECTED ERROR
			skipTest:        false,
			skipReason:      "",
		},
		{name: "Error: Special Status, Invalid Symbol", targetNodeID: "node-3", newStatus: "special", specialSymbol: pstr("xx"), expectError: true, expectedErrorIs: core.ErrInvalidArgument, skipTest: false, skipReason: ""}, // <<< Use local pstr
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipTest {
				t.Skip(tc.skipReason)
			}
			// Create a FRESH interpreter WITH ALL TOOLS for EACH test case for isolation
			interp, registry := newTestInterpreterWithAllTools(t) // Use the local helper

			// --- Get Tool Funcs from the isolated registry ---
			toolSetStatusImpl, foundSet := registry.GetTool("ChecklistSetItemStatus")
			assertToolFound(t, foundSet, "ChecklistSetItemStatus") // <<< CORRECTED HELPER
			setStatusToolFunc := toolSetStatusImpl.Func

			toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree") // <<< CORRECTED HELPER
			loadToolFunc := toolLoadTreeImpl.Func

			toolUpdateStatusImpl, foundUpdate := registry.GetTool("Checklist.UpdateStatus")
			assertToolFound(t, foundUpdate, "Checklist.UpdateStatus") // <<< CORRECTED HELPER
			updateToolFunc := toolUpdateStatusImpl.Func
			// --- End Get Tool Funcs ---

			// Setup: Load fixture and perform initial update
			result, loadErr := loadToolFunc(interp, core.MakeArgs(fixtureChecklist))
			// Use assertNoErrorSetup for fatal setup errors
			assertNoErrorSetup(t, loadErr, "Setup: Failed to load fixture checklist")

			handleID, ok := result.(string)
			if !ok {
				t.Fatalf("Setup: ChecklistLoadTree did not return a string handle, got %T", result)
			}
			_, initialUpdateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			assertNoErrorSetup(t, initialUpdateErr, "Setup: Initial Checklist.UpdateStatus failed")
			// End Setup

			// Construct args for SetItemStatus
			args := []interface{}{handleID, tc.targetNodeID, tc.newStatus}
			if tc.specialSymbol != nil {
				args = append(args, *tc.specialSymbol)
			}

			// Call SetItemStatus
			_, err := setStatusToolFunc(interp, args)

			// --- Assertions for SetItemStatus call ---
			// (Assertions remain the same)
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

			// --- Explicitly call UpdateStatus ---
			_, updateErr := updateToolFunc(interp, core.MakeArgs(handleID))
			if updateErr != nil {
				t.Fatalf("Checklist.UpdateStatus failed unexpectedly after SetItemStatus: %v", updateErr)
			}

			// --- Verification after UpdateStatus ---
			// (Verification logic remains the same)
			nodeObj, treeErr := interp.GetHandleValue(handleID, core.GenericTreeHandleType)
			if treeErr != nil {
				t.Fatalf("Failed to get tree from handle %q after successful update: %v", handleID, treeErr)
			}
			tree, ok := nodeObj.(*core.GenericTree)
			if !ok || tree == nil {
				t.Fatalf("Handle %q did not contain a valid tree after successful update", handleID)
			}
			if tc.expectedTargetAttrs != nil {
				actualAttrs := getNodeAttributes(t, interp, handleID, tc.targetNodeID)
				if actualAttrs == nil {
					t.Errorf("Target node %s not found or attributes could not be retrieved after successful call", tc.targetNodeID)
				} else if diff := cmp.Diff(tc.expectedTargetAttrs, actualAttrs); diff != "" {
					t.Errorf("Target node (%s) attributes mismatch (-want +got):\n%s", tc.targetNodeID, diff)
				}
			}
			if tc.expectedParentAttrs != nil {
				var parentID string
				if targetNode, exists := tree.NodeMap[tc.targetNodeID]; exists {
					parentID = targetNode.ParentID
				}
				if parentID != "" && parentID != tree.RootID {
					actualAttrs := getNodeAttributes(t, interp, handleID, parentID)
					if actualAttrs == nil {
						t.Errorf("Parent node %s not found after successful call", parentID)
					} else if diff := cmp.Diff(tc.expectedParentAttrs, actualAttrs); diff != "" {
						t.Errorf("Parent node (%s) attributes mismatch (-want +got):\n%s", parentID, diff)
					}
				} else if len(tc.expectedParentAttrs) > 0 && parentID != tree.RootID {
					t.Errorf("Test case expected parent attributes, but could not determine valid non-root parent ID for target %s", tc.targetNodeID)
				} else if len(tc.expectedParentAttrs) > 0 && parentID == tree.RootID {
					t.Logf("Note: Expected parent attributes for target %s which is a direct child of the root.", tc.targetNodeID)
				}
			}
			if tc.expectedGrandparentAttrs != nil {
				var parentID, grandparentID string
				targetNode, targetExists := tree.NodeMap[tc.targetNodeID]
				if targetExists {
					parentID = targetNode.ParentID
					parentNode, parentExists := tree.NodeMap[parentID]
					if parentExists {
						grandparentID = parentNode.ParentID
					}
				}
				if grandparentID != "" && grandparentID != tree.RootID {
					actualAttrs := getNodeAttributes(t, interp, handleID, grandparentID)
					if actualAttrs == nil {
						t.Errorf("Grandparent node %s not found after successful call", grandparentID)
					} else if diff := cmp.Diff(tc.expectedGrandparentAttrs, actualAttrs); diff != "" {
						t.Errorf("Grandparent node (%s) attributes mismatch (-want +got):\n%s", grandparentID, diff)
					}
				} else if len(tc.expectedGrandparentAttrs) > 0 && grandparentID != tree.RootID {
					t.Errorf("Test case expected grandparent attributes, but could not determine valid non-root grandparent ID for target %s", tc.targetNodeID)
				} else if len(tc.expectedGrandparentAttrs) > 0 && grandparentID == tree.RootID {
					t.Logf("Note: Expected grandparent attributes for target %s whose parent is a direct child of the root.", tc.targetNodeID)
				}
			}

		})
	}
}

// TestChecklistFormatTreeTool - Modified to use assertToolFound
func TestChecklistFormatTreeTool(t *testing.T) {

	testCases := []struct {
		name            string
		inputChecklist  string // Used only if inputHandle is empty
		inputHandle     string // Use this handle if provided
		expectedOutput  string
		expectError     bool
		expectedErrorIs error
	}{
		// (Test cases remain the same)
		{name: "Basic Round Trip with Metadata Sort", inputChecklist: ":: version: 1.0\n:: title: Simple List\n\n- [ ] Item 1\n- [x] Item 2\n", expectedOutput: ":: title: Simple List\n:: version: 1.0\n\n- [ ] Item 1\n- [x] Item 2\n", expectError: false},
		{name: "Nested and Auto Round Trip", inputChecklist: ":: type: Nested Example\n:: author: Test\n\n- [?] Manual Question\n  - | | Auto Open Child\n    - |x| Auto Done Grandchild\n  - [-] Manual Skipped Child\n- [>] Manual In Progress\n- |-| Auto Partial\n- [*] Manual Special Star\n- |!| Auto Blocked\n", expectedOutput: ":: author: Test\n:: type: Nested Example\n\n- [?] Manual Question\n  - | | Auto Open Child\n    - |x| Auto Done Grandchild\n  - [-] Manual Skipped Child\n- [>] Manual In Progress\n- |-| Auto Partial\n- [*] Manual Special Star\n- |!| Auto Blocked\n", expectError: false},
		{name: "Metadata Only", inputChecklist: ":: Zkey: valZ\n:: Akey: valA\n", expectedOutput: ":: Akey: valA\n:: Zkey: valZ\n", expectError: false},
		{name: "Items Only", inputChecklist: "- [ ] Item A\n  - [x] Item B\n", expectedOutput: "- [ ] Item A\n  - [x] Item B\n", expectError: false},
		{name: "Empty Checklist (Load Fails)", inputChecklist: ``, expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "No Content Checklist (Load Fails)", inputChecklist: "# Just a comment", expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Invalid Handle Format", inputHandle: "bad-handle", expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Handle Wrong Type", inputHandle: "WrongType::1234-abcd", expectError: true, expectedErrorIs: core.ErrHandleWrongType},
		{name: "Error: Handle Not Found", inputHandle: core.GenericTreeHandleType + "::no-such-uuid", expectError: true, expectedErrorIs: core.ErrNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a FRESH interpreter WITH ALL TOOLS for EACH test case for isolation
			interp, registry := newTestInterpreterWithAllTools(t) // Use local helper

			// Get Tool Funcs from the isolated registry
			toolFormatImpl, foundFormat := registry.GetTool("ChecklistFormatTree")
			assertToolFound(t, foundFormat, "ChecklistFormatTree") // <<< CORRECTED HELPER
			formatToolFunc := toolFormatImpl.Func

			toolLoadImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree") // <<< CORRECTED HELPER
			loadToolFunc := toolLoadImpl.Func
			// --- End Tool Func Get ---

			var handleID string

			// Setup: Load checklist ONLY if handle isn't provided directly
			if tc.inputHandle == "" {
				result, loadErr := loadToolFunc(interp, core.MakeArgs(tc.inputChecklist))
				// (Loading logic remains the same)
				if tc.expectError && loadErr != nil {
					if tc.expectedErrorIs != nil && errors.Is(loadErr, tc.expectedErrorIs) {
						t.Logf("Got expected error during checklist loading: %v", loadErr)
						return
					} else if tc.expectedErrorIs == nil {
						t.Logf("Got expected error during checklist loading (no specific type check): %v", loadErr)
						return
					} else {
						t.Fatalf("Checklist loading failed, but with unexpected error type. Got %v, want %v", loadErr, tc.expectedErrorIs)
					}
				}
				if loadErr != nil {
					t.Fatalf("Failed to load input checklist for test setup: %v", loadErr)
				}
				if tc.expectError {
					t.Logf("Loading succeeded for %q, expecting error during Format call.", tc.name)
					var ok bool
					handleID, ok = result.(string)
					if !ok {
						t.Fatalf("ChecklistLoadTree did not return a string handle during setup")
					}
				} else {
					var ok bool
					handleID, ok = result.(string)
					if !ok {
						t.Fatalf("ChecklistLoadTree did not return a string handle during setup")
					}
				}
			} else {
				handleID = tc.inputHandle
				if !tc.expectError {
					t.Logf("Using provided handle %q for test %q, expecting success.", handleID, tc.name)
				} else {
					t.Logf("Using provided handle %q for test %q, expecting error during Format call.", handleID, tc.name)
				}
			}

			// --- Call the tool function ---
			output, err := formatToolFunc(interp, core.MakeArgs(handleID))

			// --- Assertions ---
			// (Assertions remain the same)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error from Format call for handle %q, but got nil", handleID)
				} else if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
					t.Errorf("Expected Format error wrapping [%v] for handle %q, got: %v (Type: %T)", tc.expectedErrorIs, handleID, err, err)
				} else {
					t.Logf("Got expected Format error for handle %q: %v", handleID, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error from Format call for handle %q: %v", handleID, err)
					return
				}
				outputStr, ok := output.(string)
				if !ok {
					t.Fatalf("Expected string output from Format, got %T: %v", output, output)
				}
				normalizedWant := strings.TrimSpace(tc.expectedOutput)
				if normalizedWant != "" {
					normalizedWant += "\n"
				}
				normalizedGot := strings.TrimSpace(outputStr)
				if normalizedGot != "" {
					normalizedGot += "\n"
				}
				if diff := cmp.Diff(normalizedWant, normalizedGot); diff != "" {
					t.Errorf("ChecklistFormatTree() output mismatch for handle %q (-want +got):\n%s", handleID, diff)
				}
			}
		})
	}
}

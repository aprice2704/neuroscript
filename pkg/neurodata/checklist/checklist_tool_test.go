// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Updated error expectation in TestChecklistFormatTreeTool.
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

func TestChecklistLoadTree(t *testing.T) {
	interp, registry := newTestInterpreterWithAllTools(t)

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
		verifyFunc  func(t *testing.T, interp *core.Interpreter, handleID string)
	}{
		{
			name:        "Valid Checklist",
			content:     validChecklistContent,
			expectError: false,
			verifyFunc: func(t *testing.T, interp *core.Interpreter, handleID string) {
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

	toolLoadTreeImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
	assertToolFound(t, foundLoad, "ChecklistLoadTree")
	toolFunc := toolLoadTreeImpl.Func

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toolFunc(interp, core.MakeArgs(tc.content))
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
		expectedTargetAttrs      map[string]string
		expectedParentAttrs      map[string]string
		expectedGrandparentAttrs map[string]string
		skipTest                 bool
		skipReason               string
	}{
		{name: "Set Manual Open -> Done", targetNodeID: "node-3", newStatus: "done",
			expectedTargetAttrs: map[string]string{"status": "done"},
			expectedParentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Set Manual Done -> Skipped", targetNodeID: "node-6", newStatus: "skipped",
			expectedTargetAttrs:      map[string]string{"status": "skipped"},
			expectedParentAttrs:      map[string]string{"status": "partial", "is_automatic": "true"},
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set last L2 Open -> Done => L1A becomes Done", targetNodeID: "node-5", newStatus: "done",
			expectedTargetAttrs:      map[string]string{"status": "done"},
			expectedParentAttrs:      map[string]string{"status": "done", "is_automatic": "true"},
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set L1 Question -> Done => L0 becomes Partial", targetNodeID: "node-7", newStatus: "done",
			expectedTargetAttrs: map[string]string{"status": "done"},
			expectedParentAttrs: map[string]string{"status": "partial", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Rollup: Set L1 Manual Open -> Blocked => L0 becomes Blocked", targetNodeID: "node-3", newStatus: "blocked",
			expectedTargetAttrs: map[string]string{"status": "blocked"},
			expectedParentAttrs: map[string]string{"status": "blocked", "is_automatic": "true"},
			skipTest:            false, skipReason: ""},
		{name: "Rollup: Set L2 Manual Done -> Special * => L1A Special*, L0 Question", targetNodeID: "node-6", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs:      map[string]string{"status": "special", "special_symbol": "*"},
			expectedParentAttrs:      map[string]string{"status": "special", "is_automatic": "true", "special_symbol": "*"},
			expectedGrandparentAttrs: map[string]string{"status": "question", "is_automatic": "true"},
			skipTest:                 false, skipReason: ""},
		{name: "Rollup: Set L1 Question -> Special * => L0 becomes Special *", targetNodeID: "node-7", newStatus: "special", specialSymbol: pstr("*"),
			expectedTargetAttrs: map[string]string{"status": "special", "special_symbol": "*"},
			expectedParentAttrs: map[string]string{"status": "special", "is_automatic": "true", "special_symbol": "*"},
			skipTest:            false, skipReason: ""},
		{name: "Error: Invalid Node ID", targetNodeID: "node-99", newStatus: "done", expectError: true, expectedErrorIs: core.ErrNotFound, skipTest: false, skipReason: ""},
		{name: "Error: Invalid Status", targetNodeID: "node-3", newStatus: "invalid-status", expectError: true, expectedErrorIs: core.ErrInvalidArgument, skipTest: false, skipReason: ""},
		{
			name:            "Error: Special Status, Missing Symbol",
			targetNodeID:    "node-3",
			newStatus:       "special",
			expectError:     true,
			expectedErrorIs: core.ErrValidationRequiredArgNil,
			skipTest:        false,
			skipReason:      "",
		},
		{name: "Error: Special Status, Invalid Symbol", targetNodeID: "node-3", newStatus: "special", specialSymbol: pstr("xx"), expectError: true, expectedErrorIs: core.ErrInvalidArgument, skipTest: false, skipReason: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

func TestChecklistFormatTreeTool(t *testing.T) {
	testCases := []struct {
		name            string
		inputChecklist  string
		inputHandle     string
		expectedOutput  string
		expectError     bool
		expectedErrorIs error
	}{
		{name: "Basic Round Trip with Metadata Sort", inputChecklist: ":: version: 1.0\n:: title: Simple List\n\n- [ ] Item 1\n- [x] Item 2\n", expectedOutput: ":: title: Simple List\n:: version: 1.0\n\n- [ ] Item 1\n- [x] Item 2\n", expectError: false},
		{name: "Nested and Auto Round Trip", inputChecklist: ":: type: Nested Example\n:: author: Test\n\n- [?] Manual Question\n  - | | Auto Open Child\n    - |x| Auto Done Grandchild\n  - [-] Manual Skipped Child\n- [>] Manual In Progress\n- |-| Auto Partial\n- [*] Manual Special Star\n- |!| Auto Blocked\n", expectedOutput: ":: author: Test\n:: type: Nested Example\n\n- [?] Manual Question\n  - | | Auto Open Child\n    - |x| Auto Done Grandchild\n  - [-] Manual Skipped Child\n- [>] Manual In Progress\n- |-| Auto Partial\n- [*] Manual Special Star\n- |!| Auto Blocked\n", expectError: false},
		{name: "Metadata Only", inputChecklist: ":: Zkey: valZ\n:: Akey: valA\n", expectedOutput: ":: Akey: valA\n:: Zkey: valZ\n", expectError: false},
		{name: "Items Only", inputChecklist: "- [ ] Item A\n  - [x] Item B\n", expectedOutput: "- [ ] Item A\n  - [x] Item B\n", expectError: false},
		{name: "Empty Checklist (Load Fails)", inputChecklist: ``, expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "No Content Checklist (Load Fails)", inputChecklist: "# Just a comment", expectError: true, expectedErrorIs: core.ErrInvalidArgument},
		{name: "Error: Invalid Handle Format", inputHandle: "bad-handle", expectError: true, expectedErrorIs: core.ErrInvalidArgument}, // Or ErrHandleInvalid if more specific
		{name: "Error: Handle Wrong Type", inputHandle: "WrongType::1234-abcd", expectError: true, expectedErrorIs: core.ErrHandleWrongType},
		{name: "Error: Handle Not Found", inputHandle: core.GenericTreeHandleType + "::no-such-uuid", expectError: true, expectedErrorIs: core.ErrHandleNotFound}, // MODIFIED HERE
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, registry := newTestInterpreterWithAllTools(t)

			toolFormatImpl, foundFormat := registry.GetTool("ChecklistFormatTree")
			assertToolFound(t, foundFormat, "ChecklistFormatTree")
			formatToolFunc := toolFormatImpl.Func

			toolLoadImpl, foundLoad := registry.GetTool("ChecklistLoadTree")
			assertToolFound(t, foundLoad, "ChecklistLoadTree")
			loadToolFunc := toolLoadImpl.Func

			var handleID string

			if tc.inputHandle == "" {
				result, loadErr := loadToolFunc(interp, core.MakeArgs(tc.inputChecklist))
				if tc.expectError && loadErr != nil {
					if tc.expectedErrorIs != nil && errors.Is(loadErr, tc.expectedErrorIs) {
						t.Logf("Got expected error during checklist loading: %v", loadErr)
						return
					} else if tc.expectedErrorIs == nil {
						t.Logf("Got expected error during checklist loading (no specific type check): %v", loadErr)
						return
					} else {
						// If loadErr is not nil, and we expected an error for the Format call, this might be a setup issue
						// or the load itself was the one failing as expected.
						t.Logf("Checklist loading failed for test %q (error: %v), this might be the expected error if test targets load phase.", tc.name, loadErr)
						if tc.expectedErrorIs != nil && errors.Is(loadErr, tc.expectedErrorIs) {
							return // Error was expected during load
						}
						// If we didn't expect this specific load error, but expected a format error later,
						// it's a test setup problem or an unexpected load error.
						if !errors.Is(loadErr, tc.expectedErrorIs) && tc.expectedErrorIs != core.ErrInvalidArgument && tc.expectedErrorIs != core.ErrHandleWrongType && tc.expectedErrorIs != core.ErrHandleNotFound {
							t.Fatalf("Failed to load input checklist for test setup %q: %v", tc.name, loadErr)
						}
						// If load failed as expected, and it's one of the direct error types, then the test is done.
						if errors.Is(loadErr, core.ErrInvalidArgument) || errors.Is(loadErr, core.ErrHandleWrongType) || errors.Is(loadErr, core.ErrHandleNotFound) {
							return
						}
					}
				}
				// If loadErr is not nil here, but we didn't expect an error (tc.expectError is false)
				if loadErr != nil && !tc.expectError {
					t.Fatalf("Failed to load input checklist for test setup %q: %v", tc.name, loadErr)
				}

				if result != nil {
					var ok bool
					handleID, ok = result.(string)
					if !ok {
						t.Fatalf("ChecklistLoadTree did not return a string handle during setup for test %q", tc.name)
					}
				} else if !tc.expectError { // result is nil but no error was expected from load
					t.Fatalf("ChecklistLoadTree returned nil handle without error during setup for test %q", tc.name)
				}
				// If tc.expectError is true but loadErr was nil, we proceed to format call to get the expected error
				if tc.expectError && loadErr == nil {
					t.Logf("Loading succeeded for %q, expecting error during Format call.", tc.name)
				}

			} else {
				handleID = tc.inputHandle
				if !tc.expectError {
					t.Logf("Using provided handle %q for test %q, expecting success.", handleID, tc.name)
				} else {
					t.Logf("Using provided handle %q for test %q, expecting error during Format call.", handleID, tc.name)
				}
			}

			// Ensure handleID is valid before calling Format, unless an error is expected from Format due to bad handle
			if handleID == "" && tc.expectError && (errors.Is(tc.expectedErrorIs, core.ErrInvalidArgument) || errors.Is(tc.expectedErrorIs, core.ErrHandleWrongType) || errors.Is(tc.expectedErrorIs, core.ErrHandleNotFound)) {
				// This case means loading failed as expected, and the test is about that load failure.
				// Or, inputHandle was not set and load was supposed to provide it.
				// If inputHandle was "", and load failed as expected, we should have returned.
				// If inputHandle itself is invalid, tc.inputHandle != "" and handleID = tc.inputHandle.
				// This path means: handleID is empty because loading it failed, but the test expected error from Format tool.
				// This indicates a test logic issue if an error was expected at format time but load already failed.
				// For now, let the Format call proceed to see its error.
				if tc.inputHandle == "" { // if handle was from load and load failed as expected
					t.Logf("Skipping Format call as loading already failed as expected for test %q", tc.name)
					return
				}
			}

			output, err := formatToolFunc(interp, core.MakeArgs(handleID))

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

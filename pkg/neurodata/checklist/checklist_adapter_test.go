// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 19:50:50 PM PDT // Fix map assignment syntax in test
// filename: pkg/neurodata/checklist/checklist_adapter_test.go

package checklist

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters" // For NoOpLogger
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp" // For better diffs
)

// Helper to verify node properties in the tree (Existing - unchanged)
func verifyTreeNode(t *testing.T, tree *core.GenericTree, nodeID string, expectedProps map[string]interface{}) {
	t.Helper()

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		t.Errorf("verifyTreeNode: Node %q not found in tree map", nodeID)
		return
	}

	for key, expectedValue := range expectedProps {
		switch key {
		case "Type":
			if node.Type != expectedValue.(string) {
				t.Errorf("verifyTreeNode: Node %q Type mismatch. got=%q, want=%q", nodeID, node.Type, expectedValue)
			}
		case "Value":
			if _, ok := node.Value.(float64); ok {
				if iVal, ok := expectedValue.(int); ok {
					expectedValue = float64(iVal)
				} else if i64Val, ok := expectedValue.(int64); ok {
					expectedValue = float64(i64Val)
				}
			}
			if !reflect.DeepEqual(node.Value, expectedValue) {
				t.Errorf("verifyTreeNode: Node %q Value mismatch. got=%#v (%T), want=%#v (%T)", nodeID, node.Value, node.Value, expectedValue, expectedValue)
			}
		case "ParentID":
			if node.ParentID != expectedValue.(string) {
				t.Errorf("verifyTreeNode: Node %q ParentID mismatch. got=%q, want=%q", nodeID, node.ParentID, expectedValue)
			}
		case "Attributes":
			expectedAttrs := expectedValue.(map[string]string)
			if node.Attributes == nil && len(expectedAttrs) == 0 {
				// Okay
			} else if !reflect.DeepEqual(node.Attributes, expectedAttrs) {
				t.Errorf("verifyTreeNode: Node %q Attributes mismatch. Diff:\n%s", nodeID, cmp.Diff(expectedAttrs, node.Attributes))
			}
		case "ChildIDs":
			expectedChildren := expectedValue.([]string)
			actualChildren := node.ChildIDs
			if actualChildren == nil {
				actualChildren = []string{}
			}
			if len(expectedChildren) == 0 && len(actualChildren) == 0 {
				// Okay
			} else if !reflect.DeepEqual(actualChildren, expectedChildren) {
				t.Errorf("verifyTreeNode: Node %q ChildIDs mismatch. Diff:\n%s", nodeID, cmp.Diff(expectedChildren, node.ChildIDs))
			}
		default:
			t.Errorf("verifyTreeNode: Unknown key %q in expectedProps for node %q", key, nodeID)
		}
	}
}

// TestChecklistToTree (Existing - tests the conversion to Tree)
func TestChecklistToTree(t *testing.T) {
	// ... (test cases remain the same as generated previously) ...

	t.Run("BasicListWithMetadata", func(t *testing.T) {
		items := []ChecklistItem{
			{Text: "Item 1", Status: "pending", Symbol: ' ', Indent: 0, LineNumber: 3, IsAutomatic: false},
			{Text: "Item 2", Status: "done", Symbol: 'x', Indent: 0, LineNumber: 4, IsAutomatic: false},
		}
		metadata := map[string]string{"title": "Test List", "version": "1.0"}

		tree, err := ChecklistToTree(items, metadata)
		if err != nil {
			t.Fatalf("ChecklistToTree failed: %v", err)
		}
		if tree == nil || tree.RootID == "" || tree.NodeMap == nil {
			t.Fatalf("Tree is nil or not initialized properly")
		}
		if len(tree.NodeMap) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(tree.NodeMap))
		}

		rootID := tree.RootID
		item1ID := tree.NodeMap[rootID].ChildIDs[0] // Assume order based on input items
		item2ID := tree.NodeMap[rootID].ChildIDs[1]

		verifyTreeNode(t, tree, rootID, map[string]interface{}{
			"Type": "checklist_root", "ParentID": "", "Value": nil,
			"Attributes": metadata, "ChildIDs": []string{item1ID, item2ID},
		})
		verifyTreeNode(t, tree, item1ID, map[string]interface{}{
			"Type": "checklist_item", "ParentID": rootID, "Value": "Item 1",
			"Attributes": map[string]string{"status": "open"}, // No is_automatic
			"ChildIDs":   []string{},
		})
		verifyTreeNode(t, tree, item2ID, map[string]interface{}{
			"Type": "checklist_item", "ParentID": rootID, "Value": "Item 2",
			"Attributes": map[string]string{"status": "done"}, // No is_automatic
			"ChildIDs":   []string{},
		})
	})

	t.Run("NestedListAndStatuses", func(t *testing.T) {
		items := []ChecklistItem{
			{Text: "Parent 1 *(Anno1)*", Status: "pending", Symbol: ' ', Indent: 0, LineNumber: 1, IsAutomatic: false},
			{Text: "Child 1.1", Status: "partial", Symbol: '-', Indent: 2, LineNumber: 2, IsAutomatic: false},
			{Text: "Child 1.2 **(Anno2)**", Status: "special", Symbol: '!', Indent: 2, LineNumber: 3, IsAutomatic: false},
			{Text: "Child 1.3", Status: "special", Symbol: '*', Indent: 2, LineNumber: 4, IsAutomatic: false},
			{Text: "Parent 2", Status: "pending", Symbol: ' ', Indent: 0, LineNumber: 5, IsAutomatic: true},
			{Text: "Child 2.1", Status: "partial", Symbol: '-', Indent: 2, LineNumber: 6, IsAutomatic: true},
		}
		metadata := map[string]string{"type": "Nested"}

		tree, err := ChecklistToTree(items, metadata)
		if err != nil {
			t.Fatalf("ChecklistToTree failed: %v", err)
		}
		if tree == nil {
			t.Fatal("Tree is nil")
		}
		if len(tree.NodeMap) != 7 {
			t.Fatalf("Expected 7 nodes, got %d", len(tree.NodeMap))
		}

		// Assuming sequential IDs for verification simplicity
		rootID, p1ID, c11ID, c12ID, c13ID, p2ID, c21ID := "node-1", "node-2", "node-3", "node-4", "node-5", "node-6", "node-7"

		verifyTreeNode(t, tree, rootID, map[string]interface{}{"Type": "checklist_root", "ParentID": "", "Value": nil, "Attributes": metadata, "ChildIDs": []string{p1ID, p2ID}})
		verifyTreeNode(t, tree, p1ID, map[string]interface{}{"Type": "checklist_item", "ParentID": rootID, "Value": "Parent 1 *(Anno1)*", "Attributes": map[string]string{"status": "open"}, "ChildIDs": []string{c11ID, c12ID, c13ID}})
		verifyTreeNode(t, tree, c11ID, map[string]interface{}{"Type": "checklist_item", "ParentID": p1ID, "Value": "Child 1.1", "Attributes": map[string]string{"status": "skipped"}, "ChildIDs": []string{}})
		verifyTreeNode(t, tree, c12ID, map[string]interface{}{"Type": "checklist_item", "ParentID": p1ID, "Value": "Child 1.2 **(Anno2)**", "Attributes": map[string]string{"status": "blocked"}, "ChildIDs": []string{}})
		verifyTreeNode(t, tree, c13ID, map[string]interface{}{"Type": "checklist_item", "ParentID": p1ID, "Value": "Child 1.3", "Attributes": map[string]string{"status": "special", "special_symbol": "*"}, "ChildIDs": []string{}})
		verifyTreeNode(t, tree, p2ID, map[string]interface{}{"Type": "checklist_item", "ParentID": rootID, "Value": "Parent 2", "Attributes": map[string]string{"status": "open", "is_automatic": "true"}, "ChildIDs": []string{c21ID}})
		verifyTreeNode(t, tree, c21ID, map[string]interface{}{"Type": "checklist_item", "ParentID": p2ID, "Value": "Child 2.1", "Attributes": map[string]string{"status": "partial", "is_automatic": "true"}, "ChildIDs": []string{}})
	})

	t.Run("EmptyItemsList", func(t *testing.T) {
		items := []ChecklistItem{}
		metadata := map[string]string{"status": "empty"}
		tree, err := ChecklistToTree(items, metadata)
		if err != nil {
			t.Fatalf("ChecklistToTree failed: %v", err)
		}
		if tree == nil || len(tree.NodeMap) != 1 {
			t.Fatalf("Expected 1 node (root), got %d", len(tree.NodeMap))
		}
		rootID := tree.RootID
		verifyTreeNode(t, tree, rootID, map[string]interface{}{"Type": "checklist_root", "ParentID": "", "Value": nil, "Attributes": metadata, "ChildIDs": []string{}})
	})

	t.Run("NoMetadata", func(t *testing.T) {
		items := []ChecklistItem{{Text: "Only Item", Status: "special", Symbol: '>', Indent: 0, LineNumber: 1, IsAutomatic: false}}
		metadata := map[string]string{}
		tree, err := ChecklistToTree(items, metadata)
		if err != nil {
			t.Fatalf("ChecklistToTree failed: %v", err)
		}
		if tree == nil || len(tree.NodeMap) != 2 {
			t.Fatalf("Expected 2 nodes, got %d", len(tree.NodeMap))
		}
		rootID := tree.RootID
		itemID := tree.NodeMap[rootID].ChildIDs[0]
		verifyTreeNode(t, tree, rootID, map[string]interface{}{"Type": "checklist_root", "ParentID": "", "Value": nil, "Attributes": map[string]string{}, "ChildIDs": []string{itemID}})
		verifyTreeNode(t, tree, itemID, map[string]interface{}{"Type": "checklist_item", "ParentID": rootID, "Value": "Only Item", "Attributes": map[string]string{"status": "inprogress"}, "ChildIDs": []string{}})
	})
}

// --- NEW: Tests for TreeToChecklistString ---

func TestTreeToChecklistString(t *testing.T) {
	noopLogger := adapters.NewNoOpLogger()

	testCases := []struct {
		name              string
		inputChecklist    string                   // Input checklist string for round-trip tests
		buildTreeManually func() *core.GenericTree // For error cases
		expectedOutput    string                   // Expected output string (after formatting)
		expectError       bool
		expectedErrorIs   error // Specific error type to check (using errors.Is)
	}{
		{
			name: "Basic Round Trip",
			inputChecklist: `:: title: Simple List
:: version: 1.0

- [ ] Item 1
- [x] Item 2
`,
			// Expected output matches input because metadata keys are already sorted
			expectedOutput: `:: title: Simple List
:: version: 1.0

- [ ] Item 1
- [x] Item 2
`,
			expectError: false,
		},
		{
			name: "Nested and Auto Round Trip",
			inputChecklist: `:: type: Nested Example
:: author: Test

# Section 1
- [?] Manual Question
  - | | Auto Open Child
    - |x| Auto Done Grandchild
  - [-] Manual Skipped Child

# Section 2
- [>] Manual In Progress
- |-| Auto Partial
- [*] Manual Special Star
- |!| Auto Blocked
`,
			// Expected output has metadata sorted and blank line added after metadata
			expectedOutput: `:: author: Test
:: type: Nested Example

- [?] Manual Question
  - | | Auto Open Child
    - |x| Auto Done Grandchild
  - [-] Manual Skipped Child
- [>] Manual In Progress
- |-| Auto Partial
- [*] Manual Special Star
- |!| Auto Blocked
`,
			expectError: false,
		},
		{
			name: "Metadata Only",
			inputChecklist: `:: key1: value1
:: key2: value2
`,
			expectedOutput: `:: key1: value1
:: key2: value2
`, // No trailing newline if no items
			expectError: false,
		},
		{
			name: "Items Only",
			inputChecklist: `- [ ] Item A
- [x] Item B
`,
			expectedOutput: `- [ ] Item A
- [x] Item B
`,
			expectError: false,
		},
		{
			name:           "Empty Input String (Parser Error)", // Parser returns ErrNoContent
			inputChecklist: ``,
			// We don't call TreeToChecklistString here, expect error during setup
			expectError:     true,         // Expect error during the setup (ParseChecklist)
			expectedErrorIs: ErrNoContent, // Parser should return this
		},
		{
			name: "Minimal Tree (Root Only)", // Test formatting just a root node
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				root.Attributes["test"] = "value" // CORRECTED Syntax
				tree.RootID = root.ID
				return tree
			},
			expectedOutput: ":: test: value\n",
			expectError:    false,
		},
		{
			name:              "Error - Nil Tree",
			buildTreeManually: func() *core.GenericTree { return nil },
			expectError:       true,
			expectedErrorIs:   ErrInvalidChecklistTree,
		},
		{
			name: "Error - Tree No Root",
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				// No root assigned
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrInvalidChecklistTree,
		},
		{
			name: "Error - Wrong Root Type",
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				root := tree.NewNode("", "wrong_type") // Wrong type
				tree.RootID = root.ID
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrInvalidChecklistTree,
		},
		{
			name: "Error - Item Missing Status",
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Missing Status Item"
				// Missing item.Attributes["status"]
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrMissingStatusAttribute,
		},
		{
			name: "Error - Item Unknown Status",
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Unknown Status Item"
				item.Attributes["status"] = "invalid_status" // Set invalid status
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrUnknownStatus, // Should wrap ErrUnknownStatus
		},
		{
			name: "Error - Special Status Missing Symbol",
			buildTreeManually: func() *core.GenericTree {
				tree := core.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Special Missing Symbol"
				item.Attributes["status"] = "special" // Set special status
				// Missing item.Attributes["special_symbol"]
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrMissingSpecialSymbol, // Should wrap ErrMissingSpecialSymbol
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var tree *core.GenericTree
			var err error
			var setupErr error // Capture errors during setup (Parse/Adapt)

			if tc.buildTreeManually != nil {
				tree = tc.buildTreeManually()
			} else {
				// Round-trip test: Parse -> Adapt
				parsedData, parseErr := ParseChecklist(tc.inputChecklist, noopLogger)
				if parseErr != nil {
					if tc.expectError && errors.Is(parseErr, tc.expectedErrorIs) {
						t.Logf("Got expected setup error: %v", parseErr)
						return // Test passes
					}
					setupErr = fmt.Errorf("ParseChecklist failed during test setup: %w", parseErr)
				} else {
					tree, setupErr = ChecklistToTree(parsedData.Items, parsedData.Metadata)
					if setupErr != nil {
						if tc.expectError && errors.Is(setupErr, tc.expectedErrorIs) {
							t.Logf("Got expected setup error: %v", setupErr)
							return // Test passes
						}
						setupErr = fmt.Errorf("ChecklistToTree failed during test setup: %w", setupErr)
					}
				}
			}

			if setupErr != nil {
				// If setup failed, only proceed if we are specifically testing TreeToChecklistString errors
				// And we weren't expecting *this* setup error specifically
				if !tc.expectError || (tc.expectError && !errors.Is(setupErr, tc.expectedErrorIs) && tc.buildTreeManually == nil) {
					t.Fatalf("Unexpected error during test setup: %v", setupErr)
				} else if tc.expectError && !errors.Is(setupErr, tc.expectedErrorIs) && tc.buildTreeManually != nil {
					// Allow continuing if we are testing TreeToChecklistString error and setup error was different
					t.Logf("Ignoring setup error (%v) because manually built tree is used for testing TreeToChecklistString error (%v)", setupErr, tc.expectedErrorIs)
				} else {
					// Setup failed as expected for this test case (e.g., ParseChecklist error)
					return
				}
			}

			// --- Call the function under test ---
			output, err := TreeToChecklistString(tree)

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
				}
				if diff := cmp.Diff(tc.expectedOutput, output); diff != "" {
					t.Errorf("TreeToChecklistString() output mismatch (-want +got):\n%s", diff)
					t.Logf("WANT:\n%s\nGOT:\n%s", tc.expectedOutput, output)
				}
			}
		})
	}
}

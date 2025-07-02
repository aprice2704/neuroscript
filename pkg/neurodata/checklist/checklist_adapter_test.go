// NeuroScript Version: 0.3.1
// File version: 2.0.0
// Purpose: Updated test cases and helpers to use  TreeAttrs (map[string]interface{}) instead of map[string]string.
// filename: pkg/neurodata/checklist/checklist_adapter_test.go

package checklist

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	// For NoOpLogger
	"github.com/aprice2704/neuroscript/pkg/utils"
	"github.com/google/go-cmp/cmp" // For better diffs
)

// Helper to verify node properties in the tree
func verifyTreeNode(t *testing.T, tree *utils.GenericTree, nodeID string, expectedProps map[string]interface{}) {
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
			// FIX: The expected value is now  TreeAttrs (map[string]interface{})
			expectedAttrs, ok := expectedValue.(utils.TreeAttrs)
			if !ok {
				t.Errorf("verifyTreeNode: Invalid type for expected Attributes, want  TreeAttrs, got %T", expectedValue)
				continue
			}

			if (node.Attributes == nil || len(node.Attributes) == 0) && len(expectedAttrs) == 0 {
				// Okay
			} else if !reflect.DeepEqual(node.Attributes, expectedAttrs) {
				t.Errorf("verifyTreeNode: Node %q Attributes mismatch. Diff:\n%s", nodeID, cmp.Diff(expectedAttrs, node.Attributes))
			}
		case "ChildIDs":
			expectedChildren := expectedValue.([]string)
			actualChildren := node.ChildIDs
			if len(expectedChildren) == 0 && (actualChildren == nil || len(actualChildren) == 0) {
				// Okay
			} else if !reflect.DeepEqual(actualChildren, expectedChildren) {
				t.Errorf("verifyTreeNode: Node %q ChildIDs mismatch. Diff:\n%s", nodeID, cmp.Diff(expectedChildren, node.ChildIDs))
			}
		default:
			t.Errorf("verifyTreeNode: Unknown key %q in expectedProps for node %q", key, nodeID)
		}
	}
}

func TestChecklistToTree(t *testing.T) {
	t.Run("BasicListWithMetadata", func(t *testing.T) {
		items := []ChecklistItem{
			{Text: "Item 1", Status: "pending", Symbol: ' ', Indent: 0, LineNumber: 3, IsAutomatic: false},
			{Text: "Item 2", Status: "done", Symbol: 'x', Indent: 0, LineNumber: 4, IsAutomatic: false},
		}
		// FIX: Use map[string]interface{} for metadata to match  TreeAttrs
		metadata := map[string]interface{}{"title": "Test List", "version": "1.0"}

		// This test passes a map[string]string to a function expecting map[string]string, so ChecklistToTree is the boundary
		// We will test the output tree's attributes.
		stringMetadata := map[string]string{"title": "Test List", "version": "1.0"}
		tree, err := ChecklistToTree(items, stringMetadata)

		if err != nil {
			t.Fatalf("ChecklistToTree failed: %v", err)
		}
		if tree == nil || tree.RootID == "" || tree.NodeMap == nil {
			t.Fatalf("Tree is nil or not initialized properly")
		}
		if len(tree.NodeMap) != 3 { // root + 2 items
			t.Fatalf("Expected 3 nodes, got %d", len(tree.NodeMap))
		}

		rootID := tree.RootID
		if len(tree.NodeMap[rootID].ChildIDs) != 2 {
			t.Fatalf("Expected 2 children for root, got %d", len(tree.NodeMap[rootID].ChildIDs))
		}
		item1ID := tree.NodeMap[rootID].ChildIDs[0]
		item2ID := tree.NodeMap[rootID].ChildIDs[1]

		verifyTreeNode(t, tree, rootID, map[string]interface{}{
			"Type": "checklist_root", "ParentID": "", "Value": nil,
			"Attributes": utils.TreeAttrs(metadata), "ChildIDs": []string{item1ID, item2ID},
		})
		verifyTreeNode(t, tree, item1ID, map[string]interface{}{
			"Type": "checklist_item", "ParentID": rootID, "Value": "Item 1",
			"Attributes": utils.TreeAttrs{"status": "open"},
			"ChildIDs":   []string{},
		})
		verifyTreeNode(t, tree, item2ID, map[string]interface{}{
			"Type": "checklist_item", "ParentID": rootID, "Value": "Item 2",
			"Attributes": utils.TreeAttrs{"status": "done"},
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

		rootNode := tree.NodeMap[tree.RootID]
		p1ID := rootNode.ChildIDs[0]
		p2ID := rootNode.ChildIDs[1]
		p1Node := tree.NodeMap[p1ID]
		p2Node := tree.NodeMap[p2ID]
		c11ID := p1Node.ChildIDs[0]
		c12ID := p1Node.ChildIDs[1]
		c13ID := p1Node.ChildIDs[2]
		c21ID := p2Node.ChildIDs[0]

		verifyTreeNode(t, tree, tree.RootID, map[string]interface{}{"Type": "checklist_root", "Attributes": utils.TreeAttrs{"type": "Nested"}})
		verifyTreeNode(t, tree, p1ID, map[string]interface{}{"Type": "checklist_item", "Value": "Parent 1 *(Anno1)*", "Attributes": utils.TreeAttrs{"status": "open"}})
		verifyTreeNode(t, tree, c11ID, map[string]interface{}{"Type": "checklist_item", "Value": "Child 1.1", "Attributes": utils.TreeAttrs{"status": "skipped"}})
		verifyTreeNode(t, tree, c12ID, map[string]interface{}{"Type": "checklist_item", "Value": "Child 1.2 **(Anno2)**", "Attributes": utils.TreeAttrs{"status": "blocked"}})
		verifyTreeNode(t, tree, c13ID, map[string]interface{}{"Type": "checklist_item", "Value": "Child 1.3", "Attributes": utils.TreeAttrs{"status": "special", "special_symbol": "*"}})
		verifyTreeNode(t, tree, p2ID, map[string]interface{}{"Type": "checklist_item", "Value": "Parent 2", "Attributes": utils.TreeAttrs{"status": "open", "is_automatic": true}})
		verifyTreeNode(t, tree, c21ID, map[string]interface{}{"Type": "checklist_item", "Value": "Child 2.1", "Attributes": utils.TreeAttrs{"status": "partial", "is_automatic": true}})
	})

}

func TestTreeToChecklistString(t *testing.T) {
	noopLogger := logging.NewNoLogger()

	testCases := []struct {
		name              string
		inputChecklist    string
		buildTreeManually func() *utils.GenericTree
		expectedOutput    string
		expectError       bool
		expectedErrorIs   error
	}{
		{
			name: "Minimal Tree (Root Only)",
			buildTreeManually: func() *utils.GenericTree {
				tree := utils.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				// FIX: Attributes now map[string]interface{}
				root.Attributes = utils.TreeAttrs{"test": "value"}
				tree.RootID = root.ID
				return tree
			},
			expectedOutput: `:: test: value
`,
			expectError: false,
		},
		{
			name: "Error - Item Missing Status",
			buildTreeManually: func() *utils.GenericTree {
				tree := utils.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Missing Status Item"
				// FIX: Initialize attributes map
				item.Attributes = make(utils.TreeAttrs)
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrMissingStatusAttribute,
		},
		{
			name: "Error - Item Unknown Status",
			buildTreeManually: func() *utils.GenericTree {
				tree := utils.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Unknown Status Item"
				// FIX: Attributes now map[string]interface{}
				item.Attributes = utils.TreeAttrs{"status": "invalid_status"}
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrUnknownStatus,
		},
		{
			name: "Error - Special Status Missing Symbol",
			buildTreeManually: func() *utils.GenericTree {
				tree := utils.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "checklist_item")
				item.Value = "Special Missing Symbol"
				// FIX: Attributes now map[string]interface{}
				item.Attributes = utils.TreeAttrs{"status": "special"}
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrMissingSpecialSymbol,
		},
		{
			name: "Error - Item Wrong Type (Should Fail Formatting)",
			buildTreeManually: func() *utils.GenericTree {
				tree := utils.NewGenericTree()
				root := tree.NewNode("", "checklist_root")
				tree.RootID = root.ID
				item := tree.NewNode(root.ID, "not_a_checklist_item")
				item.Value = "Wrong Type Item"
				// FIX: Attributes now map[string]interface{}
				item.Attributes = utils.TreeAttrs{"status": "open"}
				root.ChildIDs = append(root.ChildIDs, item.ID)
				return tree
			},
			expectError:     true,
			expectedErrorIs: ErrInvalidChecklistTree,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test logic for round-trip and manual tree tests...
			// This part of the test runner logic doesn't need changes.
			// The changes are in the test case definitions above.
			var tree *utils.GenericTree
			var err error
			var setupErr error

			if tc.buildTreeManually != nil {
				tree = tc.buildTreeManually()
			} else {
				parsedData, parseErr := ParseChecklist(tc.inputChecklist, noopLogger)
				if parseErr != nil {
					if tc.expectError && errors.Is(parseErr, tc.expectedErrorIs) {
						t.Logf("Got expected setup error (ParseChecklist): %v", parseErr)
						return
					}
					setupErr = fmt.Errorf("ParseChecklist failed during test setup: %w", parseErr)
				} else if parsedData == nil {
					setupErr = fmt.Errorf("ParseChecklist returned nil data without error for input: %q", tc.inputChecklist)
				} else {
					tree, setupErr = ChecklistToTree(parsedData.Items, parsedData.Metadata)
					if setupErr != nil {
						if tc.expectError && errors.Is(setupErr, tc.expectedErrorIs) {
							t.Logf("Got expected setup error (ChecklistToTree): %v", setupErr)
							return
						}
						setupErr = fmt.Errorf("ChecklistToTree failed during test setup: %w", setupErr)
					}
				}
			}

			if setupErr != nil {
				if !tc.expectError {
					t.Fatalf("Unexpected error during test setup: %v", setupErr)
				}
				if tc.buildTreeManually == nil && !errors.Is(setupErr, tc.expectedErrorIs) {
					t.Fatalf("Setup failed with unexpected error: %v (expected: %v)", setupErr, tc.expectedErrorIs)
				}
			}

			output, err := TreeToChecklistString(tree)

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

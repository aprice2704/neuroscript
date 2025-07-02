// NeuroScript Version: 0.3.1
// File version: 2.0.0
// Purpose: Updated to align with  TreeAttrs (map[string]interface{}) and added safe type assertions.
// filename: pkg/neurodata/checklist/checklist_adapter.go

package checklist

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Ensure errors are correctly defined for this state
var (
	ErrInvalidChecklistTree   = errors.New("invalid generic tree structure for checklist formatting")
	ErrMissingStatusAttribute = errors.New("checklist item node missing 'status' attribute")
	ErrUnknownStatus          = errors.New("unknown status value in checklist item node")
	ErrMissingSpecialSymbol   = errors.New("checklist item node with status 'special' missing 'special_symbol' attribute")
)

// ChecklistToTree converts parsed checklist items and metadata into a GenericTree structure.
// Nodes representing checklist items will have Type="checklist_item".
func ChecklistToTree(items []ChecklistItem, metadata map[string]string) (*utils.GenericTree, error) {
	tree := utils.NewGenericTree()

	// 1. Create Root Node
	rootNode := tree.NewNode("", "checklist_root")
	tree.RootID = rootNode.ID
	rootNode.Value = nil
	if rootNode.Attributes == nil {
		// FIX: Align with the new  TreeAttrs type (map[string]interface{})
		rootNode.Attributes = make(utils.TreeAttrs)
	}
	for k, v := range metadata {
		rootNode.Attributes[k] = v
	}

	// 2. Process Items
	indentMap := map[int]string{
		-1: rootNode.ID,
	}

	for _, item := range items {
		currentItemIndent := item.Indent
		parentID := rootNode.ID
		bestParentIndent := -1
		for indentLevel, nodeID := range indentMap {
			if indentLevel < currentItemIndent && indentLevel >= bestParentIndent {
				bestParentIndent = indentLevel
				parentID = nodeID
			}
		}
		for indentLevel := range indentMap {
			if indentLevel >= currentItemIndent {
				delete(indentMap, indentLevel)
			}
		}

		newNode := tree.NewNode(parentID, "checklist_item")

		newNode.Value = item.Text // Store text in Value

		if newNode.Attributes == nil {
			// FIX: Align with the new  TreeAttrs type (map[string]interface{})
			newNode.Attributes = make(utils.TreeAttrs)
		}

		// Map status and handle special symbol
		statusStr := mapParserStatusToTreeStatus(item)
		newNode.Attributes["status"] = statusStr

		if item.IsAutomatic {
			// Storing a boolean directly is now possible and preferred.
			newNode.Attributes["is_automatic"] = true
		}

		if statusStr == "special" {
			isStandardSpecial := item.Symbol == '>' || item.Symbol == '!' || item.Symbol == '?'
			if !isStandardSpecial {
				newNode.Attributes["special_symbol"] = string(item.Symbol)
			}
		}

		parentNode := tree.NodeMap[parentID]
		if parentNode == nil {
			return nil, fmt.Errorf("internal error: parent node %q not found in map for item line %d", parentID, item.LineNumber)
		}
		if parentNode.ChildIDs == nil {
			parentNode.ChildIDs = make([]string, 0, 1)
		}
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNode.ID)

		indentMap[currentItemIndent] = newNode.ID
	}

	return tree, nil
}

// mapParserStatusToTreeStatus (Unchanged)
func mapParserStatusToTreeStatus(item ChecklistItem) string {
	switch item.Status {
	case "pending":
		return "open"
	case "done":
		return "done"
	case "partial":
		if item.IsAutomatic {
			return "partial"
		}
		return "skipped"
	case "special":
		switch item.Symbol {
		case '>':
			return "inprogress"
		case '!':
			return "blocked"
		case '?':
			return "question"
		default:
			return "special"
		}
	default:
		return "unknown"
	}
}

// --- Tree to Checklist String Formatting ---

// TreeToChecklistString converts a GenericTree (representing a checklist) back into Markdown format.
func TreeToChecklistString(tree *utils.GenericTree) (string, error) {
	if tree == nil || tree.RootID == "" || tree.NodeMap == nil {
		return "", fmt.Errorf("%w: input tree is nil or invalid", ErrInvalidChecklistTree)
	}

	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists || rootNode.Type != "checklist_root" {
		return "", fmt.Errorf("%w: root node %q not found or not type 'checklist_root'", ErrInvalidChecklistTree, tree.RootID)
	}

	var builder strings.Builder

	// 1. Format Metadata from Root Attributes
	if len(rootNode.Attributes) > 0 {
		keys := make([]string, 0, len(rootNode.Attributes))
		for k := range rootNode.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			builder.WriteString(":: ")
			builder.WriteString(k)
			builder.WriteString(": ")
			// FIX: Safely convert attribute value to string for display.
			builder.WriteString(fmt.Sprintf("%v", rootNode.Attributes[k]))
			builder.WriteString("\n")
		}
		if len(rootNode.ChildIDs) > 0 {
			builder.WriteString("\n")
		}
	}

	// 2. Format Items Recursively
	for _, childID := range rootNode.ChildIDs {
		err := formatChecklistNodeRecursive(&builder, tree, childID, 0)
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}

// formatChecklistNodeRecursive recursively formats checklist item nodes.
func formatChecklistNodeRecursive(builder *strings.Builder, tree *utils.GenericTree, nodeID string, depth int) error {
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return fmt.Errorf("%w: child node %q not found in tree map", ErrInvalidChecklistTree, nodeID)
	}

	if node.Type != "checklist_item" {
		return fmt.Errorf("%w: node %q has unexpected type %q during formatting", ErrInvalidChecklistTree, nodeID, node.Type)
	}

	if node.Attributes == nil {
		// FIX: Align with the new  TreeAttrs type (map[string]interface{})
		node.Attributes = make(utils.TreeAttrs)
	}

	// 1. Calculate Indentation
	indent := strings.Repeat("  ", depth)

	// 2. Determine Item Prefix from Attributes
	statusVal, ok := node.Attributes["status"]
	if !ok {
		return fmt.Errorf("%w: node %q", ErrMissingStatusAttribute, nodeID)
	}
	// FIX: Safely assert status to a string.
	statusStr, ok := statusVal.(string)
	if !ok {
		return fmt.Errorf("node %q: status attribute is not a string (type: %T)", nodeID, statusVal)
	}

	// FIX: Safely assert is_automatic to a bool.
	isAutomatic, _ := node.Attributes["is_automatic"].(bool)

	// FIX: Safely assert special_symbol to a string.
	specialSymbolVal, _ := node.Attributes["special_symbol"].(string)

	prefix, err := mapTreeStatusToMarkdown(statusStr, specialSymbolVal, isAutomatic)
	if err != nil {
		return fmt.Errorf("node %q: %w", nodeID, err)
	}

	// 3. Get Value (Item Text)
	itemText := ""
	if node.Value != nil {
		if text, ok := node.Value.(string); ok {
			itemText = text
		} else {
			itemText = fmt.Sprintf("%v", node.Value)
		}
	}

	// 4. Append Formatted Line
	builder.WriteString(indent)
	builder.WriteString(prefix)
	builder.WriteString(itemText)
	builder.WriteString("\n")

	// 5. Recurse for Children
	for _, childID := range node.ChildIDs {
		err := formatChecklistNodeRecursive(builder, tree, childID, depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

// mapTreeStatusToMarkdown maps tree attributes back to the Markdown checklist item prefix.
// (Unchanged)
func mapTreeStatusToMarkdown(status, specialSymbol string, isAutomatic bool) (string, error) {
	switch status {
	case "open":
		if isAutomatic {
			return "- | | ", nil
		}
		return "- [ ] ", nil
	case "done":
		if isAutomatic {
			return "- |x| ", nil
		}
		return "- [x] ", nil
	case "skipped":
		return "- [-] ", nil
	case "partial":
		if !isAutomatic {
			// Should not happen for manual items ideally
		}
		return "- |-| ", nil
	case "inprogress":
		if isAutomatic {
			return "- |>| ", nil
		}
		return "- [>] ", nil
	case "blocked":
		if isAutomatic {
			return "- |!| ", nil
		}
		return "- [!] ", nil
	case "question":
		if isAutomatic {
			return "- |?| ", nil
		}
		return "- [?] ", nil
	case "special":
		if specialSymbol == "" {
			return "", ErrMissingSpecialSymbol
		}
		if isAutomatic {
			return fmt.Sprintf("- |%s| ", specialSymbol), nil
		}
		return fmt.Sprintf("- [%s] ", specialSymbol), nil
	case "unknown":
		return "", fmt.Errorf("%w: cannot format unknown status", ErrUnknownStatus)
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownStatus, status)
	}
}

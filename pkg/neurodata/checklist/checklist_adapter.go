// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 19:47:00 PM PDT // Add TreeToChecklistString formatting
// filename: pkg/neurodata/checklist/checklist_adapter.go

package checklist

import (
	"errors" // Import errors package
	"fmt"
	"sort" // For sorting metadata keys
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core for GenericTree/Node
)

var (
	// ErrInvalidChecklistTree indicates the tree structure is not valid for checklist formatting.
	ErrInvalidChecklistTree = errors.New("invalid generic tree structure for checklist formatting")
	// ErrMissingStatusAttribute indicates a checklist item node lacks the required 'status' attribute.
	ErrMissingStatusAttribute = errors.New("checklist item node missing 'status' attribute")
	// ErrUnknownStatus indicates an unrecognized status value was found in a node attribute.
	ErrUnknownStatus = errors.New("unknown status value in checklist item node")
	// ErrMissingSpecialSymbol indicates a node with status 'special' lacks the 'special_symbol' attribute.
	ErrMissingSpecialSymbol = errors.New("checklist item node with status 'special' missing 'special_symbol' attribute")
)

// ChecklistToTree converts parsed checklist items and metadata into a GenericTree structure.
// (Existing function - no changes needed here)
func ChecklistToTree(items []ChecklistItem, metadata map[string]string) (*core.GenericTree, error) {
	// Use the constructor to initialize the tree
	tree := core.NewGenericTree()

	// 1. Create Root Node
	rootNode := tree.NewNode("", "checklist_root") // ParentID is empty for root
	tree.RootID = rootNode.ID
	rootNode.Value = nil // Root has no primary text value
	// Add metadata as attributes
	if rootNode.Attributes == nil {
		rootNode.Attributes = make(map[string]string)
	}
	for k, v := range metadata {
		rootNode.Attributes[k] = v
	}

	// 2. Process Items
	indentMap := map[int]string{
		-1: rootNode.ID, // Base level before any items
	}

	for _, item := range items {
		currentItemIndent := item.Indent

		// --- Determine Parent ---
		parentID := rootNode.ID // Default to root
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
		// --- End Determine Parent ---

		newNode := tree.NewNode(parentID, "checklist_item")

		// Set Value to the full item text from the parser
		newNode.Value = item.Text // Use the original text directly

		// Ensure attributes map exists
		if newNode.Attributes == nil {
			newNode.Attributes = make(map[string]string)
		}

		// Map status and handle special symbol
		statusStr := mapParserStatusToTreeStatus(item)
		newNode.Attributes["status"] = statusStr

		// Add is_automatic attribute if the item was parsed as automatic
		if item.IsAutomatic {
			newNode.Attributes["is_automatic"] = "true"
		}

		// Add special_symbol if status is "special" and symbol is not standard
		if statusStr == "special" {
			isStandardSpecial := item.Symbol == '>' || item.Symbol == '!' || item.Symbol == '?'
			if !isStandardSpecial {
				newNode.Attributes["special_symbol"] = string(item.Symbol) // Store the actual symbol
			}
		}

		// Link child to parent
		parentNode := tree.NodeMap[parentID]
		if parentNode == nil {
			return nil, fmt.Errorf("internal error: parent node %q not found in map for item line %d", parentID, item.LineNumber)
		}
		if parentNode.ChildIDs == nil {
			parentNode.ChildIDs = make([]string, 0, 1)
		}
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNode.ID)

		// Update mapping for current indent level
		indentMap[currentItemIndent] = newNode.ID
	}

	return tree, nil
}

// mapParserStatusToTreeStatus converts the parser's output status/symbol
// to the standardized status string defined for the GenericTree attributes.
// (Existing function - no changes needed here)
func mapParserStatusToTreeStatus(item ChecklistItem) string {
	switch item.Status {
	case "pending":
		return "open"
	case "done":
		return "done"
	case "partial":
		if item.IsAutomatic {
			return "partial" // Automatic items with '-' symbol are partial
		}
		return "skipped" // Manual items with '-' symbol are skipped
	case "special":
		// Map specific symbols for known special statuses
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

// --- NEW: Tree to Checklist String Formatting ---

// TreeToChecklistString converts a GenericTree (representing a checklist) back into Markdown format.
func TreeToChecklistString(tree *core.GenericTree) (string, error) {
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
		// Sort keys for consistent output
		keys := make([]string, 0, len(rootNode.Attributes))
		for k := range rootNode.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			builder.WriteString(":: ")
			builder.WriteString(k)
			builder.WriteString(": ")
			builder.WriteString(rootNode.Attributes[k])
			builder.WriteString("\n")
		}
		// Add a blank line after metadata if there are items
		if len(rootNode.ChildIDs) > 0 {
			builder.WriteString("\n")
		}
	}

	// 2. Format Items Recursively
	for _, childID := range rootNode.ChildIDs {
		err := formatChecklistNodeRecursive(&builder, tree, childID, 0)
		if err != nil {
			return "", err // Propagate the first error encountered
		}
	}

	return builder.String(), nil
}

// formatChecklistNodeRecursive recursively formats checklist item nodes.
func formatChecklistNodeRecursive(builder *strings.Builder, tree *core.GenericTree, nodeID string, depth int) error {
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return fmt.Errorf("%w: child node %q not found in tree map", ErrInvalidChecklistTree, nodeID)
	}
	if node.Type != "checklist_item" {
		// Silently ignore non-item nodes that might somehow be children? Or error? Let's error.
		return fmt.Errorf("%w: node %q has type %q, expected 'checklist_item'", ErrInvalidChecklistTree, nodeID, node.Type)
	}

	// 1. Calculate Indentation (2 spaces per depth level)
	indent := strings.Repeat("  ", depth)

	// 2. Determine Item Prefix from Attributes
	status, ok := node.Attributes["status"]
	if !ok {
		return fmt.Errorf("%w: node %q", ErrMissingStatusAttribute, nodeID)
	}
	isAutomatic := node.Attributes["is_automatic"] == "true" // Defaults to false if missing
	specialSymbol := node.Attributes["special_symbol"]       // Defaults to "" if missing

	prefix, err := mapTreeStatusToMarkdown(status, specialSymbol, isAutomatic)
	if err != nil {
		return fmt.Errorf("node %q: %w", nodeID, err) // Wrap error with node context
	}

	// 3. Get Value (Item Text)
	itemText := ""
	if node.Value != nil {
		if text, ok := node.Value.(string); ok {
			itemText = text
		} else {
			// Handle non-string value? For now, format it generically.
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
			return err // Propagate the first error encountered
		}
	}

	return nil
}

// mapTreeStatusToMarkdown maps tree attributes back to the Markdown checklist item prefix.
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
		// Assume 'skipped' always corresponds to manual '[-]'. If automatic, maybe log warning?
		if isAutomatic {
			// This state shouldn't ideally happen if ChecklistToTree logic is correct
			// But we format it predictably anyway.
			// logger.Warn("Formatting node with status 'skipped' but is_automatic=true as '- [-] '")
		}
		return "- [-] ", nil
	case "partial":
		// Assume 'partial' always corresponds to automatic '|-|'.
		if !isAutomatic {
			// This state shouldn't ideally happen. Format predictably.
			// logger.Warn("Formatting node with status 'partial' but is_automatic=false as '- |-| '")
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
		// How to format unknown? Maybe default to open? Or error? Let's error.
		return "", fmt.Errorf("%w: cannot format unknown status", ErrUnknownStatus)
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownStatus, status)
	}
}

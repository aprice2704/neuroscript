// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 14:58:13 PDT // Split from tools_tree.go
// filename: pkg/core/tree_types.go

package core

import (
	"strconv"
)

// --- Generic Tree Representation ---

const GenericTreeHandleType = "GenericTree"
const defaultIndent = "  " // Default indentation string (used by render)

// GenericTreeNode represents a node within our generic tree structure.
type GenericTreeNode struct {
	ID         string            `json:"id"`         // Unique ID within the tree instance (e.g., "node-1", "node-2")
	Type       string            `json:"type"`       // Node type (e.g., "object", "array", "string", "number", "boolean", "null")
	Value      interface{}       `json:"value"`      // Holds the value for simple types (string, number, bool, nil)
	Attributes map[string]string `json:"attributes"` // For object keys (maps string keys to child node IDs)
	ChildIDs   []string          `json:"children"`   // Ordered list of child node IDs (for arrays)
	ParentID   string            `json:"-"`          // ID of the parent node ("" for root), excluded from direct node JSON
	Tree       *GenericTree      `json:"-"`          // Back-pointer to the tree (for convenience, excluded)
}

// GenericTree holds the entire tree structure associated with a handle.
type GenericTree struct {
	RootID  string                      `json:"rootId"` // ID of the root node
	NodeMap map[string]*GenericTreeNode `json:"-"`      // Map from node ID to the node struct, excluded from direct tree JSON
	nextID  int                         // Internal counter for generating IDs
}

// newNode creates a new node and adds it to the tree's NodeMap.
func (gt *GenericTree) newNode(parentID string, nodeType string) *GenericTreeNode {
	nodeID := "node-" + strconv.Itoa(gt.nextID)
	gt.nextID++
	node := &GenericTreeNode{
		ID:         nodeID,
		Type:       nodeType,
		Attributes: make(map[string]string), // Initialize maps/slices
		ChildIDs:   make([]string, 0),
		ParentID:   parentID,
		Tree:       gt,
	}
	gt.NodeMap[nodeID] = node
	return node
}

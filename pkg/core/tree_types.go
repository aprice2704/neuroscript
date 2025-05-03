// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 18:12:31 PDT // Add NewGenericTree constructor
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
	Attributes map[string]string `json:"attributes"` // For object keys (maps string keys to child node IDs) or node metadata
	ChildIDs   []string          `json:"children"`   // Ordered list of child node IDs (for arrays or nested items)
	ParentID   string            `json:"-"`          // ID of the parent node ("" for root), excluded from direct node JSON
	Tree       *GenericTree      `json:"-"`          // Back-pointer to the tree (for convenience, excluded)
}

// GenericTree holds the entire tree structure associated with a handle.
type GenericTree struct {
	RootID  string                      `json:"rootId"` // ID of the root node
	NodeMap map[string]*GenericTreeNode `json:"-"`      // Map from node ID to the node struct, excluded from direct tree JSON
	nextID  int                         // Internal counter for generating IDs
}

// NewGenericTree creates and initializes a new GenericTree.
func NewGenericTree() *GenericTree {
	return &GenericTree{
		NodeMap: make(map[string]*GenericTreeNode),
		nextID:  1, // Start ID counter at 1
	}
}

// NewNode creates a new node, adds it to the tree's NodeMap, and returns it.
// It initializes Attributes and ChildIDs slices/maps.
// Renamed from newNode to be exported.
func (gt *GenericTree) NewNode(parentID string, nodeType string) *GenericTreeNode {
	// Simple sequential ID generation for now
	nodeID := "node-" + strconv.Itoa(gt.nextID)
	gt.nextID++ // Increment the internal counter

	node := &GenericTreeNode{
		ID:         nodeID,
		Type:       nodeType,
		Attributes: make(map[string]string), // Initialize map
		ChildIDs:   make([]string, 0),       // Initialize slice
		ParentID:   parentID,
		Tree:       gt, // Set back-pointer
		// Value is left as nil initially
	}

	// Ensure NodeMap is initialized (should be by NewGenericTree, but safe)
	if gt.NodeMap == nil {
		gt.NodeMap = make(map[string]*GenericTreeNode)
	}
	gt.NodeMap[nodeID] = node // Add the new node to the map

	return node
}

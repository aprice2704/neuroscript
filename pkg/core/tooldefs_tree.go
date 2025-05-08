// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Verified alignment with refactored tool implementations.
// nlines: 170 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tooldefs_tree.go

package core

// treeToolsToRegister defines the ToolImplementation structs for core tree manipulation tools.
// This variable is used by zz_core_tools_registrar.go to register the tools.
var treeToolsToRegister = []ToolImplementation{
	// --- Load/Save ---
	{
		Spec: ToolSpec{
			Name:        "Tree.LoadJSON",
			Description: "Loads a JSON string into a new tree structure and returns a tree handle.",
			Args: []ArgSpec{
				{Name: "json_string", Type: ArgTypeString, Required: true, Description: "The JSON data as a string."},
			},
			ReturnType: ArgTypeString, // Returns tree handle string
		},
		Func: toolTreeLoadJSON,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.ToJSON",
			Description: "Converts a tree structure (identified by tree handle) back into a JSON string. Output is pretty-printed.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
			},
			ReturnType: ArgTypeString, // Returns JSON string
		},
		Func: toolTreeFormatJSON, // from tools_tree_render.go
	},

	// --- Navigation & Access ---
	{
		Spec: ToolSpec{
			Name:        "Tree.GetNode",
			Description: "Retrieves detailed information about a specific node within a tree, returned as a map. The map includes 'id', 'type', 'value', 'attributes' (map), 'children' (slice of IDs), and 'parentId'.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "The unique ID of the node to retrieve."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolTreeGetNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.GetParent",
			Description: "Gets the node ID of the parent of a given node. Returns an empty string for the root node or if the node has no parent.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node whose parent is sought."},
			},
			ReturnType: ArgTypeString, // Returns parent node ID string
		},
		Func: toolTreeGetParent,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.GetChildren",
			Description: "Gets a list of node IDs of the children of a given node. For object nodes, children are determined by attribute values that are node IDs. For array nodes, children are from the ordered list. Other node types return an empty list.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the parent node."},
			},
			ReturnType: ArgTypeSliceString, // Returns list of child node ID strings
		},
		Func: toolTreeGetChildren,
	},
	// Note: Tree.GetValue, Tree.GetRoot, Tree.GetMetadata (single attribute) are considered covered by Tree.GetNode
	// and direct map access on its result, or would require new specialized functions.

	// --- Modification ---
	{
		Spec: ToolSpec{
			Name:        "Tree.SetValue",
			Description: "Sets the value of an existing leaf or simple-type node (e.g., string, number, boolean, null, checklist_item). Cannot set value on 'object' or 'array' type nodes using this tool.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the leaf or simple-type node to modify."},
				{Name: "value", Type: ArgTypeAny, Required: true, Description: "The new value for the node."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeModifyNode, // toolTreeModifyNode now expects args (handle, nodeID, value)
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.AddChildNode",
			Description: "Adds a new child node to an existing parent node. Returns the ID of the newly created child node.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "parent_node_id", Type: ArgTypeString, Required: true, Description: "ID of the node that will become the parent."},
				{Name: "new_node_id_suggestion", Type: ArgTypeString, Required: false, Description: "Optional suggested unique ID for the new node. If empty or nil, an ID will be auto-generated. Must be unique if provided."},
				{Name: "node_type", Type: ArgTypeString, Required: true, Description: "Type of the new child (e.g., 'object', 'array', 'string', 'number', 'boolean', 'null', 'checklist_item')."},
				{Name: "value", Type: ArgTypeAny, Required: false, Description: "Initial value if the node_type is a leaf or simple type. Ignored for 'object' and 'array' types."},
				{Name: "key_for_object_parent", Type: ArgTypeString, Required: false, Description: "If the parent is an 'object' node, this key is used to link the new child in the parent's attributes. Required for object parents."},
			},
			ReturnType: ArgTypeString, // toolTreeAddNode now returns the new node's ID string
		},
		Func: toolTreeAddNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveNode",
			Description: "Removes a node (specified by ID) and all its descendants from the tree. Cannot remove the root node.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to remove."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeRemoveNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.SetObjectAttribute",
			Description: "Sets or updates an attribute on an 'object' type node, mapping the attribute key to an existing child node's ID. This is for establishing parent-child relationships in object nodes.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to set."},
				{Name: "child_node_id", Type: ArgTypeString, Required: true, Description: "The ID of an *existing* node within the same tree to associate with the key."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeSetAttribute,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveObjectAttribute",
			Description: "Removes an attribute (a key mapping to a child node ID) from an 'object' type node. This unlinks the child but does not delete the child node itself.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to remove."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeRemoveAttribute,
	},

	// --- Metadata (Stored as string-string pairs in node.Attributes) ---
	{
		Spec: ToolSpec{
			Name:        "Tree.SetNodeMetadata",
			Description: "Sets a metadata attribute as a key-value string pair on any node. This is separate from object attributes that link to child nodes.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to set metadata on."},
				{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key of the metadata attribute (string)."},
				{Name: "metadata_value", Type: ArgTypeString, Required: true, Description: "The value of the metadata attribute (string)."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeSetNodeMetadata,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveNodeMetadata",
			Description: "Removes a metadata attribute (a key-value string pair) from a node.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to remove metadata from."},
				{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key of the metadata attribute to remove."},
			},
			ReturnType: ArgTypeNil,
		},
		Func: toolTreeRemoveNodeMetadata,
	},

	// --- Find & Render ---
	{
		Spec: ToolSpec{
			Name:        "Tree.FindNodes",
			Description: "Finds nodes within a tree (starting from a specified node) that match specific criteria. Returns a list of matching node IDs.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "start_node_id", Type: ArgTypeString, Required: true, Description: "ID of the node within the tree to start searching from. The search includes this node."},
				{Name: "query_map", Type: ArgTypeMap, Required: true, Description: "Map defining search criteria. Supported keys: 'type' (string), 'value' (any), 'metadata' (map of string:string)."},
				{Name: "max_depth", Type: ArgTypeInt, Required: false, Description: "Maximum depth to search relative to the start node (0 for start node only, -1 for unlimited). Default: -1."},
				{Name: "max_results", Type: ArgTypeInt, Required: false, Description: "Maximum number of matching node IDs to return (-1 for unlimited). Default: -1."},
			},
			ReturnType: ArgTypeSliceString, // Returns list of matching node ID strings
		},
		Func: toolTreeFindNodes,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RenderText",
			Description: "Renders a visual text representation of the entire tree structure identified by the given tree handle.",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure to render."},
			},
			ReturnType: ArgTypeString, // Returns the rendered tree as a string
		},
		Func: toolTreeRenderText,
	},
}

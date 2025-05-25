// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Populated Category, Example, ReturnHelp, ErrorConditions fields.
// nlines: 300 // Approximate
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
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "json_string", Type: ArgTypeString, Required: true, Description: "The JSON data as a string."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a string handle representing the loaded tree. This handle is used in subsequent tree operations.",
			Example:         `TOOL.Tree.LoadJSON(json_string: "{ \"name\": \"example\" }") // Returns a tree handle like "tree_uuid_1"`,
			ErrorConditions: "ErrInvalidArgType if json_string is not a string; ErrMissingArg if json_string is not provided; ErrTreeLoadFailed if JSON parsing fails (e.g. malformed JSON).",
		},
		Func: toolTreeLoadJSON,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.ToJSON",
			Description: "Converts a tree structure (identified by tree handle) back into a JSON string. Output is pretty-printed.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a pretty-printed JSON string representation of the tree.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"key\":\"value\"}"); TOOL.Tree.ToJSON(tree_handle: handle) // Returns \"{\\n  \\\"key\\\": \\\"value\\\"\\n}\"`,
			ErrorConditions: "ErrInvalidArgType if tree_handle is not a string; ErrMissingArg if tree_handle is not provided; ErrTreeNotFound if the handle is invalid or tree does not exist; ErrTreeSerializationFailed if converting tree to JSON fails.",
		},
		Func: toolTreeFormatJSON,
	},

	// --- Navigation & Access ---
	{
		Spec: ToolSpec{
			Name:        "Tree.GetNode",
			Description: "Retrieves detailed information about a specific node within a tree, returned as a map. The map includes 'id', 'type', 'value', 'attributes' (map), 'children' (slice of IDs), and 'parentId'.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "The unique ID of the node to retrieve."},
			},
			ReturnType:      ArgTypeMap,
			ReturnHelp:      "Returns a map containing details of the specified node. Structure: {'id': string, 'type': string, 'value': any, 'attributes': map[string]any, 'children': []string, 'parentId': string}.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{\"child\":\"value\"}}"); TOOL.Tree.GetNode(tree_handle: handle, node_id: "root")`,
			ErrorConditions: "ErrInvalidArgType for incorrect argument types; ErrMissingArg for missing arguments; ErrTreeNotFound if handle is invalid; ErrNodeNotFound if node_id does not exist in the tree.",
		},
		Func: toolTreeGetNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.GetParent",
			Description: "Gets the node ID of the parent of a given node. Returns an empty string for the root node or if the node has no parent.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node whose parent is sought."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the string ID of the parent node. Returns an empty string if the node is the root or has no parent.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{\"child\":\"value\"}}"); TOOL.Tree.GetParent(tree_handle: handle, node_id: "child_node_id_placeholder") // Assuming 'child' has a known ID`,
			ErrorConditions: "ErrInvalidArgType for incorrect argument types; ErrMissingArg for missing arguments; ErrTreeNotFound if handle is invalid; ErrNodeNotFound if node_id does not exist.",
		},
		Func: toolTreeGetParent,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.GetChildren",
			Description: "Gets a list of node IDs of the children of a given node. For object nodes, children are determined by attribute values that are node IDs. For array nodes, children are from the ordered list. Other node types return an empty list.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the parent node."},
			},
			ReturnType:      ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a child node ID. Returns an empty slice if the node has no children or is not a container type.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{\"child1\":\"val1\", \"child2\":\"val2\"}}"); TOOL.Tree.GetChildren(tree_handle: handle, node_id: "root_node_id_placeholder")`,
			ErrorConditions: "ErrInvalidArgType for incorrect argument types; ErrMissingArg for missing arguments; ErrTreeNotFound if handle is invalid; ErrNodeNotFound if node_id does not exist.",
		},
		Func: toolTreeGetChildren,
	},

	// --- Modification ---
	{
		Spec: ToolSpec{
			Name:        "Tree.SetValue",
			Description: "Sets the value of an existing leaf or simple-type node (e.g., string, number, boolean, null, checklist_item). Cannot set value on 'object' or 'array' type nodes using this tool.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the leaf or simple-type node to modify."},
				{Name: "value", Type: ArgTypeAny, Required: true, Description: "The new value for the node."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Modifies the node's value in place.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"key\":\"old_value\"}"); TOOL.Tree.SetValue(tree_handle: handle, node_id: "node_id_of_key", value: "new_value")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound; ErrNodeTypeMismatch if trying to set value on an 'object' or 'array' node.",
		},
		Func: toolTreeModifyNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.AddChildNode",
			Description: "Adds a new child node to an existing parent node. Returns the ID of the newly created child node.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "parent_node_id", Type: ArgTypeString, Required: true, Description: "ID of the node that will become the parent."},
				{Name: "new_node_id_suggestion", Type: ArgTypeString, Required: false, Description: "Optional suggested unique ID for the new node. If empty or nil, an ID will be auto-generated. Must be unique if provided."},
				{Name: "node_type", Type: ArgTypeString, Required: true, Description: "Type of the new child (e.g., 'object', 'array', 'string', 'number', 'boolean', 'null', 'checklist_item')."},
				{Name: "value", Type: ArgTypeAny, Required: false, Description: "Initial value if the node_type is a leaf or simple type. Ignored for 'object' and 'array' types."},
				{Name: "key_for_object_parent", Type: ArgTypeString, Required: false, Description: "If the parent is an 'object' node, this key is used to link the new child in the parent's attributes. Required for object parents."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the string ID of the newly created child node.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{}}"); TOOL.Tree.AddChildNode(tree_handle: handle, parent_node_id: "root_id", node_type: "string", value: "new child", key_for_object_parent: "childKey")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound (for parent); ErrNodeTypeMismatch if parent cannot have children of the specified type or if key_for_object_parent is missing for an object parent; ErrDuplicateNodeID if suggested ID already exists.",
		},
		Func: toolTreeAddNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveNode",
			Description: "Removes a node (specified by ID) and all its descendants from the tree. Cannot remove the root node.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to remove."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes the node and its descendants.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{\"child_to_remove\":\"value\"}}"); TOOL.Tree.RemoveNode(tree_handle: handle, node_id: "id_of_child_to_remove")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound; ErrCannotRemoveRoot if attempting to remove the root node.",
		},
		Func: toolTreeRemoveNode,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.SetObjectAttribute",
			Description: "Sets or updates an attribute on an 'object' type node, mapping the attribute key to an existing child node's ID. This is for establishing parent-child relationships in object nodes.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to set."},
				{Name: "child_node_id", Type: ArgTypeString, Required: true, Description: "The ID of an *existing* node within the same tree to associate with the key."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Sets an attribute on the object node, linking it to the child node.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"obj\":{}, \"child\":{}}"); TOOL.Tree.SetObjectAttribute(tree_handle: handle, object_node_id: "obj_id", attribute_key: "myChild", child_node_id: "child_id")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound (for object_node_id or child_node_id); ErrNodeTypeMismatch if object_node_id is not an 'object' type.",
		},
		Func: toolTreeSetAttribute,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveObjectAttribute",
			Description: "Removes an attribute (a key mapping to a child node ID) from an 'object' type node. This unlinks the child but does not delete the child node itself.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to remove."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes the attribute link from the object node.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"obj\":{\"myChild\":\"child_id_val\"}}"); TOOL.Tree.RemoveObjectAttribute(tree_handle: handle, object_node_id: "obj_id", attribute_key: "myChild")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound (for object_node_id); ErrNodeTypeMismatch if object_node_id is not an 'object' type; ErrAttributeNotFound if key does not exist.",
		},
		Func: toolTreeRemoveAttribute,
	},

	// --- Metadata (Stored as string-string pairs in node.Attributes) ---
	{
		Spec: ToolSpec{
			Name:        "Tree.SetNodeMetadata",
			Description: "Sets a metadata attribute as a key-value string pair on any node. This is separate from object attributes that link to child nodes.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to set metadata on."},
				{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key of the metadata attribute (string)."},
				{Name: "metadata_value", Type: ArgTypeString, Required: true, Description: "The value of the metadata attribute (string)."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Adds or updates a string key-value pair in the node's metadata attributes.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"node\":{}}"); TOOL.Tree.SetNodeMetadata(tree_handle: handle, node_id: "node_id_val", metadata_key: "version", metadata_value: "1.0")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound.",
		},
		Func: toolTreeSetNodeMetadata,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RemoveNodeMetadata",
			Description: "Removes a metadata attribute (a key-value string pair) from a node.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: ArgTypeString, Required: true, Description: "ID of the node to remove metadata from."},
				{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key of the metadata attribute to remove."},
			},
			ReturnType:      ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes a metadata key-value pair from the node's attributes.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"node\":{\"_metadata_version\":\"1.0\"}}"); TOOL.Tree.RemoveNodeMetadata(tree_handle: handle, node_id: "node_id_val", metadata_key: "version")`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound; ErrAttributeNotFound if metadata_key does not exist.",
		},
		Func: toolTreeRemoveNodeMetadata,
	},

	// --- Find & Render ---
	{
		Spec: ToolSpec{
			Name:        "Tree.FindNodes",
			Description: "Finds nodes within a tree (starting from a specified node) that match specific criteria. Returns a list of matching node IDs.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "start_node_id", Type: ArgTypeString, Required: true, Description: "ID of the node within the tree to start searching from. The search includes this node."},
				{Name: "query_map", Type: ArgTypeMap, Required: true, Description: "Map defining search criteria. Supported keys: 'type' (string), 'value' (any), 'metadata' (map of string:string)."},
				{Name: "max_depth", Type: ArgTypeInt, Required: false, Description: "Maximum depth to search relative to the start node (0 for start node only, -1 for unlimited). Default: -1."},
				{Name: "max_results", Type: ArgTypeInt, Required: false, Description: "Maximum number of matching node IDs to return (-1 for unlimited). Default: -1."},
			},
			ReturnType:      ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a node ID matching the query criteria.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"root\":{\"type\":\"folder\", \"child\":{\"type\":\"file\"}}}"); TOOL.Tree.FindNodes(tree_handle: handle, start_node_id: "root_id", query_map: {\"type\":\"file\"})`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound; ErrNodeNotFound (for start_node_id); ErrInvalidQuery if query_map is malformed.",
		},
		Func: toolTreeFindNodes,
	},
	{
		Spec: ToolSpec{
			Name:        "Tree.RenderText",
			Description: "Renders a visual text representation of the entire tree structure identified by the given tree handle.",
			Category:    "Tree Manipulation",
			Args: []ArgSpec{
				{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle to the tree structure to render."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a string containing a human-readable, indented text representation of the tree.",
			Example:         `handle = TOOL.Tree.LoadJSON(json_string: "{\"a\":{\"b\":\"c\"}}"); TOOL.Tree.RenderText(tree_handle: handle)`,
			ErrorConditions: "ErrInvalidArgType; ErrMissingArg; ErrTreeNotFound.",
		},
		Func: toolTreeRenderText,
	},
}

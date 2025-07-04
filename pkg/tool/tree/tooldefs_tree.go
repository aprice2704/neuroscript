// NeuroScript Version: 0.3.1
// File version: 0.0.6
// Purpose: Reviewed and refined Category, Example, ReturnHelp, and ErrorConditions for all tree tool specs.
// filename: pkg/tool/tree/tooldefs_tree.go
// nlines: 314
// risk_rating: MEDIUM

package tree

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// treeToolsToRegister defines the ToolImplementation structs for core tree manipulation tools.
// This variable is used by zz_core_tools_registrar.go to register the tools.
var treeToolsToRegister = []tool.ToolImplementation{
	// --- Load/Save ---
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.LoadJSON",
			Description: "Loads a JSON string into a new tree structure and returns a tree handle.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "json_string", Type: tool.ArgTypeString, Required: true, Description: "The JSON data as a string."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a string handle representing the loaded tree. This handle is used in subsequent tree operations.",
			Example:         `tool.Tree.LoadJSON("{\"name\": \"example\"}") // Returns a tree handle like "tree_handle_XYZ"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` for incorrect argument count. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `json_string` is not a string. Returns `ErrTreeJSONUnmarshal` (with `ErrorCodeSyntax`) if JSON parsing fails. Returns `ErrInternal` for failures in tree building or handle registration.",
		},
		Func: toolTreeLoadJSON,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.ToJSON",
			Description: "Converts a tree structure (identified by tree handle) back into a JSON string. Output is pretty-printed.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a pretty-printed JSON string representation of the tree.",
			Example:         `handle = tool.Tree.LoadJSON("{\"key\":\"value\"}"); tool.Tree.ToJSON(handle) // Returns a pretty-printed JSON string.`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrTreeJSONMarshal` (with `ErrorCodeInternal`) if marshalling to JSON fails. Returns `ErrInternal` for internal tree consistency issues (e.g., missing root node).",
		},
		Func: toolTreeFormatJSON,
	},

	// --- Navigation & Access ---
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.GetNode",
			Description: "Retrieves detailed information about a specific node within a tree, returned as a map. The map includes 'id', 'type', 'value', 'attributes' (map), 'children' (slice of IDs), 'parent_id', and 'parent_attribute_key'.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "The unique ID of the node to retrieve."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing details of the specified node. Structure: {'id': string, 'type': string, 'value': any, 'attributes': map[string]string, 'children': []string, 'parent_id': string, 'parent_attribute_key': string}. 'attributes' for non-object nodes will be their metadata. 'children' is primarily for array-like nodes.",
			Example:         `handle = tool.Tree.LoadJSON("{\"root\":{\"child\":\"value\"}}"); tool.Tree.GetNode(handle, "root_node_id") // Replace root_node_id with actual ID of the 'root' node`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist in the tree.",
		},
		Func: toolTreeGetNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.GetParent",
			Description: "Gets the node ID of the parent of a given node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node whose parent is sought."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string ID of the parent node. Returns nil if the node is the root or has no explicitly set parent (which can occur if the node was detached or is the root).",
			Example:         `handle = tool.Tree.LoadJSON("{\"root\":{\"childKey\": {}}}"); tool.Tree.GetParent(handle, "child_node_id") // Assuming child_node_id is the ID of the node under 'childKey'`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist.",
		},
		Func: toolTreeGetParent,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.GetChildren",
			Description: "Gets a list of node IDs of the children of a given 'array' type node. Other node types will result in an error.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the 'array' type parent node."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a child node ID from the specified 'array' node. Returns an empty slice if the array node has no children.",
			Example:         `handle = tool.Tree.LoadJSON("{\"myArray\":[{\"id\":\"child1\"}, {\"id\":\"child2\"}]}"); tool.Tree.GetChildren(handle, "id_of_myArray_node") // Returns ["child1", "child2"] if those are their actual IDs.`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrNodeWrongType` if the specified `node_id` is not an 'array' type node.",
		},
		Func: toolTreeGetChildren,
	},

	// --- Modification ---
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.SetValue",
			Description: "Sets the value of an existing leaf or simple-type node (e.g., string, number, boolean, null, checklist_item). Cannot set value on 'object' or 'array' type nodes using this tool.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the leaf or simple-type node to modify."},
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The new value for the node."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Modifies the node's value in place.",
			Example:         `handle = tool.Tree.LoadJSON("{\"keyNode\":\"old_value\"}"); tool.Tree.SetValue(handle, "id_of_keyNode", "new_value")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrCannotSetValueOnType` (with `ErrorCodeTreeConstraintViolation`) if attempting to set value on an 'object' or 'array' node.",
		},
		Func: toolTreeModifyNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.AddChildNode",
			Description: "Adds a new child node to an existing parent node. Returns the ID of the newly created child node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "parent_node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node that will become the parent."},
				{Name: "new_node_id_suggestion", Type: tool.ArgTypeString, Required: false, Description: "Optional suggested unique ID for the new node. If empty or nil, an ID will be auto-generated. Must be unique if provided."},
				{Name: "node_type", Type: tool.ArgTypeString, Required: true, Description: "Type of the new child (e.g., 'object', 'array', 'string', 'number', 'boolean', 'null', 'checklist_item')."},
				{Name: "value", Type: tool.ArgTypeAny, Required: false, Description: "Initial value if the node_type is a leaf or simple type. Ignored for 'object' and 'array' types."},
				{Name: "key_for_object_parent", Type: tool.ArgTypeString, Required: false, Description: "If the parent is an 'object' node, this key is used to link the new child in the parent's attributes. Required for object parents."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string ID of the newly created child node.",
			Example:         `handle = tool.Tree.LoadJSON("{\"root\":{}}"); tool.Tree.AddChildNode(handle, "actual_root_id", "newChildNodeID", "string", "new child value", "childAttributeKey")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/missing arguments (e.g., invalid `node_type`, missing `key_for_object_parent` when parent is 'object'). Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `parent_node_id` does not exist. Returns `ErrNodeWrongType` if parent node type cannot accept children in the specified manner. Returns `ErrNodeIDExists` (with `ErrorCodeTreeConstraintViolation`) if `new_node_id_suggestion` (if provided) already exists.",
		},
		Func: toolTreeAddNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.RemoveNode",
			Description: "Removes a node (specified by ID) and all its descendants from the tree. Cannot remove the root node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes the node and its descendants.",
			Example:         `handle = tool.Tree.LoadJSON("{\"root\":{\"childKey\": {}}}"); tool.Tree.RemoveNode(handle, "id_of_child_node_under_childKey")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrCannotRemoveRoot` (with `ErrorCodeTreeConstraintViolation`) if attempting to remove the root node. May return `ErrInternal` for inconsistent tree states (e.g., non-root node without a parent).",
		},
		Func: toolTreeRemoveNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.SetObjectAttribute",
			Description: "Sets or updates an attribute on an 'object' type node, mapping the attribute key to an existing child node's ID. This is for establishing parent-child relationships in object nodes.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: tool.ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: tool.ArgTypeString, Required: true, Description: "The key (name) of the attribute to set."},
				{Name: "child_node_id", Type: tool.ArgTypeString, Required: true, Description: "The ID of an *existing* node within the same tree to associate with the key."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Sets an attribute on the object node, linking it to the child node.",
			Example:         `handle = tool.Tree.LoadJSON("{\"objNode\":{}, \"childNode\":{}}"); tool.Tree.SetObjectAttribute(handle, "id_of_objNode", "myChildAttribute", "id_of_childNode")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `object_node_id` or `child_node_id` does not exist. Returns `ErrTreeNodeNotObject` (with `ErrorCodeNodeWrongType`) if `object_node_id` does not refer to an 'object' type node.",
		},
		Func: toolTreeSetAttribute,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.RemoveObjectAttribute",
			Description: "Removes an attribute (a key mapping to a child node ID) from an 'object' type node. This unlinks the child but does not delete the child node itself.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: tool.ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: tool.ArgTypeString, Required: true, Description: "The key (name) of the attribute to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes the attribute link from the object node.",
			Example:         `handle = tool.Tree.LoadJSON("{\"objNode\":{\"myChildAttribute\":\"some_child_id\"}}"); tool.Tree.RemoveObjectAttribute(handle, "id_of_objNode", "myChildAttribute")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `object_node_id` does not exist. Returns `ErrTreeNodeNotObject` (with `ErrorCodeNodeWrongType`) if `object_node_id` is not an 'object' type. Returns `ErrAttributeNotFound` if the `attribute_key` does not exist on the object node.",
		},
		Func: toolTreeRemoveAttribute,
	},

	// --- Metadata (Stored as string-string pairs in node.Attributes) ---
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.SetNodeMetadata",
			Description: "Sets a metadata attribute as a key-value string pair on any node. This is separate from object attributes that link to child nodes.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to set metadata on."},
				{Name: "metadata_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the metadata attribute (string)."},
				{Name: "metadata_value", Type: tool.ArgTypeString, Required: true, Description: "The value of the metadata attribute (string)."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Adds or updates a string key-value pair in the node's metadata attributes.",
			Example:         `handle = tool.Tree.LoadJSON("{\"myNode\":{}}"); tool.Tree.SetNodeMetadata(handle, "id_of_myNode", "version", "1.0")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist.",
		},
		Func: toolTreeSetNodeMetadata,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.RemoveNodeMetadata",
			Description: "Removes a metadata attribute (a key-value string pair) from a node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to remove metadata from."},
				{Name: "metadata_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the metadata attribute to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success. Removes a metadata key-value pair from the node's attributes.",
			Example:         `handle = tool.Tree.LoadJSON("{\"myNode\":{}}"); tool.Tree.SetNodeMetadata(handle, "id_of_myNode", "customData", "someValue"); tool.Tree.RemoveNodeMetadata(handle, "id_of_myNode", "customData")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrAttributeNotFound` if the `metadata_key` does not exist in the node's attributes.",
		},
		Func: toolTreeRemoveNodeMetadata,
	},

	// --- Find & Render ---
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.FindNodes",
			Description: "Finds nodes within a tree (starting from a specified node) that match specific criteria. Returns a list of matching node IDs.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "start_node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node within the tree to start searching from. The search includes this node."},
				{Name: "query_map", Type: tool.ArgTypeMap, Required: true, Description: "Map defining search criteria. Supported keys: 'id' (string), 'type' (string), 'value' (any), 'attributes' (map of string:string for child node ID checks), 'metadata' (map of string:string for direct string value metadata checks). Other keys are treated as direct metadata attribute checks."},
				{Name: "max_depth", Type: tool.ArgTypeInt, Required: false, Description: "Maximum depth to search relative to the start node (0 for start node only, -1 for unlimited). Default: -1."},
				{Name: "max_results", Type: tool.ArgTypeInt, Required: false, Description: "Maximum number of matching node IDs to return (-1 for unlimited). Default: -1."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a node ID matching the query criteria.",
			Example:         `handle = tool.Tree.LoadJSON("{\"root\":{\"type\":\"folder\", \"data\":{\"id\":\"child1\", \"type\":\"file\"}}}"); tool.Tree.FindNodes(handle, "id_of_root_node", {\"type\":\"file\"})`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/missing arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `start_node_id` does not exist. Returns `ErrTreeInvalidQuery` (with `ErrorCodeArgMismatch`) if `query_map` is malformed (e.g., incorrect value type for a query key). May return `ErrInternal` for other unexpected errors during the recursive search.",
		},
		Func: toolTreeFindNodes,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Tree.RenderText",
			Description: "Renders a visual text representation of the entire tree structure identified by the given tree handle.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure to render."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a string containing a human-readable, indented text representation of the tree.",
			Example:         `handle = tool.Tree.LoadJSON("{\"a\":{\"b\":\"c\"}}"); tool.Tree.RenderText(handle) // Returns a human-readable text tree`,
			ErrorConditions: "Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. May return `ErrInternal` for issues like a missing root node or other unexpected errors during the rendering process.",
		},
		Func: toolTreeRenderText,
	},
}

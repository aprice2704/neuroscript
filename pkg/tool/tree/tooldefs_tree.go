// NeuroScript Version: 0.6.5
// File version: 2
// Purpose: Added definitions for the missing tools (GetRoot, GetNodeByPath, GetNodeMetadata) to the registration list.
// filename: pkg/tool/tree/tooldefs_tree.go
// nlines: 350+
// risk_rating: MEDIUM

package tree

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "tree"

// treeToolsToRegister defines the ToolImplementation structs for core tree manipulation tools.
// This variable is used by zz_core_tools_registrar.go to register the tools.
var treeToolsToRegister = []tool.ToolImplementation{
	// --- Load/Save ---
	{
		Spec: tool.ToolSpec{
			Name:        "LoadJSON",
			Group:       group,
			Description: "Loads a JSON string into a new tree structure and returns a tree handle.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "json_string", Type: tool.ArgTypeString, Required: true, Description: "The JSON data as a string."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a string handle representing the loaded tree.",
			Example:         `tool.Tree.LoadJSON("{\"name\": \"example\"}")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeJSONUnmarshal`, `ErrInternal`.",
		},
		Func: toolTreeLoadJSON,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToJSON",
			Group:       group,
			Description: "Converts a tree structure back into a pretty-printed JSON string.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a pretty-printed JSON string representation of the tree.",
			Example:         `handle = tool.Tree.LoadJSON("{\"key\":\"value\"}"); tool.Tree.ToJSON(handle)`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrTreeJSONMarshal`, `ErrInternal`.",
		},
		Func: toolTreeFormatJSON,
	},

	// --- Navigation & Access ---
	{
		Spec: tool.ToolSpec{
			Name:        "GetRoot",
			Group:       group,
			Description: "Retrieves the root node of the tree as a map.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing details of the root node.",
			Example:         `handle = tool.Tree.LoadJSON("{}"); tool.Tree.GetRoot(handle)`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrInternal`.",
		},
		Func: toolTreeGetRoot,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetNode",
			Group:       group,
			Description: "Retrieves detailed information about a specific node within a tree, returned as a map.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "The unique ID of the node to retrieve."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing details of the specified node.",
			Example:         `tool.Tree.GetNode(handle, "node_id_123")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`.",
		},
		Func: toolTreeGetNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetNodeByPath",
			Group:       group,
			Description: "Retrieves a node from a tree using a dot-separated path expression.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Dot-separated path (e.g., 'key.0.name')."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing details of the node found at the specified path.",
			Example:         `tool.Tree.GetNodeByPath(handle, "data.users.1")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrKeyNotFound`, `ErrNodeWrongType`.",
		},
		Func: toolTreeGetNodeByPath,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetParent",
			Group:       group,
			Description: "Gets the parent of a given node as a map.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node whose parent is sought."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map of the parent node, or nil if the node is the root.",
			Example:         `tool.Tree.GetParent(handle, "child_node_id")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`.",
		},
		Func: toolTreeGetParent,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetChildren",
			Group:       group,
			Description: "Gets a list of node IDs of the children of a given 'array' type node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the 'array' type parent node."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of child node IDs.",
			Example:         `tool.Tree.GetChildren(handle, "array_node_id")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrNodeWrongType`.",
		},
		Func: toolTreeGetChildren,
	},

	// --- Modification ---
	{
		Spec: tool.ToolSpec{
			Name:        "SetValue",
			Group:       group,
			Description: "Sets the value of an existing leaf or simple-type node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the leaf or simple-type node to modify."},
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The new value for the node."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.SetValue(handle, "id_of_keyNode", "new_value")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrCannotSetValueOnType`.",
		},
		Func: toolTreeModifyNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "AddChildNode",
			Group:       group,
			Description: "Adds a new child node to an existing parent node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "parent_node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node that will become the parent."},
				{Name: "new_node_id_suggestion", Type: tool.ArgTypeString, Required: false, Description: "Optional suggested unique ID for the new node."},
				{Name: "node_type", Type: tool.ArgTypeString, Required: true, Description: "Type of the new child (e.g., 'object', 'array', 'string')."},
				{Name: "value", Type: tool.ArgTypeAny, Required: false, Description: "Initial value for simple types."},
				{Name: "key_for_object_parent", Type: tool.ArgTypeString, Required: false, Description: "Required if the parent is an 'object' node."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string ID of the newly created child node.",
			Example:         `tool.Tree.AddChildNode(handle, "root_id", "newChild", "string", "hello", "message")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrNodeWrongType`, `ErrNodeIDExists`.",
		},
		Func: toolTreeAddNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "RemoveNode",
			Group:       group,
			Description: "Removes a node and all its descendants from the tree.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.RemoveNode(handle, "node_to_delete_id")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrCannotRemoveRoot`, `ErrInternal`.",
		},
		Func: toolTreeRemoveNode,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SetObjectAttribute",
			Group:       group,
			Description: "Sets or updates an attribute on an 'object' type node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: tool.ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the attribute to set."},
				{Name: "child_node_id", Type: tool.ArgTypeString, Required: true, Description: "The ID of an existing node to associate with the key."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.SetObjectAttribute(handle, "obj_id", "myChild", "child_id")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrTreeNodeNotObject`.",
		},
		Func: toolTreeSetAttribute,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "RemoveObjectAttribute",
			Group:       group,
			Description: "Removes an attribute from an 'object' type node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle for the tree structure."},
				{Name: "object_node_id", Type: tool.ArgTypeString, Required: true, Description: "Unique ID of the 'object' type node to modify."},
				{Name: "attribute_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the attribute to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.RemoveObjectAttribute(handle, "obj_id", "myChild")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrTreeNodeNotObject`, `ErrAttributeNotFound`.",
		},
		Func: toolTreeRemoveAttribute,
	},

	// --- Metadata ---
	{
		Spec: tool.ToolSpec{
			Name:        "GetNodeMetadata",
			Group:       group,
			Description: "Retrieves the metadata attributes of a specific node as a map.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to get metadata from."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map of the node's metadata attributes.",
			Example:         `tool.Tree.GetNodeMetadata(handle, "node_id")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`.",
		},
		Func: toolTreeGetNodeMetadata,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SetNodeMetadata",
			Group:       group,
			Description: "Sets a metadata attribute as a key-value string pair on any node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to set metadata on."},
				{Name: "metadata_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the metadata attribute (string)."},
				{Name: "metadata_value", Type: tool.ArgTypeString, Required: true, Description: "The value of the metadata attribute (string)."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.SetNodeMetadata(handle, "my_node_id", "version", "1.0")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`.",
		},
		Func: toolTreeSetNodeMetadata,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "RemoveNodeMetadata",
			Group:       group,
			Description: "Removes a metadata attribute from a node.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to remove metadata from."},
				{Name: "metadata_key", Type: tool.ArgTypeString, Required: true, Description: "The key of the metadata attribute to remove."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `tool.Tree.RemoveNodeMetadata(handle, "my_node_id", "version")`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrAttributeNotFound`.",
		},
		Func: toolTreeRemoveNodeMetadata,
	},

	// --- Find & Render ---
	{
		Spec: tool.ToolSpec{
			Name:        "FindNodes",
			Group:       group,
			Description: "Finds nodes within a tree that match specific criteria.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure."},
				{Name: "start_node_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the node to start searching from."},
				{Name: "query_map", Type: tool.ArgTypeMap, Required: true, Description: "Map defining search criteria."},
				{Name: "max_depth", Type: tool.ArgTypeInt, Required: false, Description: "Maximum depth to search."},
				{Name: "max_results", Type: tool.ArgTypeInt, Required: false, Description: "Maximum number of results to return."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of node IDs matching the query.",
			Example:         `tool.Tree.FindNodes(handle, "start_node_id", {\"type\":\"file\"})`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrNodeNotFound`, `ErrTreeInvalidQuery`, `ErrInternal`.",
		},
		Func: toolTreeFindNodes,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "RenderText",
			Group:       group,
			Description: "Renders a visual text representation of the entire tree structure.",
			Category:    "Tree Manipulation",
			Args: []tool.ArgSpec{
				{Name: "tree_handle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the tree structure to render."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a human-readable, indented text representation of the tree.",
			Example:         `tool.Tree.RenderText(handle)`,
			ErrorConditions: "Returns `ErrArgumentMismatch`, `ErrInvalidArgument`, `ErrTreeNotFound`, `ErrInternal`.",
		},
		Func: toolTreeRenderText,
	},
}

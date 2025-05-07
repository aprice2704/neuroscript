// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for Tree tools.
// filename: pkg/core/tooldefs_tree.go

package core

// treeToolsToRegister contains ToolImplementation definitions for Tree tools.
// These are typically global ToolImplementation variables defined in their
// respective tools_tree_*.go files (e.g., toolTreeLoadJSONImpl in tools_tree_load.go).
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
var treeToolsToRegister = []ToolImplementation{
	// From tools_tree_load.go
	toolTreeLoadJSONImpl,

	// From tools_tree_nav.go (assuming Impl variables exist)
	toolTreeGetNodeImpl,
	toolTreeGetChildrenImpl,
	toolTreeGetParentImpl,

	// From tools_tree_find.go (assuming Impl variable exists)
	toolTreeFindNodesImpl,

	// From tools_tree_modify.go (assuming Impl variables exist)
	toolTreeModifyNodeImpl,
	toolTreeSetAttributeImpl,
	toolTreeRemoveAttributeImpl,
	toolTreeAddNodeImpl,
	toolTreeRemoveNodeImpl,

	// From tools_tree_metadata.go (assuming Impl variables exist)
	toolTreeSetNodeMetadataImpl,
	toolTreeRemoveNodeMetadataImpl,

	// From tools_tree_render.go (assuming Impl variables exist)
	toolTreeFormatJSONImpl,
	toolTreeRenderTextImpl,
}

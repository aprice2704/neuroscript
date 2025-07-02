// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 18:41:15 PM PDT // Fix GetNodeFromHandle call
// filename: pkg/neurodata/checklist/checklist_tool.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- Allowed Statuses ---
var allowedStatuses = map[string]bool{
	"open":       true,
	"done":       true,
	"skipped":    true,
	"partial":    true, // Note: Usually calculated, but allowed for direct setting
	"inprogress": true,
	"question":   true,
	"blocked":    true,
	"special":    true,
}

// --- Tool Implementation Struct Definitions ---
// (Tool specs remain unchanged)

var toolChecklistLoadTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistLoadTree",
		Description: "Parses a checklist string (in Markdown format with :: metadata) and loads it into a GenericTree handle. Returns the handle ID string.",
		Args:        []core.ArgSpec{{Name: "checklist_string", Type: core.ArgTypeString, Required: true, Description: "The checklist content as a string."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistLoadTree,
}

var toolChecklistFormatTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistFormatTree",
		Description: "Formats a checklist GenericTree handle back into its Markdown string representation.",
		Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistFormatTree,
}

var toolChecklistSetItemTextImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistSetItemText",
		Description: "Sets the text (Value) of a specific checklist item node using TreeModifyNode. Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "The unique ID of the checklist item node."},
			{Name: "new_text", Type: core.ArgTypeString, Required: true, Description: "The new text content for the item."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistSetItemText,
}

var toolChecklistAddItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistAddItem",
		Description: "Adds a new checklist item node as a child of the specified parent node ID using TreeAddNode and TreeSetNodeMetadata. Does NOT automatically update parent statuses; call Checklist.UpdateStatus explicitly afterwards. Returns the new node's ID string.", // Updated description
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID."},
			{Name: "parent_id", Type: core.ArgTypeString, Required: true, Description: "ID of the parent node (can be root or another item)."},
			{Name: "new_item_text", Type: core.ArgTypeString, Required: true, Description: "Text content for the new item."},
			{Name: "new_item_status", Type: core.ArgTypeString, Required: false, Description: "Initial status (default 'open'). Use allowed leaf statuses only."},
			{Name: "is_automatic", Type: core.ArgTypeBool, Required: false, Description: "Whether the new item is automatic (default false)."},
			{Name: "special_symbol", Type: core.ArgTypeString, Required: false, Description: "Required if status is 'special'."},
			{Name: "index", Type: core.ArgTypeInt, Required: false, Description: "Insertion index in parent's children (-1 or omitted to append)."},
		},
		ReturnType: core.ArgTypeString,
	},
	Func: toolChecklistAddItem, // Refactored function in checklist_tool_add.go
}

var toolChecklistRemoveItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistRemoveItem",
		Description: "Removes a checklist item node (and all its descendants) from the tree using TreeRemoveNode. Does NOT automatically update parent statuses; call Checklist.UpdateStatus explicitly afterwards. Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "The unique ID of the checklist item node to remove. Cannot be the root node."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistRemoveItem,
}

var toolChecklistUpdateStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "Checklist.UpdateStatus",
		Description: "Recursively updates the status of all automatic checklist items based on their children's current statuses.",
		Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."}},
		ReturnType:  core.ArgTypeNil,
	},
	Func: toolChecklistUpdateStatus,
}

// --- Registration ---

const ToolsetName = "Checklist"

func init() {
	toolsets.AddToolsetRegistration(ToolsetName, RegisterChecklistTools)
	fmt.Println("Checklist package init() running...") // Simple log to confirm init
}

// RegisterChecklistTools registers all tools in this package with the interpreter.
func RegisterChecklistTools(registry core.ToolRegistrar) error {
	fmt.Println("Checklist tools registered via RegisterChecklistTools.") // Simple log
	if registry == nil {
		return errors.New("registry cannot be nil for RegisterChecklistTools")
	}

	tools := []core.ToolImplementation{
		toolChecklistLoadTreeImpl,
		toolChecklistFormatTreeImpl,
		toolChecklistSetItemStatusImpl, // Uses refactored func below
		toolChecklistSetItemTextImpl,
		toolChecklistAddItemImpl,
		toolChecklistRemoveItemImpl,
		toolChecklistUpdateStatusImpl,
	}

	var registrationErrors []error
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register tool %q in toolset %q: %w", tool.Spec.Name, ToolsetName, err))
		}
	}
	return errors.Join(registrationErrors...)
}

// --- Tool Implementations ---

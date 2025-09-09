// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Defines the tool specifications for handling and extracting data from event objects.
// filename: pkg/tool/ns_event/tooldefs_event.go
// nlines: 170
// risk_rating: MEDIUM
package ns_event

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "ns_event"

var EventToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "GetEventShape",
			Group:       Group,
			Description: "Returns the canonical Shape-Lite definition for a standard ns_event object.",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeMap,
			Example:     `set shape = ns_event.GetEventShape()`,
		},
		Func:          toolGetEventShape,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Compose",
			Group:       Group,
			Description: "Creates a valid ns standard event from its constituent parts.",
			Args: []tool.ArgSpec{
				{Name: "kind", Type: tool.ArgTypeString, Description: "The event kind (e.g., 'start.ping').", Required: true},
				{Name: "payload", Type: tool.ArgTypeMap, Description: "The data payload of the event.", Required: true},
				{Name: "id", Type: tool.ArgTypeString, Description: "Optional event ID. If omitted, a new one is generated.", Required: false},
				{Name: "agent_id", Type: tool.ArgTypeString, Description: "Optional agent ID.", Required: false},
			},
			ReturnType: tool.ArgTypeMap,
			Example:    `ns_event.Compose("user.created", {"user_id": 123})`,
		},
		Func:          toolComposeEvent,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetPayload",
			Group:       Group,
			Description: "Extracts the core payload from a raw ns standard event, unwrapping the outer envelope.",
			Args: []tool.ArgSpec{
				{Name: "event_object", Type: tool.ArgTypeMap, Description: "The event object, typically from an 'on event' handler.", Required: true},
			},
			ReturnType: tool.ArgTypeMap,
			Example:    `ns_event.GetPayload(ev)`,
		},
		Func:          toolGetPayload,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetAllPayloads",
			Group:       Group,
			Description: "Extracts all coalesced payloads from a raw ns standard event into a list of maps.",
			Args: []tool.ArgSpec{
				{Name: "event_object", Type: tool.ArgTypeMap, Description: "The event object, typically from an 'on event' handler.", Required: true},
			},
			ReturnType: tool.ArgTypeSlice,
			Example:    `ns_event.GetAllPayloads(ev)`,
		},
		Func:          toolGetAllPayloads,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetID",
			Group:       Group,
			Description: "Extracts the event ID from the first envelope in an ns standard event.",
			Args: []tool.ArgSpec{
				{Name: "event_object", Type: tool.ArgTypeMap, Description: "The event object.", Required: true},
			},
			ReturnType: tool.ArgTypeString,
			Example:    `ns_event.GetID(ev)`,
		},
		Func:          toolGetID,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetKind",
			Group:       Group,
			Description: "Extracts the event Kind from the first envelope in an ns standard event.",
			Args: []tool.ArgSpec{
				{Name: "event_object", Type: tool.ArgTypeMap, Description: "The event object.", Required: true},
			},
			ReturnType: tool.ArgTypeString,
			Example:    `ns_event.GetKind(ev)`,
		},
		Func:          toolGetKind,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetTimestamp",
			Group:       Group,
			Description: "Extracts the event Timestamp (TS) from the first envelope in an ns standard event.",
			Args: []tool.ArgSpec{
				{Name: "event_object", Type: tool.ArgTypeMap, Description: "The event object.", Required: true},
			},
			ReturnType: tool.ArgTypeInt,
			Example:    `ns_event.GetTimestamp(ev)`,
		},
		Func:          toolGetTimestamp,
		RequiresTrust: false,
		RequiredCaps:  []capability.Capability{},
		Effects:       []string{"readonly"},
	},
}

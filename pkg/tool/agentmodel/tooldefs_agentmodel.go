// NeuroScript Version: 0.6.0
// File version: 2.0.0
// Purpose: Defines the tool specifications for managing agent models.
// filename: pkg/tool/agentmodel/tooldefs_agentmodel.go
// nlines: 120
// risk_rating: HIGH

package agentmodel

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "agentmodel"

var AgentModelToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Register",
			Group:       Group,
			Description: "Registers a new AgentModel configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true},
				{Name: "config", Type: tool.ArgTypeMap, Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
		Func: toolRegisterAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Update",
			Group:       Group,
			Description: "Updates an existing AgentModel's configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true},
				{Name: "updates", Type: tool.ArgTypeMap, Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
		Func: toolUpdateAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       Group,
			Description: "Deletes an AgentModel configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
		Func: toolDeleteAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       Group,
			Description: "Lists the names of all available AgentModels.",
			ReturnType:  tool.ArgTypeSliceString,
		},
		RequiresTrust: false,
		Func:          toolListAgentModels,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Select",
			Group:       Group,
			Description: "Selects the first available AgentModel. Criteria argument is a placeholder.",
			Args: []tool.ArgSpec{
				{Name: "criteria", Type: tool.ArgTypeMap, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		RequiresTrust: false,
		Func:          toolSelectAgentModel,
	},
}

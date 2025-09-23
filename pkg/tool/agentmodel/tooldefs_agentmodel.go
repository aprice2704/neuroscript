// NeuroScript Version: 0.7.3
// File version: 4
// Purpose: Added the 'Exists' tool definition.
// filename: pkg/tool/agentmodel/tooldefs_agentmodel.go
// nlines: 137
// risk_rating: HIGH
package agentmodel

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "agentmodel"

var AgentModelToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Register",
			Group:       Group,
			Description: "Registers a new agent model configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name for the agent model (e.g., 'gpt-4-turbo').", Required: true},
				{Name: "config", Type: tool.ArgTypeMap, Description: "A map containing model details like 'provider' and 'model'.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func:          toolRegisterAgentModel,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       Group,
			Description: "Deletes an agent model configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the model to delete.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func:          toolDeleteAgentModel,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       Group,
			Description: "Lists the names of all configured agent models.",
			ReturnType:  tool.ArgTypeSliceString,
		},
		Func:          toolListAgentModels,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Get",
			Group:       Group,
			Description: "Retrieves the full configuration of a registered agent model.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the model to retrieve.", Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          toolGetAgentModel,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Update",
			Group:       Group,
			Description: "Updates an existing agent model configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the agent model to update.", Required: true},
				{Name: "updates", Type: tool.ArgTypeMap, Description: "A map containing the fields to update.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func:          toolUpdateAgentModel,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Select",
			Group:       Group,
			Description: "Selects a model by name, or the default if no name is provided.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the model to select. If empty, selects the default.", Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolSelectAgentModel,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Exists",
			Group:       Group,
			Description: "Checks if an agent model with the given name is registered.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the model to check.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func:          toolAgentModelExists,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
	},
}

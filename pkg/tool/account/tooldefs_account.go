// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Defines the tool specifications for managing provider accounts, including a new 'Exists' tool.
// filename: pkg/tool/account/tooldefs_account.go
// nlines: 108
// risk_rating: HIGH
package account

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "account"

var AccountToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Register",
			Group:       Group,
			Description: "Registers a new provider account configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name for the account (e.g., 'openai-prod').", Required: true},
				{Name: "config", Type: tool.ArgTypeMap, Description: "A map containing account details like 'kind', 'provider', and 'apiKey'.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
			Example:    `account.Register("openai-prod", {"kind": "llm", "provider": "openai", "apiKey": "sk-..."})`,
		},
		Func:          toolRegisterAccount,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "account", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
		Effects: []string{"idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       Group,
			Description: "Deletes a provider account configuration.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the account to delete.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
			Example:    `account.Delete("openai-prod")`,
		},
		Func:          toolDeleteAccount,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "account", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
		Effects: []string{"idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       Group,
			Description: "Lists the names of all configured provider accounts.",
			ReturnType:  tool.ArgTypeSliceString,
			Example:     `account.List()`,
		},
		Func:          toolListAccounts,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "account", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
		Effects: []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Exists",
			Group:       Group,
			Description: "Checks if an account with the given name is registered.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the account to check.", Required: true},
			},
			ReturnType: tool.ArgTypeBool,
			Example:    `account.Exists("openai-prod")`,
		},
		Func:          toolAccountExists,
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "account", Verbs: []string{"read"}, Scopes: []string{"*"}},
		},
		Effects: []string{"readonly"},
	},
}

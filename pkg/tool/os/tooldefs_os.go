// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the specifications for OS-level tools like Getenv.
// filename: pkg/tool/os/tooldefs_os.go
// nlines: 39
// risk_rating: MEDIUM

package os

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "os"

var OsToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:            "Getenv",
			Group:           Group,
			Description:     "Gets the value of an environment variable. Requires 'env:read' capability.",
			Category:        "Operating System",
			Args:            []tool.ArgSpec{{Name: "varName", Type: tool.ArgTypeString, Required: true, Description: "The name of the environment variable."}},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the value of the environment variable as a string. Returns an empty string if the variable is not set.",
			Example:         `TOOL.OS.Getenv(varName: "HOME")`,
			ErrorConditions: "ErrArgumentMismatch if varName is empty or not a string. Returns an empty string for non-existent variables, which is not considered an error.",
		},
		Func:          toolGetenv, // from tools_os_env.go
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "env", Verbs: []string{"read"}},
		},
		Effects: []string{"readsEnv"},
	},
}

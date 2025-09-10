// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Defines specifications for OS process and time tools. Exported tool list and corrected Sleep return type.
// filename: pkg/tool/os/tooldefs_os_proc.go
// nlines: 48
// risk_rating: MEDIUM

package os

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

var OsProcToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Sleep",
			Group:       Group,
			Description: "Pauses execution for a specified duration. Requires 'os:exec:sleep' capability and is subject to policy time limits.",
			Category:    "Operating System",
			Args: []tool.ArgSpec{
				{Name: "duration_seconds", Type: tool.ArgTypeFloat, Required: true, Description: "The number of seconds to sleep."},
			},
			ReturnType:      tool.ArgTypeAny,
			ReturnHelp:      "Returns nil on completion.",
			Example:         `os.Sleep(duration_seconds: 1.5)`,
			ErrorConditions: "ErrArgumentMismatch if duration is not a number. ErrTimeExceeded if duration is longer than the policy limit.",
		},
		Func:          toolSleep,
		RequiresTrust: true,
		RequiredCaps:  []capability.Capability{capability.New(Group, capability.VerbExec, "sleep")},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Now",
			Group:       Group,
			Description: "Gets the current system time as a Unix timestamp.",
			Category:    "Operating System",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeFloat,
			ReturnHelp:  "Returns the number of seconds since the Unix epoch (1970-01-01T00:00:00Z UTC).",
			Example:     `os.Now()`,
		},
		Func: toolNow,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Hostname",
			Group:       Group,
			Description: "Gets the hostname of the machine.",
			Category:    "Operating System",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeString,
			ReturnHelp:  "Returns the kernel's hostname.",
			Example:     `os.Hostname()`,
		},
		Func: toolHostname,
	},
}

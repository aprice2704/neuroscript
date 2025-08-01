// filename: pkg/tool/time/tooldefs_time.go
// version: 6
// purpose: Corrected function signatures to match the ToolFunc type, resolving compiler errors. The adapters now correctly handle raw interface{} types.

package time

import "github.com/aprice2704/neuroscript/pkg/tool"

const group = "time"

// timeToolsToRegister contains the ToolImplementation definitions for Time tools.
var timeToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Now",
			Group:       group,
			Description: "Returns the current system time as a 'timedate' value.",
			Category:    "Time",
			Args:        []tool.ArgSpec{},
			ReturnType:  "timedate",
			ReturnHelp:  "A 'timedate' value representing the moment the tool was called.",
			Example:     "`set right_now = tool.Time.Now()`",
		},
		Func: adaptToolTimeNow,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Sleep",
			Group:       group,
			Description: "Pauses the script execution for a specified duration.",
			Category:    "Time",
			Args: []tool.ArgSpec{
				{Name: "duration_seconds", Type: "number", Description: "The number of seconds to sleep (can be a fraction)."},
			},
			ReturnType: "boolean",
			ReturnHelp: "Returns true on successful completion of the sleep duration.",
			Example:    "`call tool.Time.Sleep(1.5)`",
		},
		Func: adaptToolTimeSleep,
	},
}

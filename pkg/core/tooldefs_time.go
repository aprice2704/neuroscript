// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Defines the ToolImplementation slice for the core Time tool.
// filename: core/tooldefs_time.go
// nlines: 20
// risk_rating: LOW

package core

// timeToolsToRegister contains the ToolImplementation definitions for Time tools.
var timeToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Time.Now",
			Description: "Returns the current system time as a 'timedate' value.",
			Category:    "Time",
			Args:        []ArgSpec{},
			ReturnType:  "timedate",
			ReturnHelp:  "A 'timedate' value representing the moment the tool was called.",
			Example:     `set right_now = must tool.Time.Now()`,
		},
		Func: toolTimeNow,
	},
}

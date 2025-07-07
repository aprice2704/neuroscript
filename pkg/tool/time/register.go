// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Implements self-registration for the time toolset to break an import cycle.
// filename: pkg/tool/time/register.go
package time

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the time package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"time",
		tool.CreateRegistrationFunc("time", timeToolsToRegister),
	)
}

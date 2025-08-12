// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Registers the 'os' toolset with the NeuroScript runtime.
// filename: pkg/tool/os/register.go
// nlines: 15
// risk_rating: LOW

package os

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the os package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"os",
		tool.CreateRegistrationFunc("os", OsToolsToRegister),
	)
}

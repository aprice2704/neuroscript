// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the shell toolset.
// filename: pkg/tool/shell/register.go
package shell

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the shell package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"shell",
		tool.CreateRegistrationFunc("shell", shellToolsToRegister),
	)
}

// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the list toolset.
// filename: pkg/tool/list/register.go
package list

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the list package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"list",
		tool.CreateRegistrationFunc("list", listToolsToRegister),
	)
}

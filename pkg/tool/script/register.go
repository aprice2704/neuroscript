// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the script toolset.
// filename: pkg/tool/script/register.go
package script

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the script package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"script",
		tool.CreateRegistrationFunc("script", scriptToolsToRegister),
	)
}

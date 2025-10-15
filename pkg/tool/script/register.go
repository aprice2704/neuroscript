// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Updates self-registration to use the exported tool list.
// filename: pkg/tool/script/register.go
package script

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the script package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"script",
		tool.CreateRegistrationFunc("script", ToolsToRegister),
	)
}

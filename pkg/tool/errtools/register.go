// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the errtools toolset.
// filename: pkg/tool/errtools/register.go
package errtools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the errtools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"errtools",
		tool.CreateRegistrationFunc("errtools", errorToolsToRegister),
	)
}
